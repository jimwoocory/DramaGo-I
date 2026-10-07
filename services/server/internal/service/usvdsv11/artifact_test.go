package usvdsv11

import (
	"testing"

	mediamcp "github.com/mediago-dev/mediago-drama/packages/mcp/pkg/mcp"
	servicecanon "github.com/mediago-dev/mediago-drama/services/server/internal/service/canon"
)

func TestAdoptArtifactUsesExistingDocumentMetadata(t *testing.T) {
	documents := &fakeApprovalDocuments{documents: []mediamcp.WorkspaceDocument{{
		ID:       "script-1",
		Title:    "EP01 Draft",
		Category: "reference",
		Content:  "script",
		Version:  1,
	}}}
	service := NewProjectGateService(
		documents,
		fakeCanon{assets: map[string][]servicecanon.AssetRecord{}},
		fakeShots{},
	)

	result, err := service.AdoptArtifact("project-1", ArtifactScreenplay, "script-1", 1)
	if err != nil {
		t.Fatalf("AdoptArtifact() error = %v", err)
	}
	if result.Document.Category != "screenplay" {
		t.Fatalf("category = %q, want screenplay", result.Document.Category)
	}
	if !hasTag(result.Document.Tags, ScreenplayArtifactTag) {
		t.Fatalf("tags = %#v, want screenplay artifact tag", result.Document.Tags)
	}
	if result.Document.Version != 3 {
		t.Fatalf("version = %d, want 3 after category + tag mutations", result.Document.Version)
	}
	if result.Report.Summary.ScreenplayDocumentID != "script-1" {
		t.Fatalf("summary = %+v, want adopted screenplay", result.Report.Summary)
	}
}

func TestExplicitArtifactTagWinsOverLegacyTitleGuess(t *testing.T) {
	documents := []mediamcp.WorkspaceDocument{
		{ID: "legacy", Title: "Story Package", Content: "legacy", Version: 1},
		{ID: "bound", Title: "完全不同的标题", Content: "bound", Version: 1, Tags: []string{StoryArtifactTag}},
	}
	got := selectStoryPackageDocument(documents)
	if got.ID != "bound" {
		t.Fatalf("selected story = %q, want explicit artifact binding", got.ID)
	}
}
