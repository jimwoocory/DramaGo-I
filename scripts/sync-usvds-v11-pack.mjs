import { execFileSync } from "node:child_process";
import fs from "node:fs";
import path from "node:path";
import process from "node:process";

const EXPECTED_COMMIT = "0c68d9cbf525eee6f2a98d3098294cdedbb29e77";
const EXPECTED_VERSION = "1.2.6";

const repoRoot = path.resolve(import.meta.dirname, "..");
const sourceRepo = path.resolve(
	process.env.USVDS_V11_DIR || path.join(repoRoot, "..", "US-Vertical-Drama-Studio-V11"),
);
const pluginRoot = path.join(sourceRepo, "plugins", "us-vertical-drama-studio");
const skillsRoot = path.join(pluginRoot, "skills");
const destination = path.join(
	repoRoot,
	"packages",
	"instructions",
	"pkg",
	"pack",
	"usvdsv11",
	"assets",
);
const destinationSkills = path.join(destination, "skills");

const titles = {
	"usvd-v10-controller": "USVDS V11 · 总控",
	"usvd-v10-00-intake-adaptation": "USVDS V11 · Intake / Adaptation",
	"usvd-v10-01-story-architect": "USVDS V11 · Story Architect",
	"usvd-v10-02-episode-architect": "USVDS V11 · Episode Architect",
	"usvd-v10-03-creator-script-draft": "USVDS V11 · Creator Script Draft",
	"usvd-v10-04-review-continuity": "USVDS V11 · Independent Review",
	"us-vertical-drama-studio": "USVDS V11 · Studio",
	"us-vertical-drama-adapter": "USVDS V11 · Adapter",
	"us-vertical-drama-showrunner": "USVDS V11 · Showrunner",
	"us-vertical-drama-episode-architect": "USVDS V11 · Legacy Episode Architect",
	"us-vertical-drama-screenwriter": "USVDS V11 · Legacy Screenwriter",
	"us-vertical-drama-script-doctor": "USVDS V11 · Script Doctor",
	"us-vertical-drama-continuity-editor": "USVDS V11 · Continuity Editor",
	"us-vertical-drama-storyboard-director": "USVDS V11 · Storyboard Director",
};

const stageHints = {
	"usvd-v10-controller": "controller",
	"usvd-v10-00-intake-adaptation": "intake",
	"usvd-v10-01-story-architect": "story",
	"usvd-v10-02-episode-architect": "episode",
	"usvd-v10-03-creator-script-draft": "screenplay",
	"usvd-v10-04-review-continuity": "review",
	"us-vertical-drama-studio": "controller",
	"us-vertical-drama-adapter": "intake",
	"us-vertical-drama-showrunner": "story",
	"us-vertical-drama-episode-architect": "episode",
	"us-vertical-drama-screenwriter": "screenplay",
	"us-vertical-drama-script-doctor": "review",
	"us-vertical-drama-continuity-editor": "continuity",
	"us-vertical-drama-storyboard-director": "storyboard",
};

const outputCategories = {
	"usvd-v10-03-creator-script-draft": "screenplay",
	"us-vertical-drama-screenwriter": "screenplay",
	"us-vertical-drama-storyboard-director": "storyboard",
};

const artifactTags = {
	"usvd-v10-01-story-architect": "usvds:artifact:story-package",
	"us-vertical-drama-showrunner": "usvds:artifact:story-package",
	"usvd-v10-02-episode-architect": "usvds:artifact:episode-architecture",
	"us-vertical-drama-episode-architect": "usvds:artifact:episode-architecture",
	"usvd-v10-03-creator-script-draft": "usvds:artifact:screenplay",
	"us-vertical-drama-screenwriter": "usvds:artifact:screenplay",
	"us-vertical-drama-storyboard-director": "usvds:artifact:storyboard",
};

const commit = execFileSync("git", ["-C", sourceRepo, "rev-parse", "HEAD"], {
	encoding: "utf8",
}).trim();
if (commit !== EXPECTED_COMMIT) {
	throw new Error(`USVDS V11 commit mismatch: expected ${EXPECTED_COMMIT}, got ${commit}`);
}
const plugin = JSON.parse(fs.readFileSync(path.join(pluginRoot, "plugin.json"), "utf8"));
if (plugin.version !== EXPECTED_VERSION) {
	throw new Error(
		`USVDS V11 plugin version mismatch: expected ${EXPECTED_VERSION}, got ${plugin.version}`,
	);
}

fs.rmSync(destination, { force: true, recursive: true });
fs.mkdirSync(destinationSkills, { recursive: true });
fs.writeFileSync(
	path.join(destination, "pack.json"),
	JSON.stringify(
		{
			id: "usvds-v11",
			name: "US Vertical Drama Studio V11",
			version: EXPECTED_VERSION,
			author: plugin.author?.name || "MediaGo Drama",
			description:
				"Read-only DramaGo skill snapshot compiled from USVDS v11@" +
				EXPECTED_COMMIT.slice(0, 7) +
				". References and contracts are inlined so the runtime remains self-contained.",
		},
		null,
		2,
	) + "\n",
);

const dependencyPattern =
	/(?:\.\.\/\.\.\/)?(?:references|contracts)\/[A-Za-z0-9_.-]+/g;

const skillDirs = fs
	.readdirSync(skillsRoot, { withFileTypes: true })
	.filter((entry) => entry.isDirectory() && fs.existsSync(path.join(skillsRoot, entry.name, "SKILL.md")))
	.map((entry) => entry.name)
	.sort();

const parseName = (raw) => {
	const match = raw.match(/^---\s*\r?\n([\s\S]*?)\r?\n---/);
	if (!match) throw new Error("missing YAML frontmatter");
	const nameMatch = match[1].match(/^name:\s*(.+?)\s*$/m);
	if (!nameMatch) throw new Error("missing skill name");
	return nameMatch[1].trim();
};

const injectMetadata = (raw, name) => {
	const end = raw.indexOf("\n---", 4);
	if (end < 0) throw new Error(`invalid frontmatter for ${name}`);
	const frontmatter = raw.slice(0, end);
	const rest = raw.slice(end);
	const additions = [];
	if (!/^title:/m.test(frontmatter)) {
		additions.push(`title: ${titles[name] || name}`);
	}
	if (outputCategories[name] && !/^document_category:/m.test(frontmatter)) {
		additions.push(`document_category: ${outputCategories[name]}`);
	}
	if (!/^hint:/m.test(frontmatter)) {
		additions.push("hint:");
		additions.push("  workflow: usvds-v11");
		additions.push(`  stage: ${stageHints[name] || "specialist"}`);
		additions.push(`  baseline: ${EXPECTED_COMMIT.slice(0, 7)}`);
	}
	return frontmatter + "\n" + additions.join("\n") + rest;
};

const findUniqueSkillReference = (fileName) => {
	const matches = [];
	for (const entry of fs.readdirSync(skillsRoot, { withFileTypes: true })) {
		if (!entry.isDirectory()) continue;
		const candidate = path.join(skillsRoot, entry.name, "references", fileName);
		if (fs.existsSync(candidate)) matches.push(candidate);
	}
	if (matches.length > 1) {
		throw new Error(`ambiguous V11 reference ${fileName}: ${matches.join(", ")}`);
	}
	return matches[0] || null;
};

const resolveDependency = (skillDir, token) => {
	if (token.startsWith("../../references/")) {
		return path.join(pluginRoot, "references", path.basename(token));
	}
	if (token.startsWith("../../contracts/")) {
		return path.join(pluginRoot, "contracts", path.basename(token));
	}
	if (token.startsWith("references/")) {
		const local = path.join(skillDir, token);
		if (fs.existsSync(local)) return local;
		return findUniqueSkillReference(path.basename(token));
	}
	if (token.startsWith("contracts/")) {
		return path.join(pluginRoot, token);
	}
	return null;
};

let emitted = 0;
for (const dirName of skillDirs) {
	const skillDir = path.join(skillsRoot, dirName);
	const sourcePath = path.join(skillDir, "SKILL.md");
	const sourceRaw = fs.readFileSync(sourcePath, "utf8").replace(/\r\n/g, "\n");
	const name = parseName(sourceRaw);
	let compiled = injectMetadata(sourceRaw, name).trimEnd();

	const dependencies = [...new Set(sourceRaw.match(dependencyPattern) || [])].sort();
	if (dependencies.length > 0) {
		compiled +=
			"\n\n---\n\n## DramaGo bundled V11 references\n\n" +
			"This section is generated from the exact pinned USVDS V11 snapshot. Treat it as supporting reference material; the skill rules above remain authoritative.\n";
		for (const token of dependencies) {
			const dependencyPath = resolveDependency(skillDir, token);
			if (!dependencyPath || !fs.existsSync(dependencyPath)) {
				throw new Error(`missing dependency for ${name}: ${token}`);
			}
			const content = fs.readFileSync(dependencyPath, "utf8").replace(/\r\n/g, "\n").trim();
			compiled += `\n\n### Bundled: ${token}\n\n${content}\n`;
		}
	}

	const artifactTagContract = artifactTags[name]
		? `- When creating/updating this stage's authoritative artifact, preserve the document tag \`${artifactTags[name]}\` on that DramaGo Document so Gate routing can locate it deterministically.\n`
		: "";

	compiled +=
		`\n\n---\n\n## DramaGo runtime integration contract\n\n` +
		`- Baseline: USVDS v11@${EXPECTED_COMMIT}, plugin ${EXPECTED_VERSION}.\n` +
		`- Write creator/development artifacts into the existing DramaGo project Documents; update the current artifact instead of inventing a parallel project store.\n` +
		`- Use DramaGo document categories for executable artifacts: screenplay for screenplay output and storyboard for storyboard/shot output when applicable.\n` +
		artifactTagContract +
		`- Character/scene/prop identity and variants remain owned by DramaGo Canon/Variant; shot execution and continuity remain owned by ShotManifest/ResolvedState; generation execution remains owned by GenerationTask/Asset.\n` +
		`- Never create shadow USVDS project, asset, shot, approval, job, or generation state.\n` +
		`- Legacy words such as APPROVED/PASS inside this Skill are narrative/workflow guidance only and must not set DramaGo human/system Gate tags by themselves. Gate state changes require the existing DramaGo approval/user action path.\n` +
		`- Before downstream production work, inspect the current DramaGo USVDS V11 Gate state and stop when the required gate is blocked.\n`;
	fs.writeFileSync(path.join(destinationSkills, `${name}.skill.md`), compiled);
	emitted += 1;
}

fs.writeFileSync(
	path.join(destination, "SOURCE.json"),
	JSON.stringify(
		{
			source_repository: "jimwoocory/US-Vertical-Drama-Studio",
			branch: "v11",
			commit: EXPECTED_COMMIT,
			plugin_version: EXPECTED_VERSION,
			skill_count: emitted,
			compilation: "SKILL.md plus directly referenced references/contracts inlined",
		},
		null,
		2,
	) + "\n",
);

process.stdout.write(
	`Synced ${emitted} USVDS V11 skills from ${EXPECTED_COMMIT.slice(0, 7)} / plugin ${EXPECTED_VERSION}\n`,
);
