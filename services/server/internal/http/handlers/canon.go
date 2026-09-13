package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mediago-dev/mediago-drama/services/server/internal/domain"
	httpresponse "github.com/mediago-dev/mediago-drama/services/server/internal/http/response"
	servicecanon "github.com/mediago-dev/mediago-drama/services/server/internal/service/canon"
)

// CanonStore supplies project Canon operations.
type CanonStore interface {
	List(projectID string, resourceType string) ([]servicecanon.AssetRecord, error)
	Get(projectID string, id string) (servicecanon.AssetRecord, error)
	SyncProject(projectID string) (servicecanon.ProjectSyncSummary, error)
	CreateVariant(input servicecanon.VariantInput) (domain.CanonAssetModel, error)
	BindReference(input servicecanon.ReferenceInput) (servicecanon.ReferenceRecord, error)
	UpdateStatus(projectID string, id string, status string) (servicecanon.AssetRecord, error)
}

// Canon handles project Canon/Variant/Reference HTTP routes.
type Canon struct {
	store      CanonStore
	isNotFound func(error) bool
}

// NewCanon returns a Canon route handler.
func NewCanon(store CanonStore, isNotFound func(error) bool) Canon {
	return Canon{store: store, isNotFound: isNotFound}
}

type createCanonVariantRequest struct {
	VariantKind string `json:"variantKind"`
	Name        string `json:"name"`
	SpecJSON    string `json:"specJson"`
	PromptText  string `json:"promptText"`
	Status      string `json:"status"`
}

type bindCanonReferenceRequest struct {
	AssetID  string `json:"assetId"`
	Role     string `json:"role"`
	Priority int    `json:"priority"`
	Locked   bool   `json:"locked"`
}

type updateCanonStatusRequest struct {
	Status string `json:"status"`
}

// HandleList godoc
// @Summary 获取项目 Canon 资产
// @Description 返回项目中的角色、场景、道具 Canon Core/Variant 及其参考资产。
// @Tags Canon
// @Produce json
// @Param projectId path string true "Project ID"
// @Param resourceType query string false "character/scene/prop"
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/v1/projects/{projectId}/canon [get]
func (handler Canon) HandleList(context *gin.Context) {
	projectID, ok := requiredProjectID(context)
	if !ok {
		return
	}
	records, err := handler.store.List(projectID, strings.TrimSpace(context.Query("resourceType")))
	if err != nil {
		httpresponse.ErrorFromStatus(context, http.StatusBadRequest, err)
		return
	}
	httpresponse.OK(context, records)
}

// HandleSync godoc
// @Summary 同步项目 Canon
// @Description 从角色、场景、道具文档资源惰性同步 Canon，并桥接当前已选生成资产。
// @Tags Canon
// @Produce json
// @Param projectId path string true "Project ID"
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/v1/projects/{projectId}/canon/sync [post]
func (handler Canon) HandleSync(context *gin.Context) {
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

// HandleCreateVariant godoc
// @Summary 创建 Canon Variant
// @Description 在角色、场景或道具 Canon 下创建造型/时间/状态等变体。
// @Tags Canon
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param canonId path string true "Canon ID"
// @Param payload body SwaggerObject true "Variant payload"
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 404 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/v1/projects/{projectId}/canon/{canonId}/variants [post]
func (handler Canon) HandleCreateVariant(context *gin.Context) {
	projectID, ok := requiredProjectID(context)
	if !ok {
		return
	}
	canonID, ok := requiredPathParam(context, "canonId", "canonId")
	if !ok {
		return
	}
	payload, err := decodeJSON[createCanonVariantRequest](context)
	if err != nil {
		httpresponse.ErrorFromStatus(context, http.StatusBadRequest, err)
		return
	}
	variant, err := handler.store.CreateVariant(servicecanon.VariantInput{
		ProjectID:   projectID,
		ParentID:    canonID,
		VariantKind: payload.VariantKind,
		Name:        payload.Name,
		SpecJSON:    payload.SpecJSON,
		PromptText:  payload.PromptText,
		Status:      payload.Status,
	})
	if err != nil {
		if handler.matchesNotFound(err) {
			httpresponse.Error(context, http.StatusNotFound, "Canon 不存在")
			return
		}
		httpresponse.ErrorFromStatus(context, http.StatusBadRequest, err)
		return
	}
	record, err := handler.store.Get(projectID, variant.ID)
	if err != nil {
		httpresponse.Fail(context, http.StatusInternalServerError, "internal error", err)
		return
	}
	httpresponse.OK(context, record)
}

// HandleBindReference godoc
// @Summary 绑定 Canon 参考资产
// @Description 将当前项目已有媒体资产绑定到 Canon Core/Variant，并设置 role/priority/locked。
// @Tags Canon
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param canonId path string true "Canon ID"
// @Param payload body SwaggerObject true "Reference payload"
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 404 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/v1/projects/{projectId}/canon/{canonId}/references [post]
func (handler Canon) HandleBindReference(context *gin.Context) {
	projectID, ok := requiredProjectID(context)
	if !ok {
		return
	}
	canonID, ok := requiredPathParam(context, "canonId", "canonId")
	if !ok {
		return
	}
	payload, err := decodeJSON[bindCanonReferenceRequest](context)
	if err != nil {
		httpresponse.ErrorFromStatus(context, http.StatusBadRequest, err)
		return
	}
	reference, err := handler.store.BindReference(servicecanon.ReferenceInput{
		ProjectID:    projectID,
		CanonAssetID: canonID,
		AssetID:      payload.AssetID,
		Role:         payload.Role,
		Priority:     payload.Priority,
		Locked:       payload.Locked,
	})
	if err != nil {
		if handler.matchesNotFound(err) {
			httpresponse.Error(context, http.StatusNotFound, "Canon 或参考资产不存在")
			return
		}
		httpresponse.ErrorFromStatus(context, http.StatusBadRequest, err)
		return
	}
	httpresponse.OK(context, reference)
}

// HandleUpdateStatus godoc
// @Summary 更新 Canon 状态
// @Description 显式设置 draft/approved/locked/deprecated；locked 状态不会被文档惰性同步覆盖。
// @Tags Canon
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param canonId path string true "Canon ID"
// @Param payload body SwaggerObject true "Status payload"
// @Success 200 {object} SwaggerEnvelope
// @Failure 400 {object} SwaggerEnvelope
// @Failure 404 {object} SwaggerEnvelope
// @Failure 500 {object} SwaggerEnvelope
// @Router /api/v1/projects/{projectId}/canon/{canonId}/status [patch]
func (handler Canon) HandleUpdateStatus(context *gin.Context) {
	projectID, ok := requiredProjectID(context)
	if !ok {
		return
	}
	canonID, ok := requiredPathParam(context, "canonId", "canonId")
	if !ok {
		return
	}
	payload, err := decodeJSON[updateCanonStatusRequest](context)
	if err != nil {
		httpresponse.ErrorFromStatus(context, http.StatusBadRequest, err)
		return
	}
	record, err := handler.store.UpdateStatus(projectID, canonID, payload.Status)
	if err != nil {
		if handler.matchesNotFound(err) {
			httpresponse.Error(context, http.StatusNotFound, "Canon 不存在")
			return
		}
		httpresponse.ErrorFromStatus(context, http.StatusBadRequest, err)
		return
	}
	httpresponse.OK(context, record)
}

func (handler Canon) matchesNotFound(err error) bool {
	return handler.isNotFound != nil && handler.isNotFound(err)
}
