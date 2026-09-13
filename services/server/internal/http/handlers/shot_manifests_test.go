package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	serviceshotmanifest "github.com/mediago-dev/mediago-drama/services/server/internal/service/shotmanifest"
)

type fakeShotManifestStore struct {
	listProjectID  string
	listDocumentID string
	upsertInput    serviceshotmanifest.UpsertInput
	compileProject string
	compileShotID  string
	listResponse   []serviceshotmanifest.Record
	upsertResponse serviceshotmanifest.Record
	compileResult  serviceshotmanifest.CompileResult
	syncResponse   serviceshotmanifest.ProjectSyncSummary
	syncProjectID  string
}

func (fake *fakeShotManifestStore) ListDocument(projectID string, documentID string) ([]serviceshotmanifest.Record, error) {
	fake.listProjectID = projectID
	fake.listDocumentID = documentID
	return fake.listResponse, nil
}

func (fake *fakeShotManifestStore) SyncProject(projectID string) (serviceshotmanifest.ProjectSyncSummary, error) {
	fake.syncProjectID = projectID
	return fake.syncResponse, nil
}

func (fake *fakeShotManifestStore) Upsert(input serviceshotmanifest.UpsertInput) (serviceshotmanifest.Record, error) {
	fake.upsertInput = input
	return fake.upsertResponse, nil
}

func (fake *fakeShotManifestStore) CompileAndPersist(projectID string, id string) (serviceshotmanifest.CompileResult, error) {
	fake.compileProject = projectID
	fake.compileShotID = id
	return fake.compileResult, nil
}

func TestShotManifestHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeShotManifestStore{
		listResponse: []serviceshotmanifest.Record{{
			ID:         "shot-1",
			ProjectID:  "project-a",
			DocumentID: "storyboard-ep1",
			SectionID:  "section-shot-1",
			Sequence:   1,
			Status:     serviceshotmanifest.StatusReady,
		}},
		upsertResponse: serviceshotmanifest.Record{
			ID:                "shot-1",
			ProjectID:         "project-a",
			DocumentID:        "storyboard-ep1",
			SectionID:         "section-shot-1",
			Sequence:          1,
			ResolvedStateJSON: `{"characters":{"char-a":{"look":{"wardrobe":"wedding"}}}}`,
			Status:            serviceshotmanifest.StatusDraft,
		},
		compileResult: serviceshotmanifest.CompileResult{
			DocumentID:        "storyboard-ep1",
			SectionID:         "section-shot-1",
			Prompt:            "compiled prompt",
			ReferenceAssetIDs: []string{"asset-a"},
		},
		syncResponse: serviceshotmanifest.ProjectSyncSummary{Created: 2, Reused: 1},
	}
	handler := NewShotManifests(store)
	router := gin.New()
	router.GET("/projects/:projectId/shot-manifests", handler.HandleList)
	router.POST("/projects/:projectId/shot-manifests/sync", handler.HandleSync)
	router.PUT("/projects/:projectId/shot-manifests", handler.HandleUpsert)
	router.POST("/projects/:projectId/shot-manifests/:shotId/compile", handler.HandleCompile)

	t.Run("list", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/projects/project-a/shot-manifests?documentId=storyboard-ep1", nil))
		if recorder.Code != http.StatusOK {
			t.Fatalf("list status=%d body=%s", recorder.Code, recorder.Body.String())
		}
		if store.listProjectID != "project-a" || store.listDocumentID != "storyboard-ep1" {
			t.Fatalf("list args=%q/%q", store.listProjectID, store.listDocumentID)
		}
	})

	t.Run("sync", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/projects/project-a/shot-manifests/sync", nil))
		if recorder.Code != http.StatusOK {
			t.Fatalf("sync status=%d body=%s", recorder.Code, recorder.Body.String())
		}
		if store.syncProjectID != "project-a" {
			t.Fatalf("sync project=%q, want project-a", store.syncProjectID)
		}
		var envelope struct {
			Data serviceshotmanifest.ProjectSyncSummary `json:"data"`
		}
		if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
			t.Fatalf("decode sync response: %v", err)
		}
		if envelope.Data.Created != 2 || envelope.Data.Reused != 1 {
			t.Fatalf("sync response=%+v", envelope.Data)
		}
	})

	t.Run("upsert", func(t *testing.T) {
		body := bytes.NewBufferString(`{
			"documentId":"storyboard-ep1",
			"sectionId":"section-shot-1",
			"shotKey":"shot-1",
			"sequence":1,
			"actionText":"向前走",
			"bindings":{"characters":[{"canonId":"char-a","variantId":"look-wedding"}]},
			"stateChanges":{"characters":{"char-a":{"look":{"wardrobe":"wedding"}}}}
		}`)
		request := httptest.NewRequest(http.MethodPut, "/projects/project-a/shot-manifests", body)
		request.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("upsert status=%d body=%s", recorder.Code, recorder.Body.String())
		}
		if store.upsertInput.ProjectID != "project-a" || store.upsertInput.DocumentID != "storyboard-ep1" || store.upsertInput.SectionID != "section-shot-1" {
			t.Fatalf("upsert identity=%+v", store.upsertInput)
		}
		if store.upsertInput.Bindings.Characters[0].CanonID != "char-a" || store.upsertInput.Bindings.Characters[0].VariantID != "look-wedding" {
			t.Fatalf("upsert bindings=%+v", store.upsertInput.Bindings)
		}
		var changes map[string]any
		if err := json.Unmarshal([]byte(store.upsertInput.StateChangesJSON), &changes); err != nil {
			t.Fatalf("stateChangesJson invalid: %v", err)
		}
	})

	t.Run("compile", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/projects/project-a/shot-manifests/shot-1/compile", nil))
		if recorder.Code != http.StatusOK {
			t.Fatalf("compile status=%d body=%s", recorder.Code, recorder.Body.String())
		}
		if store.compileProject != "project-a" || store.compileShotID != "shot-1" {
			t.Fatalf("compile args=%q/%q", store.compileProject, store.compileShotID)
		}
	})
}
