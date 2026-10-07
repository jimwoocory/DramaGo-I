package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	serviceusvdsv11 "github.com/mediago-dev/mediago-drama/services/server/internal/service/usvdsv11"
)

type fakeUSVDSV11GateStore struct {
	projectID        string
	gate             serviceusvdsv11.GateID
	documentID       string
	expectedVersion  int
	artifact         serviceusvdsv11.ArtifactKind
	mutation         serviceusvdsv11.GateMutationResult
	artifactMutation serviceusvdsv11.ArtifactMutationResult
	report           serviceusvdsv11.ProjectGateReport
	err              error
}

func (fake *fakeUSVDSV11GateStore) EvaluateProject(projectID string) (serviceusvdsv11.ProjectGateReport, error) {
	fake.projectID = projectID
	return fake.report, fake.err
}

func (fake *fakeUSVDSV11GateStore) ApproveGate(projectID string, gate serviceusvdsv11.GateID, documentID string, expectedVersion int) (serviceusvdsv11.GateMutationResult, error) {
	fake.projectID = projectID
	fake.gate = gate
	fake.documentID = documentID
	fake.expectedVersion = expectedVersion
	return fake.mutation, fake.err
}

func (fake *fakeUSVDSV11GateStore) RevokeGate(projectID string, gate serviceusvdsv11.GateID, documentID string, expectedVersion int) (serviceusvdsv11.GateMutationResult, error) {
	return fake.ApproveGate(projectID, gate, documentID, expectedVersion)
}

func (fake *fakeUSVDSV11GateStore) AdoptArtifact(projectID string, artifact serviceusvdsv11.ArtifactKind, documentID string, expectedVersion int) (serviceusvdsv11.ArtifactMutationResult, error) {
	fake.projectID = projectID
	fake.artifact = artifact
	fake.documentID = documentID
	fake.expectedVersion = expectedVersion
	return fake.artifactMutation, fake.err
}

func TestUSVDSV11GateHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &fakeUSVDSV11GateStore{report: serviceusvdsv11.ProjectGateReport{
		ProjectID: "project-a",
		Baseline:  serviceusvdsv11.CurrentBaseline,
		Gates: []serviceusvdsv11.GateResult{{
			Gate:  serviceusvdsv11.GateGenerationReady,
			Ready: true,
		}},
	}}
	handler := NewUSVDSV11Gates(store)
	router := gin.New()
	router.GET("/projects/:projectId/usvds-v11/gates", handler.HandleGet)
	router.POST("/projects/:projectId/usvds-v11/gates/:gate/approve", handler.HandleApprove)
	router.POST("/projects/:projectId/usvds-v11/gates/:gate/revoke", handler.HandleRevoke)
	router.POST("/projects/:projectId/usvds-v11/artifacts/:artifact/adopt", handler.HandleAdoptArtifact)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/projects/project-a/usvds-v11/gates", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if store.projectID != "project-a" {
		t.Fatalf("project id = %q", store.projectID)
	}
	var envelope struct {
		Data serviceusvdsv11.ProjectGateReport `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if envelope.Data.Baseline.Commit != serviceusvdsv11.CurrentBaseline.Commit {
		t.Fatalf("baseline = %+v", envelope.Data.Baseline)
	}

	store.mutation = serviceusvdsv11.GateMutationResult{Report: store.report}
	approve := httptest.NewRecorder()
	approveRequest := httptest.NewRequest(
		http.MethodPost,
		"/projects/project-a/usvds-v11/gates/story_approved/approve",
		strings.NewReader(`{"documentId":"story-1","expectedVersion":7}`),
	)
	approveRequest.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(approve, approveRequest)
	if approve.Code != http.StatusOK {
		t.Fatalf("approve status=%d body=%s", approve.Code, approve.Body.String())
	}
	if store.projectID != "project-a" || store.gate != serviceusvdsv11.GateStoryApproved || store.documentID != "story-1" || store.expectedVersion != 7 {
		t.Fatalf("approve args project=%q gate=%q document=%q version=%d", store.projectID, store.gate, store.documentID, store.expectedVersion)
	}

	revoke := httptest.NewRecorder()
	revokeRequest := httptest.NewRequest(
		http.MethodPost,
		"/projects/project-a/usvds-v11/gates/screenplay_reviewed/revoke",
		strings.NewReader(`{"documentId":"script-1","expectedVersion":9}`),
	)
	revokeRequest.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(revoke, revokeRequest)
	if revoke.Code != http.StatusOK {
		t.Fatalf("revoke status=%d body=%s", revoke.Code, revoke.Body.String())
	}
	if store.gate != serviceusvdsv11.GateScreenplayReviewed || store.documentID != "script-1" || store.expectedVersion != 9 {
		t.Fatalf("revoke args gate=%q document=%q version=%d", store.gate, store.documentID, store.expectedVersion)
	}

	store.artifactMutation = serviceusvdsv11.ArtifactMutationResult{Report: store.report}
	adopt := httptest.NewRecorder()
	adoptRequest := httptest.NewRequest(
		http.MethodPost,
		"/projects/project-a/usvds-v11/artifacts/screenplay/adopt",
		strings.NewReader(`{"documentId":"script-2","expectedVersion":4}`),
	)
	adoptRequest.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(adopt, adoptRequest)
	if adopt.Code != http.StatusOK {
		t.Fatalf("adopt status=%d body=%s", adopt.Code, adopt.Body.String())
	}
	if store.artifact != serviceusvdsv11.ArtifactScreenplay || store.documentID != "script-2" || store.expectedVersion != 4 {
		t.Fatalf("adopt args artifact=%q document=%q version=%d", store.artifact, store.documentID, store.expectedVersion)
	}
}
