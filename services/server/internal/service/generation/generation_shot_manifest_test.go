package generation

import (
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/mediago-dev/mediago-drama/services/server/internal/service/shotmanifest"
)

type fakeShotManifestCompiler struct {
	projectID string
	shotID    string
	result    shotmanifest.CompileResult
	err       error
}

func (fake *fakeShotManifestCompiler) CompileAndPersist(projectID string, id string) (shotmanifest.CompileResult, error) {
	fake.projectID = projectID
	fake.shotID = id
	return fake.result, fake.err
}

func TestApplyShotManifestCompilationOverridesFreeFormInputs(t *testing.T) {
	compiler := &fakeShotManifestCompiler{result: shotmanifest.CompileResult{
		DocumentID:        "storyboard-ep1",
		SectionID:         "section-shot-7",
		Prompt:            "compiled canonical prompt",
		ReferenceAssetIDs: []string{"asset-char", "asset-scene"},
		References: []shotmanifest.CompiledReference{
			{Kind: "section", DocumentID: "characters", BlockID: "section-char", AssetID: "asset-char", Role: "identity"},
			{Kind: "section", DocumentID: "scenes", BlockID: "section-scene", AssetID: "asset-scene", Role: "scene_master"},
		},
	}}
	workflow := &GenerationService{shotManifestCompiler: compiler}
	payload := generationMessageRequest{
		ProjectID:         "project-a",
		ShotManifestID:    " shot-7 ",
		Prompt:            "free form prompt that must be discarded",
		PromptSupplements: []GenerationPromptSupplementRequest{{ReferenceID: "legacy-style", ReferencePrompt: "cinematic lighting"}},
		ReferenceURLs:     []string{"https://untrusted.example/extra.png"},
		ReferenceAssetIDs: []string{"asset-extra"},
		ReferenceBindings: []GenerationReferenceBinding{{Kind: "asset", AssetID: "asset-extra"}},
	}

	status, err := workflow.applyShotManifestCompilation(&payload)
	if err != nil || status != 0 {
		t.Fatalf("applyShotManifestCompilation() status=%d error=%v", status, err)
	}
	if compiler.projectID != "project-a" || compiler.shotID != "shot-7" {
		t.Fatalf("compiler args = %q/%q", compiler.projectID, compiler.shotID)
	}
	if payload.Prompt != "compiled canonical prompt" {
		t.Fatalf("Prompt = %q, want compiled prompt", payload.Prompt)
	}
	if len(payload.PromptSupplements) != 0 {
		t.Fatalf("PromptSupplements = %#v, want manifest-backed request to clear legacy supplements", payload.PromptSupplements)
	}
	if len(payload.ReferenceURLs) != 0 {
		t.Fatalf("ReferenceURLs = %#v, want authoritative manifest to clear them", payload.ReferenceURLs)
	}
	if !reflect.DeepEqual(payload.ReferenceAssetIDs, []string{"asset-char", "asset-scene"}) {
		t.Fatalf("ReferenceAssetIDs = %#v", payload.ReferenceAssetIDs)
	}
	if len(payload.ReferenceBindings) != 2 || payload.ReferenceBindings[0].AssetID != "asset-char" || payload.ReferenceBindings[1].AssetID != "asset-scene" {
		t.Fatalf("ReferenceBindings = %#v", payload.ReferenceBindings)
	}
	if payload.DocumentID != "storyboard-ep1" || payload.SectionID != "section-shot-7" || payload.ResourceType != "storyboard" {
		t.Fatalf("document context = document:%q section:%q resource:%q", payload.DocumentID, payload.SectionID, payload.ResourceType)
	}
	if payload.DocumentContext == nil || payload.DocumentContext.ProjectID != "project-a" || payload.DocumentContext.DocumentID != "storyboard-ep1" || payload.DocumentContext.SectionID != "section-shot-7" {
		t.Fatalf("DocumentContext = %#v", payload.DocumentContext)
	}
}

func TestApplyShotManifestCompilationFailsClosed(t *testing.T) {
	t.Run("compiler missing", func(t *testing.T) {
		workflow := &GenerationService{}
		payload := generationMessageRequest{ProjectID: "project-a", ShotManifestID: "shot-1", Prompt: "fallback must not run"}
		status, err := workflow.applyShotManifestCompilation(&payload)
		if status != http.StatusServiceUnavailable || err == nil {
			t.Fatalf("status=%d error=%v, want service unavailable", status, err)
		}
	})

	t.Run("project missing", func(t *testing.T) {
		workflow := &GenerationService{shotManifestCompiler: &fakeShotManifestCompiler{}}
		payload := generationMessageRequest{ShotManifestID: "shot-1", Prompt: "fallback must not run"}
		status, err := workflow.applyShotManifestCompilation(&payload)
		if status != http.StatusBadRequest || err == nil {
			t.Fatalf("status=%d error=%v, want bad request", status, err)
		}
	})

	t.Run("compiler error", func(t *testing.T) {
		workflow := &GenerationService{shotManifestCompiler: &fakeShotManifestCompiler{err: errors.New("shot not found")}}
		payload := generationMessageRequest{ProjectID: "project-a", ShotManifestID: "shot-missing", Prompt: "fallback must not run"}
		status, err := workflow.applyShotManifestCompilation(&payload)
		if status != http.StatusBadRequest || err == nil {
			t.Fatalf("status=%d error=%v, want bad request", status, err)
		}
	})
}
