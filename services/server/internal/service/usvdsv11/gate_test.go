package usvdsv11

import "testing"

func TestCurrentBaselinePinsReviewedV11(t *testing.T) {
	if CurrentBaseline.Branch != "v11" {
		t.Fatalf("branch = %q", CurrentBaseline.Branch)
	}
	if CurrentBaseline.Commit != "0c68d9cbf525eee6f2a98d3098294cdedbb29e77" {
		t.Fatalf("commit = %q", CurrentBaseline.Commit)
	}
	if CurrentBaseline.PluginVersion != "1.2.6" {
		t.Fatalf("plugin version = %q", CurrentBaseline.PluginVersion)
	}
}

func TestOwnershipMapUsesDramaGoAggregates(t *testing.T) {
	got := OwnershipMap()
	if len(got) == 0 {
		t.Fatal("ownership map is empty")
	}
	for _, item := range got {
		if item.Owner == "USVDSProject" || item.Owner == "USVDSAsset" || item.Owner == "USVDSShot" {
			t.Fatalf("shadow owner introduced: %+v", item)
		}
	}
}

func TestEvaluateBlocksGenerationUntilExistingDramaGoStateIsReady(t *testing.T) {
	results := Evaluate(Snapshot{
		ProjectID:  "project-1",
		DocumentID: "doc-1",
	})
	last := results[len(results)-1]
	if last.Gate != GateGenerationReady || last.Ready {
		t.Fatalf("generation gate = %+v", last)
	}
	if len(last.Blockers) == 0 {
		t.Fatal("expected generation blockers")
	}
}

func TestEvaluateAllowsReadyShotWithoutGenerationShadowRecord(t *testing.T) {
	results := Evaluate(Snapshot{
		ProjectID:           "project-1",
		DocumentID:          "doc-1",
		StoryApproved:       true,
		ScreenplayReviewed:  true,
		CanonApproved:       true,
		ContinuityResolved:  true,
		ShotManifestID:      "shot-1",
		ShotStatus:          "ready",
		CompiledPrompt:      "prompt",
	})
	for _, result := range results {
		if !result.Ready {
			t.Fatalf("gate %s blocked: %v", result.Gate, result.Blockers)
		}
	}
}
