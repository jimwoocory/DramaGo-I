package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	serviceusvdsv11 "github.com/mediago-dev/mediago-drama/services/server/internal/service/usvdsv11"
)

type fakeUSVDSV11GateStore struct {
	projectID string
	report    serviceusvdsv11.ProjectGateReport
	err       error
}

func (fake *fakeUSVDSV11GateStore) EvaluateProject(projectID string) (serviceusvdsv11.ProjectGateReport, error) {
	fake.projectID = projectID
	return fake.report, fake.err
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
}
