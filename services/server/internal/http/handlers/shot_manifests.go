package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	httpresponse "github.com/mediago-dev/mediago-drama/services/server/internal/http/response"
	serviceshotmanifest "github.com/mediago-dev/mediago-drama/services/server/internal/service/shotmanifest"
)

// ShotManifestStore supplies structured storyboard execution operations.
type ShotManifestStore interface {
	ListDocument(projectID string, documentID string) ([]serviceshotmanifest.Record, error)
	SyncProject(projectID string) (serviceshotmanifest.ProjectSyncSummary, error)
	Upsert(input serviceshotmanifest.UpsertInput) (serviceshotmanifest.Record, error)
	CompileAndPersist(projectID string, id string) (serviceshotmanifest.CompileResult, error)
}

// ShotManifests handles project shot-manifest HTTP routes.
type ShotManifests struct {
	store ShotManifestStore
}

// NewShotManifests returns a shot-manifest handler.
func NewShotManifests(store ShotManifestStore) ShotManifests {
	return ShotManifests{store: store}
}

type upsertShotManifestRequest struct {
	DocumentID     string                       `json:"documentId"`
	SectionID      string                       `json:"sectionId"`
	ShotKey        string                       `json:"shotKey,omitempty"`
	Sequence       int                          `json:"sequence"`
	InheritsFromID string                       `json:"inheritsFromId,omitempty"`
	ActionText     string                       `json:"actionText,omitempty"`
	CameraText     string                       `json:"cameraText,omitempty"`
	AudioText      string                       `json:"audioText,omitempty"`
	StyleProfileID string                       `json:"styleProfileId,omitempty"`
	Bindings       serviceshotmanifest.Bindings `json:"bindings"`
	StateChanges   json.RawMessage              `json:"stateChanges,omitempty"`
	SourceHash     string                       `json:"sourceHash,omitempty"`
	Status         string                       `json:"status,omitempty"`
}

// HandleList godoc
// @Summary 获取 Shot Manifest
// @Description 返回指定 storyboard 文档的结构化镜头执行清单。
// @Tags Shot Manifest
// @Produce json
// @Param projectId path string true "Project ID"
// @Param documentId query string true "Storyboard Document ID"
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/v1/projects/{projectId}/shot-manifests [get]
func (handler ShotManifests) HandleList(context *gin.Context) {
	projectID, ok := requiredProjectID(context)
	if !ok {
		return
	}
	documentID := strings.TrimSpace(context.Query("documentId"))
	if documentID == "" {
		httpresponse.Error(context, http.StatusBadRequest, "缺少 documentId")
		return
	}
	records, err := handler.store.ListDocument(projectID, documentID)
	if err != nil {
		httpresponse.ErrorFromStatus(context, http.StatusBadRequest, err)
		return
	}
	httpresponse.OK(context, records)
}

// HandleSync godoc
// @Summary 同步旧分镜为 Shot Manifest
// @Description 不改写 Markdown，仅为缺失的 storyboard 分组惰性创建 draft Shot Manifest，并同步 Canon/资源绑定。
// @Tags Shot Manifest
// @Produce json
// @Param projectId path string true "Project ID"
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/v1/projects/{projectId}/shot-manifests/sync [post]
func (handler ShotManifests) HandleSync(context *gin.Context) {
	projectID, ok := requiredProjectID(context)
	if !ok {
		return
	}
	summary, err := handler.store.SyncProject(projectID)
	if err != nil {
		httpresponse.Fail(context, http.StatusInternalServerError, "internal error", err)
		return
	}
	httpresponse.OK(context, summary)
}

// HandleUpsert godoc
// @Summary 保存 Shot Manifest
// @Description 保存/更新一条镜头执行合同，并自动解析上一镜连续性状态与 Canon 绑定。
// @Tags Shot Manifest
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param payload body SwaggerObject true "Shot Manifest payload"
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/v1/projects/{projectId}/shot-manifests [put]
func (handler ShotManifests) HandleUpsert(context *gin.Context) {
	projectID, ok := requiredProjectID(context)
	if !ok {
		return
	}
	payload, err := decodeJSON[upsertShotManifestRequest](context)
	if err != nil {
		httpresponse.ErrorFromStatus(context, http.StatusBadRequest, err)
		return
	}
	stateChanges := strings.TrimSpace(string(payload.StateChanges))
	if stateChanges == "" || stateChanges == "null" {
		stateChanges = "{}"
	}
	record, err := handler.store.Upsert(serviceshotmanifest.UpsertInput{
		ProjectID:        projectID,
		DocumentID:       payload.DocumentID,
		SectionID:        payload.SectionID,
		ShotKey:          payload.ShotKey,
		Sequence:         payload.Sequence,
		InheritsFromID:   payload.InheritsFromID,
		ActionText:       payload.ActionText,
		CameraText:       payload.CameraText,
		AudioText:        payload.AudioText,
		StyleProfileID:   payload.StyleProfileID,
		Bindings:         payload.Bindings,
		StateChangesJSON: stateChanges,
		SourceHash:       payload.SourceHash,
		Status:           payload.Status,
	})
	if err != nil {
		httpresponse.ErrorFromStatus(context, http.StatusBadRequest, err)
		return
	}
	httpresponse.OK(context, record)
}

// HandleCompile godoc
// @Summary 编译 Shot Manifest
// @Description 使用 Canon/Variant/Resolved State 确定性编译最终生成提示词和参考资产，不调用 LLM。
// @Tags Shot Manifest
// @Produce json
// @Param projectId path string true "Project ID"
// @Param shotId path string true "Shot Manifest ID"
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/v1/projects/{projectId}/shot-manifests/{shotId}/compile [post]
func (handler ShotManifests) HandleCompile(context *gin.Context) {
	projectID, ok := requiredProjectID(context)
	if !ok {
		return
	}
	shotID, ok := requiredPathParam(context, "shotId", "shotId")
	if !ok {
		return
	}
	compiled, err := handler.store.CompileAndPersist(projectID, shotID)
	if err != nil {
		httpresponse.ErrorFromStatus(context, http.StatusBadRequest, err)
		return
	}
	httpresponse.OK(context, compiled)
}
