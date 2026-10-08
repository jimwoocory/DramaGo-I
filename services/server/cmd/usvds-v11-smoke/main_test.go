package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mediago-dev/mediago-drama/services/server/internal/domain"
	"github.com/mediago-dev/mediago-drama/services/server/internal/repository"
	serviceshared "github.com/mediago-dev/mediago-drama/services/server/internal/service/shared"
)

func TestSmokeRunnerUsesSnapshotWithoutMutatingSource(t *testing.T) {
	sourceRoot := t.TempDir()
	projectDir := filepath.Join(t.TempDir(), "real-project")
	if err := os.MkdirAll(filepath.Join(projectDir, "work"), 0o755); err != nil {
		t.Fatalf("mkdir work: %v", err)
	}
	storyPath := filepath.Join(projectDir, "work", "story.md")
	story := strings.Join([]string{
		"---",
		"id: story-1",
		"title: Story Package",
		"category: reference",
		"tags:",
		"  - usvds:artifact:story-package",
		"version: 1",
		"---",
		"",
		"# Story Package",
		"",
		"真实项目副本 smoke。",
		"",
	}, "\n")
	if err := os.WriteFile(storyPath, []byte(story), 0o644); err != nil {
		t.Fatalf("write story: %v", err)
	}

	if err := serviceshared.EnsureWorkspaceLayout(sourceRoot); err != nil {
		t.Fatalf("EnsureWorkspaceLayout() error = %v", err)
	}
	repos, err := repository.OpenWorkspaceRepositories(serviceshared.WorkspacePathsFor(sourceRoot).DatabasePath())
	if err != nil {
		t.Fatalf("OpenWorkspaceRepositories() error = %v", err)
	}
	if err := repos.DB.Create(&domain.WorkspaceProjectModel{
		ID:          "project-real",
		Name:        "Real Project",
		Category:    "drama",
		Status:      "active",
		ProjectDir:  projectDir,
		RelativeDir: "external/project-real",
	}).Error; err != nil {
		t.Fatalf("create project: %v", err)
	}
	closeGorm(repos.DB)

	before, err := os.ReadFile(storyPath)
	if err != nil {
		t.Fatalf("read source before smoke: %v", err)
	}
	report, _, err := run(options{
		workspace:  sourceRoot,
		projectID:  "project-real",
		jsonOutput: true,
	})
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	after, err := os.ReadFile(storyPath)
	if err != nil {
		t.Fatalf("read source after smoke: %v", err)
	}
	if string(before) != string(after) {
		t.Fatal("source project markdown changed during smoke")
	}
	if !report.OriginalWorkspaceSafe {
		t.Fatal("OriginalWorkspaceSafe = false")
	}
	if report.ProductionEvidenceReady || len(report.ProductionEvidenceBlockers) == 0 {
		t.Fatalf("production evidence = %v blockers = %v, want incomplete real-data fixture", report.ProductionEvidenceReady, report.ProductionEvidenceBlockers)
	}
	if report.Project.ID != "project-real" {
		t.Fatalf("project = %+v", report.Project)
	}
	if report.DocumentCount == 0 {
		t.Fatal("document count = 0, want copied project document")
	}
	if report.Gates.Summary.StoryDocumentID != "story-1" {
		t.Fatalf("story document = %q, want story-1", report.Gates.Summary.StoryDocumentID)
	}
	if report.Gates.NextWorkflow == nil || report.Gates.NextWorkflow.ID != "review" {
		t.Fatalf("next workflow = %+v, want review for unapproved copied story", report.Gates.NextWorkflow)
	}
	if _, err := os.Stat(report.SnapshotWorkspace); !os.IsNotExist(err) {
		t.Fatalf("temporary snapshot still exists after run: %s", report.SnapshotWorkspace)
	}
}

func TestSmokeRunnerExternalScreenplayOverlayNeverGrantsApproval(t *testing.T) {
	sourceRoot := t.TempDir()
	projectDir := filepath.Join(t.TempDir(), "project")
	if err := os.MkdirAll(filepath.Join(projectDir, "work"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := serviceshared.EnsureWorkspaceLayout(sourceRoot); err != nil {
		t.Fatal(err)
	}
	repos, err := repository.OpenWorkspaceRepositories(serviceshared.WorkspacePathsFor(sourceRoot).DatabasePath())
	if err != nil {
		t.Fatal(err)
	}
	if err := repos.DB.Create(&domain.WorkspaceProjectModel{
		ID: "project-overlay", Name: "External Screenplay Smoke", Category: "drama",
		Status: "active", ProjectDir: projectDir, RelativeDir: "external/project-overlay",
	}).Error; err != nil {
		t.Fatal(err)
	}
	closeGorm(repos.DB)

	overlay := filepath.Join(t.TempDir(), "screenplay.md")
	body := "---\nid: script-external\ntitle: External draft\ncategory: screenplay\ntags:\n  - usvds:artifact:screenplay\nversion: 1\n---\n\n# External draft\n\nCreator authorized draft, NOT APPROVED.\n"
	if err := os.WriteFile(overlay, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	report, _, err := run(options{workspace: sourceRoot, projectID: "project-overlay", overlayScreenplay: overlay})
	if err != nil {
		t.Fatalf("run overlay: %v", err)
	}
	if report.DocumentCount != 1 || report.Gates.Summary.ScreenplayDocumentID != "script-external" {
		t.Fatalf("document count=%d summary=%+v", report.DocumentCount, report.Gates.Summary)
	}
	if report.ExternalScreenplayOverlay == nil || !report.ExternalScreenplayOverlay.ApprovalProhibited {
		t.Fatalf("overlay provenance missing: %+v", report.ExternalScreenplayOverlay)
	}
	if report.GenerationReady || report.ProductionEvidenceReady || len(report.ProductionEvidenceBlockers) == 0 {
		t.Fatalf("source-only smoke must remain blocked: %+v", report)
	}
	for _, state := range report.Gates.Approvals {
		if state.Approved {
			t.Fatalf("smoke overlay unexpectedly approved gate: %+v", state)
		}
	}
	if _, err := os.Stat(report.SnapshotWorkspace); !os.IsNotExist(err) {
		t.Fatalf("smoke snapshot not removed: %s", report.SnapshotWorkspace)
	}
	if entries, err := os.ReadDir(filepath.Join(projectDir, "work")); err != nil || len(entries) != 0 {
		t.Fatalf("source project was written: entries=%v error=%v", entries, err)
	}
}

func TestUnapprovedScreenplayOverlayRejectsSpoofedApprovalAndInvalidSource(t *testing.T) {
	safe := "---\nid: source-one\ncategory: screenplay\ntags:\n  - usvds:artifact:screenplay\nversion: 1\n---\n\n# Draft\n"
	if err := validateUnapprovedScreenplayOverlay([]byte(safe)); err != nil {
		t.Fatalf("valid unapproved screenplay rejected: %v", err)
	}
	for _, sample := range []string{
		strings.Replace(safe, "  - usvds:artifact:screenplay", "  - usvds:artifact:screenplay\n  - usvds:approval:screenplay_reviewed:v2:sha256:fake", 1),
		strings.Replace(safe, "category: screenplay", "category: storyboard", 1),
		strings.Replace(safe, "  - usvds:artifact:screenplay", "  - unrelated", 1),
	} {
		if err := validateUnapprovedScreenplayOverlay([]byte(sample)); err == nil {
			t.Fatalf("unsafe overlay accepted: %q", sample)
		}
	}
}

func TestProductionEvidenceRequiresPopulatedShotPreflight(t *testing.T) {
	if blockers := productionEvidenceBlockers(0, 0, 0, 0); len(blockers) != 4 {
		t.Fatalf("empty project blockers = %v, want all 4 prerequisites", blockers)
	}
	if blockers := productionEvidenceBlockers(2, 0, 0, 0); len(blockers) != 3 {
		t.Fatalf("story-only project blockers = %v, want 3 production prerequisites", blockers)
	}
	if blockers := productionEvidenceBlockers(3, 1, 1, 1); len(blockers) != 0 {
		t.Fatalf("ready project blockers = %v, want none", blockers)
	}
}

func TestSelectProjectPrefersNewestActiveAccessibleProject(t *testing.T) {
	activeOld := t.TempDir()
	activeNew := t.TempDir()
	archived := t.TempDir()
	projects := []domain.WorkspaceProjectModel{
		{ID: "archived", Status: "archived", ProjectDir: archived},
		{ID: "active-old", Status: "active", ProjectDir: activeOld},
		{ID: "active-new", Status: "active", ProjectDir: activeNew},
	}
	projects[0].UpdatedAt = projects[0].UpdatedAt.Add(3)
	projects[1].UpdatedAt = projects[1].UpdatedAt.Add(1)
	projects[2].UpdatedAt = projects[2].UpdatedAt.Add(2)

	selected, err := selectProject(projects, "")
	if err != nil {
		t.Fatalf("selectProject() error = %v", err)
	}
	if selected.ID != "active-new" {
		t.Fatalf("selected = %q, want active-new", selected.ID)
	}
}
