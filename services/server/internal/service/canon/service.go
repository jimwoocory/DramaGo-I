package canon

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/mediago-dev/mediago-drama/services/server/internal/domain"
	"github.com/mediago-dev/mediago-drama/services/server/internal/repository"
	"github.com/mediago-dev/mediago-drama/services/server/internal/service/model"
	"github.com/mediago-dev/mediago-drama/services/server/internal/service/shared"
)

const (
	ResourceTypeCharacter = "character"
	ResourceTypeScene     = "scene"
	ResourceTypeProp      = "prop"

	StatusDraft      = "draft"
	StatusApproved   = "approved"
	StatusLocked     = "locked"
	StatusDeprecated = "deprecated"
)

// Service owns Canon lifecycle rules above the persistence layer.
type documentResourceProvider interface {
	ListWorkspaceDocumentResources(projectID string) (model.WorkspaceDocumentResourcesResponse, error)
}

type selectedAssetProvider interface {
	ListProjectSelectedAssets(projectID string) ([]domain.ProjectSelectedAssetModel, error)
}

// Service owns Canon lifecycle rules above the persistence layer.
type Service struct {
	repo       *repository.CanonRepository
	documents  documentResourceProvider
	selections selectedAssetProvider
	initErr    error
}

// SourceResource describes one document resource that should have a Canon core.
type SourceResource struct {
	ProjectID        string
	ResourceType     string
	ResourceID       string
	SourceDocumentID string
	Name             string
	SpecJSON         string
	PromptText       string
	SourceHash       string
}

// SyncResult reports whether source synchronization changed the Canon core.
type SyncResult struct {
	Asset   domain.CanonAssetModel
	Created bool
	Changed bool
	Locked  bool
}

// VariantInput creates a mutable appearance/state variant under a Canon object.
type VariantInput struct {
	ProjectID   string
	ParentID    string
	VariantKind string
	Name        string
	SpecJSON    string
	PromptText  string
	Status      string
}

// ReferenceInput binds an existing project media asset to a Canon core or variant.
type ReferenceInput struct {
	ProjectID    string
	CanonAssetID string
	AssetID      string
	Role         string
	Priority     int
	Locked       bool
}

// ResourceSyncSummary reports one lazy document-resource synchronization pass.
type ResourceSyncSummary struct {
	Created int                      `json:"created"`
	Updated int                      `json:"updated"`
	Locked  int                      `json:"locked"`
	Skipped int                      `json:"skipped"`
	Assets  []domain.CanonAssetModel `json:"-"`
}

// ProjectSyncSummary reports a full lazy sync from document resources and
// selected generation assets.
type ProjectSyncSummary struct {
	Resources         ResourceSyncSummary `json:"resources"`
	ReferencesCreated int                 `json:"referencesCreated"`
	ReferencesReused  int                 `json:"referencesReused"`
	SelectionsSkipped int                 `json:"selectionsSkipped"`
}

// ReferenceRecord is the API-safe projection of a Canon media reference.
type ReferenceRecord struct {
	ID        string `json:"id"`
	AssetID   string `json:"assetId"`
	Role      string `json:"role"`
	Priority  int    `json:"priority"`
	Locked    bool   `json:"locked"`
	URL       string `json:"url,omitempty"`
	PosterURL string `json:"posterUrl,omitempty"`
}

// AssetRecord is the API-safe projection of a Canon core/variant.
type AssetRecord struct {
	ID               string            `json:"id"`
	ProjectID        string            `json:"projectId"`
	ResourceType     string            `json:"resourceType"`
	ResourceID       string            `json:"resourceId"`
	SourceDocumentID string            `json:"sourceDocumentId"`
	ParentID         string            `json:"parentId,omitempty"`
	VariantKind      string            `json:"variantKind,omitempty"`
	Name             string            `json:"name"`
	SpecJSON         string            `json:"specJson"`
	PromptText       string            `json:"promptText"`
	Status           string            `json:"status"`
	Version          int               `json:"version"`
	SourceHash       string            `json:"sourceHash,omitempty"`
	References       []ReferenceRecord `json:"references"`
	CreatedAt        string            `json:"createdAt"`
	UpdatedAt        string            `json:"updatedAt"`
}

// NewService returns a Canon service.
func NewService(repo *repository.CanonRepository, initErr error) *Service {
	service := &Service{repo: repo, initErr: initErr}
	if service.initErr == nil && service.repo == nil {
		service.initErr = errors.New("canon repository is nil")
	}
	return service
}

// SetDocumentResourceProvider wires the existing workspace resource parser into Canon.
func (service *Service) SetDocumentResourceProvider(provider documentResourceProvider) {
	service.documents = provider
}

// SetSelectedAssetProvider wires MediaGo's existing selected-generation-asset store into Canon.
func (service *Service) SetSelectedAssetProvider(provider selectedAssetProvider) {
	service.selections = provider
}

// List returns Canon cores/variants and their reference bindings for one project.
func (service *Service) List(projectID string, resourceType string) ([]AssetRecord, error) {
	if service.initErr != nil {
		return nil, service.initErr
	}
	projectID = domain.CleanProjectID(projectID)
	resourceType = strings.ToLower(strings.TrimSpace(resourceType))
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}
	if resourceType != "" && !validResourceType(resourceType) {
		return nil, fmt.Errorf("invalid canon resource type %q", resourceType)
	}
	assets, err := service.repo.ListCanonAssets(projectID, resourceType)
	if err != nil {
		return nil, err
	}
	records := make([]AssetRecord, 0, len(assets))
	for _, asset := range assets {
		references, err := service.repo.ListCanonReferences(projectID, asset.ID)
		if err != nil {
			return nil, err
		}
		records = append(records, assetRecord(asset, references))
	}
	return records, nil
}

// Get returns one Canon core/variant with its reference bindings.
func (service *Service) Get(projectID string, id string) (AssetRecord, error) {
	if service.initErr != nil {
		return AssetRecord{}, service.initErr
	}
	projectID = domain.CleanProjectID(projectID)
	id = strings.TrimSpace(id)
	if projectID == "" || id == "" {
		return AssetRecord{}, errors.New("projectId and canon id are required")
	}
	asset, err := service.repo.GetCanonAsset(projectID, id)
	if err != nil {
		return AssetRecord{}, err
	}
	references, err := service.repo.ListCanonReferences(projectID, id)
	if err != nil {
		return AssetRecord{}, err
	}
	return assetRecord(asset, references), nil
}

// UpdateStatus performs an explicit user/system status transition. Source sync
// still respects locked status and will not overwrite it.
func (service *Service) UpdateStatus(projectID string, id string, status string) (AssetRecord, error) {
	if service.initErr != nil {
		return AssetRecord{}, service.initErr
	}
	projectID = domain.CleanProjectID(projectID)
	id = strings.TrimSpace(id)
	status = normalizeStatus(status)
	if projectID == "" || id == "" {
		return AssetRecord{}, errors.New("projectId and canon id are required")
	}
	if !validStatus(status) {
		return AssetRecord{}, fmt.Errorf("invalid canon status %q", status)
	}
	updated, err := service.repo.UpdateCanonAsset(projectID, id, map[string]any{"status": status})
	if err != nil {
		return AssetRecord{}, err
	}
	if !updated {
		return AssetRecord{}, repository.ErrRecordNotFound
	}
	asset, err := service.repo.GetCanonAsset(projectID, id)
	if err != nil {
		return AssetRecord{}, err
	}
	references, err := service.repo.ListCanonReferences(projectID, id)
	if err != nil {
		return AssetRecord{}, err
	}
	return assetRecord(asset, references), nil
}

// EnsureCoreFromSource creates or refreshes a Canon core from a document
// resource. A locked Canon is never overwritten by source synchronization.
func (service *Service) EnsureCoreFromSource(input SourceResource) (SyncResult, error) {
	if service.initErr != nil {
		return SyncResult{}, service.initErr
	}
	input = normalizeSourceResource(input)
	if err := validateSourceResource(input); err != nil {
		return SyncResult{}, err
	}

	existing, err := service.repo.FindCanonCoreBySource(
		input.ProjectID,
		input.ResourceType,
		input.SourceDocumentID,
		input.ResourceID,
	)
	if repository.IsRecordNotFound(err) {
		model := domain.CanonAssetModel{
			ID:               coreIDForSource(input),
			ProjectID:        input.ProjectID,
			ResourceType:     input.ResourceType,
			ResourceID:       input.ResourceID,
			SourceDocumentID: input.SourceDocumentID,
			Name:             input.Name,
			SpecJSON:         input.SpecJSON,
			PromptText:       input.PromptText,
			Status:           StatusDraft,
			Version:          1,
			SourceHash:       input.SourceHash,
		}
		if err := service.repo.CreateCanonAsset(model); err != nil {
			return SyncResult{}, err
		}
		created, err := service.repo.GetCanonAsset(input.ProjectID, model.ID)
		if err != nil {
			return SyncResult{}, err
		}
		return SyncResult{Asset: created, Created: true, Changed: true}, nil
	}
	if err != nil {
		return SyncResult{}, err
	}

	if existing.Status == StatusLocked {
		return SyncResult{Asset: existing, Locked: true}, nil
	}
	if existing.SourceHash == input.SourceHash {
		return SyncResult{Asset: existing}, nil
	}

	updated, err := service.repo.UpdateCanonAsset(input.ProjectID, existing.ID, map[string]any{
		"name":        input.Name,
		"spec_json":   input.SpecJSON,
		"prompt_text": input.PromptText,
		"source_hash": input.SourceHash,
		"version":     existing.Version + 1,
	})
	if err != nil {
		return SyncResult{}, err
	}
	if !updated {
		return SyncResult{}, repository.ErrRecordNotFound
	}
	next, err := service.repo.GetCanonAsset(input.ProjectID, existing.ID)
	if err != nil {
		return SyncResult{}, err
	}
	return SyncResult{Asset: next, Changed: true}, nil
}

// CreateVariant creates a child variant while preserving the parent's identity
// and source linkage.
func (service *Service) CreateVariant(input VariantInput) (domain.CanonAssetModel, error) {
	if service.initErr != nil {
		return domain.CanonAssetModel{}, service.initErr
	}
	input.ProjectID = domain.CleanProjectID(input.ProjectID)
	input.ParentID = strings.TrimSpace(input.ParentID)
	input.VariantKind = strings.TrimSpace(input.VariantKind)
	input.Name = strings.TrimSpace(input.Name)
	input.SpecJSON = normalizeSpecJSON(input.SpecJSON)
	input.PromptText = strings.TrimSpace(input.PromptText)
	input.Status = normalizeStatus(input.Status)
	if input.ProjectID == "" || input.ParentID == "" {
		return domain.CanonAssetModel{}, errors.New("projectId and parentId are required")
	}
	if input.VariantKind == "" {
		return domain.CanonAssetModel{}, errors.New("variantKind is required")
	}
	if input.Status == "" {
		input.Status = StatusDraft
	}
	if !validStatus(input.Status) {
		return domain.CanonAssetModel{}, fmt.Errorf("invalid canon status %q", input.Status)
	}

	parent, err := service.repo.GetCanonAsset(input.ProjectID, input.ParentID)
	if err != nil {
		return domain.CanonAssetModel{}, err
	}
	parentID := parent.ID
	model := domain.CanonAssetModel{
		ID:               shared.MustRandomID("canon"),
		ProjectID:        parent.ProjectID,
		ResourceType:     parent.ResourceType,
		ResourceID:       parent.ResourceID,
		SourceDocumentID: parent.SourceDocumentID,
		ParentID:         &parentID,
		VariantKind:      input.VariantKind,
		Name:             input.Name,
		SpecJSON:         input.SpecJSON,
		PromptText:       input.PromptText,
		Status:           input.Status,
		Version:          1,
		SourceHash:       sourceHash(input.SpecJSON, input.PromptText, input.Name, input.VariantKind),
	}
	if err := service.repo.CreateCanonAsset(model); err != nil {
		return domain.CanonAssetModel{}, err
	}
	return service.repo.GetCanonAsset(input.ProjectID, model.ID)
}

// BindReference attaches or refreshes a project-local physical media asset on a Canon core/variant.
func (service *Service) BindReference(input ReferenceInput) (ReferenceRecord, error) {
	if service.initErr != nil {
		return ReferenceRecord{}, service.initErr
	}
	input.ProjectID = domain.CleanProjectID(input.ProjectID)
	input.CanonAssetID = strings.TrimSpace(input.CanonAssetID)
	input.AssetID = strings.TrimSpace(input.AssetID)
	input.Role = strings.ToLower(strings.TrimSpace(input.Role))
	if input.ProjectID == "" || input.CanonAssetID == "" || input.AssetID == "" {
		return ReferenceRecord{}, errors.New("projectId, canonAssetId and assetId are required")
	}
	canonAsset, err := service.repo.GetCanonAsset(input.ProjectID, input.CanonAssetID)
	if err != nil {
		return ReferenceRecord{}, err
	}
	belongs, err := service.repo.MediaAssetBelongsToProject(input.ProjectID, input.AssetID)
	if err != nil {
		return ReferenceRecord{}, err
	}
	if !belongs {
		return ReferenceRecord{}, fmt.Errorf("media asset %s does not belong to project %s", input.AssetID, input.ProjectID)
	}
	if input.Role == "" {
		input.Role = defaultReferenceRoleForCanon(canonAsset)
	}
	if input.Priority == 0 {
		input.Priority = defaultReferencePriorityForCanon(canonAsset)
	}

	existing, err := service.repo.FindCanonReferenceByAsset(input.ProjectID, input.CanonAssetID, input.AssetID)
	if err == nil {
		updated, updateErr := service.repo.UpdateCanonReference(input.ProjectID, existing.ID, map[string]any{
			"role":     input.Role,
			"priority": input.Priority,
			"locked":   input.Locked,
		})
		if updateErr != nil {
			return ReferenceRecord{}, updateErr
		}
		if !updated {
			return ReferenceRecord{}, repository.ErrRecordNotFound
		}
		refreshed, getErr := service.repo.FindCanonReferenceByAsset(input.ProjectID, input.CanonAssetID, input.AssetID)
		if getErr != nil {
			return ReferenceRecord{}, getErr
		}
		return referenceRecord(refreshed), nil
	}
	if !repository.IsRecordNotFound(err) {
		return ReferenceRecord{}, err
	}

	model := domain.CanonReferenceModel{
		ID:           shared.MustRandomID("canon-ref"),
		ProjectID:    input.ProjectID,
		CanonAssetID: input.CanonAssetID,
		AssetID:      input.AssetID,
		Role:         input.Role,
		Priority:     input.Priority,
		Locked:       input.Locked,
	}
	if err := service.repo.CreateCanonReference(model); err != nil {
		return ReferenceRecord{}, err
	}
	created, err := service.repo.FindCanonReferenceByAsset(input.ProjectID, input.CanonAssetID, input.AssetID)
	if err != nil {
		return ReferenceRecord{}, err
	}
	return referenceRecord(created), nil
}

// SyncDocumentResources lazily materializes character/scene/prop document
// sections as Canon cores. Storyboard sections are execution units and are
// intentionally excluded from Canon resources.
func (service *Service) SyncDocumentResources(projectID string, resources []model.WorkspaceDocumentResourceRecord) (ResourceSyncSummary, error) {
	if service.initErr != nil {
		return ResourceSyncSummary{}, service.initErr
	}
	projectID = domain.CleanProjectID(projectID)
	if projectID == "" {
		return ResourceSyncSummary{}, errors.New("projectId is required")
	}

	summary := ResourceSyncSummary{Assets: []domain.CanonAssetModel{}}
	for _, resource := range resources {
		resourceType := strings.ToLower(strings.TrimSpace(resource.Type))
		if !validResourceType(resourceType) {
			summary.Skipped++
			continue
		}
		parts := splitSourceResource(resourceType, resource.Markdown, resource.Prompt)
		result, err := service.EnsureCoreFromSource(SourceResource{
			ProjectID:        projectID,
			ResourceType:     resourceType,
			ResourceID:       resource.SectionID,
			SourceDocumentID: resource.DocumentID,
			Name:             resource.Title,
			SpecJSON:         canonSourceSpec(parts.CorePrompt),
			PromptText:       parts.CorePrompt,
			SourceHash:       sourceHash(parts.CoreMarkdown),
		})
		if err != nil {
			return ResourceSyncSummary{}, err
		}
		syncSummaryAdd(&summary, result)
		if err := syncSourceVariants(service, projectID, result.Asset.ID, parts.Variants, &summary); err != nil {
			return ResourceSyncSummary{}, err
		}
	}
	return summary, nil
}

// SyncProject performs the full lazy project materialization used by the HTTP
// API and future Shot/Continuity resolver.
func (service *Service) SyncProject(projectID string) (ProjectSyncSummary, error) {
	if service.initErr != nil {
		return ProjectSyncSummary{}, service.initErr
	}
	projectID = domain.CleanProjectID(projectID)
	if projectID == "" {
		return ProjectSyncSummary{}, errors.New("projectId is required")
	}
	if service.documents == nil {
		return ProjectSyncSummary{}, errors.New("canon document resource provider is not configured")
	}
	resources, err := service.documents.ListWorkspaceDocumentResources(projectID)
	if err != nil {
		return ProjectSyncSummary{}, err
	}
	resourceSummary, err := service.SyncDocumentResources(projectID, resources.Resources)
	if err != nil {
		return ProjectSyncSummary{}, err
	}
	summary := ProjectSyncSummary{Resources: resourceSummary}
	if service.selections == nil {
		return summary, nil
	}
	selectedAssets, err := service.selections.ListProjectSelectedAssets(projectID)
	if err != nil {
		return ProjectSyncSummary{}, err
	}
	for _, selected := range selectedAssets {
		resourceType := strings.ToLower(strings.TrimSpace(selected.ResourceType))
		if !validResourceType(resourceType) ||
			strings.TrimSpace(domain.StringValue(selected.ResourceID)) == "" ||
			strings.TrimSpace(domain.StringValue(selected.SourceDocumentID)) == "" ||
			strings.TrimSpace(selected.AssetID) == "" {
			summary.SelectionsSkipped++
			continue
		}
		_, created, err := service.EnsureReferenceFromSelectedAsset(selected)
		if repository.IsRecordNotFound(err) {
			summary.SelectionsSkipped++
			continue
		}
		if err != nil {
			return ProjectSyncSummary{}, err
		}
		if created {
			summary.ReferencesCreated++
		} else {
			summary.ReferencesReused++
		}
	}
	return summary, nil
}

func (service *Service) EnsureReferenceFromSelectedAsset(selected domain.ProjectSelectedAssetModel) (domain.CanonReferenceModel, bool, error) {
	if service.initErr != nil {
		return domain.CanonReferenceModel{}, false, service.initErr
	}
	projectID := domain.CleanProjectID(selected.ProjectID)
	resourceType := strings.ToLower(strings.TrimSpace(selected.ResourceType))
	resourceID := strings.TrimSpace(domain.StringValue(selected.ResourceID))
	sourceDocumentID := strings.TrimSpace(domain.StringValue(selected.SourceDocumentID))
	assetID := strings.TrimSpace(selected.AssetID)
	if projectID == "" || !validResourceType(resourceType) || resourceID == "" || sourceDocumentID == "" || assetID == "" {
		return domain.CanonReferenceModel{}, false, errors.New("selected asset is missing canonical resource identity")
	}

	core, err := service.repo.FindCanonCoreBySource(projectID, resourceType, sourceDocumentID, resourceID)
	if err != nil {
		return domain.CanonReferenceModel{}, false, err
	}
	existing, err := service.repo.ListCanonReferences(projectID, core.ID)
	if err != nil {
		return domain.CanonReferenceModel{}, false, err
	}
	for _, reference := range existing {
		if reference.AssetID == assetID {
			return reference, false, nil
		}
	}

	reference := domain.CanonReferenceModel{
		ID:           shared.MustRandomID("canon-ref"),
		ProjectID:    projectID,
		CanonAssetID: core.ID,
		AssetID:      assetID,
		Role:         defaultReferenceRole(resourceType),
		Priority:     100,
		Locked:       false,
	}
	if err := service.repo.CreateCanonReference(reference); err != nil {
		return domain.CanonReferenceModel{}, false, err
	}
	created, err := service.repo.ListCanonReferences(projectID, core.ID)
	if err != nil {
		return domain.CanonReferenceModel{}, false, err
	}
	for _, item := range created {
		if item.ID == reference.ID {
			return item, true, nil
		}
	}
	return domain.CanonReferenceModel{}, false, repository.ErrRecordNotFound
}

func normalizeSourceResource(input SourceResource) SourceResource {
	input.ProjectID = domain.CleanProjectID(input.ProjectID)
	input.ResourceType = strings.ToLower(strings.TrimSpace(input.ResourceType))
	input.ResourceID = strings.TrimSpace(input.ResourceID)
	input.SourceDocumentID = strings.TrimSpace(input.SourceDocumentID)
	input.Name = strings.TrimSpace(input.Name)
	input.SpecJSON = normalizeSpecJSON(input.SpecJSON)
	input.PromptText = strings.TrimSpace(input.PromptText)
	input.SourceHash = strings.TrimSpace(input.SourceHash)
	if input.SourceHash == "" {
		input.SourceHash = sourceHash(input.ResourceType, input.ResourceID, input.SourceDocumentID, input.Name, input.SpecJSON, input.PromptText)
	}
	return input
}

func validateSourceResource(input SourceResource) error {
	if input.ProjectID == "" || input.ResourceID == "" || input.SourceDocumentID == "" {
		return errors.New("projectId, resourceId and sourceDocumentId are required")
	}
	if !validResourceType(input.ResourceType) {
		return fmt.Errorf("invalid canon resource type %q", input.ResourceType)
	}
	return nil
}

func validResourceType(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case ResourceTypeCharacter, ResourceTypeScene, ResourceTypeProp:
		return true
	default:
		return false
	}
}

func normalizeStatus(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func validStatus(value string) bool {
	switch normalizeStatus(value) {
	case StatusDraft, StatusApproved, StatusLocked, StatusDeprecated:
		return true
	default:
		return false
	}
}

func defaultReferenceRole(resourceType string) string {
	switch strings.ToLower(strings.TrimSpace(resourceType)) {
	case ResourceTypeCharacter:
		return "identity"
	case ResourceTypeScene:
		return "scene_master"
	case ResourceTypeProp:
		return "prop_master"
	default:
		return "custom"
	}
}

func defaultReferenceRoleForCanon(asset domain.CanonAssetModel) string {
	if asset.ParentID == nil {
		return defaultReferenceRole(asset.ResourceType)
	}
	switch strings.ToLower(strings.TrimSpace(asset.ResourceType)) {
	case ResourceTypeCharacter:
		return "look"
	case ResourceTypeScene:
		return "scene_variant"
	case ResourceTypeProp:
		return "prop_variant"
	default:
		return "custom"
	}
}

func defaultReferencePriorityForCanon(asset domain.CanonAssetModel) int {
	if asset.ParentID == nil {
		switch strings.ToLower(strings.TrimSpace(asset.ResourceType)) {
		case ResourceTypeCharacter:
			return 1000
		case ResourceTypeScene:
			return 700
		case ResourceTypeProp:
			return 400
		default:
			return 100
		}
	}
	switch strings.ToLower(strings.TrimSpace(asset.ResourceType)) {
	case ResourceTypeCharacter:
		return 900
	case ResourceTypeScene:
		return 800
	case ResourceTypeProp:
		return 500
	default:
		return 100
	}
}

func assetRecord(asset domain.CanonAssetModel, references []domain.CanonReferenceModel) AssetRecord {
	record := AssetRecord{
		ID:               asset.ID,
		ProjectID:        asset.ProjectID,
		ResourceType:     asset.ResourceType,
		ResourceID:       asset.ResourceID,
		SourceDocumentID: asset.SourceDocumentID,
		ParentID:         domain.StringValue(asset.ParentID),
		VariantKind:      asset.VariantKind,
		Name:             asset.Name,
		SpecJSON:         asset.SpecJSON,
		PromptText:       asset.PromptText,
		Status:           asset.Status,
		Version:          asset.Version,
		SourceHash:       asset.SourceHash,
		References:       make([]ReferenceRecord, 0, len(references)),
		CreatedAt:        domain.StringFromTime(asset.CreatedAt),
		UpdatedAt:        domain.StringFromTime(asset.UpdatedAt),
	}
	for _, reference := range references {
		record.References = append(record.References, referenceRecord(reference))
	}
	return record
}

func referenceRecord(reference domain.CanonReferenceModel) ReferenceRecord {
	return ReferenceRecord{
		ID:        reference.ID,
		AssetID:   reference.AssetID,
		Role:      reference.Role,
		Priority:  reference.Priority,
		Locked:    reference.Locked,
		URL:       reference.Asset.URL,
		PosterURL: reference.Asset.PosterURL,
	}
}

func normalizeSpecJSON(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "{}"
	}
	var payload any
	if json.Unmarshal([]byte(value), &payload) != nil {
		return value
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return value
	}
	return string(encoded)
}

func coreIDForSource(input SourceResource) string {
	digest := sha256.Sum256([]byte(strings.Join([]string{
		input.ProjectID,
		input.ResourceType,
		input.SourceDocumentID,
		input.ResourceID,
	}, "\x00")))
	return "canon-" + hex.EncodeToString(digest[:])[:24]
}

func sourceHash(values ...string) string {
	digest := sha256.Sum256([]byte(strings.Join(values, "\x00")))
	return hex.EncodeToString(digest[:])
}
