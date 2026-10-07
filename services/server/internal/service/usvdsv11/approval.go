package usvdsv11

import (
	"crypto/sha256"
	"fmt"
	"strings"

	mediamcp "github.com/mediago-dev/mediago-drama/packages/mcp/pkg/mcp"
	servicedocument "github.com/mediago-dev/mediago-drama/services/server/internal/service/document"
	"github.com/mediago-dev/mediago-drama/services/server/internal/service/model"
)

const (
	StoryArtifactTag               = "usvds:artifact:story-package"
	EpisodeArchitectureArtifactTag = "usvds:artifact:episode-architecture"
	ScreenplayArtifactTag          = "usvds:artifact:screenplay"
	StoryboardArtifactTag          = "usvds:artifact:storyboard"
)

type gateApprovalDocumentStore interface {
	RequireWorkspaceDocument(projectID string, documentID string) (mediamcp.WorkspaceDocument, error)
	SetWorkspaceDocumentCategory(projectID string, input servicedocument.SetDocumentCategoryInput, expectedVersion int) (servicedocument.WorkspaceDocumentMetadataMutationResult, error)
	SetWorkspaceDocumentTags(projectID string, input servicedocument.SetDocumentTagsInput, expectedVersion int) (servicedocument.WorkspaceDocumentMetadataMutationResult, error)
}

type ApprovalState struct {
	Gate            GateID `json:"gate"`
	DocumentID      string `json:"documentId,omitempty"`
	DocumentTitle   string `json:"documentTitle,omitempty"`
	DocumentVersion int    `json:"documentVersion,omitempty"`
	ContentDigest   string `json:"contentDigest,omitempty"`
	Approved        bool   `json:"approved"`
	Stale           bool   `json:"stale"`
}

type GateMutationResult struct {
	Document mediamcp.WorkspaceDocument `json:"document"`
	Report   ProjectGateReport          `json:"report"`
}

func approvalPrefix(gate GateID) string {
	return "usvds:approval:" + string(gate) + ":"
}

func approvalDigest(content string) string {
	sum := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", sum[:])
}

func approvalToken(gate GateID, document mediamcp.WorkspaceDocument, version int) string {
	return fmt.Sprintf(
		"%sv%d:sha256:%s",
		approvalPrefix(gate),
		model.NormalizedDocumentVersion(version),
		approvalDigest(document.Content),
	)
}

func approvalState(gate GateID, document mediamcp.WorkspaceDocument) ApprovalState {
	state := ApprovalState{
		Gate:            gate,
		DocumentID:      strings.TrimSpace(document.ID),
		DocumentTitle:   strings.TrimSpace(document.Title),
		DocumentVersion: model.NormalizedDocumentVersion(document.Version),
		ContentDigest:   approvalDigest(document.Content),
	}
	if state.DocumentID == "" {
		return state
	}
	expected := approvalToken(gate, document, state.DocumentVersion)
	prefix := approvalPrefix(gate)
	for _, tag := range document.Tags {
		tag = strings.TrimSpace(tag)
		if tag == expected {
			state.Approved = true
			return state
		}
		if strings.HasPrefix(tag, prefix) {
			state.Stale = true
		}
	}
	return state
}

func replaceGateApprovalTag(tags []string, gate GateID, replacement string) []string {
	prefix := approvalPrefix(gate)
	result := make([]string, 0, len(tags)+1)
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" || strings.HasPrefix(tag, prefix) {
			continue
		}
		if gate == GateStoryApproved && tag == StoryApprovedTag {
			continue
		}
		if gate == GateScreenplayReviewed && tag == ScreenplayReviewedTag {
			continue
		}
		result = append(result, tag)
	}
	if replacement != "" {
		result = append(result, replacement)
	}
	return model.NormalizeDocumentTags(result)
}

func (service *ProjectGateService) ApproveGate(
	projectID string,
	gate GateID,
	documentID string,
	expectedVersion int,
) (GateMutationResult, error) {
	if err := validateHumanGate(gate); err != nil {
		return GateMutationResult{}, err
	}
	store, ok := service.documents.(gateApprovalDocumentStore)
	if !ok {
		return GateMutationResult{}, fmt.Errorf("USVDS V11 document approval store is not configured")
	}
	document, err := store.RequireWorkspaceDocument(projectID, strings.TrimSpace(documentID))
	if err != nil {
		return GateMutationResult{}, err
	}
	if err := validateGateDocument(gate, document); err != nil {
		return GateMutationResult{}, err
	}
	currentVersion := model.NormalizedDocumentVersion(document.Version)
	if expectedVersion != currentVersion {
		return GateMutationResult{}, model.WorkspaceVersionConflictError{
			DocumentID: document.ID,
			Expected:   expectedVersion,
			Current:    currentVersion,
		}
	}

	tags := replaceGateApprovalTag(document.Tags, gate, "")
	switch gate {
	case GateStoryApproved:
		tags = model.NormalizeDocumentTags(append(tags, StoryArtifactTag))
	case GateScreenplayReviewed:
		tags = model.NormalizeDocumentTags(append(tags, ScreenplayArtifactTag))
	}
	token := approvalToken(gate, document, currentVersion+1)
	tags = model.NormalizeDocumentTags(append(tags, token))

	mutation, err := store.SetWorkspaceDocumentTags(projectID, servicedocument.SetDocumentTagsInput{
		DocumentID:      document.ID,
		Tags:            tags,
		ExpectedVersion: currentVersion,
		Summary:         "Approve current USVDS V11 document gate",
	}, currentVersion)
	if err != nil {
		return GateMutationResult{}, err
	}
	report, err := service.EvaluateProject(projectID)
	if err != nil {
		return GateMutationResult{}, err
	}
	return GateMutationResult{Document: mutation.Document, Report: report}, nil
}

func (service *ProjectGateService) RevokeGate(
	projectID string,
	gate GateID,
	documentID string,
	expectedVersion int,
) (GateMutationResult, error) {
	if err := validateHumanGate(gate); err != nil {
		return GateMutationResult{}, err
	}
	store, ok := service.documents.(gateApprovalDocumentStore)
	if !ok {
		return GateMutationResult{}, fmt.Errorf("USVDS V11 document approval store is not configured")
	}
	document, err := store.RequireWorkspaceDocument(projectID, strings.TrimSpace(documentID))
	if err != nil {
		return GateMutationResult{}, err
	}
	if err := validateGateDocument(gate, document); err != nil {
		return GateMutationResult{}, err
	}
	currentVersion := model.NormalizedDocumentVersion(document.Version)
	if expectedVersion != currentVersion {
		return GateMutationResult{}, model.WorkspaceVersionConflictError{
			DocumentID: document.ID,
			Expected:   expectedVersion,
			Current:    currentVersion,
		}
	}

	tags := replaceGateApprovalTag(document.Tags, gate, "")
	mutation, err := store.SetWorkspaceDocumentTags(projectID, servicedocument.SetDocumentTagsInput{
		DocumentID:      document.ID,
		Tags:            tags,
		ExpectedVersion: currentVersion,
		Summary:         "Revoke current USVDS V11 document gate approval",
	}, currentVersion)
	if err != nil {
		return GateMutationResult{}, err
	}
	report, err := service.EvaluateProject(projectID)
	if err != nil {
		return GateMutationResult{}, err
	}
	return GateMutationResult{Document: mutation.Document, Report: report}, nil
}

func validateHumanGate(gate GateID) error {
	switch gate {
	case GateStoryApproved, GateScreenplayReviewed:
		return nil
	default:
		return fmt.Errorf("gate %q is not a human document approval gate", gate)
	}
}

func validateGateDocument(gate GateID, document mediamcp.WorkspaceDocument) error {
	switch gate {
	case GateStoryApproved:
		return nil
	case GateScreenplayReviewed:
		if model.NormalizeDocumentCategoryValue(document.Category) != "screenplay" {
			return fmt.Errorf("screenplay review requires a screenplay document")
		}
		return nil
	default:
		return validateHumanGate(gate)
	}
}
