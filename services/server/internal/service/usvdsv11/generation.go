package usvdsv11

import (
	"fmt"
	"strings"

	serviceshotmanifest "github.com/mediago-dev/mediago-drama/services/server/internal/service/shotmanifest"
)

// ValidateGenerationCompileResult enforces the deterministic V11 generation
// preconditions on an existing DramaGo ShotManifest compilation result.
func ValidateGenerationCompileResult(compiled serviceshotmanifest.CompileResult) error {
	blockers := []string{}
	if strings.TrimSpace(compiled.ShotManifestID) == "" {
		blockers = append(blockers, "ShotManifest is required")
	}
	if strings.TrimSpace(compiled.Status) != serviceshotmanifest.StatusReady {
		blockers = append(blockers, "ShotManifest status must be ready")
	}
	if strings.TrimSpace(compiled.ResolvedStateJSON) == "" {
		blockers = append(blockers, "continuity state must be resolved")
	}
	if strings.TrimSpace(compiled.Prompt) == "" {
		blockers = append(blockers, "compiled prompt is required")
	}
	if len(blockers) > 0 {
		return fmt.Errorf("USVDS V11 generation gate blocked: %s", strings.Join(blockers, "; "))
	}
	return nil
}
