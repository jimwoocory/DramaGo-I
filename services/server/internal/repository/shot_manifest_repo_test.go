package repository

import (
	"path/filepath"
	"testing"

	"github.com/mediago-dev/mediago-drama/services/server/internal/domain"
	"github.com/mediago-dev/mediago-drama/services/server/internal/testutil"
)

func TestShotManifestRepositoryLifecycle(t *testing.T) {
	db, err := OpenWorkspaceDB(filepath.Join(t.TempDir(), "workspace.db"))
	if err != nil {
		t.Fatalf("OpenWorkspaceDB() error = %v", err)
	}
	testutil.CloseDB(t, db)
	seedRepositoryProject(t, db, "project-a")
	repo := NewShotManifestRepositoryFromDB(db)

	first := domain.ShotManifestModel{
		ID:                "shot-001",
		ProjectID:         "project-a",
		DocumentID:        "storyboard-ep1",
		SectionID:         "section-shot-1",
		ShotKey:           "shot-1",
		Sequence:          1,
		ActionText:        "韩三河抬手拉开外套。",
		CameraText:        "中景缓慢推进",
		AudioText:         "无台词",
		BindingsJSON:      `{"characters":[{"canonId":"char-han","variantId":"look-default"}]}`,
		StateChangesJSON:  `{"characters":{"char-han":{"look":{"outerwear":"half_removed"}}}}`,
		ResolvedStateJSON: `{"characters":{"char-han":{"look":{"outerwear":"half_removed"}}}}`,
		SourceHash:        "hash-v1",
		Version:           1,
		Status:            "draft",
	}
	if err := repo.Upsert(first); err != nil {
		t.Fatalf("Upsert(first) error = %v", err)
	}

	got, err := repo.Get("project-a", first.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Sequence != 1 || got.ActionText != first.ActionText || got.ResolvedStateJSON != first.ResolvedStateJSON {
		t.Fatalf("got = %+v, want first manifest", got)
	}

	second := domain.ShotManifestModel{
		ID:                "shot-002",
		ProjectID:         "project-a",
		DocumentID:        "storyboard-ep1",
		SectionID:         "section-shot-2",
		ShotKey:           "shot-2",
		Sequence:          2,
		InheritsFromID:    domain.StringPtr(first.ID),
		ActionText:        "韩三河继续向前走。",
		BindingsJSON:      first.BindingsJSON,
		StateChangesJSON:  `{}`,
		ResolvedStateJSON: first.ResolvedStateJSON,
		SourceHash:        "hash-v2",
		Version:           1,
		Status:            "draft",
	}
	if err := repo.Upsert(second); err != nil {
		t.Fatalf("Upsert(second) error = %v", err)
	}

	listed, err := repo.ListDocument("project-a", "storyboard-ep1")
	if err != nil {
		t.Fatalf("ListDocument() error = %v", err)
	}
	if len(listed) != 2 || listed[0].ID != first.ID || listed[1].ID != second.ID {
		t.Fatalf("ListDocument() = %+v, want ordered first/second", listed)
	}

	first.ActionText = "韩三河把外套拉到肩下。"
	first.Version = 2
	first.SourceHash = "hash-v1-updated"
	if err := repo.Upsert(first); err != nil {
		t.Fatalf("Upsert(updated first) error = %v", err)
	}
	updated, err := repo.FindBySource("project-a", "storyboard-ep1", "section-shot-1", "shot-1")
	if err != nil {
		t.Fatalf("FindBySource() error = %v", err)
	}
	if updated.ID != first.ID || updated.Version != 2 || updated.ActionText != first.ActionText {
		t.Fatalf("updated = %+v, want v2 action", updated)
	}
}

func TestWorkspaceSchemaAutoMigratesShotManifestTable(t *testing.T) {
	db, err := OpenWorkspaceDB(filepath.Join(t.TempDir(), "workspace.db"))
	if err != nil {
		t.Fatalf("OpenWorkspaceDB() error = %v", err)
	}
	testutil.CloseDB(t, db)
	if !db.Migrator().HasTable("shot_manifests") {
		t.Fatal("workspace schema missing shot_manifests table")
	}
}
