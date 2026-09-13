package domain

import "time"

// CanonAssetModel stores the authoritative structured identity for one creative
// resource. Core resources have ParentID=nil. Variants point at a core/parent
// CanonAssetModel through ParentID and describe only the mutable appearance or
// state delta for that variant.
type CanonAssetModel struct {
	ID               string    `gorm:"column:id;primaryKey"`
	ProjectID        string    `gorm:"column:project_id;not null;index:canon_assets_project_type_idx,priority:1;index:canon_assets_source_idx,priority:1"`
	ResourceType     string    `gorm:"column:resource_type;not null;index:canon_assets_project_type_idx,priority:2;index:canon_assets_source_idx,priority:2"`
	ResourceID       string    `gorm:"column:resource_id;not null;default:'';index:canon_assets_source_idx,priority:4"`
	SourceDocumentID string    `gorm:"column:source_document_id;not null;default:'';index:canon_assets_source_idx,priority:3"`
	ParentID         *string   `gorm:"column:parent_id;index:canon_assets_parent_idx"`
	VariantKind      string    `gorm:"column:variant_kind;not null;default:'';index:canon_assets_variant_kind_idx"`
	Name             string    `gorm:"column:name;not null;default:''"`
	SpecJSON         string    `gorm:"column:spec_json;not null;type:text;default:'{}'"`
	PromptText       string    `gorm:"column:prompt_text;not null;type:text;default:''"`
	Status           string    `gorm:"column:status;not null;default:'draft';index:canon_assets_status_idx"`
	Version          int       `gorm:"column:version;not null;default:1"`
	SourceHash       string    `gorm:"column:source_hash;not null;default:'';index:canon_assets_source_hash_idx"`
	CreatedAt        time.Time `gorm:"column:created_at;not null;autoCreateTime:nano"`
	UpdatedAt        time.Time `gorm:"column:updated_at;not null;autoUpdateTime:nano"`

	Project WorkspaceProjectModel `gorm:"foreignKey:ProjectID;references:ID;constraint:OnDelete:CASCADE"`
	Parent  *CanonAssetModel      `gorm:"foreignKey:ParentID;references:ID;constraint:OnDelete:CASCADE"`
}

// TableName returns the backing table name.
func (CanonAssetModel) TableName() string {
	return "canon_assets"
}

// CanonReferenceModel binds one Canon core/variant to an existing physical
// media asset. No media bytes are duplicated; AssetID references assets.id.
type CanonReferenceModel struct {
	ID           string    `gorm:"column:id;primaryKey"`
	ProjectID    string    `gorm:"column:project_id;not null;index:canon_references_project_idx"`
	CanonAssetID string    `gorm:"column:canon_asset_id;not null;index:canon_references_canon_idx"`
	AssetID      string    `gorm:"column:asset_id;not null;index:canon_references_asset_idx"`
	Role         string    `gorm:"column:role;not null;default:'custom';index:canon_references_role_idx"`
	Priority     int       `gorm:"column:priority;not null;default:0"`
	Locked       bool      `gorm:"column:locked;not null;default:false"`
	CreatedAt    time.Time `gorm:"column:created_at;not null;autoCreateTime:nano"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null;autoUpdateTime:nano"`

	Project    WorkspaceProjectModel `gorm:"foreignKey:ProjectID;references:ID;constraint:OnDelete:CASCADE"`
	CanonAsset CanonAssetModel       `gorm:"foreignKey:CanonAssetID;references:ID;constraint:OnDelete:CASCADE"`
	Asset      AssetModel            `gorm:"foreignKey:AssetID;references:ID;constraint:OnDelete:CASCADE"`
}

// TableName returns the backing table name.
func (CanonReferenceModel) TableName() string {
	return "canon_references"
}
