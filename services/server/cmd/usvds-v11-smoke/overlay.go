package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const maxSmokeScreenplayBytes = 2 * 1024 * 1024

// screenplayOverlayReport proves that an external source was exercised on the
// disposable workspace. It deliberately contains no screenplay prose or path.
type screenplayOverlayReport struct {
	Mode               string `json:"mode"`
	SHA256             string `json:"sha256"`
	Bytes              int    `json:"bytes"`
	ApprovalProhibited bool   `json:"approvalProhibited"`
}

// copyUnapprovedScreenplayOverlay only writes into the temporary project copy.
// It never changes the user's source workspace or grants a V11 human approval.
func copyUnapprovedScreenplayOverlay(source string, snapshotProjectDir string) (*screenplayOverlayReport, error) {
	source = strings.TrimSpace(source)
	if source == "" {
		return nil, nil
	}
	if !strings.EqualFold(filepath.Ext(source), ".md") {
		return nil, fmt.Errorf("smoke screenplay overlay must be a Markdown file")
	}
	info, err := os.Lstat(source)
	if err != nil {
		return nil, fmt.Errorf("reading smoke screenplay overlay: %w", err)
	}
	if !info.Mode().IsRegular() || info.Size() > maxSmokeScreenplayBytes {
		return nil, fmt.Errorf("smoke screenplay overlay must be a regular file not exceeding 2 MiB")
	}
	raw, err := os.ReadFile(source)
	if err != nil {
		return nil, fmt.Errorf("reading screenplay overlay: %w", err)
	}
	if err := validateUnapprovedScreenplayOverlay(raw); err != nil {
		return nil, err
	}
	destination := filepath.Join(snapshotProjectDir, "work", "__p6_unapproved_screenplay_overlay__.md")
	if _, err := os.Lstat(destination); err == nil {
		return nil, fmt.Errorf("smoke destination already exists in project snapshot")
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
		return nil, err
	}
	if err := os.WriteFile(destination, raw, 0o600); err != nil {
		return nil, fmt.Errorf("writing screenplay to temporary snapshot: %w", err)
	}
	sum := sha256.Sum256(raw)
	return &screenplayOverlayReport{
		Mode:               "temporary_unapproved_screenplay",
		SHA256:             hex.EncodeToString(sum[:]),
		Bytes:              len(raw),
		ApprovalProhibited: true,
	}, nil
}

// validateUnapprovedScreenplayOverlay checks the minimal DramaGo frontmatter
// contract and rejects any attempt to smuggle human-approval metadata.
func validateUnapprovedScreenplayOverlay(raw []byte) error {
	normalized := strings.ReplaceAll(string(raw), "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	if len(lines) < 5 || strings.TrimSpace(lines[0]) != "---" {
		return fmt.Errorf("smoke screenplay requires Markdown YAML frontmatter")
	}
	end := -1
	for i := 1; i < len(lines) && i < 200; i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			break
		}
	}
	if end < 0 {
		return fmt.Errorf("smoke screenplay frontmatter was not closed")
	}
	frontmatter := strings.Join(lines[1:end], "\n")
	categoryFound := false
	artifactTagFound := false
	for _, line := range lines[1:end] {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "category:") {
			category := strings.Trim(strings.TrimSpace(strings.TrimPrefix(trimmed, "category:")), "'\"")
			categoryFound = strings.EqualFold(category, "screenplay")
		}
		if trimmed == "- usvds:artifact:screenplay" || trimmed == "- 'usvds:artifact:screenplay'" || trimmed == "- \"usvds:artifact:screenplay\"" {
			artifactTagFound = true
		}
	}
	lower := strings.ToLower(frontmatter)
	for _, forbidden := range []string{"usvds:approval:", "usvds:story-approved", "usvds:screenplay-reviewed"} {
		if strings.Contains(lower, forbidden) {
			return fmt.Errorf("external smoke screenplay must not carry approval tag %q", forbidden)
		}
	}
	if !categoryFound || !artifactTagFound {
		return fmt.Errorf("external screenplay must declare category: screenplay and tag usvds:artifact:screenplay")
	}
	return nil
}
