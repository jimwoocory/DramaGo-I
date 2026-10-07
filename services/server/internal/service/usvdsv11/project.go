package usvdsv11

import (
	"fmt"
	"strings"

	mediamcp "github.com/mediago-dev/mediago-drama/packages/mcp/pkg/mcp"
	servicecanon "github.com/mediago-dev/mediago-drama/services/server/internal/service/canon"
	"github.com/mediago-dev/mediago-drama/services/server/internal/service/model"
	serviceshotmanifest "github.com/mediago-dev/mediago-drama/services/server/internal/service/shotmanifest"
)

const (
	StoryApprovedTag      = "usvds:story-approved"
	ScreenplayReviewedTag = "usvds:screenplay-reviewed"
)

type projectDocumentProvider interface {
	ListWorkspaceDocuments(projectID string) (model.WorkspaceDocumentsResponse, error)
}

type projectCanonProvider interface {
	List(projectID string, resourceType string) ([]servicecanon.AssetRecord, error)
}

type projectShotProvider interface {
	ListDocument(projectID string, documentID string) ([]serviceshotmanifest.Record, error)
}

// GenerationProjectState is a read-only summary of existing DramaGo GenerationTask records.
type GenerationProjectState struct {
	Total         int    `json:"total"`
	Pending       int    `json:"pending"`
	Running       int    `json:"running"`
	Completed     int    `json:"completed"`
	Failed        int    `json:"failed"`
	LastTaskID    string `json:"lastTaskId,omitempty"`
	LastUpdatedAt string `json:"lastUpdatedAt,omitempty"`
}

// GenerationStateResolver projects existing DramaGo GenerationTask records into V11 gate reporting.
type GenerationStateResolver func(projectID string) (GenerationProjectState, error)

// ProjectGateSummary explains the existing DramaGo state used to derive V11 gates.
type ProjectGateSummary struct {
	StoryDocumentID               string `json:"storyDocumentId,omitempty"`
	EpisodeArchitectureDocumentID string `json:"episodeArchitectureDocumentId,omitempty"`
	ScreenplayDocumentID          string `json:"screenplayDocumentId,omitempty"`
	StoryboardDocuments           int    `json:"storyboardDocuments"`
	CanonCoreCount                int    `json:"canonCoreCount"`
	CanonApprovedCount            int    `json:"canonApprovedCount"`
	ShotCount                     int    `json:"shotCount"`
	ReadyShotCount                int    `json:"readyShotCount"`
	ResolvedShotCount             int    `json:"resolvedShotCount"`
	CompiledShotCount             int    `json:"compiledShotCount"`
}

// ProjectGateReport is the project-level USVDS V11 readiness projection.
type ProjectGateReport struct {
	ProjectID    string                 `json:"projectId"`
	Baseline     Baseline               `json:"baseline"`
	Ownership    []Ownership            `json:"ownership"`
	Gates        []GateResult           `json:"gates"`
	Approvals    []ApprovalState        `json:"approvals"`
	Workflows    []WorkflowDescriptor   `json:"workflows"`
	NextWorkflow *WorkflowDescriptor    `json:"nextWorkflow,omitempty"`
	Summary      ProjectGateSummary     `json:"summary"`
	Generation   GenerationProjectState `json:"generation"`
}

// ProjectGateService derives V11 readiness from existing DramaGo domain records.
type ProjectGateService struct {
	documents  projectDocumentProvider
	canon      projectCanonProvider
	shots      projectShotProvider
	generation GenerationStateResolver
}

// NewProjectGateService returns a non-persistent V11 project gate service.
func NewProjectGateService(
	documents projectDocumentProvider,
	canon projectCanonProvider,
	shots projectShotProvider,
) *ProjectGateService {
	return &ProjectGateService{documents: documents, canon: canon, shots: shots}
}

// SetGenerationStateResolver attaches a read-only GenerationTask projection.
func (service *ProjectGateService) SetGenerationStateResolver(resolver GenerationStateResolver) {
	if service == nil {
		return
	}
	service.generation = resolver
}

// EvaluateProject derives all V11 gates from current DramaGo-owned state.
func (service *ProjectGateService) EvaluateProject(projectID string) (ProjectGateReport, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return ProjectGateReport{}, fmt.Errorf("project id is required")
	}
	if service == nil || service.documents == nil || service.canon == nil || service.shots == nil {
		return ProjectGateReport{}, fmt.Errorf("USVDS V11 gate service is not configured")
	}

	documents, err := service.documents.ListWorkspaceDocuments(projectID)
	if err != nil {
		return ProjectGateReport{}, fmt.Errorf("listing project documents: %w", err)
	}

	storyDocument := selectStoryPackageDocument(documents.Documents)
	episodeArchitectureDocument := selectEpisodeArchitectureDocument(documents.Documents)
	screenplayDocument := selectScreenplayDocument(documents.Documents)
	storyboardIDs := make([]string, 0)
	for _, document := range documents.Documents {
		if model.NormalizeDocumentCategoryValue(document.Category) == "storyboard" ||
			hasTag(document.Tags, StoryboardArtifactTag) {
			storyboardIDs = append(storyboardIDs, document.ID)
		}
	}
	storyApproval := approvalState(GateStoryApproved, storyDocument)
	screenplayApproval := approvalState(GateScreenplayReviewed, screenplayDocument)
	storyApproved := storyApproval.Approved
	screenplayReviewed := screenplayApproval.Approved

	canonCoreCount := 0
	canonApprovedCount := 0
	for _, resourceType := range []string{
		servicecanon.ResourceTypeCharacter,
		servicecanon.ResourceTypeScene,
		servicecanon.ResourceTypeProp,
	} {
		assets, listErr := service.canon.List(projectID, resourceType)
		if listErr != nil {
			return ProjectGateReport{}, fmt.Errorf("listing %s Canon: %w", resourceType, listErr)
		}
		for _, asset := range assets {
			if strings.TrimSpace(asset.ParentID) != "" || asset.Status == servicecanon.StatusDeprecated {
				continue
			}
			canonCoreCount++
			if asset.Status == servicecanon.StatusApproved || asset.Status == servicecanon.StatusLocked {
				canonApprovedCount++
			}
		}
	}
	canonApproved := canonCoreCount > 0 && canonApprovedCount == canonCoreCount

	shotCount := 0
	readyShotCount := 0
	resolvedShotCount := 0
	compiledShotCount := 0
	for _, documentID := range storyboardIDs {
		shots, listErr := service.shots.ListDocument(projectID, documentID)
		if listErr != nil {
			return ProjectGateReport{}, fmt.Errorf("listing ShotManifest for %s: %w", documentID, listErr)
		}
		for _, shot := range shots {
			shotCount++
			if shot.Status == serviceshotmanifest.StatusReady {
				readyShotCount++
			}
			if shot.Status != serviceshotmanifest.StatusConflict &&
				strings.TrimSpace(shot.ResolvedStateJSON) != "" {
				resolvedShotCount++
			}
			if strings.TrimSpace(shot.CompiledPrompt) != "" {
				compiledShotCount++
			}
		}
	}

	continuityResolved := shotCount > 0 && resolvedShotCount == shotCount
	storyboardReady := shotCount > 0 && readyShotCount == shotCount
	generationReady := storyboardReady && compiledShotCount == shotCount

	snapshot := Snapshot{
		ProjectID:          projectID,
		DocumentID:         firstNonEmpty(storyDocument.ID, screenplayDocument.ID),
		StoryApproved:      storyApproved,
		ScreenplayReviewed: screenplayReviewed,
		CanonApproved:      canonApproved,
		ContinuityResolved: continuityResolved,
	}
	if shotCount > 0 {
		snapshot.ShotManifestID = fmt.Sprintf("project:%d", shotCount)
	}
	if storyboardReady {
		snapshot.ShotStatus = serviceshotmanifest.StatusReady
	}
	if generationReady {
		snapshot.CompiledPrompt = "compiled"
	}

	generation := GenerationProjectState{}
	if service.generation != nil {
		generation, err = service.generation(projectID)
		if err != nil {
			return ProjectGateReport{}, fmt.Errorf("reading generation state: %w", err)
		}
	}

	report := ProjectGateReport{
		ProjectID: projectID,
		Baseline:  CurrentBaseline,
		Ownership: OwnershipMap(),
		Gates:     Evaluate(snapshot),
		Approvals: []ApprovalState{storyApproval, screenplayApproval},
		Workflows: WorkflowCatalog(),
		Summary: ProjectGateSummary{
			StoryDocumentID:               storyDocument.ID,
			EpisodeArchitectureDocumentID: episodeArchitectureDocument.ID,
			ScreenplayDocumentID:          screenplayDocument.ID,
			StoryboardDocuments:           len(storyboardIDs),
			CanonCoreCount:                canonCoreCount,
			CanonApprovedCount:            canonApprovedCount,
			ShotCount:                     shotCount,
			ReadyShotCount:                readyShotCount,
			ResolvedShotCount:             resolvedShotCount,
			CompiledShotCount:             compiledShotCount,
		},
		Generation: generation,
	}
	report.NextWorkflow = recommendedWorkflow(report)
	return report, nil
}

func selectStoryPackageDocument(documents []mediamcp.WorkspaceDocument) mediamcp.WorkspaceDocument {
	for _, document := range documents {
		if hasTag(document.Tags, StoryArtifactTag) {
			return document
		}
		for _, tag := range document.Tags {
			if strings.HasPrefix(strings.TrimSpace(tag), approvalPrefix(GateStoryApproved)) {
				return document
			}
		}
	}
	for _, document := range documents {
		if isStoryPackageDocument(document) {
			return document
		}
	}
	return mediamcp.WorkspaceDocument{}
}

func selectEpisodeArchitectureDocument(documents []mediamcp.WorkspaceDocument) mediamcp.WorkspaceDocument {
	for _, document := range documents {
		if hasTag(document.Tags, EpisodeArchitectureArtifactTag) {
			return document
		}
	}
	for _, document := range documents {
		if isEpisodeArchitectureDocument(document) {
			return document
		}
	}
	return mediamcp.WorkspaceDocument{}
}

func selectScreenplayDocument(documents []mediamcp.WorkspaceDocument) mediamcp.WorkspaceDocument {
	for _, document := range documents {
		if hasTag(document.Tags, ScreenplayArtifactTag) {
			return document
		}
	}
	for _, document := range documents {
		if model.NormalizeDocumentCategoryValue(document.Category) == "screenplay" {
			return document
		}
	}
	return mediamcp.WorkspaceDocument{}
}

func isStoryPackageDocument(document mediamcp.WorkspaceDocument) bool {
	if hasTag(document.Tags, StoryArtifactTag) || hasTag(document.Tags, StoryApprovedTag) {
		return true
	}
	for _, tag := range document.Tags {
		if strings.HasPrefix(strings.TrimSpace(tag), approvalPrefix(GateStoryApproved)) {
			return true
		}
	}
	title := strings.ToLower(strings.TrimSpace(document.Title))
	return strings.Contains(title, "story package") ||
		strings.Contains(title, "故事包") ||
		strings.Contains(title, "故事大纲") ||
		strings.Contains(title, "故事架构")
}

func isEpisodeArchitectureDocument(document mediamcp.WorkspaceDocument) bool {
	if hasTag(document.Tags, EpisodeArchitectureArtifactTag) {
		return true
	}
	title := strings.ToLower(strings.TrimSpace(document.Title))
	return strings.Contains(title, "episode architecture") ||
		strings.Contains(title, "episode map") ||
		strings.Contains(title, "分集架构") ||
		strings.Contains(title, "分集大纲")
}

func hasTag(tags []string, target string) bool {
	target = strings.ToLower(strings.TrimSpace(target))
	for _, tag := range tags {
		if strings.ToLower(strings.TrimSpace(tag)) == target {
			return true
		}
	}
	return false
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}
