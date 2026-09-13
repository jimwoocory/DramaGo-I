package generation

import (
	"fmt"
	"net/http"
	"strings"
)

func (workflow *GenerationService) applyShotManifestCompilation(payload *generationMessageRequest) (int, error) {
	if payload == nil {
		return http.StatusBadRequest, fmt.Errorf("generation payload is nil")
	}
	payload.ShotManifestID = strings.TrimSpace(payload.ShotManifestID)
	if payload.ShotManifestID == "" {
		return 0, nil
	}
	if workflow == nil || workflow.shotManifestCompiler == nil {
		return http.StatusServiceUnavailable, fmt.Errorf("Shot Manifest compiler 尚未配置")
	}
	if strings.TrimSpace(payload.ProjectID) == "" {
		return http.StatusBadRequest, fmt.Errorf("使用 shotManifestId 时必须提供 projectId")
	}

	compiled, err := workflow.shotManifestCompiler.CompileAndPersist(payload.ProjectID, payload.ShotManifestID)
	if err != nil {
		return http.StatusBadRequest, err
	}
	if strings.TrimSpace(compiled.Prompt) == "" {
		return http.StatusBadRequest, fmt.Errorf("Shot Manifest 编译结果为空")
	}

	// Shot Manifest is the execution contract. When it is present, discard free-form
	// prompt/reference inputs so callers cannot bypass Canon/continuity constraints.
	payload.Prompt = compiled.Prompt
	payload.PromptSupplements = nil
	payload.ReferenceURLs = nil
	payload.ReferenceAssetIDs = append([]string(nil), compiled.ReferenceAssetIDs...)
	payload.ReferenceBindings = make([]GenerationReferenceBinding, 0, len(compiled.References))
	for _, reference := range compiled.References {
		payload.ReferenceBindings = append(payload.ReferenceBindings, GenerationReferenceBinding{
			Kind:       reference.Kind,
			DocumentID: reference.DocumentID,
			BlockID:    reference.BlockID,
			AssetID:    reference.AssetID,
		})
	}
	payload.DocumentID = strings.TrimSpace(compiled.DocumentID)
	payload.SectionID = strings.TrimSpace(compiled.SectionID)
	payload.DocumentContext = &GenerationDocumentContext{
		ProjectID:  payload.ProjectID,
		DocumentID: payload.DocumentID,
		SectionID:  payload.SectionID,
	}
	payload.ResourceType = "storyboard"
	return 0, nil
}
