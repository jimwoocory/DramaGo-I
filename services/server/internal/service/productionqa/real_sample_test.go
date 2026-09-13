package productionqa

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

// TestRealRegressionSample validates the original project when the developer
// supplies MEDIAGO_REGRESSION_SAMPLE_PROJECT. CI can run the portable fixture
// without requiring a private/local project directory.
func TestRealRegressionSample(t *testing.T) {
	projectDir := os.Getenv("MEDIAGO_REGRESSION_SAMPLE_PROJECT")
	if projectDir == "" {
		t.Skip("MEDIAGO_REGRESSION_SAMPLE_PROJECT is not configured")
	}
	workDir := filepath.Join(projectDir, "work")
	read := func(name string) string {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(workDir, name))
		if err != nil {
			t.Fatalf("reading %s: %v", name, err)
		}
		return string(data)
	}

	report := AuditLegacyDocuments(
		read("第一集 陷阵木勺.md"),
		read("第一集 场景设定.md"),
		read("第一集 道具设定.md"),
		read("第一集 分镜.md"),
	)
	if report.TargetDurationSeconds != 720 {
		t.Fatalf("real sample target duration = %.1f, want 720", report.TargetDurationSeconds)
	}
	if math.Abs(report.StoryboardDurationSeconds-236.8) > 0.001 {
		t.Fatalf("real sample storyboard duration = %.2f, want 236.8", report.StoryboardDurationSeconds)
	}
	if report.StoryboardGroupCount != 16 {
		t.Fatalf("real sample group count = %d, want 16", report.StoryboardGroupCount)
	}

	codes := map[string]bool{}
	for _, finding := range report.Findings {
		codes[finding.Code] = true
	}
	for _, code := range []string{
		CodeDurationCompression,
		CodeSceneCanonConflict,
		CodeRepeatedGlobalBoilerplate,
		CodeUnsupportedPropUsage,
	} {
		if !codes[code] {
			t.Fatalf("real sample findings = %+v, missing %s", report.Findings, code)
		}
	}
}
