package generation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	coregeneration "github.com/mediago-dev/mediago-drama/packages/core/pkg/generation"
	"github.com/mediago-dev/mediago-drama/services/server/internal/platform/timestamp"
)

const (
	generationIdempotencyKeyParam  = "_mediago_idempotency_key"
	generationRequestHashParam     = "_mediago_request_hash"
	minGenerationIdempotencyLength = 16
	maxGenerationIdempotencyLength = 128
)

var ErrGenerationIdempotencyConflict = errors.New("generation idempotency conflict")

type generationIdempotencyConflictError struct {
	Key string
}

func (err generationIdempotencyConflictError) Error() string {
	return fmt.Sprintf("%s: key %q was already used for another generation request", ErrGenerationIdempotencyConflict, err.Key)
}

func (err generationIdempotencyConflictError) Unwrap() error {
	return ErrGenerationIdempotencyConflict
}

func generationIdempotencyTaskID(scope string, key string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(scope) + "\x00" + strings.TrimSpace(key)))
	return "generation-idem-" + hex.EncodeToString(sum[:16])
}

func generationIdempotencyScope(request GenerationMessageRequest) string {
	if projectID := strings.TrimSpace(request.ProjectID); projectID != "" {
		return "project:" + projectID
	}
	if conversationID := strings.TrimSpace(request.ConversationID); conversationID != "" {
		return "conversation:" + conversationID
	}
	return "global"
}

func validateGenerationIdempotencyKey(key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil
	}
	if len(key) < minGenerationIdempotencyLength || len(key) > maxGenerationIdempotencyLength {
		return fmt.Errorf(
			"idempotencyKey must be %d..%d characters",
			minGenerationIdempotencyLength,
			maxGenerationIdempotencyLength,
		)
	}
	return nil
}

func generationSemanticRequestHash(
	request GenerationMessageRequest,
	route coregeneration.ModelRoute,
	resolvedReferenceURLs []string,
) (string, error) {
	params := providerGenerationParams(request.Params)
	type semanticRequest struct {
		ProjectID         string
		ConversationID    string
		DocumentID        string
		SectionID         string
		ShotManifestID    string
		CapabilityID      string
		ResourceType      string
		RouteID           string
		VersionID         string
		Provider          string
		ModelID           string
		Model             string
		Prompt            string
		ReferenceURLs     []string
		ReferenceAssetIDs []string
		Params            map[string]any
		SourceRefs        []ContentSourceRef
	}
	payload := semanticRequest{
		ProjectID:         strings.TrimSpace(request.ProjectID),
		ConversationID:    strings.TrimSpace(request.ConversationID),
		DocumentID:        strings.TrimSpace(request.DocumentID),
		SectionID:         strings.TrimSpace(request.SectionID),
		ShotManifestID:    strings.TrimSpace(request.ShotManifestID),
		CapabilityID:      strings.TrimSpace(request.CapabilityID),
		ResourceType:      strings.TrimSpace(request.ResourceType),
		RouteID:           route.ID,
		VersionID:         route.VersionID,
		Provider:          route.Provider,
		ModelID:           strings.TrimSpace(request.ModelID),
		Model:             strings.TrimSpace(request.Model),
		Prompt:            request.Prompt,
		ReferenceURLs:     append([]string(nil), resolvedReferenceURLs...),
		ReferenceAssetIDs: CompactStrings(request.ReferenceAssetIDs),
		Params:            params,
		SourceRefs:        append([]ContentSourceRef(nil), request.SourceRefs...),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("encoding generation idempotency request: %w", err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func generationParamsWithIdempotency(params map[string]any, key string, requestHash string) map[string]any {
	next := make(map[string]any, len(params)+2)
	for name, value := range params {
		next[name] = value
	}
	if key = strings.TrimSpace(key); key != "" {
		next[generationIdempotencyKeyParam] = key
	}
	if requestHash = strings.TrimSpace(requestHash); requestHash != "" {
		next[generationRequestHashParam] = requestHash
	}
	return next
}

func generationIdempotencyMetadata(task GenerationTaskRecord) (key string, requestHash string) {
	if task.Params == nil {
		return "", ""
	}
	if value, ok := task.Params[generationIdempotencyKeyParam].(string); ok {
		key = strings.TrimSpace(value)
	}
	if value, ok := task.Params[generationRequestHashParam].(string); ok {
		requestHash = strings.TrimSpace(value)
	}
	return key, requestHash
}

func (service *GenerationTaskService) ReserveIdempotent(task GenerationTaskRecord) (GenerationTaskRecord, bool, error) {
	if service == nil {
		return GenerationTaskRecord{}, false, errors.New("generation task service is nil")
	}
	if service.initErr != nil {
		return GenerationTaskRecord{}, false, service.initErr
	}
	key, requestHash := generationIdempotencyMetadata(task)
	if key == "" || requestHash == "" {
		if err := service.Upsert(task); err != nil {
			return GenerationTaskRecord{}, false, err
		}
		return task, true, nil
	}

	now := timestamp.NowRFC3339Nano()
	if strings.TrimSpace(task.CreatedAt) == "" {
		task.CreatedAt = now
	}
	task.UpdatedAt = now
	if task.ReferenceURLs == nil {
		task.ReferenceURLs = []string{}
	}
	if task.ReferenceAssetIDs == nil {
		task.ReferenceAssetIDs = []string{}
	}
	if task.SourceRefs == nil {
		task.SourceRefs = []ContentSourceRef{}
	}
	if task.Params == nil {
		task.Params = map[string]any{}
	}
	if task.Assets == nil {
		task.Assets = []GenerationAsset{}
	}
	paramsJSON, err := json.Marshal(task.Params)
	if err != nil {
		return GenerationTaskRecord{}, false, err
	}
	sourceRefsJSON, err := json.Marshal(task.SourceRefs)
	if err != nil {
		return GenerationTaskRecord{}, false, err
	}

	service.idempotencyMu.Lock()
	defer service.idempotencyMu.Unlock()

	service.mu.Lock()
	if err := service.ensureTaskConversationLocked(task); err != nil {
		service.mu.Unlock()
		return GenerationTaskRecord{}, false, err
	}
	created, err := service.repo.CreateGenerationTaskIfAbsent(
		generationTaskPersistenceModel(task, string(paramsJSON), string(sourceRefsJSON)),
	)
	if err != nil {
		service.mu.Unlock()
		return GenerationTaskRecord{}, false, err
	}
	if created {
		if err := service.syncNormalizedTaskReferenceRowsLocked(task); err != nil {
			service.mu.Unlock()
			return GenerationTaskRecord{}, false, err
		}
		if err := service.syncNormalizedTaskAssetRowsLocked(task); err != nil {
			service.mu.Unlock()
			return GenerationTaskRecord{}, false, err
		}
		service.mu.Unlock()
		if IsActiveGenerationStatus(task.Status) && service.onTaskStarted != nil {
			service.onTaskStarted(task)
		}
		return task, true, nil
	}
	model, err := service.repo.GetGenerationTask(task.ID)
	service.mu.Unlock()
	if err != nil {
		return GenerationTaskRecord{}, false, err
	}
	existing, err := generationTaskRecordFromModel(model)
	if err != nil {
		return GenerationTaskRecord{}, false, err
	}
	existingKey, existingHash := generationIdempotencyMetadata(existing)
	if existingKey != key || existingHash != requestHash {
		return GenerationTaskRecord{}, false, generationIdempotencyConflictError{Key: key}
	}
	if err := service.attachAttemptSummary(&existing); err != nil {
		return GenerationTaskRecord{}, false, err
	}
	return existing, false, nil
}

func (workflow *GenerationService) reserveGenerationIdempotency(
	request *GenerationMessageRequest,
	route coregeneration.ModelRoute,
	resolvedReferenceURLs []string,
) (GenerationTaskRecord, bool, error) {
	if request == nil {
		return GenerationTaskRecord{}, false, errors.New("generation request is nil")
	}
	key := strings.TrimSpace(request.IdempotencyKey)
	if key == "" {
		return GenerationTaskRecord{}, false, nil
	}
	if err := validateGenerationIdempotencyKey(key); err != nil {
		return GenerationTaskRecord{}, false, err
	}
	requestHash, err := generationSemanticRequestHash(*request, route, resolvedReferenceURLs)
	if err != nil {
		return GenerationTaskRecord{}, false, err
	}
	taskID := generationIdempotencyTaskID(generationIdempotencyScope(*request), key)
	request.ReservedTaskID = taskID
	request.Params = generationParamsWithIdempotency(request.Params, key, requestHash)
	reservedResponse := GenerationMessageResponse{
		ID:      taskID,
		Role:    "assistant",
		Status:  "queued",
		Message: "生成请求已预留，等待提交到模型服务。",
		Assets:  []GenerationAsset{},
		Usage:   GenerationUsage{},
	}
	task := GenerationTaskFromMessage(*request, route, reservedResponse)
	return workflow.generationTasks.ReserveIdempotent(task)
}

func generationResponseWithReservedTaskID(
	request GenerationMessageRequest,
	response GenerationMessageResponse,
) GenerationMessageResponse {
	if taskID := strings.TrimSpace(request.ReservedTaskID); taskID != "" {
		response.ID = taskID
	}
	return response
}
