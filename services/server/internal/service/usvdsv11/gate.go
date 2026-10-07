package usvdsv11

import "strings"

type GateID string

const (
	GateStoryApproved      GateID = "story_approved"
	GateScreenplayReviewed GateID = "screenplay_reviewed"
	GateCanonLocked        GateID = "canon_locked"
	GateContinuityResolved GateID = "continuity_resolved"
	GateStoryboardReady    GateID = "storyboard_ready"
	GateGenerationReady    GateID = "generation_ready"
)

// Snapshot is a non-persistent projection of existing DramaGo state.
// IDs point to current DramaGo entities; no USVDS shadow entities are created.
type Snapshot struct {
	ProjectID          string
	DocumentID         string
	StoryApproved      bool
	ScreenplayReviewed bool
	CanonApproved      bool
	ContinuityResolved bool

	ShotManifestID string
	ShotStatus     string
	CompiledPrompt string

	GenerationTaskID string
}

// GateResult reports deterministic readiness for one V11 gate.
type GateResult struct {
	Gate     GateID
	Ready    bool
	Blockers []string
}

// Evaluate returns the deterministic P0 V11 gate state from DramaGo-owned
// records. Model quality review remains in workflow/skill layers.
func Evaluate(snapshot Snapshot) []GateResult {
	results := make([]GateResult, 0, 6)
	results = append(results, result(GateStoryApproved,
		blockIf(strings.TrimSpace(snapshot.ProjectID) == "", "project is required"),
		blockIf(strings.TrimSpace(snapshot.DocumentID) == "", "document is required"),
		blockIf(!snapshot.StoryApproved, "story approval is required"),
	))
	results = append(results, result(GateScreenplayReviewed,
		blockIf(!snapshot.StoryApproved, "story approval is required"),
		blockIf(!snapshot.ScreenplayReviewed, "screenplay review is required"),
	))
	results = append(results, result(GateCanonLocked,
		blockIf(!snapshot.ScreenplayReviewed, "screenplay review is required"),
		blockIf(!snapshot.CanonApproved, "Canon approval is required"),
	))
	results = append(results, result(GateContinuityResolved,
		blockIf(!snapshot.CanonApproved, "Canon approval is required"),
		blockIf(!snapshot.ContinuityResolved, "continuity resolution is required"),
	))
	results = append(results, result(GateStoryboardReady,
		blockIf(!snapshot.ContinuityResolved, "continuity resolution is required"),
		blockIf(strings.TrimSpace(snapshot.ShotManifestID) == "", "ShotManifest is required"),
		blockIf(strings.TrimSpace(snapshot.ShotStatus) != "ready", "ShotManifest status must be ready"),
	))
	results = append(results, result(GateGenerationReady,
		blockIf(strings.TrimSpace(snapshot.ShotManifestID) == "", "ShotManifest is required"),
		blockIf(strings.TrimSpace(snapshot.ShotStatus) != "ready", "ShotManifest status must be ready"),
		blockIf(strings.TrimSpace(snapshot.CompiledPrompt) == "", "compiled prompt is required"),
	))
	return results
}

func result(gate GateID, candidates ...string) GateResult {
	blockers := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate != "" {
			blockers = append(blockers, candidate)
		}
	}
	return GateResult{Gate: gate, Ready: len(blockers) == 0, Blockers: blockers}
}

func blockIf(condition bool, message string) string {
	if condition {
		return message
	}
	return ""
}
