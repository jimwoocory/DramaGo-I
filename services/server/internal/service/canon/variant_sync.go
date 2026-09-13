package canon

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/mediago-dev/mediago-drama/services/server/internal/domain"
	"github.com/mediago-dev/mediago-drama/services/server/internal/repository"
)

var canonMarkdownHeadingPattern = regexp.MustCompile(`^(#{1,6})\s+(.+?)\s*$`)

type sourceVariant struct {
	Kind       string
	Name       string
	Markdown   string
	PromptText string
	SourceHash string
}

type sourceResourceParts struct {
	CoreMarkdown string
	CorePrompt   string
	Variants     []sourceVariant
}

type VariantSourceInput struct {
	ProjectID   string
	ParentID    string
	VariantKind string
	Name        string
	SpecJSON    string
	PromptText  string
	SourceHash  string
}

// EnsureVariantFromSource creates or refreshes one H3-derived Canon Variant.
// Its identity is stable for parent + kind + name, and a locked Variant is never
// overwritten by subsequent document synchronization.
func (service *Service) EnsureVariantFromSource(input VariantSourceInput) (SyncResult, error) {
	if service.initErr != nil {
		return SyncResult{}, service.initErr
	}
	input.ProjectID = domain.CleanProjectID(input.ProjectID)
	input.ParentID = strings.TrimSpace(input.ParentID)
	input.VariantKind = strings.ToLower(strings.TrimSpace(input.VariantKind))
	input.Name = strings.TrimSpace(input.Name)
	input.SpecJSON = normalizeSpecJSON(input.SpecJSON)
	input.PromptText = strings.TrimSpace(input.PromptText)
	input.SourceHash = strings.TrimSpace(input.SourceHash)
	if input.ProjectID == "" || input.ParentID == "" || input.VariantKind == "" || input.Name == "" {
		return SyncResult{}, errors.New("projectId, parentId, variantKind and name are required")
	}
	if input.SourceHash == "" {
		input.SourceHash = sourceHash(input.VariantKind, input.Name, input.SpecJSON, input.PromptText)
	}

	parent, err := service.repo.GetCanonAsset(input.ProjectID, input.ParentID)
	if err != nil {
		return SyncResult{}, err
	}
	existing, err := service.repo.FindCanonVariant(input.ProjectID, parent.ID, input.VariantKind, input.Name)
	if repository.IsRecordNotFound(err) {
		parentID := parent.ID
		model := domain.CanonAssetModel{
			ID:               variantIDForSource(parent.ID, input.VariantKind, input.Name),
			ProjectID:        parent.ProjectID,
			ResourceType:     parent.ResourceType,
			ResourceID:       parent.ResourceID,
			SourceDocumentID: parent.SourceDocumentID,
			ParentID:         &parentID,
			VariantKind:      input.VariantKind,
			Name:             input.Name,
			SpecJSON:         input.SpecJSON,
			PromptText:       input.PromptText,
			Status:           StatusDraft,
			Version:          1,
			SourceHash:       input.SourceHash,
		}
		if err := service.repo.CreateCanonAsset(model); err != nil {
			return SyncResult{}, err
		}
		created, err := service.repo.GetCanonAsset(input.ProjectID, model.ID)
		if err != nil {
			return SyncResult{}, err
		}
		return SyncResult{Asset: created, Created: true, Changed: true}, nil
	}
	if err != nil {
		return SyncResult{}, err
	}
	if existing.Status == StatusLocked {
		return SyncResult{Asset: existing, Locked: true}, nil
	}
	if existing.SourceHash == input.SourceHash {
		return SyncResult{Asset: existing}, nil
	}
	updated, err := service.repo.UpdateCanonAsset(input.ProjectID, existing.ID, map[string]any{
		"spec_json":   input.SpecJSON,
		"prompt_text": input.PromptText,
		"source_hash": input.SourceHash,
		"version":     existing.Version + 1,
	})
	if err != nil {
		return SyncResult{}, err
	}
	if !updated {
		return SyncResult{}, repository.ErrRecordNotFound
	}
	next, err := service.repo.GetCanonAsset(input.ProjectID, existing.ID)
	if err != nil {
		return SyncResult{}, err
	}
	return SyncResult{Asset: next, Changed: true}, nil
}

func splitSourceResource(resourceType string, markdown string, fallbackPrompt string) sourceResourceParts {
	lines := strings.Split(markdown, "\n")
	coreLines := make([]string, 0, len(lines))
	variants := []sourceVariant{}
	var activeKind string
	var activeName string
	activeLines := []string{}

	flushVariant := func() {
		if activeKind == "" || activeName == "" {
			activeKind, activeName = "", ""
			activeLines = nil
			return
		}
		variantMarkdown := strings.TrimSpace(strings.Join(activeLines, "\n"))
		variants = append(variants, sourceVariant{
			Kind:       activeKind,
			Name:       activeName,
			Markdown:   variantMarkdown,
			PromptText: canonPromptFromMarkdown(variantMarkdown),
			SourceHash: sourceHash(variantMarkdown),
		})
		activeKind, activeName = "", ""
		activeLines = nil
	}

	for _, line := range lines {
		match := canonMarkdownHeadingPattern.FindStringSubmatch(strings.TrimSpace(line))
		if len(match) == 3 {
			level := len(match[1])
			if activeKind != "" && level <= 3 {
				flushVariant()
			}
			if level == 3 {
				if kind, name, ok := canonVariantHeading(resourceType, match[2]); ok {
					activeKind = kind
					activeName = name
					activeLines = []string{line}
					continue
				}
			}
		}
		if activeKind != "" {
			activeLines = append(activeLines, line)
		} else {
			coreLines = append(coreLines, line)
		}
	}
	flushVariant()

	coreMarkdown := strings.TrimSpace(strings.Join(coreLines, "\n"))
	corePrompt := canonPromptFromMarkdown(coreMarkdown)
	if corePrompt == "" {
		corePrompt = strings.TrimSpace(fallbackPrompt)
	}
	return sourceResourceParts{CoreMarkdown: coreMarkdown, CorePrompt: corePrompt, Variants: variants}
}

func canonVariantHeading(resourceType string, heading string) (string, string, bool) {
	heading = strings.TrimSpace(heading)
	if heading == "" {
		return "", "", false
	}
	type rule struct {
		prefix string
		kind   string
	}
	var rules []rule
	switch strings.ToLower(strings.TrimSpace(resourceType)) {
	case ResourceTypeCharacter:
		rules = []rule{{"造型变体", "look"}, {"外观变体", "look"}, {"时期变体", "look"}, {"look variant", "look"}}
	case ResourceTypeScene:
		rules = []rule{{"场景变体", "scene"}, {"时间变体", "scene"}, {"scene variant", "scene"}, {"区域", "zone"}, {"zone", "zone"}}
	case ResourceTypeProp:
		rules = []rule{{"状态变体", "state"}, {"道具变体", "state"}, {"state variant", "state"}}
	default:
		return "", "", false
	}
	lower := strings.ToLower(heading)
	for _, candidate := range rules {
		prefix := strings.ToLower(candidate.prefix)
		if !strings.HasPrefix(lower, prefix) {
			continue
		}
		remainder := strings.TrimSpace(heading[len([]byte(candidate.prefix)):])
		remainder = strings.TrimSpace(strings.TrimLeft(remainder, ":：-—"))
		if remainder == "" {
			continue
		}
		return candidate.kind, remainder, true
	}
	return "", "", false
}

func canonPromptFromMarkdown(markdown string) string {
	lines := []string{}
	for _, raw := range strings.Split(markdown, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "<!--") && strings.Contains(strings.ToLower(line), "section-id") {
			continue
		}
		if canonMarkdownHeadingPattern.MatchString(line) {
			continue
		}
		if strings.HasPrefix(line, "![") {
			continue
		}
		line = strings.ReplaceAll(line, "**", "")
		lines = append(lines, line)
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func canonSourceSpec(prompt string) string {
	encoded, err := json.Marshal(map[string]any{
		"summary":    compactCanonSummary(prompt),
		"sourceText": strings.TrimSpace(prompt),
	})
	if err != nil {
		return "{}"
	}
	return string(encoded)
}

func compactCanonSummary(value string) string {
	runes := []rune(strings.TrimSpace(value))
	const limit = 180
	if len(runes) <= limit {
		return string(runes)
	}
	return string(runes[:limit]) + "…"
}

func variantIDForSource(parentID string, kind string, name string) string {
	digest := sha256Digest(strings.Join([]string{strings.TrimSpace(parentID), strings.TrimSpace(kind), strings.TrimSpace(name)}, "\x00"))
	return "canon-variant-" + digest[:24]
}

func sha256Digest(value string) string {
	return sourceHash(value)
}

func syncSummaryAdd(summary *ResourceSyncSummary, result SyncResult) {
	if summary == nil {
		return
	}
	summary.Assets = append(summary.Assets, result.Asset)
	switch {
	case result.Created:
		summary.Created++
	case result.Locked:
		summary.Locked++
	case result.Changed:
		summary.Updated++
	}
}

func syncSourceVariants(service *Service, projectID string, parentID string, variants []sourceVariant, summary *ResourceSyncSummary) error {
	for _, variant := range variants {
		result, err := service.EnsureVariantFromSource(VariantSourceInput{
			ProjectID:   projectID,
			ParentID:    parentID,
			VariantKind: variant.Kind,
			Name:        variant.Name,
			SpecJSON:    canonSourceSpec(variant.PromptText),
			PromptText:  variant.PromptText,
			SourceHash:  variant.SourceHash,
		})
		if err != nil {
			return fmt.Errorf("syncing Canon variant %s/%s: %w", variant.Kind, variant.Name, err)
		}
		syncSummaryAdd(summary, result)
	}
	return nil
}
