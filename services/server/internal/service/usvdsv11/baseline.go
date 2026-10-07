package usvdsv11

// Baseline identifies the exact USVDS V11 source reviewed for this adapter.
// Branch names are not sufficient because the V11 branch contains historical
// V9/V10 material and generated distribution surfaces.
type Baseline struct {
	Branch        string
	Commit        string
	PluginVersion string
}

var CurrentBaseline = Baseline{
	Branch:        "v11",
	Commit:        "0c68d9cbf525eee6f2a98d3098294cdedbb29e77",
	PluginVersion: "1.2.6",
}

// Ownership documents which existing DramaGo aggregate owns each V11 concept.
// These values are intentionally descriptive: this package does not introduce
// persistence models or repositories.
type Ownership struct {
	Concept string
	Owner   string
}

func OwnershipMap() []Ownership {
	return []Ownership{
		{Concept: "project", Owner: "WorkspaceProject"},
		{Concept: "story_screenplay", Owner: "Document"},
		{Concept: "character_scene_prop_identity", Owner: "CanonAsset"},
		{Concept: "look_state_variant", Owner: "CanonAsset.ParentID"},
		{Concept: "reference_media", Owner: "Asset+CanonReference"},
		{Concept: "shot_execution_contract", Owner: "ShotManifest"},
		{Concept: "continuity_state", Owner: "ShotManifest.ResolvedStateJSON"},
		{Concept: "compiled_prompt", Owner: "ShotManifest.CompiledPrompt"},
		{Concept: "generation_execution", Owner: "GenerationTask"},
		{Concept: "generated_output", Owner: "GenerationTaskAsset+Asset"},
	}
}
