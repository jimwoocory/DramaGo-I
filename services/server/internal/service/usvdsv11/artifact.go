package usvdsv11

import (
	"fmt"
	"strings"

	mediamcp "github.com/mediago-dev/mediago-drama/packages/mcp/pkg/mcp"
	servicedocument "github.com/mediago-dev/mediago-drama/services/server/internal/service/document"
	"github.com/mediago-dev/mediago-drama/services/server/internal/service/model"
)

type ArtifactKind string

const (
	ArtifactStoryPackage        ArtifactKind = "story-package"
	ArtifactEpisodeArchitecture ArtifactKind = "episode-architecture"
	ArtifactScreenplay          ArtifactKind = "screenplay"
	ArtifactStoryboard          ArtifactKind = "storyboard"
)

type ArtifactDescriptor struct {
	Kind             ArtifactKind `json:"kind"`
	Tag              string       `json:"tag"`
	DocumentCategory string       `json:"documentCategory,omitempty"`
}

type ArtifactMutationResult struct {
	Artifact ArtifactDescriptor         `json:"artifact"`
	Document mediamcp.WorkspaceDocument `json:"document"`
	Report   ProjectGateReport          `json:"report"`
}

func ArtifactCatalog() []ArtifactDescriptor {
	return []ArtifactDescriptor{
		{Kind: ArtifactStoryPackage, Tag: StoryArtifactTag, DocumentCategory: "reference"},
		{Kind: ArtifactEpisodeArchitecture, Tag: EpisodeArchitectureArtifactTag, DocumentCategory: "reference"},
		{Kind: ArtifactScreenplay, Tag: ScreenplayArtifactTag, DocumentCategory: "screenplay"},
		{Kind: ArtifactStoryboard, Tag: StoryboardArtifactTag, DocumentCategory: "storyboard"},
	}
}

func artifactDescriptor(kind ArtifactKind) (ArtifactDescriptor, error) {
	kind = ArtifactKind(strings.TrimSpace(string(kind)))
	for _, descriptor := range ArtifactCatalog() {
		if descriptor.Kind == kind {
			return descriptor, nil
		}
	}
	return ArtifactDescriptor{}, fmt.Errorf("unsupported USVDS V11 artifact %q", kind)
}

// AdoptArtifact binds an existing DramaGo Document to one V11 artifact role.
// Persistence remains entirely inside the existing Document category/tags fields.
func (service *ProjectGateService) AdoptArtifact(
	projectID string,
	kind ArtifactKind,
	documentID string,
	expectedVersion int,
) (ArtifactMutationResult, error) {
	descriptor, err := artifactDescriptor(kind)
	if err != nil {
		return ArtifactMutationResult{}, err
	}
	store, ok := service.documents.(gateApprovalDocumentStore)
	if !ok {
		return ArtifactMutationResult{}, fmt.Errorf("USVDS V11 document artifact store is not configured")
	}
	document, err := store.RequireWorkspaceDocument(projectID, strings.TrimSpace(documentID))
	if err != nil {
		return ArtifactMutationResult{}, err
	}
	currentVersion := model.NormalizedDocumentVersion(document.Version)
	if expectedVersion != currentVersion {
		return ArtifactMutationResult{}, model.WorkspaceVersionConflictError{
			DocumentID: document.ID,
			Expected:   expectedVersion,
			Current:    currentVersion,
		}
	}

	if category := model.NormalizeDocumentCategoryValue(descriptor.DocumentCategory); category != "" &&
		model.NormalizeDocumentCategoryValue(document.Category) != category {
		categoryResult, categoryErr := store.SetWorkspaceDocumentCategory(projectID, servicedocument.SetDocumentCategoryInput{
			DocumentID:      document.ID,
			Category:        category,
			ExpectedVersion: currentVersion,
			Summary:         "Bind document to USVDS V11 artifact category",
		}, currentVersion)
		if categoryErr != nil {
			return ArtifactMutationResult{}, categoryErr
		}
		document = categoryResult.Document
		currentVersion = model.NormalizedDocumentVersion(document.Version)
	}

	tags := model.NormalizeDocumentTags(append(document.Tags, descriptor.Tag))
	tagResult, err := store.SetWorkspaceDocumentTags(projectID, servicedocument.SetDocumentTagsInput{
		DocumentID:      document.ID,
		Tags:            tags,
		ExpectedVersion: currentVersion,
		Summary:         "Bind document to USVDS V11 artifact",
	}, currentVersion)
	if err != nil {
		return ArtifactMutationResult{}, err
	}
	report, err := service.EvaluateProject(projectID)
	if err != nil {
		return ArtifactMutationResult{}, err
	}
	return ArtifactMutationResult{
		Artifact: descriptor,
		Document: tagResult.Document,
		Report:   report,
	}, nil
}
