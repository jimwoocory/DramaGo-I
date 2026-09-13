package shotmanifest

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/mediago-dev/mediago-drama/services/server/internal/repository"
	servicecanon "github.com/mediago-dev/mediago-drama/services/server/internal/service/canon"
)

// CompiledReference is provider-agnostic reference metadata that maps directly
// to MediaGo generation reference bindings at the generation integration layer.
type CompiledReference struct {
	Kind         string `json:"kind"`
	DocumentID   string `json:"documentId,omitempty"`
	BlockID      string `json:"blockId,omitempty"`
	AssetID      string `json:"assetId,omitempty"`
	Role         string `json:"role,omitempty"`
	ResourceType string `json:"resourceType,omitempty"`
	Scope        string `json:"scope,omitempty"`
	Priority     int    `json:"priority,omitempty"`
	Locked       bool   `json:"locked,omitempty"`
}

// CompileResult is the deterministic generation payload for one Shot Manifest.
type CompileResult struct {
	DocumentID        string              `json:"documentId"`
	SectionID         string              `json:"sectionId"`
	Prompt            string              `json:"prompt"`
	ReferenceAssetIDs []string            `json:"referenceAssetIds"`
	References        []CompiledReference `json:"references"`
}

// CompileAndPersist compiles one persisted shot and stores the deterministic prompt.
func (service *Service) CompileAndPersist(projectID string, id string) (CompileResult, error) {
	if service == nil || service.repo == nil {
		return CompileResult{}, errors.New("shot manifest repository is not configured")
	}
	model, err := service.repo.Get(projectID, id)
	if err != nil {
		return CompileResult{}, err
	}
	record, err := recordFromModel(model)
	if err != nil {
		return CompileResult{}, err
	}
	compiled, err := service.Compile(record)
	if err != nil {
		return CompileResult{}, err
	}
	updated, err := service.repo.UpdateCompiledPrompt(projectID, id, compiled.Prompt)
	if err != nil {
		return CompileResult{}, err
	}
	if !updated {
		return CompileResult{}, repository.ErrRecordNotFound
	}
	return compiled, nil
}

// Compile builds a provider-ready prompt from Canon + Variant + resolved shot
// state without calling an LLM. Identical inputs always produce identical output.
func (service *Service) Compile(record Record) (CompileResult, error) {
	if service == nil || service.canon == nil {
		return CompileResult{}, errors.New("canon provider is not configured")
	}
	bindings, err := service.effectiveCompileBindings(record)
	if err != nil {
		return CompileResult{}, err
	}
	sections := make([]string, 0, 10)
	references := []CompiledReference{}

	for _, binding := range bindings.Characters {
		core, variant, err := service.compileBinding(record.ProjectID, binding.CanonID, binding.VariantID, servicecanon.ResourceTypeCharacter)
		if err != nil {
			return CompileResult{}, err
		}
		text := compileCanonText("角色", core, variant)
		if text != "" {
			sections = append(sections, text)
		}
		references = append(references, compileReferences(core, variant, binding.ReferenceAssetIDs)...)
	}
	if bindings.Scene != nil {
		binding := *bindings.Scene
		core, variant, err := service.compileBinding(record.ProjectID, binding.CanonID, binding.VariantID, servicecanon.ResourceTypeScene)
		if err != nil {
			return CompileResult{}, err
		}
		text := compileCanonText("场景", core, variant)
		if text != "" {
			sections = append(sections, text)
		}
		sceneReferences := compileReferences(core, variant, binding.ReferenceAssetIDs)
		references = append(references, suppressRedundantSceneCoreReferences(sceneReferences)...)
	}
	for _, binding := range bindings.Props {
		core, variant, err := service.compileBinding(record.ProjectID, binding.CanonID, binding.VariantID, servicecanon.ResourceTypeProp)
		if err != nil {
			return CompileResult{}, err
		}
		text := compileCanonText("道具", core, variant)
		if text != "" {
			sections = append(sections, text)
		}
		references = append(references, compileReferences(core, variant, binding.ReferenceAssetIDs)...)
	}
	if len(bindings.Props) > 0 {
		sections = append(sections, "道具一致性约束：已绑定道具的形状、材质、颜色和识别细节以对应 Prop Master 参考为视觉权威；若角色参考图中也出现同一随身道具，应视为同一个物理实体，不得因多张参考图重复生成；道具数量以动作文本为准，不得无依据增加副本或替换为近似材质/器型。")
	}
	continuityReference, ok, err := service.previousShotContinuityReference(record)
	if err != nil {
		return CompileResult{}, err
	}
	if ok {
		references = append(references, continuityReference)
	}

	if state := canonicalJSON(record.ResolvedStateJSON); state != "{}" {
		sections = append(sections, "连续性状态："+state)
	}
	if record.ActionText != "" {
		sections = append(sections, "动作："+record.ActionText)
	}
	if record.CameraText != "" {
		sections = append(sections, "镜头："+record.CameraText)
	}
	if record.AudioText != "" {
		sections = append(sections, "音频："+record.AudioText)
	}
	if record.StyleProfileID != "" {
		sections = append(sections, "风格配置："+record.StyleProfileID)
	}

	references = uniqueCompiledReferences(references)
	assetIDs := make([]string, 0, len(references))
	for _, reference := range references {
		if reference.AssetID != "" {
			assetIDs = append(assetIDs, reference.AssetID)
		}
	}
	assetIDs = orderedUniqueStrings(assetIDs)
	return CompileResult{
		DocumentID:        record.DocumentID,
		SectionID:         record.SectionID,
		Prompt:            strings.Join(sections, "\n"),
		ReferenceAssetIDs: assetIDs,
		References:        references,
	}, nil
}

type compileCanonListProvider interface {
	List(projectID string, resourceType string) ([]servicecanon.AssetRecord, error)
}

func (service *Service) effectiveCompileBindings(record Record) (Bindings, error) {
	bindings := normalizeBindings(record.Bindings)
	if strings.TrimSpace(record.ShotKey) != "" || strings.TrimSpace(record.Status) != StatusDraft || strings.TrimSpace(record.ActionText) == "" {
		return bindings, nil
	}
	catalog, ok := service.canon.(compileCanonListProvider)
	if !ok || catalog == nil {
		return bindings, nil
	}
	props, err := catalog.List(record.ProjectID, servicecanon.ResourceTypeProp)
	if err != nil {
		return Bindings{}, fmt.Errorf("listing Prop Canon for legacy compile enrichment: %w", err)
	}
	existing := map[string]bool{}
	for _, binding := range bindings.Props {
		if canonID := strings.TrimSpace(binding.CanonID); canonID != "" {
			existing[canonID] = true
		}
	}
	normalizedText := normalizeLegacySearchText(record.ActionText)
	for _, prop := range props {
		if prop.ParentID != "" || prop.ID == "" || existing[prop.ID] {
			continue
		}
		if _, matched := legacyUniquePropAliasPosition(normalizedText, prop, props); !matched {
			continue
		}
		bindings.Props = append(bindings.Props, ResourceBinding{CanonID: prop.ID})
		existing[prop.ID] = true
	}
	return normalizeBindings(bindings), nil
}

func (service *Service) previousShotContinuityReference(record Record) (CompiledReference, bool, error) {
	inheritsFromID := strings.TrimSpace(record.InheritsFromID)
	if inheritsFromID == "" || service == nil || service.repo == nil || service.continuityAssets == nil {
		return CompiledReference{}, false, nil
	}
	previousModel, err := service.repo.Get(record.ProjectID, inheritsFromID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return CompiledReference{}, false, nil
		}
		return CompiledReference{}, false, err
	}
	previous, err := recordFromModel(previousModel)
	if err != nil {
		return CompiledReference{}, false, err
	}
	assetID := ""
	ok := false
	if strings.TrimSpace(previous.ShotKey) != "" {
		assetID, ok, err = service.continuityAssets.LatestImageAssetIDForShotManifest(record.ProjectID, previous.ID)
	} else {
		assetID, ok, err = service.continuityAssets.LatestImageAssetIDForSection(record.ProjectID, previous.DocumentID, previous.SectionID)
	}
	if err != nil {
		return CompiledReference{}, false, err
	}
	assetID = strings.TrimSpace(assetID)
	if !ok || assetID == "" {
		return CompiledReference{}, false, nil
	}
	return CompiledReference{
		Kind:         "asset",
		DocumentID:   previous.DocumentID,
		BlockID:      previous.SectionID,
		AssetID:      assetID,
		Role:         "previous_shot",
		ResourceType: "continuity",
		Scope:        "continuity",
		Priority:     1000,
	}, true, nil
}

func (service *Service) compileBinding(projectID string, canonID string, variantID string, wantType string) (servicecanon.AssetRecord, *servicecanon.AssetRecord, error) {
	core, err := service.canon.Get(projectID, strings.TrimSpace(canonID))
	if err != nil {
		return servicecanon.AssetRecord{}, nil, err
	}
	if core.ParentID != "" || core.ResourceType != wantType {
		return servicecanon.AssetRecord{}, nil, fmt.Errorf("Canon %s is not a %s core", core.ID, wantType)
	}
	variantID = strings.TrimSpace(variantID)
	if variantID == "" {
		return core, nil, nil
	}
	variant, err := service.canon.Get(projectID, variantID)
	if err != nil {
		return servicecanon.AssetRecord{}, nil, err
	}
	if variant.ParentID != core.ID || variant.ResourceType != wantType {
		return servicecanon.AssetRecord{}, nil, fmt.Errorf("variant %s does not belong to Canon %s", variant.ID, core.ID)
	}
	return core, &variant, nil
}

func compileCanonText(label string, core servicecanon.AssetRecord, variant *servicecanon.AssetRecord) string {
	parts := []string{}
	if core.PromptText != "" {
		parts = append(parts, core.PromptText)
	}
	if variant != nil && variant.PromptText != "" {
		parts = append(parts, variant.PromptText)
	}
	if len(parts) == 0 {
		return ""
	}
	name := strings.TrimSpace(core.Name)
	if name == "" {
		name = core.ID
	}
	return label + "「" + name + "」：" + strings.Join(parts, "；")
}

func compileReferences(core servicecanon.AssetRecord, variant *servicecanon.AssetRecord, explicit []string) []CompiledReference {
	allowed := map[string]CompiledReference{}
	add := func(reference servicecanon.ReferenceRecord, scope string) {
		assetID := strings.TrimSpace(reference.AssetID)
		if assetID == "" {
			return
		}
		candidate := CompiledReference{
			Kind:         "section",
			DocumentID:   core.SourceDocumentID,
			BlockID:      core.ResourceID,
			AssetID:      assetID,
			Role:         strings.TrimSpace(reference.Role),
			ResourceType: core.ResourceType,
			Scope:        scope,
			Priority:     reference.Priority,
			Locked:       reference.Locked,
		}
		if existing, ok := allowed[assetID]; !ok || compiledReferenceLess(candidate, existing) {
			allowed[assetID] = candidate
		}
	}
	for _, reference := range core.References {
		add(reference, "core")
	}
	if variant != nil {
		for _, reference := range variant.References {
			add(reference, "variant")
		}
	}

	selected := map[string]bool{}
	for _, assetID := range orderedUniqueStrings(explicit) {
		selected[assetID] = true
	}
	result := make([]CompiledReference, 0, len(allowed))
	for assetID, reference := range allowed {
		if len(selected) > 0 && !selected[assetID] {
			continue
		}
		result = append(result, reference)
	}
	sort.SliceStable(result, func(i, j int) bool { return compiledReferenceLess(result[i], result[j]) })
	return result
}

func suppressRedundantSceneCoreReferences(values []CompiledReference) []CompiledReference {
	hasVariantAnchor := false
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value.ResourceType), servicecanon.ResourceTypeScene) &&
			strings.EqualFold(strings.TrimSpace(value.Scope), "variant") && strings.TrimSpace(value.AssetID) != "" {
			hasVariantAnchor = true
			break
		}
	}
	if !hasVariantAnchor {
		return values
	}
	result := make([]CompiledReference, 0, len(values))
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value.ResourceType), servicecanon.ResourceTypeScene) &&
			strings.EqualFold(strings.TrimSpace(value.Scope), "core") {
			continue
		}
		result = append(result, value)
	}
	return result
}

func uniqueCompiledReferences(values []CompiledReference) []CompiledReference {
	bestByAsset := map[string]CompiledReference{}
	for _, value := range values {
		value.AssetID = strings.TrimSpace(value.AssetID)
		if value.AssetID == "" {
			continue
		}
		if existing, ok := bestByAsset[value.AssetID]; !ok || compiledReferenceLess(value, existing) {
			bestByAsset[value.AssetID] = value
		}
	}
	result := make([]CompiledReference, 0, len(bestByAsset))
	for _, value := range bestByAsset {
		result = append(result, value)
	}
	sort.SliceStable(result, func(i, j int) bool { return compiledReferenceLess(result[i], result[j]) })
	return result
}

func compiledReferenceLess(left CompiledReference, right CompiledReference) bool {
	leftRank := compiledReferencePolicyRank(left)
	rightRank := compiledReferencePolicyRank(right)
	if leftRank != rightRank {
		return leftRank > rightRank
	}
	if left.Priority != right.Priority {
		return left.Priority > right.Priority
	}
	if left.Locked != right.Locked {
		return left.Locked
	}
	leftKey := strings.Join([]string{left.DocumentID, left.BlockID, left.Role, left.AssetID}, "\x00")
	rightKey := strings.Join([]string{right.DocumentID, right.BlockID, right.Role, right.AssetID}, "\x00")
	return leftKey < rightKey
}

func compiledReferencePolicyRank(reference CompiledReference) int {
	resourceType := strings.ToLower(strings.TrimSpace(reference.ResourceType))
	scope := strings.ToLower(strings.TrimSpace(reference.Scope))
	role := strings.ToLower(strings.TrimSpace(reference.Role))
	if role == "continuity" || role == "previous_shot" || role == "previous-shot" {
		return 550
	}
	switch resourceType {
	case servicecanon.ResourceTypeCharacter:
		if scope == "core" && (role == "identity" || role == "face" || role == "character_identity") {
			return 700
		}
		if scope == "variant" {
			return 600
		}
		return 650
	case servicecanon.ResourceTypeScene:
		if scope == "variant" {
			return 500
		}
		return 400
	case servicecanon.ResourceTypeProp:
		if scope == "variant" {
			return 300
		}
		return 250
	default:
		return 100
	}
}

func orderedUniqueStrings(values []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}

func canonicalJSON(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "{}"
	}
	var value any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return raw
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return raw
	}
	return string(encoded)
}
