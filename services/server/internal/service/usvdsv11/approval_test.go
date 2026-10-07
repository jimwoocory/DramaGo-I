package usvdsv11

import (
	"fmt"
	"testing"

	mediamcp "github.com/mediago-dev/mediago-drama/packages/mcp/pkg/mcp"
	servicecanon "github.com/mediago-dev/mediago-drama/services/server/internal/service/canon"
	servicedocument "github.com/mediago-dev/mediago-drama/services/server/internal/service/document"
	"github.com/mediago-dev/mediago-drama/services/server/internal/service/model"
)

type fakeApprovalDocuments struct {
	documents []mediamcp.WorkspaceDocument
}

func (fake *fakeApprovalDocuments) ListWorkspaceDocuments(projectID string) (model.WorkspaceDocumentsResponse, error) {
	return model.WorkspaceDocumentsResponse{ProjectID: projectID, Documents: append([]mediamcp.WorkspaceDocument(nil), fake.documents...)}, nil
}

func (fake *fakeApprovalDocuments) RequireWorkspaceDocument(_ string, documentID string) (mediamcp.WorkspaceDocument, error) {
	for _, document := range fake.documents {
		if document.ID == documentID {
			return document, nil
		}
	}
	return mediamcp.WorkspaceDocument{}, fmt.Errorf("document not found")
}

func (fake *fakeApprovalDocuments) SetWorkspaceDocumentTags(
	_ string,
	input servicedocument.SetDocumentTagsInput,
	expectedVersion int,
) (servicedocument.WorkspaceDocumentMetadataMutationResult, error) {
	for index, document := range fake.documents {
		if document.ID != input.DocumentID {
			continue
		}
		current := model.NormalizedDocumentVersion(document.Version)
		if current != expectedVersion {
			return servicedocument.WorkspaceDocumentMetadataMutationResult{}, model.WorkspaceVersionConflictError{
				DocumentID: document.ID,
				Expected:   expectedVersion,
				Current:    current,
			}
		}
		before := document
		document.Tags = model.NormalizeDocumentTags(input.Tags)
		document.Version = current + 1
		fake.documents[index] = document
		return servicedocument.WorkspaceDocumentMetadataMutationResult{
			Before:   before,
			Document: document,
		}, nil
	}
	return servicedocument.WorkspaceDocumentMetadataMutationResult{}, fmt.Errorf("document not found")
}

func TestApproveGateBindsCurrentDocumentRevision(t *testing.T) {
	documents := &fakeApprovalDocuments{documents: []mediamcp.WorkspaceDocument{{
		ID:      "story-1",
		Title:   "Story Package",
		Content: "revision two",
		Version: 2,
		Tags:    []string{StoryArtifactTag},
	}}}
	service := NewProjectGateService(documents, fakeCanon{assets: map[string][]servicecanon.AssetRecord{}}, fakeShots{})

	result, err := service.ApproveGate("project-1", GateStoryApproved, "story-1", 2)
	if err != nil {
		t.Fatalf("ApproveGate() error = %v", err)
	}
	if result.Document.Version != 3 {
		t.Fatalf("approved document version = %d, want 3", result.Document.Version)
	}
	state := approvalForGate(result.Report.Approvals, GateStoryApproved)
	if !state.Approved || state.Stale || state.DocumentVersion != 3 {
		t.Fatalf("approval state = %+v", state)
	}
	if !hasTag(result.Document.Tags, approvalToken(GateStoryApproved, result.Document, 3)) {
		t.Fatalf("tags = %#v, want revision-bound approval token", result.Document.Tags)
	}
}

func TestDocumentEditMakesApprovalStale(t *testing.T) {
	document := mediamcp.WorkspaceDocument{
		ID:      "story-1",
		Title:   "Story Package",
		Content: "approved",
		Version: 3,
		Tags:    []string{StoryArtifactTag},
	}
	document.Tags = append(document.Tags, approvalToken(GateStoryApproved, document, 3))
	documents := &fakeApprovalDocuments{documents: []mediamcp.WorkspaceDocument{document}}
	service := NewProjectGateService(documents, fakeCanon{assets: map[string][]servicecanon.AssetRecord{}}, fakeShots{})

	documents.documents[0].Content = "edited after approval"
	documents.documents[0].Version = 4

	report, err := service.EvaluateProject("project-1")
	if err != nil {
		t.Fatalf("EvaluateProject() error = %v", err)
	}
	state := approvalForGate(report.Approvals, GateStoryApproved)
	if state.Approved || !state.Stale {
		t.Fatalf("approval state = %+v, want stale", state)
	}
	for _, gate := range report.Gates {
		if gate.Gate == GateStoryApproved && gate.Ready {
			t.Fatal("story gate remained ready after document edit")
		}
	}
}

func TestRevokeGateRemovesRevisionApproval(t *testing.T) {
	document := mediamcp.WorkspaceDocument{
		ID:       "script-1",
		Title:    "EP01 Screenplay",
		Category: "screenplay",
		Content:  "script",
		Version:  5,
		Tags:     []string{ScreenplayArtifactTag},
	}
	document.Tags = append(document.Tags, approvalToken(GateScreenplayReviewed, document, 5))
	documents := &fakeApprovalDocuments{documents: []mediamcp.WorkspaceDocument{document}}
	service := NewProjectGateService(documents, fakeCanon{assets: map[string][]servicecanon.AssetRecord{}}, fakeShots{})

	result, err := service.RevokeGate("project-1", GateScreenplayReviewed, "script-1", 5)
	if err != nil {
		t.Fatalf("RevokeGate() error = %v", err)
	}
	if result.Document.Version != 6 {
		t.Fatalf("revoked document version = %d, want 6", result.Document.Version)
	}
	state := approvalForGate(result.Report.Approvals, GateScreenplayReviewed)
	if state.Approved || state.Stale {
		t.Fatalf("approval state = %+v, want clean revoked state", state)
	}
}

func TestApproveGateRejectsNonHumanGate(t *testing.T) {
	documents := &fakeApprovalDocuments{documents: []mediamcp.WorkspaceDocument{{ID: "story-1", Version: 1}}}
	service := NewProjectGateService(documents, fakeCanon{assets: map[string][]servicecanon.AssetRecord{}}, fakeShots{})
	if _, err := service.ApproveGate("project-1", GateGenerationReady, "story-1", 1); err == nil {
		t.Fatal("expected non-human gate approval to fail")
	}
}
