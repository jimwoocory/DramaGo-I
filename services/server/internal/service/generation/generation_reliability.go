package generation

import (
	"errors"
	"fmt"
	"strings"

	coregeneration "github.com/mediago-dev/mediago-drama/packages/core/pkg/generation"
	"github.com/mediago-dev/mediago-drama/services/server/internal/service/shared"
)

const (
	GenerationStatusUnknown     = "unknown"
	GenerationStatusReconciling = "reconciling"
)

func isGenerationSubmissionUncertainStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case GenerationStatusUnknown, GenerationStatusReconciling:
		return true
	default:
		return false
	}
}

func isAmbiguousProviderSubmissionError(err error) bool {
	if err == nil {
		return false
	}
	if isGenerationTimeoutError(err) {
		return true
	}
	var httpErr *coregeneration.HTTPError
	if errors.As(err, &httpErr) {
		if httpErr.StatusCode >= 500 || httpErr.Retryable {
			return true
		}
	}
	if failure, ok := coregeneration.FailureFromError(err); ok {
		return failure.Retryable &&
			(failure.Reason == coregeneration.FailureProviderError ||
				failure.Reason == coregeneration.FailureTimeout)
	}
	return false
}

func UncertainSubmissionGenerationResponse(id string, providerTaskID string, err error) GenerationMessageResponse {
	failure := GenerationFailureDetailsFromError(err)
	raw := generationSafeFailureError(failure)
	status := GenerationStatusUnknown
	message := "供应商提交结果不确定；为避免重复生成，系统不会自动重新提交。"
	if strings.TrimSpace(providerTaskID) != "" {
		status = GenerationStatusReconciling
		message = "供应商提交结果不确定，已保留任务 ID 并进入核对状态；系统不会重复提交。"
	}
	return GenerationMessageResponse{
		ID:        ValueOrFallback(id, shared.MustRandomID("generation")),
		Role:      "assistant",
		Status:    status,
		Message:   message,
		Assets:    []GenerationAsset{},
		Usage:     GenerationUsage{},
		Error:     raw,
		ErrorCode: "provider_submission_unknown",
		ErrorType: "provider_unknown",
		Retryable: false,
	}
}

func generationReconcileRequiredError(task GenerationTaskRecord) error {
	return fmt.Errorf(
		"generation task %s has uncertain provider submission state %q; reconcile provider status before retrying",
		task.ID,
		task.Status,
	)
}
