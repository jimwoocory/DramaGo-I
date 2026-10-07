package usvdsv11

import (
	"context"
	"strings"
	"testing"

	"github.com/mediago-dev/mediago-drama/packages/instructions/pkg/pack"
)

func TestSnapshotLoadsPinnedV11Skills(t *testing.T) {
	bundle, err := Snapshot(context.Background())
	if err != nil {
		t.Fatalf("Snapshot() error = %v", err)
	}
	if bundle.Manifest.ID != PackID || bundle.Manifest.Version != Version {
		t.Fatalf("manifest = %+v", bundle.Manifest)
	}
	skills := map[string]bool{}
	entries := map[string]pack.Entry{}
	for _, entry := range bundle.Entries {
		if entry.Kind == pack.KindSkill {
			skills[entry.Slug] = true
			entries[entry.Slug] = entry
		}
	}
	for _, name := range []string{
		"usvd-v10-controller",
		"usvd-v10-00-intake-adaptation",
		"usvd-v10-01-story-architect",
		"usvd-v10-02-episode-architect",
		"usvd-v10-03-creator-script-draft",
		"usvd-v10-04-review-continuity",
		"us-vertical-drama-storyboard-director",
	} {
		if !skills[name] {
			t.Fatalf("missing V11 skill %q", name)
		}
	}
	if len(skills) != 14 {
		t.Fatalf("skill count = %d, want 14", len(skills))
	}
	if !strings.Contains(entries["usvd-v10-03-creator-script-draft"].Raw, "document_category: screenplay") {
		t.Fatal("creator script draft must target DramaGo screenplay documents")
	}
	if !strings.Contains(entries["us-vertical-drama-storyboard-director"].Raw, "document_category: storyboard") {
		t.Fatal("storyboard director must target DramaGo storyboard documents")
	}
	if !strings.Contains(entries["usvd-v10-controller"].Body, "DramaGo runtime integration contract") ||
		!strings.Contains(entries["usvd-v10-controller"].Body, "Never create shadow USVDS") {
		t.Fatal("controller is missing the DramaGo ownership contract")
	}
	for name, tag := range map[string]string{
		"usvd-v10-01-story-architect":      "usvds:artifact:story-package",
		"usvd-v10-02-episode-architect":    "usvds:artifact:episode-architecture",
		"usvd-v10-03-creator-script-draft": "usvds:artifact:screenplay",
	} {
		if !strings.Contains(entries[name].Body, tag) {
			t.Fatalf("skill %s missing authoritative artifact tag %s", name, tag)
		}
	}
}

func TestSourceMatchesPinnedSnapshot(t *testing.T) {
	source, err := Source()
	if err != nil {
		t.Fatalf("Source() error = %v", err)
	}
	if source.Commit != Commit || source.PluginVersion != Version || source.SkillCount != 14 {
		t.Fatalf("source = %+v", source)
	}
}
