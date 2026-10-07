package generation

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	coregeneration "github.com/mediago-dev/mediago-drama/packages/core/pkg/generation"
)

func TestReserveIdempotentGenerationTask(t *testing.T) {
	repo, err := newTestGenerationTaskRepository(t, filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatalf("newTestGenerationTaskRepository() error = %v", err)
	}
	service := NewGenerationTaskServiceFromRepository(repo, nil, nil)
	task := GenerationTaskRecord{
		ID:       "generation-idem-test",
		Kind:     string(coregeneration.KindVideo),
		RouteID:  coregeneration.RouteDMXSeedance20Fast,
		Provider: coregeneration.ProviderDMX,
		Status:   "queued",
		Params: generationParamsWithIdempotency(
			map[string]any{"duration": "5"},
			"video-idempotency-key-0001",
			"hash-one",
		),
	}

	first, created, err := service.ReserveIdempotent(task)
	if err != nil || !created || first.ID != task.ID {
		t.Fatalf("first reservation = %+v created=%v err=%v", first, created, err)
	}
	second, created, err := service.ReserveIdempotent(task)
	if err != nil || created || second.ID != task.ID {
		t.Fatalf("second reservation = %+v created=%v err=%v", second, created, err)
	}

	conflict := task
	conflict.Params = generationParamsWithIdempotency(
		map[string]any{"duration": "10"},
		"video-idempotency-key-0001",
		"hash-two",
	)
	if _, _, err := service.ReserveIdempotent(conflict); !errors.Is(err, ErrGenerationIdempotencyConflict) {
		t.Fatalf("conflict error = %v, want ErrGenerationIdempotencyConflict", err)
	}
}

func TestReserveIdempotentGenerationTaskAcrossServiceInstances(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "settings.db")
	firstRepo, err := newTestGenerationTaskRepository(t, dbPath)
	if err != nil {
		t.Fatalf("first repository error = %v", err)
	}
	secondRepo, err := newTestGenerationTaskRepository(t, dbPath)
	if err != nil {
		t.Fatalf("second repository error = %v", err)
	}
	firstService := NewGenerationTaskServiceFromRepository(firstRepo, nil, nil)
	secondService := NewGenerationTaskServiceFromRepository(secondRepo, nil, nil)
	task := GenerationTaskRecord{
		ID:       "generation-idem-cross-service",
		Kind:     string(coregeneration.KindVideo),
		RouteID:  coregeneration.RouteDMXSeedance20Fast,
		Provider: coregeneration.ProviderDMX,
		Status:   "queued",
		Params: generationParamsWithIdempotency(
			map[string]any{"duration": "5"},
			"video-idempotency-key-cross-service",
			"hash-cross-service",
		),
	}

	if _, created, err := firstService.ReserveIdempotent(task); err != nil || !created {
		t.Fatalf("first reserve created=%v err=%v", created, err)
	}
	second, created, err := secondService.ReserveIdempotent(task)
	if err != nil || created {
		t.Fatalf("second reserve = %+v created=%v err=%v, want existing", second, created, err)
	}
	if second.ID != task.ID {
		t.Fatalf("second task id = %q, want %q", second.ID, task.ID)
	}
}

func TestGenerationInternalIdempotencyParamsAreNotExposedOrSent(t *testing.T) {
	params := generationParamsWithIdempotency(
		map[string]any{"duration": "5"},
		"video-idempotency-key-0001",
		"hash-one",
	)
	for _, filtered := range []map[string]any{generationParamsForClient(params), providerGenerationParams(params)} {
		if _, ok := filtered[generationIdempotencyKeyParam]; ok {
			t.Fatalf("filtered params leaked idempotency key: %#v", filtered)
		}
		if _, ok := filtered[generationRequestHashParam]; ok {
			t.Fatalf("filtered params leaked request hash: %#v", filtered)
		}
		if filtered["duration"] != "5" {
			t.Fatalf("filtered params lost public value: %#v", filtered)
		}
	}
}

func TestReliabilityClassifiesAmbiguousSubmissionFailures(t *testing.T) {
	if !isAmbiguousProviderSubmissionError(context.DeadlineExceeded) {
		t.Fatal("deadline exceeded should be treated as ambiguous submission")
	}
	if !isAmbiguousProviderSubmissionError(&coregeneration.HTTPError{StatusCode: 503, Retryable: true}) {
		t.Fatal("503 provider error should be ambiguous")
	}
	if isAmbiguousProviderSubmissionError(&coregeneration.HTTPError{StatusCode: 400, Retryable: false}) {
		t.Fatal("400 non-retryable provider error should not be ambiguous")
	}
	unknown := UncertainSubmissionGenerationResponse("task-1", "", context.DeadlineExceeded)
	if unknown.Status != GenerationStatusUnknown ||
		unknown.ErrorCode != "provider_submission_unknown" ||
		unknown.ErrorType != "provider_unknown" ||
		unknown.Retryable {
		t.Fatalf("unknown response = %+v", unknown)
	}
	reconciling := UncertainSubmissionGenerationResponse("task-2", "provider-2", context.DeadlineExceeded)
	if reconciling.Status != GenerationStatusReconciling || reconciling.Retryable {
		t.Fatalf("reconciling response = %+v", reconciling)
	}
}
