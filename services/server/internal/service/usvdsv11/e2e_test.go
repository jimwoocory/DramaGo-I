package usvdsv11

import (
	"testing"

	mediamcp "github.com/mediago-dev/mediago-drama/packages/mcp/pkg/mcp"
	servicecanon "github.com/mediago-dev/mediago-drama/services/server/internal/service/canon"
	serviceshotmanifest "github.com/mediago-dev/mediago-drama/services/server/internal/service/shotmanifest"
)

func TestUSVDSV11FixtureE2EFromStoryToGenerationPreflight(t *testing.T) {
	documents := &fakeApprovalDocuments{documents: []mediamcp.WorkspaceDocument{
		{
			ID:       "story-1",
			Title:    "Agent Story Output",
			Category: "reference",
			Content:  "story package",
			Version:  1,
		},
	}}
	shotRecords := map[string][]serviceshotmanifest.Record{}
	service := NewProjectGateService(
		documents,
		fakeCanon{assets: map[string][]servicecanon.AssetRecord{
			servicecanon.ResourceTypeCharacter: {{ID: "char-1", Status: servicecanon.StatusApproved}},
			servicecanon.ResourceTypeScene:     {{ID: "scene-1", Status: servicecanon.StatusApproved}},
			servicecanon.ResourceTypeProp:      {{ID: "prop-1", Status: servicecanon.StatusLocked}},
		}},
		fakeShots{records: shotRecords},
	)

	initial, err := service.EvaluateProject("project-1")
	if err != nil {
		t.Fatalf("initial EvaluateProject() error = %v", err)
	}
	if initial.NextWorkflow == nil || initial.NextWorkflow.ID != "story" {
		t.Fatalf("initial next workflow = %+v, want story", initial.NextWorkflow)
	}

	storyAdoption, err := service.AdoptArtifact(
		"project-1",
		ArtifactStoryPackage,
		"story-1",
		1,
	)
	if err != nil {
		t.Fatalf("adopt story: %v", err)
	}
	if storyAdoption.Report.NextWorkflow == nil || storyAdoption.Report.NextWorkflow.ID != "review" {
		t.Fatalf("after story next = %+v, want review", storyAdoption.Report.NextWorkflow)
	}
	storyApproval, err := service.ApproveGate(
		"project-1",
		GateStoryApproved,
		"story-1",
		storyAdoption.Document.Version,
	)
	if err != nil {
		t.Fatalf("approve story: %v", err)
	}
	if storyApproval.Report.NextWorkflow == nil || storyApproval.Report.NextWorkflow.ID != "episode" {
		t.Fatalf("after story approval next = %+v, want episode", storyApproval.Report.NextWorkflow)
	}

	documents.documents = append(documents.documents, mediamcp.WorkspaceDocument{
		ID:       "episodes-1",
		Title:    "Agent Episode Output",
		Category: "reference",
		Content:  "episode architecture",
		Version:  1,
	})
	episodeAdoption, err := service.AdoptArtifact(
		"project-1",
		ArtifactEpisodeArchitecture,
		"episodes-1",
		1,
	)
	if err != nil {
		t.Fatalf("adopt episodes: %v", err)
	}
	if episodeAdoption.Report.NextWorkflow == nil || episodeAdoption.Report.NextWorkflow.ID != "screenplay" {
		t.Fatalf("after episode next = %+v, want screenplay", episodeAdoption.Report.NextWorkflow)
	}

	documents.documents = append(documents.documents, mediamcp.WorkspaceDocument{
		ID:       "script-1",
		Title:    "Agent Screenplay Output",
		Category: "reference",
		Content:  "screenplay",
		Version:  1,
	})
	screenplayAdoption, err := service.AdoptArtifact(
		"project-1",
		ArtifactScreenplay,
		"script-1",
		1,
	)
	if err != nil {
		t.Fatalf("adopt screenplay: %v", err)
	}
	if screenplayAdoption.Document.Category != "screenplay" ||
		!hasTag(screenplayAdoption.Document.Tags, ScreenplayArtifactTag) {
		t.Fatalf("screenplay document = %+v, want category/tag binding", screenplayAdoption.Document)
	}
	if screenplayAdoption.Report.NextWorkflow == nil || screenplayAdoption.Report.NextWorkflow.ID != "review" {
		t.Fatalf("after screenplay next = %+v, want review", screenplayAdoption.Report.NextWorkflow)
	}
	screenplayApproval, err := service.ApproveGate(
		"project-1",
		GateScreenplayReviewed,
		"script-1",
		screenplayAdoption.Document.Version,
	)
	if err != nil {
		t.Fatalf("approve screenplay: %v", err)
	}
	if screenplayApproval.Report.NextWorkflow == nil || screenplayApproval.Report.NextWorkflow.ID != "storyboard" {
		t.Fatalf("after screenplay approval next = %+v, want storyboard", screenplayApproval.Report.NextWorkflow)
	}

	documents.documents = append(documents.documents, mediamcp.WorkspaceDocument{
		ID:       "board-1",
		Title:    "Agent Storyboard Output",
		Category: "reference",
		Content:  "storyboard",
		Version:  1,
	})
	shotRecords["board-1"] = []serviceshotmanifest.Record{{
		ID:                "shot-1",
		DocumentID:        "board-1",
		SectionID:         "shot-section-1",
		Status:            serviceshotmanifest.StatusReady,
		ResolvedStateJSON: "{}",
		CompiledPrompt:    "compiled generation prompt",
	}}
	storyboardAdoption, err := service.AdoptArtifact(
		"project-1",
		ArtifactStoryboard,
		"board-1",
		1,
	)
	if err != nil {
		t.Fatalf("adopt storyboard: %v", err)
	}
	if storyboardAdoption.Document.Category != "storyboard" ||
		!hasTag(storyboardAdoption.Document.Tags, StoryboardArtifactTag) {
		t.Fatalf("storyboard document = %+v, want category/tag binding", storyboardAdoption.Document)
	}
	for _, gate := range storyboardAdoption.Report.Gates {
		if !gate.Ready {
			t.Fatalf("final gate %s blocked: %v", gate.Gate, gate.Blockers)
		}
	}
	if storyboardAdoption.Report.NextWorkflow != nil {
		t.Fatalf("final next workflow = %+v, want completed production readiness", storyboardAdoption.Report.NextWorkflow)
	}

	if err := ValidateGenerationCompileResult(serviceshotmanifest.CompileResult{
		ShotManifestID:    "shot-1",
		Status:            serviceshotmanifest.StatusReady,
		ResolvedStateJSON: "{}",
		Prompt:            "compiled generation prompt",
		DocumentID:        "board-1",
		SectionID:         "shot-section-1",
	}); err != nil {
		t.Fatalf("generation preflight blocked after full fixture flow: %v", err)
	}
}
