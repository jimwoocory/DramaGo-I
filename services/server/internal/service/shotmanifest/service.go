package shotmanifest

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/mediago-dev/mediago-drama/services/server/internal/domain"
	"github.com/mediago-dev/mediago-drama/services/server/internal/repository"
	servicecanon "github.com/mediago-dev/mediago-drama/services/server/internal/service/canon"
	"github.com/mediago-dev/mediago-drama/services/server/internal/service/model"
)

const (
	StatusDraft    = "draft"
	StatusReady    = "ready"
	StatusConflict = "conflict"
)

var (
	legacyMentionPattern        = regexp.MustCompile(`mention://([^/)?\s]+)(?:/([^)?\s]+))?`)
	legacySceneCuePattern       = regexp.MustCompile(`「([^」]{1,40})」`)
	legacyStoryboardBeatPattern = regexp.MustCompile(`^\s*[-*]\s*([0-9]+(?:\.[0-9]+)?)\s*(?:-|~|至|—|–)\s*([0-9]+(?:\.[0-9]+)?)\s*(?:秒)?\s*[:：]\s*(.+?)\s*$`)
)

// CanonProvider exposes the authoritative Canon records needed by Shot binding validation.
type CanonProvider interface {
	Get(projectID string, id string) (servicecanon.AssetRecord, error)
}

type canonCatalogProvider interface {
	CanonProvider
	List(projectID string, resourceType string) ([]servicecanon.AssetRecord, error)
	SyncProject(projectID string) (servicecanon.ProjectSyncSummary, error)
}

type documentResourceProvider interface {
	ListWorkspaceDocumentResources(projectID string) (model.WorkspaceDocumentResourcesResponse, error)
}

type continuityAssetProvider interface {
	LatestImageAssetIDForSection(projectID string, documentID string, sectionID string) (string, bool, error)
	LatestImageAssetIDForShotManifest(projectID string, shotManifestID string) (string, bool, error)
}

// Service owns Shot Manifest persistence, continuity inheritance and Canon binding validation.
type Service struct {
	repo             *repository.ShotManifestRepository
	canon            CanonProvider
	documents        documentResourceProvider
	continuityAssets continuityAssetProvider
	initErr          error
}

// CharacterBinding pins one character identity and optional appearance variant to a shot.
type CharacterBinding struct {
	CanonID           string   `json:"canonId"`
	VariantID         string   `json:"variantId,omitempty"`
	ReferenceAssetIDs []string `json:"referenceAssetIds,omitempty"`
}

// ResourceBinding pins one scene/prop identity and optional variant to a shot.
type ResourceBinding struct {
	CanonID           string   `json:"canonId"`
	VariantID         string   `json:"variantId,omitempty"`
	ReferenceAssetIDs []string `json:"referenceAssetIds,omitempty"`
}

// Bindings is the normalized Canon binding set for a shot.
type Bindings struct {
	Characters []CharacterBinding `json:"characters,omitempty"`
	Scene      *ResourceBinding   `json:"scene,omitempty"`
	Props      []ResourceBinding  `json:"props,omitempty"`
}

// UpsertInput is the structured authoring payload for one shot/group.
type UpsertInput struct {
	ProjectID                   string
	DocumentID                  string
	SectionID                   string
	ShotKey                     string
	Sequence                    int
	StartSeconds                float64
	EndSeconds                  float64
	DurationSeconds             float64
	InheritsFromID              string
	ActionText                  string
	CameraText                  string
	AudioText                   string
	StyleProfileID              string
	Bindings                    Bindings
	StateChangesJSON            string
	SourceHash                  string
	Status                      string
	SkipImplicitPropInheritance bool
}

// Record is the API/service projection of a persisted Shot Manifest.
type Record struct {
	ID                string   `json:"id"`
	ProjectID         string   `json:"projectId"`
	DocumentID        string   `json:"documentId"`
	SectionID         string   `json:"sectionId"`
	ShotKey           string   `json:"shotKey"`
	Sequence          int      `json:"sequence"`
	StartSeconds      float64  `json:"startSeconds,omitempty"`
	EndSeconds        float64  `json:"endSeconds,omitempty"`
	DurationSeconds   float64  `json:"durationSeconds,omitempty"`
	InheritsFromID    string   `json:"inheritsFromId,omitempty"`
	ActionText        string   `json:"actionText"`
	CameraText        string   `json:"cameraText"`
	AudioText         string   `json:"audioText"`
	StyleProfileID    string   `json:"styleProfileId,omitempty"`
	Bindings          Bindings `json:"bindings"`
	StateChangesJSON  string   `json:"stateChangesJson"`
	ResolvedStateJSON string   `json:"resolvedStateJson"`
	CompiledPrompt    string   `json:"compiledPrompt,omitempty"`
	SourceHash        string   `json:"sourceHash,omitempty"`
	Version           int      `json:"version"`
	Status            string   `json:"status"`
	CreatedAt         string   `json:"createdAt"`
	UpdatedAt         string   `json:"updatedAt"`
}

// NewService returns a Shot Manifest service.
func NewService(repo *repository.ShotManifestRepository, canon CanonProvider, initErr error) *Service {
	service := &Service{repo: repo, canon: canon, initErr: initErr}
	if service.initErr == nil && service.repo == nil {
		service.initErr = errors.New("shot manifest repository is nil")
	}
	return service
}

// SetDocumentResourceProvider wires the existing workspace resource parser for lazy legacy migration.
func (service *Service) SetDocumentResourceProvider(provider documentResourceProvider) {
	service.documents = provider
}

// SetContinuityAssetProvider wires successful generation outputs as short-lived previous-shot references.
// These references are never persisted into Canon and therefore cannot replace long-term identity/scene truth.
func (service *Service) SetContinuityAssetProvider(provider continuityAssetProvider) {
	service.continuityAssets = provider
}

// ProjectSyncSummary reports lazy migration of existing storyboard resources.
type ProjectSyncSummary struct {
	Created int `json:"created"`
	Updated int `json:"updated"`
	Reused  int `json:"reused"`
	Skipped int `json:"skipped"`
}

type legacyStoryboardBeat struct {
	ShotKey         string
	StartSeconds    float64
	EndSeconds      float64
	DurationSeconds float64
	Content         string
	CameraText      string
	AudioText       string
	SourceHash      string
}

func parseLegacyStoryboardBeats(_ string, markdown string) []legacyStoryboardBeat {
	beats := []legacyStoryboardBeat{}
	for _, rawLine := range strings.Split(markdown, "\n") {
		match := legacyStoryboardBeatPattern.FindStringSubmatch(rawLine)
		if len(match) < 4 {
			continue
		}
		startSeconds, startErr := strconv.ParseFloat(strings.TrimSpace(match[1]), 64)
		endSeconds, endErr := strconv.ParseFloat(strings.TrimSpace(match[2]), 64)
		startSeconds = normalizeStoryboardSeconds(startSeconds)
		endSeconds = normalizeStoryboardSeconds(endSeconds)
		content := strings.TrimSpace(match[3])
		if startErr != nil || endErr != nil || endSeconds <= startSeconds || content == "" {
			continue
		}
		fields := parseLegacyStoryboardInlineFields(content)
		index := len(beats) + 1
		shotKey := fmt.Sprintf("beat-%03d", index)
		audioParts := []string{}
		for _, key := range []string{"音频", "台词"} {
			if value := strings.TrimSpace(fields[key]); value != "" {
				audioParts = append(audioParts, key+"："+value)
			}
		}
		cameraText := firstNonEmptyShotText(fields["画面"], fields["运镜"], fields["镜头"], fields["机位"])
		beats = append(beats, legacyStoryboardBeat{
			ShotKey:         shotKey,
			StartSeconds:    startSeconds,
			EndSeconds:      endSeconds,
			DurationSeconds: normalizeStoryboardSeconds(endSeconds - startSeconds),
			Content:         content,
			CameraText:      cameraText,
			AudioText:       strings.Join(audioParts, "；"),
			SourceHash:      productionShotSourceHash(shotKey, startSeconds, endSeconds, content),
		})
	}
	return beats
}

func parseLegacyStoryboardInlineFields(content string) map[string]string {
	fields := map[string]string{}
	for _, segment := range strings.FieldsFunc(content, func(r rune) bool { return r == '；' || r == ';' }) {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			continue
		}
		index := strings.Index(segment, "：")
		separatorWidth := len("：")
		if index < 0 {
			index = strings.Index(segment, ":")
			separatorWidth = 1
		}
		if index <= 0 || index+separatorWidth >= len(segment) {
			continue
		}
		key := strings.TrimSpace(segment[:index])
		value := strings.TrimSpace(segment[index+separatorWidth:])
		if key != "" && value != "" {
			fields[key] = value
		}
	}
	return fields
}

func normalizeStoryboardSeconds(value float64) float64 {
	return math.Round(value*1000) / 1000
}

func productionShotSourceHash(shotKey string, startSeconds float64, endSeconds float64, content string) string {
	return hashStrings(
		strings.TrimSpace(shotKey),
		strconv.FormatFloat(startSeconds, 'f', -1, 64),
		strconv.FormatFloat(endSeconds, 'f', -1, 64),
		strings.TrimSpace(content),
	)
}

func isUntouchedAutoProductionDraft(existing domain.ShotManifestModel, input UpsertInput) bool {
	if !strings.HasPrefix(strings.TrimSpace(existing.ShotKey), "beat-") || strings.TrimSpace(existing.Status) != StatusDraft {
		return false
	}
	if strings.TrimSpace(existing.CompiledPrompt) != "" || strings.TrimSpace(existing.StyleProfileID) != "" {
		return false
	}
	existingAutoHash := productionShotSourceHash(existing.ShotKey, existing.StartSeconds, existing.EndSeconds, existing.ActionText)
	if strings.TrimSpace(existing.SourceHash) == "" || strings.TrimSpace(existing.SourceHash) != existingAutoHash {
		return false
	}
	stateChanges, err := decodeStateObject(existing.StateChangesJSON)
	if err != nil || len(stateChanges) != 0 {
		return false
	}

	// When the source is unchanged, any difference in editable persisted fields is a manual
	// adjustment and must be preserved. When the source changed, the old record can still be
	// recognized as a pristine auto draft by its self-consistent source signature and safely
	// reconciled to the new parsed beat.
	if strings.TrimSpace(existing.SourceHash) == strings.TrimSpace(input.SourceHash) {
		if strings.TrimSpace(existing.ActionText) != strings.TrimSpace(input.ActionText) ||
			strings.TrimSpace(existing.CameraText) != strings.TrimSpace(input.CameraText) ||
			strings.TrimSpace(existing.AudioText) != strings.TrimSpace(input.AudioText) ||
			existing.StartSeconds != input.StartSeconds || existing.EndSeconds != input.EndSeconds ||
			existing.DurationSeconds != input.DurationSeconds {
			return false
		}
		existingBindings := strings.TrimSpace(existing.BindingsJSON)
		inputBindings, marshalErr := json.Marshal(normalizeBindings(input.Bindings))
		if marshalErr != nil || existingBindings != string(inputBindings) {
			return false
		}
		if strings.TrimSpace(domain.StringValue(existing.InheritsFromID)) != strings.TrimSpace(input.InheritsFromID) {
			return false
		}
	}
	return true
}

// Upsert resolves continuity and persists a deterministic structured shot.
func (service *Service) Upsert(input UpsertInput) (Record, error) {
	if service.initErr != nil {
		return Record{}, service.initErr
	}
	input = normalizeInput(input)
	if err := validateInput(input); err != nil {
		return Record{}, err
	}

	previous, hasPrevious, err := service.resolvePrevious(input)
	if err != nil {
		return Record{}, err
	}
	previousState := "{}"
	inheritsFromID := strings.TrimSpace(input.InheritsFromID)
	if hasPrevious {
		previousRecord, err := recordFromModel(previous)
		if err != nil {
			return Record{}, err
		}
		input.Bindings, err = inheritBindings(input.Bindings, previousRecord.Bindings, input.StateChangesJSON, !input.SkipImplicitPropInheritance)
		if err != nil {
			return Record{}, err
		}
		previousState = previous.ResolvedStateJSON
		inheritsFromID = previous.ID
	}
	bindingsJSON, err := json.Marshal(input.Bindings)
	if err != nil {
		return Record{}, fmt.Errorf("encoding shot bindings: %w", err)
	}
	if err := service.validateBindings(input.ProjectID, input.Bindings, input.StateChangesJSON); err != nil {
		return Record{}, err
	}
	resolvedState, err := ResolveState(previousState, input.StateChangesJSON)
	if err != nil {
		return Record{}, err
	}

	existing, err := service.repo.FindBySource(input.ProjectID, input.DocumentID, input.SectionID, input.ShotKey)
	if err != nil && !repository.IsRecordNotFound(err) {
		return Record{}, err
	}
	model := domain.ShotManifestModel{
		ID:                shotIDForSource(input.ProjectID, input.DocumentID, input.SectionID, input.ShotKey),
		ProjectID:         input.ProjectID,
		DocumentID:        input.DocumentID,
		SectionID:         input.SectionID,
		ShotKey:           input.ShotKey,
		Sequence:          input.Sequence,
		StartSeconds:      input.StartSeconds,
		EndSeconds:        input.EndSeconds,
		DurationSeconds:   input.DurationSeconds,
		ActionText:        input.ActionText,
		CameraText:        input.CameraText,
		AudioText:         input.AudioText,
		StyleProfileID:    input.StyleProfileID,
		BindingsJSON:      string(bindingsJSON),
		StateChangesJSON:  input.StateChangesJSON,
		ResolvedStateJSON: resolvedState,
		SourceHash:        input.SourceHash,
		Version:           1,
		Status:            input.Status,
	}
	if inheritsFromID != "" {
		model.InheritsFromID = domain.StringPtr(inheritsFromID)
	}
	if err == nil {
		model.ID = existing.ID
		model.CreatedAt = existing.CreatedAt
		model.Version = existing.Version
		if shotMaterialHash(existing) != shotMaterialHash(model) {
			model.Version++
		}
	}
	if err := service.repo.Upsert(model); err != nil {
		return Record{}, err
	}
	persisted, err := service.repo.FindBySource(input.ProjectID, input.DocumentID, input.SectionID, input.ShotKey)
	if err != nil {
		return Record{}, err
	}
	return recordFromModel(persisted)
}

// ListDocument returns all manifests in deterministic execution order.
func (service *Service) ListDocument(projectID string, documentID string) ([]Record, error) {
	if service.initErr != nil {
		return nil, service.initErr
	}
	projectID = domain.CleanProjectID(projectID)
	documentID = strings.TrimSpace(documentID)
	if projectID == "" || documentID == "" {
		return nil, errors.New("projectId and documentId are required")
	}
	models, err := service.repo.ListDocument(projectID, documentID)
	if err != nil {
		return nil, err
	}
	records := make([]Record, 0, len(models))
	for _, model := range models {
		record, err := recordFromModel(model)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

// SyncProject lazily materializes existing storyboard H2 resources into draft
// Shot Manifests. Existing manifests are never overwritten by migration.
func (service *Service) SyncProject(projectID string) (ProjectSyncSummary, error) {
	if service.initErr != nil {
		return ProjectSyncSummary{}, service.initErr
	}
	projectID = domain.CleanProjectID(projectID)
	if projectID == "" {
		return ProjectSyncSummary{}, errors.New("projectId is required")
	}
	if service.documents == nil {
		return ProjectSyncSummary{}, errors.New("shot manifest document resource provider is not configured")
	}
	catalog, ok := service.canon.(canonCatalogProvider)
	if !ok || catalog == nil {
		return ProjectSyncSummary{}, errors.New("shot manifest Canon catalog is not configured")
	}
	if _, err := catalog.SyncProject(projectID); err != nil {
		return ProjectSyncSummary{}, fmt.Errorf("syncing Canon before shot migration: %w", err)
	}
	resources, err := service.documents.ListWorkspaceDocumentResources(projectID)
	if err != nil {
		return ProjectSyncSummary{}, err
	}
	canonAssets, err := catalog.List(projectID, "")
	if err != nil {
		return ProjectSyncSummary{}, err
	}

	sequenceByDocument := map[string]int{}
	previousShotByDocument := map[string]string{}
	summary := ProjectSyncSummary{}
	for _, resource := range resources.Resources {
		if strings.TrimSpace(resource.Type) != "storyboard" {
			continue
		}
		documentID := strings.TrimSpace(resource.DocumentID)
		sectionID := strings.TrimSpace(resource.SectionID)
		if documentID == "" || sectionID == "" {
			summary.Skipped++
			continue
		}
		beats := parseLegacyStoryboardBeats(resource.Title, resource.Markdown)
		if len(beats) > 0 {
			for _, beat := range beats {
				sequenceByDocument[documentID]++
				input := UpsertInput{
					ProjectID:                   projectID,
					DocumentID:                  documentID,
					SectionID:                   sectionID,
					ShotKey:                     beat.ShotKey,
					Sequence:                    sequenceByDocument[documentID],
					StartSeconds:                beat.StartSeconds,
					EndSeconds:                  beat.EndSeconds,
					DurationSeconds:             beat.DurationSeconds,
					InheritsFromID:              previousShotByDocument[documentID],
					ActionText:                  beat.Content,
					CameraText:                  beat.CameraText,
					AudioText:                   beat.AudioText,
					Bindings:                    inferLegacyBindings(resource.Title, beat.Content, canonAssets),
					StateChangesJSON:            `{}`,
					SourceHash:                  beat.SourceHash,
					Status:                      StatusDraft,
					SkipImplicitPropInheritance: true,
				}

				existing, findErr := service.repo.FindBySource(projectID, documentID, sectionID, beat.ShotKey)
				if findErr == nil {
					if !isUntouchedAutoProductionDraft(existing, input) {
						summary.Reused++
						previousShotByDocument[documentID] = existing.ID
						continue
					}
					beforeHash := shotMaterialHash(existing)
					record, err := service.Upsert(input)
					if err != nil {
						return ProjectSyncSummary{}, err
					}
					persisted, err := service.repo.FindBySource(projectID, documentID, sectionID, beat.ShotKey)
					if err != nil {
						return ProjectSyncSummary{}, err
					}
					if shotMaterialHash(persisted) != beforeHash {
						summary.Updated++
					} else {
						summary.Reused++
					}
					previousShotByDocument[documentID] = record.ID
					continue
				}
				if !repository.IsRecordNotFound(findErr) {
					return ProjectSyncSummary{}, findErr
				}
				record, err := service.Upsert(input)
				if err != nil {
					return ProjectSyncSummary{}, err
				}
				summary.Created++
				previousShotByDocument[documentID] = record.ID
			}
			continue
		}

		sequenceByDocument[documentID]++
		actionText := firstNonEmptyShotText(resource.Prompt, resource.PlainText, resource.Markdown, resource.Title)
		sourceHash := hashStrings(resource.Markdown, resource.PlainText, resource.Prompt)
		bindings := inferLegacyBindings(resource.Title, resource.Markdown+"\n"+resource.PlainText+"\n"+resource.Prompt, canonAssets)
		input := UpsertInput{
			ProjectID:                   projectID,
			DocumentID:                  documentID,
			SectionID:                   sectionID,
			Sequence:                    sequenceByDocument[documentID],
			InheritsFromID:              previousShotByDocument[documentID],
			ActionText:                  actionText,
			Bindings:                    bindings,
			StateChangesJSON:            `{}`,
			SourceHash:                  sourceHash,
			Status:                      StatusDraft,
			SkipImplicitPropInheritance: true,
		}

		existing, findErr := service.repo.FindBySource(projectID, documentID, sectionID, "")
		if findErr == nil {
			if !isUntouchedLegacyDraft(existing, actionText, sourceHash) {
				summary.Reused++
				previousShotByDocument[documentID] = existing.ID
				continue
			}
			beforeHash := shotMaterialHash(existing)
			record, err := service.Upsert(input)
			if err != nil {
				return ProjectSyncSummary{}, err
			}
			persisted, err := service.repo.FindBySource(projectID, documentID, sectionID, "")
			if err != nil {
				return ProjectSyncSummary{}, err
			}
			if shotMaterialHash(persisted) != beforeHash {
				summary.Updated++
			} else {
				summary.Reused++
			}
			previousShotByDocument[documentID] = record.ID
			continue
		}
		if !repository.IsRecordNotFound(findErr) {
			return ProjectSyncSummary{}, findErr
		}
		record, err := service.Upsert(input)
		if err != nil {
			return ProjectSyncSummary{}, err
		}
		summary.Created++
		previousShotByDocument[documentID] = record.ID
	}
	return summary, nil
}

func isUntouchedLegacyDraft(existing domain.ShotManifestModel, actionText string, sourceHash string) bool {
	if strings.TrimSpace(existing.ShotKey) != "" || strings.TrimSpace(existing.Status) != StatusDraft {
		return false
	}
	if strings.TrimSpace(existing.CompiledPrompt) != "" || strings.TrimSpace(existing.CameraText) != "" || strings.TrimSpace(existing.AudioText) != "" || strings.TrimSpace(existing.StyleProfileID) != "" {
		return false
	}
	if strings.TrimSpace(existing.ActionText) != strings.TrimSpace(actionText) || strings.TrimSpace(existing.SourceHash) != strings.TrimSpace(sourceHash) {
		return false
	}
	stateChanges, err := decodeStateObject(existing.StateChangesJSON)
	if err != nil || len(stateChanges) != 0 {
		return false
	}
	return true
}

func (service *Service) resolvePrevious(input UpsertInput) (domain.ShotManifestModel, bool, error) {
	if input.InheritsFromID != "" {
		previous, err := service.repo.Get(input.ProjectID, input.InheritsFromID)
		if err != nil {
			return domain.ShotManifestModel{}, false, fmt.Errorf("resolving explicit inherited shot: %w", err)
		}
		if previous.DocumentID != input.DocumentID {
			return domain.ShotManifestModel{}, false, errors.New("inheritsFrom shot belongs to a different storyboard document")
		}
		if previous.Sequence >= input.Sequence {
			return domain.ShotManifestModel{}, false, errors.New("inheritsFrom shot must precede the current shot")
		}
		return previous, true, nil
	}
	manifests, err := service.repo.ListDocument(input.ProjectID, input.DocumentID)
	if err != nil {
		return domain.ShotManifestModel{}, false, err
	}
	var candidate *domain.ShotManifestModel
	for index := range manifests {
		manifest := manifests[index]
		if manifest.SectionID == input.SectionID && manifest.ShotKey == input.ShotKey {
			continue
		}
		if manifest.Sequence >= input.Sequence {
			continue
		}
		if candidate == nil || manifest.Sequence > candidate.Sequence {
			copy := manifest
			candidate = &copy
		}
	}
	if candidate == nil {
		return domain.ShotManifestModel{}, false, nil
	}
	return *candidate, true, nil
}

func (service *Service) validateBindings(projectID string, bindings Bindings, stateChangesJSON string) error {
	if service.canon == nil {
		if len(bindings.Characters) == 0 && bindings.Scene == nil && len(bindings.Props) == 0 {
			return nil
		}
		return errors.New("canon provider is not configured")
	}
	lockedCharacterIDs := map[string]bool{}
	for _, binding := range bindings.Characters {
		core, err := service.validateBinding(projectID, binding.CanonID, binding.VariantID, servicecanon.ResourceTypeCharacter)
		if err != nil {
			return err
		}
		if core.Status == servicecanon.StatusLocked {
			lockedCharacterIDs[core.ID] = true
		}
	}
	if bindings.Scene != nil {
		if _, err := service.validateBinding(projectID, bindings.Scene.CanonID, bindings.Scene.VariantID, servicecanon.ResourceTypeScene); err != nil {
			return err
		}
	}
	for _, binding := range bindings.Props {
		if _, err := service.validateBinding(projectID, binding.CanonID, binding.VariantID, servicecanon.ResourceTypeProp); err != nil {
			return err
		}
	}
	if len(lockedCharacterIDs) > 0 {
		if err := rejectLockedIdentityChanges(stateChangesJSON, lockedCharacterIDs); err != nil {
			return err
		}
	}
	return nil
}

func (service *Service) validateBinding(projectID string, canonID string, variantID string, wantType string) (servicecanon.AssetRecord, error) {
	canonID = strings.TrimSpace(canonID)
	variantID = strings.TrimSpace(variantID)
	if canonID == "" {
		return servicecanon.AssetRecord{}, fmt.Errorf("%s Canon binding is missing canonId", wantType)
	}
	core, err := service.canon.Get(projectID, canonID)
	if err != nil {
		return servicecanon.AssetRecord{}, fmt.Errorf("loading %s Canon %s: %w", wantType, canonID, err)
	}
	if core.ParentID != "" || core.ResourceType != wantType {
		return servicecanon.AssetRecord{}, fmt.Errorf("Canon %s is not a %s core", canonID, wantType)
	}
	if variantID != "" {
		variant, err := service.canon.Get(projectID, variantID)
		if err != nil {
			return servicecanon.AssetRecord{}, fmt.Errorf("loading %s variant %s: %w", wantType, variantID, err)
		}
		if variant.ParentID != core.ID || variant.ResourceType != wantType {
			return servicecanon.AssetRecord{}, fmt.Errorf("variant %s does not belong to Canon %s", variantID, core.ID)
		}
	}
	return core, nil
}

func rejectLockedIdentityChanges(raw string, lockedCharacterIDs map[string]bool) error {
	state, err := decodeStateObject(raw)
	if err != nil {
		return fmt.Errorf("decoding state changes for locked Canon validation: %w", err)
	}
	characters, _ := state["characters"].(map[string]any)
	for canonID := range lockedCharacterIDs {
		character, _ := characters[canonID].(map[string]any)
		for _, reserved := range []string{"identity", "core", "canon"} {
			if _, exists := character[reserved]; exists {
				return fmt.Errorf("locked character Canon %s cannot change %s state", canonID, reserved)
			}
		}
	}
	return nil
}

func inheritBindings(current Bindings, previous Bindings, stateChangesJSON string, inheritMissingProps bool) (Bindings, error) {
	stateChanges, err := decodeStateObject(stateChangesJSON)
	if err != nil {
		return Bindings{}, fmt.Errorf("decoding state changes for binding inheritance: %w", err)
	}
	clearedProps := map[string]bool{}
	if props, ok := stateChanges["props"].(map[string]any); ok {
		for canonID, value := range props {
			if value == nil {
				clearedProps[strings.TrimSpace(canonID)] = true
			}
		}
	}

	previousCharacters := map[string]CharacterBinding{}
	for _, binding := range previous.Characters {
		if binding.CanonID != "" {
			previousCharacters[binding.CanonID] = binding
		}
	}
	for index := range current.Characters {
		binding := &current.Characters[index]
		prior, ok := previousCharacters[binding.CanonID]
		if !ok {
			continue
		}
		if binding.VariantID == "" {
			binding.VariantID = prior.VariantID
		}
		if len(binding.ReferenceAssetIDs) == 0 && len(prior.ReferenceAssetIDs) > 0 {
			binding.ReferenceAssetIDs = append([]string(nil), prior.ReferenceAssetIDs...)
		}
	}

	if current.Scene == nil && previous.Scene != nil {
		copy := *previous.Scene
		copy.ReferenceAssetIDs = append([]string(nil), previous.Scene.ReferenceAssetIDs...)
		current.Scene = &copy
	} else if current.Scene != nil && previous.Scene != nil && current.Scene.CanonID == previous.Scene.CanonID {
		if current.Scene.VariantID == "" {
			current.Scene.VariantID = previous.Scene.VariantID
		}
		if len(current.Scene.ReferenceAssetIDs) == 0 && len(previous.Scene.ReferenceAssetIDs) > 0 {
			current.Scene.ReferenceAssetIDs = append([]string(nil), previous.Scene.ReferenceAssetIDs...)
		}
	}

	previousProps := map[string]ResourceBinding{}
	for _, binding := range previous.Props {
		if binding.CanonID != "" {
			previousProps[binding.CanonID] = binding
		}
	}
	currentPropIDs := map[string]bool{}
	for index := range current.Props {
		binding := &current.Props[index]
		currentPropIDs[binding.CanonID] = true
		prior, ok := previousProps[binding.CanonID]
		if !ok {
			continue
		}
		if binding.VariantID == "" {
			binding.VariantID = prior.VariantID
		}
		if len(binding.ReferenceAssetIDs) == 0 && len(prior.ReferenceAssetIDs) > 0 {
			binding.ReferenceAssetIDs = append([]string(nil), prior.ReferenceAssetIDs...)
		}
	}
	if inheritMissingProps {
		for _, prior := range previous.Props {
			canonID := strings.TrimSpace(prior.CanonID)
			if canonID == "" || currentPropIDs[canonID] || clearedProps[canonID] {
				continue
			}
			copy := prior
			copy.ReferenceAssetIDs = append([]string(nil), prior.ReferenceAssetIDs...)
			current.Props = append(current.Props, copy)
		}
	}
	return normalizeBindings(current), nil
}

func inferLegacyBindings(title string, text string, canonAssets []servicecanon.AssetRecord) Bindings {
	text = strings.TrimSpace(text)
	if text == "" || len(canonAssets) == 0 {
		return Bindings{}
	}

	coreBySource := map[string]servicecanon.AssetRecord{}
	variantsByParent := map[string][]servicecanon.AssetRecord{}
	positionByID := map[string]int{}
	hasExplicitSceneMention := false
	for _, asset := range canonAssets {
		if asset.ID == "" {
			continue
		}
		if asset.ParentID != "" {
			variantsByParent[asset.ParentID] = append(variantsByParent[asset.ParentID], asset)
			continue
		}
		coreBySource[asset.SourceDocumentID+"\x00"+asset.ResourceID] = asset
	}
	for _, mention := range legacyMentionPattern.FindAllStringSubmatch(text, -1) {
		if len(mention) < 3 || strings.TrimSpace(mention[2]) == "" {
			continue
		}
		documentID := decodeLegacyMentionPart(mention[1])
		sectionID := decodeLegacyMentionPart(mention[2])
		asset, ok := coreBySource[documentID+"\x00"+sectionID]
		if !ok {
			continue
		}
		position := strings.Index(text, mention[0])
		if current, exists := positionByID[asset.ID]; !exists || (position >= 0 && position < current) {
			positionByID[asset.ID] = position
		}
		if asset.ResourceType == servicecanon.ResourceTypeScene {
			hasExplicitSceneMention = true
		}
	}

	lowerText := strings.ToLower(text)
	for _, asset := range canonAssets {
		if asset.ParentID != "" || asset.ID == "" || strings.TrimSpace(asset.Name) == "" || asset.ResourceType == servicecanon.ResourceTypeScene {
			continue
		}
		if _, exists := positionByID[asset.ID]; exists {
			continue
		}
		name := strings.ToLower(strings.TrimSpace(asset.Name))
		if len([]rune(name)) < 2 {
			continue
		}
		if position := strings.Index(lowerText, name); position >= 0 {
			positionByID[asset.ID] = position
		}
	}

	// Legacy storyboard prose often mentions only the distinctive tail of a
	// qualified prop name (for example "木勺" for Canon "陷阵木勺").  Exact
	// name matching therefore misses key carried props even though the prose is
	// unambiguous.  Accept the longest contiguous alias only when that alias is
	// unique across project Prop cores; ambiguous aliases are deliberately left
	// unresolved instead of guessing.
	normalizedText := normalizeLegacySearchText(text)
	for _, asset := range canonAssets {
		if asset.ParentID != "" || asset.ID == "" || asset.ResourceType != servicecanon.ResourceTypeProp || strings.TrimSpace(asset.Name) == "" {
			continue
		}
		if _, exists := positionByID[asset.ID]; exists {
			continue
		}
		if position, ok := legacyUniquePropAliasPosition(normalizedText, asset, canonAssets); ok {
			positionByID[asset.ID] = position
		}
	}

	if !hasExplicitSceneMention {
		bestSceneID := ""
		bestSceneScore := 0
		for _, asset := range canonAssets {
			if asset.ParentID != "" || asset.ResourceType != servicecanon.ResourceTypeScene || asset.ID == "" {
				continue
			}
			score := legacySceneMatchScore(title, text, asset, canonAssets)
			if score > bestSceneScore || (score == bestSceneScore && score > 0 && (bestSceneID == "" || asset.ID < bestSceneID)) {
				bestSceneID = asset.ID
				bestSceneScore = score
			}
		}
		if bestSceneScore >= 40 && bestSceneID != "" {
			positionByID[bestSceneID] = 0
		}
	}

	type match struct {
		asset    servicecanon.AssetRecord
		position int
	}
	matches := make([]match, 0, len(positionByID))
	for _, asset := range canonAssets {
		position, ok := positionByID[asset.ID]
		if !ok || asset.ParentID != "" {
			continue
		}
		matches = append(matches, match{asset: asset, position: position})
	}
	sort.SliceStable(matches, func(first, second int) bool {
		if matches[first].position != matches[second].position {
			return matches[first].position < matches[second].position
		}
		return matches[first].asset.ID < matches[second].asset.ID
	})

	bindings := Bindings{}
	for _, matched := range matches {
		variantID := matchingVariantID(lowerText, variantsByParent[matched.asset.ID])
		binding := ResourceBinding{CanonID: matched.asset.ID, VariantID: variantID}
		switch matched.asset.ResourceType {
		case servicecanon.ResourceTypeCharacter:
			bindings.Characters = append(bindings.Characters, CharacterBinding{CanonID: matched.asset.ID, VariantID: variantID})
		case servicecanon.ResourceTypeScene:
			if bindings.Scene == nil {
				bindings.Scene = &binding
			}
		case servicecanon.ResourceTypeProp:
			bindings.Props = append(bindings.Props, binding)
		}
	}
	return normalizeBindings(bindings)
}

func legacyUniquePropAliasPosition(normalizedText string, asset servicecanon.AssetRecord, canonAssets []servicecanon.AssetRecord) (int, bool) {
	normalizedText = normalizeLegacySearchText(normalizedText)
	normalizedName := normalizeLegacySearchText(asset.Name)
	nameRunes := []rune(normalizedName)
	if len(nameRunes) < 2 || normalizedText == "" {
		return 0, false
	}

	for length := len(nameRunes); length >= 2; length-- {
		for start := 0; start+length <= len(nameRunes); start++ {
			alias := string(nameRunes[start : start+length])
			position := strings.Index(normalizedText, alias)
			if position < 0 {
				continue
			}
			matches := 0
			for _, candidate := range canonAssets {
				if candidate.ParentID != "" || candidate.ResourceType != servicecanon.ResourceTypeProp || candidate.ID == "" {
					continue
				}
				if strings.Contains(normalizeLegacySearchText(candidate.Name), alias) {
					matches++
				}
			}
			if matches == 1 {
				return position, true
			}
		}
	}
	return 0, false
}

func legacySceneMatchScore(title string, text string, asset servicecanon.AssetRecord, canonAssets []servicecanon.AssetRecord) int {
	titleSubject := legacyTitleSubject(title)
	normalizedText := normalizeLegacySearchText(text)
	normalizedName := normalizeLegacySearchText(asset.Name)
	score := 0

	for _, cue := range legacySceneStageCues(asset.PromptText) {
		normalizedCue := normalizeLegacySearchText(cue)
		if normalizedCue == "" {
			continue
		}
		if strings.Contains(titleSubject, normalizedCue) {
			score += 220
		} else if len([]rune(titleSubject)) >= 2 && strings.Contains(normalizedCue, titleSubject) {
			score += 180
		} else if overlap := longestCommonRuneSubstring(titleSubject, normalizedCue); overlap >= 3 {
			score += overlap * 45
		}
		for _, part := range splitLegacySceneParts(normalizedCue) {
			if len([]rune(part)) >= 2 && strings.Contains(titleSubject, part) {
				score += 130 + len([]rune(part))*5
			}
		}
	}

	if normalizedName != "" && strings.Contains(normalizedText, normalizedName) {
		score += 100 + len([]rune(normalizedName))*3
	}
	if overlap := longestCommonRuneSubstring(titleSubject, normalizedName); overlap >= 3 {
		score += overlap * 80
	}
	for _, part := range legacySceneNameParts(asset.Name, canonAssets) {
		if strings.Contains(normalizedText, part) {
			score += 50 + len([]rune(part))*5
		}
	}
	if overlap := longestCommonRuneSubstring(normalizedText, normalizedName); overlap >= 4 {
		score += overlap * 8
	}
	for _, item := range legacySceneCoreItems(asset.PromptText) {
		if strings.Contains(normalizedText, item) {
			length := len([]rune(item))
			if length > 8 {
				length = 8
			}
			score += 18 + length*2
		}
	}
	return score
}

func legacyTitleSubject(title string) string {
	normalized := normalizeLegacySearchText(title)
	if strings.HasPrefix(normalized, "第") {
		if index := strings.Index(normalized, "组"); index >= 0 {
			normalized = normalized[index+len("组"):]
		}
	}
	return normalized
}

func legacySceneStageCues(promptText string) []string {
	firstLine := promptText
	if index := strings.Index(firstLine, "\n"); index >= 0 {
		firstLine = firstLine[:index]
	}
	if !strings.Contains(firstLine, "对应场次") {
		return nil
	}
	matches := legacySceneCuePattern.FindAllStringSubmatch(firstLine, -1)
	result := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) >= 2 && strings.TrimSpace(match[1]) != "" {
			result = append(result, strings.TrimSpace(match[1]))
		}
	}
	return result
}

func legacySceneNameParts(name string, canonAssets []servicecanon.AssetRecord) []string {
	normalizedName := normalizeLegacySearchText(name)
	runes := []rune(normalizedName)
	longestPrefix := 0
	for length := 2; length < len(runes); length++ {
		prefix := string(runes[:length])
		matches := 0
		for _, candidate := range canonAssets {
			if candidate.ParentID != "" || candidate.ResourceType != servicecanon.ResourceTypeScene {
				continue
			}
			if strings.HasPrefix(normalizeLegacySearchText(candidate.Name), prefix) {
				matches++
			}
		}
		if matches >= 2 {
			longestPrefix = length
		}
	}
	candidate := normalizedName
	if longestPrefix > 0 && len(runes)-longestPrefix >= 3 {
		candidate = string(runes[longestPrefix:])
	}
	return splitLegacySceneParts(candidate)
}

func legacySceneCoreItems(promptText string) []string {
	const marker = "核心物件："
	index := strings.Index(promptText, marker)
	if index < 0 {
		return nil
	}
	line := promptText[index+len(marker):]
	if end := strings.Index(line, "\n"); end >= 0 {
		line = line[:end]
	}
	parts := strings.FieldsFunc(line, func(r rune) bool {
		switch r {
		case '、', '，', ',', '；', ';', '。', '与', '和', '及':
			return true
		default:
			return false
		}
	})
	result := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, part := range parts {
		part = normalizeLegacySearchText(part)
		if len([]rune(part)) < 2 || seen[part] {
			continue
		}
		seen[part] = true
		result = append(result, part)
	}
	return result
}

func splitLegacySceneParts(value string) []string {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		switch r {
		case '与', '和', '及':
			return true
		default:
			return false
		}
	})
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func normalizeLegacySearchText(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r) {
			return -1
		}
		return r
	}, value)
}

func longestCommonRuneSubstring(first string, second string) int {
	firstRunes := []rune(first)
	secondRunes := []rune(second)
	if len(firstRunes) == 0 || len(secondRunes) == 0 {
		return 0
	}
	if len(firstRunes) > len(secondRunes) {
		firstRunes, secondRunes = secondRunes, firstRunes
	}
	previous := make([]int, len(firstRunes)+1)
	best := 0
	for _, right := range secondRunes {
		current := make([]int, len(firstRunes)+1)
		for index, left := range firstRunes {
			if left != right {
				continue
			}
			current[index+1] = previous[index] + 1
			if current[index+1] > best {
				best = current[index+1]
			}
		}
		previous = current
	}
	return best
}

func matchingVariantID(lowerText string, variants []servicecanon.AssetRecord) string {
	bestID := ""
	bestLength := 0
	for _, variant := range variants {
		name := strings.ToLower(strings.TrimSpace(variant.Name))
		if name == "" || !strings.Contains(lowerText, name) {
			continue
		}
		length := len([]rune(name))
		if length > bestLength || (length == bestLength && variant.ID < bestID) {
			bestID = variant.ID
			bestLength = length
		}
	}
	return bestID
}

func decodeLegacyMentionPart(value string) string {
	value = strings.TrimSpace(value)
	decoded, err := url.PathUnescape(value)
	if err != nil {
		return value
	}
	return strings.TrimSpace(decoded)
}

func firstNonEmptyShotText(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

func normalizeInput(input UpsertInput) UpsertInput {
	input.ProjectID = domain.CleanProjectID(input.ProjectID)
	input.DocumentID = strings.TrimSpace(input.DocumentID)
	input.SectionID = strings.TrimSpace(input.SectionID)
	input.ShotKey = strings.TrimSpace(input.ShotKey)
	input.InheritsFromID = strings.TrimSpace(input.InheritsFromID)
	input.ActionText = strings.TrimSpace(input.ActionText)
	input.CameraText = strings.TrimSpace(input.CameraText)
	input.AudioText = strings.TrimSpace(input.AudioText)
	input.StyleProfileID = strings.TrimSpace(input.StyleProfileID)
	input.StateChangesJSON = strings.TrimSpace(input.StateChangesJSON)
	input.SourceHash = strings.TrimSpace(input.SourceHash)
	input.Status = strings.ToLower(strings.TrimSpace(input.Status))
	if input.StateChangesJSON == "" {
		input.StateChangesJSON = "{}"
	}
	if input.Status == "" {
		input.Status = StatusDraft
	}
	input.Bindings = normalizeBindings(input.Bindings)
	if input.SourceHash == "" {
		encoded, _ := json.Marshal(input.Bindings)
		input.SourceHash = hashStrings(input.ActionText, input.CameraText, input.AudioText, input.StyleProfileID, string(encoded), input.StateChangesJSON)
	}
	return input
}

func normalizeBindings(bindings Bindings) Bindings {
	for index := range bindings.Characters {
		bindings.Characters[index].CanonID = strings.TrimSpace(bindings.Characters[index].CanonID)
		bindings.Characters[index].VariantID = strings.TrimSpace(bindings.Characters[index].VariantID)
		bindings.Characters[index].ReferenceAssetIDs = compactStrings(bindings.Characters[index].ReferenceAssetIDs)
	}
	sort.SliceStable(bindings.Characters, func(i, j int) bool {
		return bindings.Characters[i].CanonID < bindings.Characters[j].CanonID
	})
	if bindings.Scene != nil {
		bindings.Scene.CanonID = strings.TrimSpace(bindings.Scene.CanonID)
		bindings.Scene.VariantID = strings.TrimSpace(bindings.Scene.VariantID)
		bindings.Scene.ReferenceAssetIDs = compactStrings(bindings.Scene.ReferenceAssetIDs)
	}
	for index := range bindings.Props {
		bindings.Props[index].CanonID = strings.TrimSpace(bindings.Props[index].CanonID)
		bindings.Props[index].VariantID = strings.TrimSpace(bindings.Props[index].VariantID)
		bindings.Props[index].ReferenceAssetIDs = compactStrings(bindings.Props[index].ReferenceAssetIDs)
	}
	sort.SliceStable(bindings.Props, func(i, j int) bool {
		return bindings.Props[i].CanonID < bindings.Props[j].CanonID
	})
	return bindings
}

func validateInput(input UpsertInput) error {
	if input.ProjectID == "" || input.DocumentID == "" || input.SectionID == "" {
		return errors.New("projectId, documentId and sectionId are required")
	}
	if input.Sequence < 0 {
		return errors.New("sequence must not be negative")
	}
	if input.StartSeconds < 0 || input.EndSeconds < 0 || input.DurationSeconds < 0 {
		return errors.New("shot timing must not be negative")
	}
	if input.EndSeconds > 0 && input.EndSeconds < input.StartSeconds {
		return errors.New("shot endSeconds must not precede startSeconds")
	}
	if input.Status != StatusDraft && input.Status != StatusReady && input.Status != StatusConflict {
		return fmt.Errorf("invalid shot manifest status %q", input.Status)
	}
	if _, err := decodeStateObject(input.StateChangesJSON); err != nil {
		return fmt.Errorf("invalid stateChangesJson: %w", err)
	}
	return nil
}

func recordFromModel(model domain.ShotManifestModel) (Record, error) {
	bindings := Bindings{}
	if err := json.Unmarshal([]byte(model.BindingsJSON), &bindings); err != nil {
		return Record{}, fmt.Errorf("decoding shot bindings: %w", err)
	}
	return Record{
		ID:                model.ID,
		ProjectID:         model.ProjectID,
		DocumentID:        model.DocumentID,
		SectionID:         model.SectionID,
		ShotKey:           model.ShotKey,
		Sequence:          model.Sequence,
		StartSeconds:      model.StartSeconds,
		EndSeconds:        model.EndSeconds,
		DurationSeconds:   model.DurationSeconds,
		InheritsFromID:    domain.StringValue(model.InheritsFromID),
		ActionText:        model.ActionText,
		CameraText:        model.CameraText,
		AudioText:         model.AudioText,
		StyleProfileID:    model.StyleProfileID,
		Bindings:          bindings,
		StateChangesJSON:  model.StateChangesJSON,
		ResolvedStateJSON: model.ResolvedStateJSON,
		CompiledPrompt:    model.CompiledPrompt,
		SourceHash:        model.SourceHash,
		Version:           model.Version,
		Status:            model.Status,
		CreatedAt:         domain.StringFromTime(model.CreatedAt),
		UpdatedAt:         domain.StringFromTime(model.UpdatedAt),
	}, nil
}

func shotIDForSource(projectID string, documentID string, sectionID string, shotKey string) string {
	digest := sha256.Sum256([]byte(strings.Join([]string{projectID, documentID, sectionID, shotKey}, "\x00")))
	return "shot-" + hex.EncodeToString(digest[:])[:24]
}

func shotMaterialHash(model domain.ShotManifestModel) string {
	return hashStrings(
		fmt.Sprintf("%.6f", model.StartSeconds),
		fmt.Sprintf("%.6f", model.EndSeconds),
		fmt.Sprintf("%.6f", model.DurationSeconds),
		model.ActionText,
		model.CameraText,
		model.AudioText,
		model.StyleProfileID,
		model.BindingsJSON,
		model.StateChangesJSON,
		model.ResolvedStateJSON,
		model.SourceHash,
		model.Status,
		domain.StringValue(model.InheritsFromID),
	)
}

func hashStrings(values ...string) string {
	digest := sha256.Sum256([]byte(strings.Join(values, "\x00")))
	return hex.EncodeToString(digest[:])
}

func compactStrings(values []string) []string {
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
	sort.Strings(result)
	return result
}
