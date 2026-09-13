package domain

import "time"

// ShotManifestModel is the structured execution contract for one storyboard
// shot/group. It binds Canon resources and persists both explicit state changes
// and the fully resolved continuity state used for deterministic generation.
type ShotManifestModel struct {
	ID                string    `gorm:"column:id;primaryKey"`
	ProjectID         string    `gorm:"column:project_id;not null;index:shot_manifests_project_sequence_idx,priority:1;uniqueIndex:shot_manifests_source_key_idx,priority:1"`
	DocumentID        string    `gorm:"column:document_id;not null;index:shot_manifests_document_idx;uniqueIndex:shot_manifests_source_key_idx,priority:2"`
	SectionID         string    `gorm:"column:section_id;not null;index:shot_manifests_section_idx;uniqueIndex:shot_manifests_source_key_idx,priority:3"`
	ShotKey           string    `gorm:"column:shot_key;not null;default:'';uniqueIndex:shot_manifests_source_key_idx,priority:4"`
	Sequence          int       `gorm:"column:sequence;not null;default:0;index:shot_manifests_project_sequence_idx,priority:2"`
	StartSeconds      float64   `gorm:"column:start_seconds;not null;default:0"`
	EndSeconds        float64   `gorm:"column:end_seconds;not null;default:0"`
	DurationSeconds   float64   `gorm:"column:duration_seconds;not null;default:0"`
	InheritsFromID    *string   `gorm:"column:inherits_from_id;index:shot_manifests_inherits_idx"`
	ActionText        string    `gorm:"column:action_text;not null;type:text;default:''"`
	CameraText        string    `gorm:"column:camera_text;not null;type:text;default:''"`
	AudioText         string    `gorm:"column:audio_text;not null;type:text;default:''"`
	StyleProfileID    string    `gorm:"column:style_profile_id;not null;default:''"`
	BindingsJSON      string    `gorm:"column:bindings_json;not null;type:text;default:'{}'"`
	StateChangesJSON  string    `gorm:"column:state_changes_json;not null;type:text;default:'{}'"`
	ResolvedStateJSON string    `gorm:"column:resolved_state_json;not null;type:text;default:'{}'"`
	CompiledPrompt    string    `gorm:"column:compiled_prompt;not null;type:text;default:''"`
	SourceHash        string    `gorm:"column:source_hash;not null;default:'';index:shot_manifests_source_hash_idx"`
	Version           int       `gorm:"column:version;not null;default:1"`
	Status            string    `gorm:"column:status;not null;default:'draft';index:shot_manifests_status_idx"`
	CreatedAt         time.Time `gorm:"column:created_at;not null;autoCreateTime:nano"`
	UpdatedAt         time.Time `gorm:"column:updated_at;not null;autoUpdateTime:nano"`

	Project      WorkspaceProjectModel `gorm:"foreignKey:ProjectID;references:ID;constraint:OnDelete:CASCADE"`
	InheritsFrom *ShotManifestModel    `gorm:"foreignKey:InheritsFromID;references:ID;constraint:OnDelete:SET NULL"`
}

// TableName returns the backing table name.
func (ShotManifestModel) TableName() string {
	return "shot_manifests"
}
