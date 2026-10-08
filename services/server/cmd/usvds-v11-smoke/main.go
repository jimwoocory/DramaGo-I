package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	mediamcp "github.com/mediago-dev/mediago-drama/packages/mcp/pkg/mcp"
	appworkspace "github.com/mediago-dev/mediago-drama/services/server/internal/app/workspace"
	"github.com/mediago-dev/mediago-drama/services/server/internal/domain"
	"github.com/mediago-dev/mediago-drama/services/server/internal/repository"
	servicecanon "github.com/mediago-dev/mediago-drama/services/server/internal/service/canon"
	servicegeneration "github.com/mediago-dev/mediago-drama/services/server/internal/service/generation"
	serviceshared "github.com/mediago-dev/mediago-drama/services/server/internal/service/shared"
	serviceshotmanifest "github.com/mediago-dev/mediago-drama/services/server/internal/service/shotmanifest"
	serviceusvdsv11 "github.com/mediago-dev/mediago-drama/services/server/internal/service/usvdsv11"
)

type options struct {
	workspace              string
	projectID              string
	listOnly               bool
	keepSnapshot           bool
	requireGenerationReady bool
	jsonOutput             bool
}

type projectSummary struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	ProjectDir string `json:"projectDir"`
	UpdatedAt  string `json:"updatedAt"`
}

type smokeReport struct {
	SourceWorkspace       string                                 `json:"sourceWorkspace"`
	SnapshotWorkspace     string                                 `json:"snapshotWorkspace"`
	Project               projectSummary                         `json:"project"`
	DocumentCount         int                                    `json:"documentCount"`
	CanonSync             servicecanon.ProjectSyncSummary        `json:"canonSync"`
	ShotManifestSync      serviceshotmanifest.ProjectSyncSummary `json:"shotManifestSync"`
	Gates                 serviceusvdsv11.ProjectGateReport      `json:"gates"`
	StoryboardDocuments   int                                    `json:"storyboardDocuments"`
	ShotCount             int                                    `json:"shotCount"`
	ReadyShotCount        int                                    `json:"readyShotCount"`
	PreflightPassCount    int                                    `json:"preflightPassCount"`
	PreflightFailures     []string                               `json:"preflightFailures"`
	GenerationReady       bool                                   `json:"generationReady"`
	OriginalWorkspaceSafe bool                                   `json:"originalWorkspaceSafe"`
}

func main() {
	opts := parseFlags()
	report, projects, err := run(opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, "USVDS V11 smoke failed:", err)
		os.Exit(1)
	}
	if opts.listOnly {
		writeJSON(projects)
		return
	}
	if opts.jsonOutput {
		writeJSON(report)
	} else {
		printReport(report)
	}
	if opts.requireGenerationReady && !report.GenerationReady {
		os.Exit(2)
	}
}

func parseFlags() options {
	workspace := flag.String("workspace", "", "DramaGo workspace root; defaults to the normal app workspace")
	projectID := flag.String("project", "", "Project ID to smoke; defaults to the newest active project with a valid project directory")
	listOnly := flag.Bool("list", false, "List projects from a copied workspace DB and exit")
	keepSnapshot := flag.Bool("keep-snapshot", false, "Keep the temporary smoke snapshot instead of deleting it")
	requireGenerationReady := flag.Bool("require-generation-ready", false, "Exit 2 unless the selected project reaches generation_ready")
	jsonOutput := flag.Bool("json", false, "Emit JSON")
	flag.Parse()
	return options{
		workspace:              *workspace,
		projectID:              strings.TrimSpace(*projectID),
		listOnly:               *listOnly,
		keepSnapshot:           *keepSnapshot,
		requireGenerationReady: *requireGenerationReady,
		jsonOutput:             *jsonOutput,
	}
}

func run(opts options) (smokeReport, []projectSummary, error) {
	sourceRoot := serviceshared.ResolveWorkspaceDir(opts.workspace)
	sourceDB := serviceshared.WorkspacePathsFor(sourceRoot).DatabasePath()
	if info, err := os.Stat(sourceDB); err != nil || info.IsDir() {
		if err == nil {
			err = fmt.Errorf("workspace database is a directory")
		}
		return smokeReport{}, nil, fmt.Errorf("workspace database %s is unavailable: %w", sourceDB, err)
	}

	snapshotRoot, err := os.MkdirTemp("", "dramago-v11-smoke-*")
	if err != nil {
		return smokeReport{}, nil, err
	}
	if !opts.keepSnapshot {
		defer os.RemoveAll(snapshotRoot)
	}
	if err := copyWorkspaceDatabase(sourceRoot, snapshotRoot); err != nil {
		return smokeReport{}, nil, err
	}

	repos, err := repository.OpenWorkspaceRepositories(serviceshared.WorkspacePathsFor(snapshotRoot).DatabasePath())
	if err != nil {
		return smokeReport{}, nil, err
	}
	defer closeGorm(repos.DB)
	var integrity string
	if err := repos.DB.Raw("PRAGMA integrity_check").Scan(&integrity).Error; err != nil {
		return smokeReport{}, nil, fmt.Errorf("checking snapshot database: %w", err)
	}
	if strings.TrimSpace(strings.ToLower(integrity)) != "ok" {
		return smokeReport{}, nil, fmt.Errorf("snapshot database integrity check = %q", integrity)
	}

	models, err := repos.Workspace.ListProjects()
	if err != nil {
		return smokeReport{}, nil, err
	}
	projects := summarizeProjects(models)
	if opts.listOnly {
		return smokeReport{}, projects, nil
	}

	project, err := selectProject(models, opts.projectID)
	if err != nil {
		return smokeReport{}, projects, err
	}
	sourceProjectDir := strings.TrimSpace(project.ProjectDir)
	if sourceProjectDir == "" {
		return smokeReport{}, projects, fmt.Errorf("project %s has no project_dir", project.ID)
	}
	if info, err := os.Stat(sourceProjectDir); err != nil || !info.IsDir() {
		if err == nil {
			err = fmt.Errorf("not a directory")
		}
		return smokeReport{}, projects, fmt.Errorf("project directory %s is unavailable: %w", sourceProjectDir, err)
	}

	snapshotProjectDir := filepath.Join(snapshotRoot, "projects", project.ID)
	if err := copyProjectForSmoke(sourceProjectDir, snapshotProjectDir); err != nil {
		return smokeReport{}, projects, err
	}
	if err := repos.DB.Model(&domain.WorkspaceProjectModel{}).
		Where("id = ?", project.ID).
		Updates(map[string]any{
			"project_dir":  snapshotProjectDir,
			"relative_dir": filepath.ToSlash(filepath.Join("projects", project.ID)),
		}).Error; err != nil {
		return smokeReport{}, projects, fmt.Errorf("redirecting project in snapshot DB: %w", err)
	}

	state := appworkspace.NewStateService(snapshotRoot)
	defer state.Close()
	if err := state.InitErr(); err != nil {
		return smokeReport{}, projects, err
	}
	documents := state.StateService().Documents

	canon := servicecanon.NewService(repos.Canon, nil)
	canon.SetDocumentResourceProvider(documents)
	generationTasks := servicegeneration.NewGenerationTaskServiceFromRepository(repos.GenerationTasks, nil, nil)
	canon.SetSelectedAssetProvider(repos.GenerationTasks)
	shotManifests := serviceshotmanifest.NewService(repos.ShotManifests, canon, nil)
	shotManifests.SetDocumentResourceProvider(documents)
	shotManifests.SetContinuityAssetProvider(generationTasks)

	canonSync, err := canon.SyncProject(project.ID)
	if err != nil {
		return smokeReport{}, projects, fmt.Errorf("Canon sync: %w", err)
	}
	shotSync, err := shotManifests.SyncProject(project.ID)
	if err != nil {
		return smokeReport{}, projects, fmt.Errorf("ShotManifest sync: %w", err)
	}

	gates := serviceusvdsv11.NewProjectGateService(documents, canon, shotManifests)
	gates.SetGenerationStateResolver(func(projectID string) (serviceusvdsv11.GenerationProjectState, error) {
		tasks, err := generationTasks.ListByProject("", projectID)
		if err != nil {
			return serviceusvdsv11.GenerationProjectState{}, err
		}
		state := serviceusvdsv11.GenerationProjectState{Total: len(tasks)}
		for index, task := range tasks {
			if index == 0 {
				state.LastTaskID = task.ID
				state.LastUpdatedAt = task.UpdatedAt
			}
			switch strings.ToLower(strings.TrimSpace(task.Status)) {
			case "completed", "succeeded", "success":
				state.Completed++
			case "failed", "error", "cancelled", "canceled":
				state.Failed++
			case "pending", "queued", "submitted", "waiting", "unknown", "reconciling":
				state.Pending++
			default:
				state.Running++
			}
		}
		return state, nil
	})

	documentList, err := documents.ListWorkspaceDocuments(project.ID)
	if err != nil {
		return smokeReport{}, projects, err
	}

	preflightFailures := []string{}
	storyboardDocuments := 0
	shotCount := 0
	readyShotCount := 0
	preflightPassCount := 0
	for _, document := range documentList.Documents {
		if !isStoryboardDocument(document) {
			continue
		}
		storyboardDocuments++
		shots, err := shotManifests.ListDocument(project.ID, document.ID)
		if err != nil {
			preflightFailures = append(preflightFailures, fmt.Sprintf("%s: list shots: %v", document.ID, err))
			continue
		}
		for _, shot := range shots {
			shotCount++
			if shot.Status != serviceshotmanifest.StatusReady {
				continue
			}
			readyShotCount++
			compiled, err := shotManifests.CompileAndPersist(project.ID, shot.ID)
			if err != nil {
				preflightFailures = append(preflightFailures, fmt.Sprintf("%s: compile: %v", shot.ID, err))
				continue
			}
			if err := serviceusvdsv11.ValidateGenerationCompileResult(compiled); err != nil {
				preflightFailures = append(preflightFailures, fmt.Sprintf("%s: preflight: %v", shot.ID, err))
				continue
			}
			preflightPassCount++
		}
	}

	gateReport, err := gates.EvaluateProject(project.ID)
	if err != nil {
		return smokeReport{}, projects, err
	}
	generationReady := false
	for _, gate := range gateReport.Gates {
		if gate.Gate == serviceusvdsv11.GateGenerationReady {
			generationReady = gate.Ready
			break
		}
	}

	return smokeReport{
		SourceWorkspace:       sourceRoot,
		SnapshotWorkspace:     snapshotRoot,
		Project:               summarizeProject(project),
		DocumentCount:         len(documentList.Documents),
		CanonSync:             canonSync,
		ShotManifestSync:      shotSync,
		Gates:                 gateReport,
		StoryboardDocuments:   storyboardDocuments,
		ShotCount:             shotCount,
		ReadyShotCount:        readyShotCount,
		PreflightPassCount:    preflightPassCount,
		PreflightFailures:     preflightFailures,
		GenerationReady:       generationReady,
		OriginalWorkspaceSafe: true,
	}, projects, nil
}

func copyWorkspaceDatabase(sourceRoot string, snapshotRoot string) error {
	sourcePaths := serviceshared.WorkspacePathsFor(sourceRoot)
	targetPaths := serviceshared.WorkspacePathsFor(snapshotRoot)
	if err := os.MkdirAll(targetPaths.DatabaseDir(), 0o700); err != nil {
		return err
	}
	for _, suffix := range []string{"", "-wal", "-shm"} {
		source := sourcePaths.DatabasePath() + suffix
		if _, err := os.Stat(source); errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return err
		}
		if err := copyFile(source, targetPaths.DatabasePath()+suffix); err != nil {
			return err
		}
	}
	for _, name := range []string{"media.workspace.json"} {
		source := filepath.Join(sourceRoot, name)
		if _, err := os.Stat(source); err == nil {
			if err := copyFile(source, filepath.Join(snapshotRoot, name)); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyProjectForSmoke(sourceDir string, targetDir string) error {
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return err
	}
	for _, name := range []string{"project.media.json", "agent-history.jsonl"} {
		source := filepath.Join(sourceDir, name)
		if _, err := os.Stat(source); err == nil {
			if err := copyFile(source, filepath.Join(targetDir, name)); err != nil {
				return err
			}
		}
	}
	workDir := filepath.Join(sourceDir, "work")
	if info, err := os.Stat(workDir); err == nil && info.IsDir() {
		if err := copyTree(workDir, filepath.Join(targetDir, "work")); err != nil {
			return err
		}
	}
	return nil
}

func copyTree(source string, target string) error {
	return filepath.WalkDir(source, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		destination := filepath.Join(target, relative)
		if entry.IsDir() {
			return os.MkdirAll(destination, 0o755)
		}
		return copyFile(path, destination)
	})
}

func copyFile(source string, target string) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.Create(target)
	if err != nil {
		return err
	}
	if _, err := io.Copy(output, input); err != nil {
		output.Close()
		return err
	}
	return output.Close()
}

func selectProject(projects []domain.WorkspaceProjectModel, requested string) (domain.WorkspaceProjectModel, error) {
	if requested != "" {
		for _, project := range projects {
			if project.ID == requested {
				return project, nil
			}
		}
		return domain.WorkspaceProjectModel{}, fmt.Errorf("project %q not found", requested)
	}
	candidates := append([]domain.WorkspaceProjectModel(nil), projects...)
	sort.SliceStable(candidates, func(i, j int) bool {
		leftActive := candidates[i].Status == "active"
		rightActive := candidates[j].Status == "active"
		if leftActive != rightActive {
			return leftActive
		}
		return candidates[i].UpdatedAt.After(candidates[j].UpdatedAt)
	})
	for _, project := range candidates {
		if strings.TrimSpace(project.ProjectDir) == "" {
			continue
		}
		if info, err := os.Stat(project.ProjectDir); err == nil && info.IsDir() {
			return project, nil
		}
	}
	return domain.WorkspaceProjectModel{}, errors.New("no project with an accessible project directory was found")
}

func summarizeProjects(projects []domain.WorkspaceProjectModel) []projectSummary {
	result := make([]projectSummary, 0, len(projects))
	for _, project := range projects {
		result = append(result, summarizeProject(project))
	}
	return result
}

func summarizeProject(project domain.WorkspaceProjectModel) projectSummary {
	return projectSummary{
		ID:         project.ID,
		Name:       project.Name,
		Status:     project.Status,
		ProjectDir: project.ProjectDir,
		UpdatedAt:  project.UpdatedAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"),
	}
}

func isStoryboardDocument(document mediamcp.WorkspaceDocument) bool {
	if strings.EqualFold(strings.TrimSpace(document.Category), "storyboard") {
		return true
	}
	for _, tag := range document.Tags {
		if strings.TrimSpace(tag) == serviceusvdsv11.StoryboardArtifactTag {
			return true
		}
	}
	return false
}

func closeGorm(db interface{ DB() (*sql.DB, error) }) {
	if db == nil {
		return
	}
	sqlDB, err := db.DB()
	if err == nil {
		_ = sqlDB.Close()
	}
}

func writeJSON(value any) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(value)
}

func printReport(report smokeReport) {
	fmt.Printf("USVDS V11 smoke: project=%s (%s)\n", report.Project.ID, report.Project.Name)
	fmt.Printf("snapshot=%s\n", report.SnapshotWorkspace)
	fmt.Printf("documents=%d storyboardDocuments=%d shots=%d readyShots=%d preflightPass=%d\n",
		report.DocumentCount,
		report.StoryboardDocuments,
		report.ShotCount,
		report.ReadyShotCount,
		report.PreflightPassCount,
	)
	fmt.Printf("gates=%d generationReady=%v nextWorkflow=", len(report.Gates.Gates), report.GenerationReady)
	if report.Gates.NextWorkflow == nil {
		fmt.Println("none")
	} else {
		fmt.Println(report.Gates.NextWorkflow.ID)
	}
	for _, gate := range report.Gates.Gates {
		fmt.Printf("- %s ready=%v", gate.Gate, gate.Ready)
		if len(gate.Blockers) > 0 {
			fmt.Printf(" blockers=%s", strings.Join(gate.Blockers, " | "))
		}
		fmt.Println()
	}
	for _, failure := range report.PreflightFailures {
		fmt.Println("preflight-failure:", failure)
	}
}
