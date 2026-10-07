package usvdsv11

import (
	"testing"

	mediamcp "github.com/mediago-dev/mediago-drama/packages/mcp/pkg/mcp"
	servicecanon "github.com/mediago-dev/mediago-drama/services/server/internal/service/canon"
	"github.com/mediago-dev/mediago-drama/services/server/internal/service/model"
	serviceshotmanifest "github.com/mediago-dev/mediago-drama/services/server/internal/service/shotmanifest"
)

type fakeDocuments struct {
	documents []mediamcp.WorkspaceDocument
}

func (fake fakeDocuments) ListWorkspaceDocuments(projectID string) (model.WorkspaceDocumentsResponse, error) {
	return model.WorkspaceDocumentsResponse{ProjectID: projectID, Documents: fake.documents}, nil
}

type fakeCanon struct {
	assets map[string][]servicecanon.AssetRecord
}

func (fake fakeCanon) List(_ string, resourceType string) ([]servicecanon.AssetRecord, error) {
	return fake.assets[resourceType], nil
}

type fakeShots struct {
	records map[string][]serviceshotmanifest.Record
}

func (fake fakeShots) ListDocument(_ string, documentID string) ([]serviceshotmanifest.Record, error) {
	return fake.records[documentID], nil
}

func TestEvaluateProjectUsesExistingDramaGoTruth(t *testing.T) {
	service := NewProjectGateService(
		fakeDocuments{documents: []mediamcp.WorkspaceDocument{
			{ID: "story-1", Title: "Story Package", Tags: []string{StoryApprovedTag}},
			{ID: "script-1", Category: "screenplay", Tags: []string{ScreenplayReviewedTag}},
			{ID: "board-1", Category: "storyboard"},
		}},
		fakeCanon{assets: map[string][]servicecanon.AssetRecord{
			servicecanon.ResourceTypeCharacter: {{ID: "char-1", Status: servicecanon.StatusApproved}},
			servicecanon.ResourceTypeScene:     {{ID: "scene-1", Status: servicecanon.StatusLocked}},
			servicecanon.ResourceTypeProp:      {{ID: "prop-1", Status: servicecanon.StatusApproved}},
		}},
		fakeShots{records: map[string][]serviceshotmanifest.Record{
			"board-1": {{
				ID:                "shot-1",
				Status:            serviceshotmanifest.StatusReady,
				ResolvedStateJSON: "{}",
				CompiledPrompt:    "compiled prompt",
			}},
		}},
	)
	service.SetGenerationStateResolver(func(projectID string) (GenerationProjectState, error) {
		return GenerationProjectState{Total: 1, Completed: 1, LastTaskID: "task-1"}, nil
	})

	report, err := service.EvaluateProject("project-1")
	if err != nil {
		t.Fatalf("EvaluateProject() error = %v", err)
	}
	if report.Baseline.Commit != CurrentBaseline.Commit {
		t.Fatalf("baseline = %+v", report.Baseline)
	}
	if report.Summary.CanonCoreCount != 3 || report.Summary.ReadyShotCount != 1 || report.Generation.Completed != 1 {
		t.Fatalf("summary = %+v generation = %+v", report.Summary, report.Generation)
	}
	for _, gate := range report.Gates {
		if !gate.Ready {
			t.Fatalf("gate %s blocked: %v", gate.Gate, gate.Blockers)
		}
	}
}

func TestEvaluateProjectBlocksWhenExistingStateIsIncomplete(t *testing.T) {
	service := NewProjectGateService(
		fakeDocuments{documents: []mediamcp.WorkspaceDocument{
			{ID: "script-1", Category: "screenplay"},
			{ID: "board-1", Category: "storyboard"},
		}},
		fakeCanon{assets: map[string][]servicecanon.AssetRecord{
			servicecanon.ResourceTypeCharacter: {{ID: "char-1", Status: servicecanon.StatusDraft}},
		}},
		fakeShots{records: map[string][]serviceshotmanifest.Record{
			"board-1": {{
				ID:                "shot-1",
				Status:            serviceshotmanifest.StatusDraft,
				ResolvedStateJSON: "",
			}},
		}},
	)

	report, err := service.EvaluateProject("project-1")
	if err != nil {
		t.Fatalf("EvaluateProject() error = %v", err)
	}
	ready := map[GateID]bool{}
	for _, gate := range report.Gates {
		ready[gate.Gate] = gate.Ready
	}
	if ready[GateStoryApproved] || ready[GateScreenplayReviewed] || ready[GateCanonLocked] ||
		ready[GateContinuityResolved] || ready[GateStoryboardReady] || ready[GateGenerationReady] {
		t.Fatalf("unexpected ready gates: %#v", ready)
	}
}
