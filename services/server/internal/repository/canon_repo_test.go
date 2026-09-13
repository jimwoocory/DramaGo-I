package repository

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/mediago-dev/mediago-drama/services/server/internal/domain"
	"github.com/mediago-dev/mediago-drama/services/server/internal/testutil"
)

func TestCanonRepositoryLifecycle(t *testing.T) {
	repo, err := NewCanonRepository(filepath.Join(t.TempDir(), "workspace.db"))
	if err != nil {
		t.Fatalf("NewCanonRepository() error = %v", err)
	}
	testutil.CloseDB(t, repo.db)
	seedRepositoryProject(t, repo.db, "project-a")

	core := domain.CanonAssetModel{
		ID:               "canon-character-lintong",
		ProjectID:        "project-a",
		ResourceType:     "character",
		ResourceID:       "section-lintong",
		SourceDocumentID: "characters",
		Name:             "林书彤",
		SpecJSON:         `{"age":23,"hair":"black"}`,
		PromptText:       "23岁中国女性，黑色长发，鹅蛋脸",
		Status:           "approved",
		Version:          1,
		SourceHash:       "hash-core-v1",
	}
	if err := repo.CreateCanonAsset(core); err != nil {
		t.Fatalf("CreateCanonAsset(core) error = %v", err)
	}

	found, err := repo.FindCanonCoreBySource("project-a", "character", "characters", "section-lintong")
	if err != nil {
		t.Fatalf("FindCanonCoreBySource() error = %v", err)
	}
	if found.ID != core.ID || found.Status != "approved" {
		t.Fatalf("found core = %+v, want %s approved", found, core.ID)
	}

	parentID := core.ID
	variant := domain.CanonAssetModel{
		ID:               "canon-character-lintong-look-wedding-v1",
		ProjectID:        "project-a",
		ResourceType:     "character",
		ResourceID:       "section-lintong",
		SourceDocumentID: "characters",
		ParentID:         &parentID,
		VariantKind:      "look",
		Name:             "婚礼造型",
		SpecJSON:         `{"wardrobe":"white wedding dress","hair":"updo"}`,
		PromptText:       "白色婚纱，盘发",
		Status:           "locked",
		Version:          1,
	}
	if err := repo.CreateCanonAsset(variant); err != nil {
		t.Fatalf("CreateCanonAsset(variant) error = %v", err)
	}

	variants, err := repo.ListCanonVariants("project-a", core.ID)
	if err != nil {
		t.Fatalf("ListCanonVariants() error = %v", err)
	}
	if len(variants) != 1 || variants[0].ID != variant.ID || variants[0].VariantKind != "look" {
		t.Fatalf("variants = %+v, want wedding look", variants)
	}

	asset := domain.AssetModel{
		ID:        "asset-lintong-master",
		ProjectID: domain.StringPtr("project-a"),
		Kind:      "image",
		Filename:  "lintong-master.png",
		MIMEType:  "image/png",
		SizeBytes: 123,
		RelPath:   "projects/project-a/library/lintong-master.png",
		Source:    "generated",
	}
	if err := repo.db.Create(&asset).Error; err != nil {
		t.Fatalf("creating physical asset: %v", err)
	}

	reference := domain.CanonReferenceModel{
		ID:           "canon-ref-lintong-master",
		ProjectID:    "project-a",
		CanonAssetID: core.ID,
		AssetID:      asset.ID,
		Role:         "identity",
		Priority:     100,
		Locked:       true,
	}
	if err := repo.CreateCanonReference(reference); err != nil {
		t.Fatalf("CreateCanonReference() error = %v", err)
	}

	references, err := repo.ListCanonReferences("project-a", core.ID)
	if err != nil {
		t.Fatalf("ListCanonReferences() error = %v", err)
	}
	if len(references) != 1 || references[0].Asset.ID != asset.ID || references[0].Role != "identity" || !references[0].Locked {
		t.Fatalf("references = %+v, want locked identity reference", references)
	}

	updated, err := repo.UpdateCanonAsset("project-a", core.ID, map[string]any{
		"prompt_text": "23岁中国女性，黑色长发，鹅蛋脸，固定身份锚点",
		"version":     2,
	})
	if err != nil {
		t.Fatalf("UpdateCanonAsset() error = %v", err)
	}
	if !updated {
		t.Fatal("UpdateCanonAsset() updated = false, want true")
	}
	got, err := repo.GetCanonAsset("project-a", core.ID)
	if err != nil {
		t.Fatalf("GetCanonAsset() error = %v", err)
	}
	if got.Version != 2 {
		t.Fatalf("GetCanonAsset().Version = %d, want 2", got.Version)
	}

	deleted, err := repo.DeleteCanonReference("project-a", reference.ID)
	if err != nil {
		t.Fatalf("DeleteCanonReference() error = %v", err)
	}
	if !deleted {
		t.Fatal("DeleteCanonReference() deleted = false, want true")
	}
	if _, err := repo.GetCanonAsset("project-a", "missing"); !errors.Is(err, ErrRecordNotFound) {
		t.Fatalf("GetCanonAsset(missing) error = %v, want ErrRecordNotFound", err)
	}
}

func TestWorkspaceSchemaAutoMigratesCanonTables(t *testing.T) {
	db, err := OpenWorkspaceDB(filepath.Join(t.TempDir(), "workspace.db"))
	if err != nil {
		t.Fatalf("OpenWorkspaceDB() error = %v", err)
	}
	testutil.CloseDB(t, db)

	for _, table := range []string{"canon_assets", "canon_references"} {
		if !db.Migrator().HasTable(table) {
			t.Fatalf("workspace schema missing table %q", table)
		}
	}
}

func TestWorkspaceSchemaAddsCanonWithoutDestroyingExistingProjectData(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "legacy-workspace.db")
	db, err := OpenGormSQLite(dbPath)
	if err != nil {
		t.Fatalf("OpenGormSQLite() error = %v", err)
	}
	testutil.CloseDB(t, db)

	// Simulate an existing workspace created before Canon tables existed.
	if err := db.AutoMigrate(
		&domain.WorkspaceProjectModel{},
		&domain.AssetModel{},
		&domain.ProjectSelectedAssetModel{},
	); err != nil {
		t.Fatalf("creating legacy workspace schema: %v", err)
	}
	project := domain.WorkspaceProjectModel{
		ID:          "project-existing",
		Name:        "Existing Project",
		Category:    "drama",
		Status:      "active",
		RelativeDir: "project-existing",
	}
	if err := db.Create(&project).Error; err != nil {
		t.Fatalf("creating legacy project: %v", err)
	}
	asset := domain.AssetModel{
		ID:        "asset-existing",
		ProjectID: domain.StringPtr(project.ID),
		Kind:      "image",
		Filename:  "existing.png",
		MIMEType:  "image/png",
		SizeBytes: 9,
		RelPath:   "projects/project-existing/library/existing.png",
		Source:    "generated",
	}
	if err := db.Create(&asset).Error; err != nil {
		t.Fatalf("creating legacy asset: %v", err)
	}

	if err := EnsureWorkspaceSchema(db); err != nil {
		t.Fatalf("EnsureWorkspaceSchema() error = %v", err)
	}
	for _, table := range []string{"canon_assets", "canon_references"} {
		if !db.Migrator().HasTable(table) {
			t.Fatalf("migrated workspace missing table %q", table)
		}
	}
	var gotProject domain.WorkspaceProjectModel
	if err := db.First(&gotProject, "id = ?", project.ID).Error; err != nil {
		t.Fatalf("existing project lost during migration: %v", err)
	}
	var gotAsset domain.AssetModel
	if err := db.First(&gotAsset, "id = ?", asset.ID).Error; err != nil {
		t.Fatalf("existing media asset lost during migration: %v", err)
	}
	if gotProject.Name != project.Name || gotAsset.Filename != asset.Filename {
		t.Fatalf("existing data changed: project=%+v asset=%+v", gotProject, gotAsset)
	}
}
