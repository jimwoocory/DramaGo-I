package repository

import (
	"fmt"
	"strings"

	"github.com/mediago-dev/mediago-drama/services/server/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ShotManifestRepository persists structured storyboard execution manifests.
type ShotManifestRepository struct {
	db *gorm.DB
}

func NewShotManifestRepositoryFromDB(db *gorm.DB) *ShotManifestRepository {
	return &ShotManifestRepository{db: db}
}

// Get returns one shot manifest scoped to a project.
func (repo *ShotManifestRepository) Get(projectID string, id string) (domain.ShotManifestModel, error) {
	var model domain.ShotManifestModel
	err := repo.db.First(
		&model,
		"project_id = ? AND id = ?",
		domain.CleanProjectID(projectID),
		strings.TrimSpace(id),
	).Error
	if IsRecordNotFound(err) {
		return domain.ShotManifestModel{}, ErrRecordNotFound
	}
	if err != nil {
		return domain.ShotManifestModel{}, fmt.Errorf("getting shot manifest: %w", err)
	}
	return model, nil
}

// FindBySource returns the manifest associated with one storyboard source key.
func (repo *ShotManifestRepository) FindBySource(projectID string, documentID string, sectionID string, shotKey string) (domain.ShotManifestModel, error) {
	var model domain.ShotManifestModel
	err := repo.db.Where(
		"project_id = ? AND document_id = ? AND section_id = ? AND shot_key = ?",
		domain.CleanProjectID(projectID),
		strings.TrimSpace(documentID),
		strings.TrimSpace(sectionID),
		strings.TrimSpace(shotKey),
	).First(&model).Error
	if IsRecordNotFound(err) {
		return domain.ShotManifestModel{}, ErrRecordNotFound
	}
	if err != nil {
		return domain.ShotManifestModel{}, fmt.Errorf("finding shot manifest by source: %w", err)
	}
	return model, nil
}

// ListDocument returns manifests in deterministic execution order.
func (repo *ShotManifestRepository) ListDocument(projectID string, documentID string) ([]domain.ShotManifestModel, error) {
	models := []domain.ShotManifestModel{}
	if err := repo.db.Where(
		"project_id = ? AND document_id = ?",
		domain.CleanProjectID(projectID),
		strings.TrimSpace(documentID),
	).Order("sequence ASC, section_id ASC, shot_key ASC, id ASC").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("listing shot manifests: %w", err)
	}
	return models, nil
}

// UpdateCompiledPrompt stores deterministic compiler output without changing the authored shot version.
func (repo *ShotManifestRepository) UpdateCompiledPrompt(projectID string, id string, prompt string) (bool, error) {
	result := repo.db.Model(&domain.ShotManifestModel{}).
		Where("project_id = ? AND id = ?", domain.CleanProjectID(projectID), strings.TrimSpace(id)).
		Update("compiled_prompt", prompt)
	if result.Error != nil {
		return false, fmt.Errorf("updating shot compiled prompt: %w", result.Error)
	}
	return result.RowsAffected > 0, nil
}

// Upsert persists a manifest by stable source key.
func (repo *ShotManifestRepository) Upsert(model domain.ShotManifestModel) error {
	model.ProjectID = domain.CleanProjectID(model.ProjectID)
	model.ID = strings.TrimSpace(model.ID)
	model.DocumentID = strings.TrimSpace(model.DocumentID)
	model.SectionID = strings.TrimSpace(model.SectionID)
	model.ShotKey = strings.TrimSpace(model.ShotKey)
	model.StyleProfileID = strings.TrimSpace(model.StyleProfileID)
	model.SourceHash = strings.TrimSpace(model.SourceHash)
	model.Status = strings.TrimSpace(model.Status)
	if model.BindingsJSON == "" {
		model.BindingsJSON = "{}"
	}
	if model.StateChangesJSON == "" {
		model.StateChangesJSON = "{}"
	}
	if model.ResolvedStateJSON == "" {
		model.ResolvedStateJSON = "{}"
	}
	if model.Version <= 0 {
		model.Version = 1
	}
	if model.Status == "" {
		model.Status = "draft"
	}
	if err := repo.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "project_id"},
			{Name: "document_id"},
			{Name: "section_id"},
			{Name: "shot_key"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"sequence",
			"inherits_from_id",
			"action_text",
			"camera_text",
			"audio_text",
			"style_profile_id",
			"bindings_json",
			"state_changes_json",
			"resolved_state_json",
			"compiled_prompt",
			"source_hash",
			"version",
			"status",
			"updated_at",
		}),
	}).Create(&model).Error; err != nil {
		return fmt.Errorf("upserting shot manifest: %w", err)
	}
	return nil
}
