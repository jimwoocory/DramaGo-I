package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mediago-dev/mediago-drama/services/server/internal/domain"
	servicecanon "github.com/mediago-dev/mediago-drama/services/server/internal/service/canon"
)

type fakeCanonStore struct {
	listProjectID     string
	listResourceType  string
	syncProjectID     string
	variantInput      servicecanon.VariantInput
	referenceInput    servicecanon.ReferenceInput
	statusProjectID   string
	statusCanonID     string
	statusValue       string
	listResponse      []servicecanon.AssetRecord
	syncResponse      servicecanon.ProjectSyncSummary
	variantResponse   domain.CanonAssetModel
	referenceResponse servicecanon.ReferenceRecord
	getResponse       servicecanon.AssetRecord
	statusResponse    servicecanon.AssetRecord
}

func (fake *fakeCanonStore) List(projectID string, resourceType string) ([]servicecanon.AssetRecord, error) {
	fake.listProjectID = projectID
	fake.listResourceType = resourceType
	return fake.listResponse, nil
}

func (fake *fakeCanonStore) Get(projectID string, id string) (servicecanon.AssetRecord, error) {
	return fake.getResponse, nil
}

func (fake *fakeCanonStore) SyncProject(projectID string) (servicecanon.ProjectSyncSummary, error) {
	fake.syncProjectID = projectID
	return fake.syncResponse, nil
}

func (fake *fakeCanonStore) CreateVariant(input servicecanon.VariantInput) (domain.CanonAssetModel, error) {
	fake.variantInput = input
	return fake.variantResponse, nil
}

func (fake *fakeCanonStore) BindReference(input servicecanon.ReferenceInput) (servicecanon.ReferenceRecord, error) {
	fake.referenceInput = input
	return fake.referenceResponse, nil
}

func (fake *fakeCanonStore) UpdateStatus(projectID string, id string, status string) (servicecanon.AssetRecord, error) {
	fake.statusProjectID = projectID
	fake.statusCanonID = id
	fake.statusValue = status
	return fake.statusResponse, nil
}

func TestCanonHandlersExposeProjectCanonLifecycle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeCanonStore{
		listResponse: []servicecanon.AssetRecord{{
			ID:           "canon-character-lintong",
			ProjectID:    "project-a",
			ResourceType: "character",
			ResourceID:   "section-lintong",
			Name:         "林书彤",
			Status:       servicecanon.StatusApproved,
			Version:      2,
			References: []servicecanon.ReferenceRecord{{
				ID:      "canon-ref-1",
				AssetID: "asset-1",
				Role:    "identity",
			}},
		}},
		syncResponse: servicecanon.ProjectSyncSummary{
			Resources:         servicecanon.ResourceSyncSummary{Created: 2, Skipped: 1},
			ReferencesCreated: 1,
		},
		variantResponse: domain.CanonAssetModel{ID: "canon-variant-wedding"},
		referenceResponse: servicecanon.ReferenceRecord{
			ID: "canon-ref-wedding", AssetID: "asset-wedding", Role: "look", Priority: 900, Locked: true,
		},
		getResponse: servicecanon.AssetRecord{
			ID:          "canon-variant-wedding",
			ProjectID:   "project-a",
			ParentID:    "canon-character-lintong",
			VariantKind: "look",
			Name:        "婚礼造型",
			Status:      servicecanon.StatusApproved,
			Version:     1,
		},
		statusResponse: servicecanon.AssetRecord{
			ID:        "canon-character-lintong",
			ProjectID: "project-a",
			Status:    servicecanon.StatusLocked,
			Version:   2,
		},
	}
	handler := NewCanon(store, func(error) bool { return false })
	router := gin.New()
	router.GET("/projects/:projectId/canon", handler.HandleList)
	router.POST("/projects/:projectId/canon/sync", handler.HandleSync)
	router.POST("/projects/:projectId/canon/:canonId/variants", handler.HandleCreateVariant)
	router.POST("/projects/:projectId/canon/:canonId/references", handler.HandleBindReference)
	router.PATCH("/projects/:projectId/canon/:canonId/status", handler.HandleUpdateStatus)

	t.Run("list", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/projects/project-a/canon?resourceType=character", nil))
		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200: %s", recorder.Code, recorder.Body.String())
		}
		if store.listProjectID != "project-a" || store.listResourceType != "character" {
			t.Fatalf("list args = %q/%q", store.listProjectID, store.listResourceType)
		}
		var envelope struct {
			Success bool                       `json:"success"`
			Data    []servicecanon.AssetRecord `json:"data"`
		}
		if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
			t.Fatalf("decode list response: %v", err)
		}
		if !envelope.Success || len(envelope.Data) != 1 || envelope.Data[0].References[0].AssetID != "asset-1" {
			t.Fatalf("list response = %+v", envelope)
		}
	})

	t.Run("sync", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/projects/project-a/canon/sync", nil))
		if recorder.Code != http.StatusOK || store.syncProjectID != "project-a" {
			t.Fatalf("sync status=%d project=%q body=%s", recorder.Code, store.syncProjectID, recorder.Body.String())
		}
		var envelope struct {
			Data servicecanon.ProjectSyncSummary `json:"data"`
		}
		if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
			t.Fatalf("decode sync response: %v", err)
		}
		if envelope.Data.Resources.Created != 2 || envelope.Data.ReferencesCreated != 1 {
			t.Fatalf("sync response = %+v", envelope.Data)
		}
	})

	t.Run("variant", func(t *testing.T) {
		body := bytes.NewBufferString(`{"variantKind":"look","name":"婚礼造型","specJson":"{\"wardrobe\":\"wedding\"}","promptText":"白色婚纱，盘发","status":"approved"}`)
		request := httptest.NewRequest(http.MethodPost, "/projects/project-a/canon/canon-character-lintong/variants", body)
		request.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("variant status=%d body=%s", recorder.Code, recorder.Body.String())
		}
		if store.variantInput.ProjectID != "project-a" || store.variantInput.ParentID != "canon-character-lintong" || store.variantInput.VariantKind != "look" || store.variantInput.Status != "approved" {
			t.Fatalf("variant input = %+v", store.variantInput)
		}
	})

	t.Run("reference", func(t *testing.T) {
		body := bytes.NewBufferString(`{"assetId":"asset-wedding","role":"look","priority":900,"locked":true}`)
		request := httptest.NewRequest(http.MethodPost, "/projects/project-a/canon/canon-variant-wedding/references", body)
		request.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("reference status=%d body=%s", recorder.Code, recorder.Body.String())
		}
		if store.referenceInput.ProjectID != "project-a" || store.referenceInput.CanonAssetID != "canon-variant-wedding" || store.referenceInput.AssetID != "asset-wedding" || store.referenceInput.Role != "look" || store.referenceInput.Priority != 900 || !store.referenceInput.Locked {
			t.Fatalf("reference input = %+v", store.referenceInput)
		}
	})

	t.Run("status", func(t *testing.T) {
		body := bytes.NewBufferString(`{"status":"locked"}`)
		request := httptest.NewRequest(http.MethodPatch, "/projects/project-a/canon/canon-character-lintong/status", body)
		request.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("status update code=%d body=%s", recorder.Code, recorder.Body.String())
		}
		if store.statusProjectID != "project-a" || store.statusCanonID != "canon-character-lintong" || store.statusValue != "locked" {
			t.Fatalf("status args = %q/%q/%q", store.statusProjectID, store.statusCanonID, store.statusValue)
		}
	})
}
