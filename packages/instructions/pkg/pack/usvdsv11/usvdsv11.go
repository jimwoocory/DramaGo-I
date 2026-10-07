// Package usvdsv11 exposes the pinned USVDS V11 skill snapshot used by DramaGo.
package usvdsv11

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"

	"github.com/mediago-dev/mediago-drama/packages/instructions/pkg/pack"
)

const (
	PackID  = "usvds-v11"
	Version = "1.2.6"
	Commit  = "0c68d9cbf525eee6f2a98d3098294cdedbb29e77"
)

//go:embed assets
var assets embed.FS

type Provenance struct {
	SourceRepository string `json:"source_repository"`
	Branch           string `json:"branch"`
	Commit           string `json:"commit"`
	PluginVersion    string `json:"plugin_version"`
	SkillCount       int    `json:"skill_count"`
	Compilation      string `json:"compilation"`
}

func FS() embed.FS {
	return assets
}

func Snapshot(ctx context.Context) (pack.Bundle, error) {
	packFS, err := fs.Sub(assets, "assets")
	if err != nil {
		return pack.Bundle{}, fmt.Errorf("opening USVDS V11 pack assets: %w", err)
	}
	bundle, err := pack.ParseFS(ctx, packFS)
	if err != nil {
		return pack.Bundle{}, fmt.Errorf("loading USVDS V11 prompt pack: %w", err)
	}
	if bundle.Manifest.ID != PackID || bundle.Manifest.Version != Version {
		return pack.Bundle{}, fmt.Errorf(
			"USVDS V11 snapshot identity mismatch: %s@%s",
			bundle.Manifest.ID,
			bundle.Manifest.Version,
		)
	}
	return bundle, nil
}

func Source() (Provenance, error) {
	data, err := assets.ReadFile("assets/SOURCE.json")
	if err != nil {
		return Provenance{}, fmt.Errorf("reading USVDS V11 provenance: %w", err)
	}
	var provenance Provenance
	if err := json.Unmarshal(data, &provenance); err != nil {
		return Provenance{}, fmt.Errorf("parsing USVDS V11 provenance: %w", err)
	}
	if provenance.Commit != Commit || provenance.PluginVersion != Version {
		return Provenance{}, fmt.Errorf(
			"USVDS V11 provenance mismatch: %s@%s",
			provenance.Commit,
			provenance.PluginVersion,
		)
	}
	return provenance, nil
}
