package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mediago-dev/mediago-drama/services/server/internal/domain"
	"github.com/mediago-dev/mediago-drama/services/server/internal/repository"
	"github.com/mediago-dev/mediago-drama/services/server/internal/service/shared"
)

func TestRebaseWorkspaceRelativeProjectDirsUsesCurrentWorkspace(t *testing.T) {
	workspaceDir := t.TempDir()
	paths := shared.WorkspacePathsFor(workspaceDir)
	if err := os.MkdirAll(filepath.Dir(paths.DatabasePath()), 0o755); err != nil {
		t.Fatalf("creating workspace database dir: %v", err)
	}
	db, err := repository.OpenWorkspaceDB(paths.DatabasePath())
	if err != nil {
		t.Fatalf("OpenWorkspaceDB() error = %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB() error = %v", err)
	}
	defer sqlDB.Close()
	repo := repository.NewWorkspaceRepository(db)
	repos := WorkspaceStateRepositories{Workspace: repo}
	now := domain.TimeFromString("2026-09-11T00:00:00Z")

	internalID := "project-relocated"
	internalRelative := filepath.ToSlash(filepath.Join("projects", internalID))
	currentInternalDir := filepath.Join(workspaceDir, filepath.FromSlash(internalRelative))
	if err := os.MkdirAll(currentInternalDir, 0o755); err != nil {
		t.Fatalf("creating relocated internal project: %v", err)
	}
	staleWorkspace := t.TempDir()
	staleInternalDir := filepath.Join(staleWorkspace, "projects", internalID)
	if err := os.MkdirAll(staleInternalDir, 0o755); err != nil {
		t.Fatalf("creating stale internal project: %v", err)
	}
	if err := repo.UpsertProject(domain.WorkspaceProjectModel{
		ID:          internalID,
		Name:        "Relocated",
		Category:    "agent",
		Status:      repository.ProjectStatusActive,
		ProjectDir:  staleInternalDir,
		RelativeDir: internalRelative,
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		t.Fatalf("UpsertProject(internal) error = %v", err)
	}

	externalID := "project-external"
	externalDir := filepath.Join(t.TempDir(), "external-project")
	if err := os.MkdirAll(externalDir, 0o755); err != nil {
		t.Fatalf("creating external project: %v", err)
	}
	if err := repo.UpsertProject(domain.WorkspaceProjectModel{
		ID:          externalID,
		Name:        "External",
		Category:    "agent",
		Status:      repository.ProjectStatusActive,
		ProjectDir:  externalDir,
		RelativeDir: shared.DisplayProjectDir(workspaceDir, externalDir),
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		t.Fatalf("UpsertProject(external) error = %v", err)
	}

	rebased, err := rebaseWorkspaceRelativeProjectDirs(workspaceDir, repos)
	if err != nil {
		t.Fatalf("rebaseWorkspaceRelativeProjectDirs() error = %v", err)
	}
	if rebased != 1 {
		t.Fatalf("rebased = %d, want 1", rebased)
	}
	internal, err := repo.GetProject(internalID)
	if err != nil {
		t.Fatalf("GetProject(internal) error = %v", err)
	}
	if !sameWorkspacePath(internal.ProjectDir, currentInternalDir) || internal.RelativeDir != internalRelative {
		t.Fatalf("internal project location = (%q,%q), want (%q,%q)", internal.ProjectDir, internal.RelativeDir, currentInternalDir, internalRelative)
	}
	external, err := repo.GetProject(externalID)
	if err != nil {
		t.Fatalf("GetProject(external) error = %v", err)
	}
	if !sameWorkspacePath(external.ProjectDir, externalDir) {
		t.Fatalf("external project dir = %q, want %q", external.ProjectDir, externalDir)
	}

	rebased, err = rebaseWorkspaceRelativeProjectDirs(workspaceDir, repos)
	if err != nil {
		t.Fatalf("second rebase error = %v", err)
	}
	if rebased != 0 {
		t.Fatalf("second rebased = %d, want 0", rebased)
	}
}
