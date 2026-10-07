package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	httpresponse "github.com/mediago-dev/mediago-drama/services/server/internal/http/response"
	serviceusvdsv11 "github.com/mediago-dev/mediago-drama/services/server/internal/service/usvdsv11"
)

// USVDSV11GateStore supplies project-level V11 readiness.
type USVDSV11GateStore interface {
	EvaluateProject(projectID string) (serviceusvdsv11.ProjectGateReport, error)
}

// USVDSV11Gates handles project V11 gate routes.
type USVDSV11Gates struct {
	store USVDSV11GateStore
}

// NewUSVDSV11Gates returns a V11 gate handler.
func NewUSVDSV11Gates(store USVDSV11GateStore) USVDSV11Gates {
	return USVDSV11Gates{store: store}
}

// HandleGet godoc
// @Summary 获取 USVDS V11 Gate 状态
// @Description 基于 DramaGo 现有 Document/Canon/ShotManifest/GenerationTask 真相计算项目 Gate，不创建平行 USVDS 数据。
// @Tags USVDS V11
// @Produce json
// @Param projectId path string true "Project ID"
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/v1/projects/{projectId}/usvds-v11/gates [get]
func (handler USVDSV11Gates) HandleGet(context *gin.Context) {
	projectID, ok := requiredProjectID(context)
	if !ok {
		return
	}
	if handler.store == nil {
		httpresponse.Error(context, http.StatusServiceUnavailable, "USVDS V11 Gate 服务尚未配置")
		return
	}
	report, err := handler.store.EvaluateProject(projectID)
	if err != nil {
		httpresponse.Fail(context, http.StatusInternalServerError, "internal error", err)
		return
	}
	httpresponse.OK(context, report)
}
