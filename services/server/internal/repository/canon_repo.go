package repository

import (
	"fmt"
	"strings"

	"github.com/mediago-dev/mediago-drama/services/server/internal/domain"
	"gorm.io/gorm"
)

// CanonRepository persists project-scoped canonical creative resources and
// their reference bindings.
type CanonRepository struct {
	db *gorm.DB
}

// NewCanonRepository opens the workspace database via the central schema owner.
func NewCanonRepository(dbPath string) (*CanonRepository, error) {
	db, err := OpenWorkspaceDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening canon repository database: %w", err)
	}
	return NewCanonRepositoryFromDB(db), nil
}

// NewCanonRepositoryFromDB creates a repository from an existing workspace DB.
func NewCanonRepositoryFromDB(db *gorm.DB) *CanonRepository {
	return &CanonRepository{db: db}
}

// ListCanonAssets lists cores and variants for one project, optionally filtered
// by resource type.
func (repo *CanonRepository) ListCanonAssets(projectID string, resourceType string) ([]domain.CanonAssetModel, error) {
	models := []domain.CanonAssetModel{}
	query := repo.db.Where("project_id = ?", domain.CleanProjectID(projectID))
	if resourceType = strings.TrimSpace(resourceType); resourceType != "" {
		query = query.Where("resource_type = ?", resourceType)
	}
	if err := query.Order("created_at ASC, id ASC").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("listing canon assets: %w", err)
	}
	return models, nil
}

// GetCanonAsset returns one canon core or variant within a project.
func (repo *CanonRepository) GetCanonAsset(projectID string, id string) (domain.CanonAssetModel, error) {
	var model domain.CanonAssetModel
	err := repo.db.First(
		&model,
		"project_id = ? AND id = ?",
		domain.CleanProjectID(projectID),
		strings.TrimSpace(id),
	).Error
	if IsRecordNotFound(err) {
		return domain.CanonAssetModel{}, ErrRecordNotFound
	}
	if err != nil {
		return domain.CanonAssetModel{}, fmt.Errorf("getting canon asset: %w", err)
	}
	return model, nil
}

// FindCanonCoreBySource finds the root canon object associated with a document
// resource section. Variants are intentionally excluded.
func (repo *CanonRepository) FindCanonCoreBySource(projectID string, resourceType string, sourceDocumentID string, resourceID string) (domain.CanonAssetModel, error) {
	var model domain.CanonAssetModel
	err := repo.db.Where(
		"project_id = ? AND resource_type = ? AND source_document_id = ? AND resource_id = ? AND parent_id IS NULL",
		domain.CleanProjectID(projectID),
		strings.TrimSpace(resourceType),
		strings.TrimSpace(sourceDocumentID),
		strings.TrimSpace(resourceID),
	).Order("version DESC, updated_at DESC").First(&model).Error
	if IsRecordNotFound(err) {
		return domain.CanonAssetModel{}, ErrRecordNotFound
	}
	if err != nil {
		return domain.CanonAssetModel{}, fmt.Errorf("finding canon core by source: %w", err)
	}
	return model, nil
}

// FindCanonVariant finds one direct variant by stable parent/kind/name identity.
func (repo *CanonRepository) FindCanonVariant(projectID string, parentID string, variantKind string, name string) (domain.CanonAssetModel, error) {
	var model domain.CanonAssetModel
	err := repo.db.Where(
		"project_id = ? AND parent_id = ? AND variant_kind = ? AND name = ?",
		domain.CleanProjectID(projectID),
		strings.TrimSpace(parentID),
		strings.TrimSpace(variantKind),
		strings.TrimSpace(name),
	).Order("version DESC, updated_at DESC").First(&model).Error
	if IsRecordNotFound(err) {
		return domain.CanonAssetModel{}, ErrRecordNotFound
	}
	if err != nil {
		return domain.CanonAssetModel{}, fmt.Errorf("finding canon variant: %w", err)
	}
	return model, nil
}

// ListCanonVariants lists direct variants of one canon object.
func (repo *CanonRepository) ListCanonVariants(projectID string, parentID string) ([]domain.CanonAssetModel, error) {
	models := []domain.CanonAssetModel{}
	if err := repo.db.Where(
		"project_id = ? AND parent_id = ?",
		domain.CleanProjectID(projectID),
		strings.TrimSpace(parentID),
	).Order("created_at ASC, id ASC").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("listing canon variants: %w", err)
	}
	return models, nil
}

// CreateCanonAsset inserts one core or variant.
func (repo *CanonRepository) CreateCanonAsset(model domain.CanonAssetModel) error {
	model.ProjectID = domain.CleanProjectID(model.ProjectID)
	model.ID = strings.TrimSpace(model.ID)
	model.ResourceType = strings.TrimSpace(model.ResourceType)
	model.ResourceID = strings.TrimSpace(model.ResourceID)
	model.SourceDocumentID = strings.TrimSpace(model.SourceDocumentID)
	model.VariantKind = strings.TrimSpace(model.VariantKind)
	model.Name = strings.TrimSpace(model.Name)
	model.Status = strings.TrimSpace(model.Status)
	model.SourceHash = strings.TrimSpace(model.SourceHash)
	if model.SpecJSON == "" {
		model.SpecJSON = "{}"
	}
	if model.Status == "" {
		model.Status = "draft"
	}
	if model.Version <= 0 {
		model.Version = 1
	}
	if err := repo.db.Create(&model).Error; err != nil {
		return fmt.Errorf("creating canon asset: %w", err)
	}
	return nil
}

// UpdateCanonAsset updates mutable canon fields inside one project.
func (repo *CanonRepository) UpdateCanonAsset(projectID string, id string, updates map[string]any) (bool, error) {
	if len(updates) == 0 {
		return true, nil
	}
	result := repo.db.Model(&domain.CanonAssetModel{}).
		Where("project_id = ? AND id = ?", domain.CleanProjectID(projectID), strings.TrimSpace(id)).
		Updates(updates)
	if result.Error != nil {
		return false, fmt.Errorf("updating canon asset: %w", result.Error)
	}
	return result.RowsAffected > 0, nil
}

// ListCanonReferences returns media references for one canon core/variant.
func (repo *CanonRepository) ListCanonReferences(projectID string, canonAssetID string) ([]domain.CanonReferenceModel, error) {
	models := []domain.CanonReferenceModel{}
	if err := repo.db.Preload("Asset").Where(
		"project_id = ? AND canon_asset_id = ?",
		domain.CleanProjectID(projectID),
		strings.TrimSpace(canonAssetID),
	).Order("priority DESC, created_at ASC, id ASC").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("listing canon references: %w", err)
	}
	return models, nil
}

// CreateCanonReference binds an existing physical asset to a canon object.
func (repo *CanonRepository) CreateCanonReference(model domain.CanonReferenceModel) error {
	model.ProjectID = domain.CleanProjectID(model.ProjectID)
	model.ID = strings.TrimSpace(model.ID)
	model.CanonAssetID = strings.TrimSpace(model.CanonAssetID)
	model.AssetID = strings.TrimSpace(model.AssetID)
	model.Role = strings.TrimSpace(model.Role)
	if model.Role == "" {
		model.Role = "custom"
	}
	if err := repo.db.Create(&model).Error; err != nil {
		return fmt.Errorf("creating canon reference: %w", err)
	}
	return nil
}

// DeleteCanonReference removes one reference binding without deleting the
// underlying media asset.
func (repo *CanonRepository) DeleteCanonReference(projectID string, id string) (bool, error) {
	result := repo.db.Delete(
		&domain.CanonReferenceModel{},
		"project_id = ? AND id = ?",
		domain.CleanProjectID(projectID),
		strings.TrimSpace(id),
	)
	if result.Error != nil {
		return false, fmt.Errorf("deleting canon reference: %w", result.Error)
	}
	return result.RowsAffected > 0, nil
}

// FindCanonReferenceByAsset returns an existing binding of one physical asset to one Canon object.
func (repo *CanonRepository) FindCanonReferenceByAsset(projectID string, canonAssetID string, assetID string) (domain.CanonReferenceModel, error) {
	var model domain.CanonReferenceModel
	err := repo.db.Preload("Asset").Where(
		"project_id = ? AND canon_asset_id = ? AND asset_id = ?",
		domain.CleanProjectID(projectID), strings.TrimSpace(canonAssetID), strings.TrimSpace(assetID),
	).First(&model).Error
	if IsRecordNotFound(err) {
		return domain.CanonReferenceModel{}, ErrRecordNotFound
	}
	if err != nil {
		return domain.CanonReferenceModel{}, fmt.Errorf("finding canon reference by asset: %w", err)
	}
	return model, nil
}

// UpdateCanonReference updates mutable policy fields on an existing Canon reference.
func (repo *CanonRepository) UpdateCanonReference(projectID string, id string, updates map[string]any) (bool, error) {
	if len(updates) == 0 {
		return true, nil
	}
	result := repo.db.Model(&domain.CanonReferenceModel{}).
		Where("project_id = ? AND id = ?", domain.CleanProjectID(projectID), strings.TrimSpace(id)).
		Updates(updates)
	if result.Error != nil {
		return false, fmt.Errorf("updating canon reference: %w", result.Error)
	}
	return result.RowsAffected > 0, nil
}

// MediaAssetBelongsToProject verifies that a physical media asset exists in the same project.
func (repo *CanonRepository) MediaAssetBelongsToProject(projectID string, assetID string) (bool, error) {
	var count int64
	if err := repo.db.Model(&domain.AssetModel{}).
		Where("project_id = ? AND id = ?", domain.CleanProjectID(projectID), strings.TrimSpace(assetID)).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("checking canon reference media asset: %w", err)
	}
	return count > 0, nil
}
