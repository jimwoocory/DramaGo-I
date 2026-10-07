import httpClient from "@/shared/lib/http";

const projectPath = (projectId: string, suffix: string) => {
	const id = projectId.trim();
	if (!id) throw new Error("projectId is required");
	return `/projects/${encodeURIComponent(id)}${suffix}`;
};

export type USVDSV11GateId =
	| "story_approved"
	| "screenplay_reviewed"
	| "canon_locked"
	| "continuity_resolved"
	| "storyboard_ready"
	| "generation_ready";

export interface USVDSV11Gate {
	gate: USVDSV11GateId;
	ready: boolean;
	blockers: string[];
}

export interface USVDSV11Baseline {
	branch: string;
	commit: string;
	pluginVersion: string;
}

export interface USVDSV11GateSummary {
	storyDocumentId?: string;
	screenplayDocumentId?: string;
	storyboardDocuments: number;
	canonCoreCount: number;
	canonApprovedCount: number;
	shotCount: number;
	readyShotCount: number;
	resolvedShotCount: number;
	compiledShotCount: number;
}

export interface USVDSV11Workflow {
	id: string;
	label: string;
	description: string;
	skill: string;
	stage: string;
}

export interface USVDSV11GenerationState {
	total: number;
	pending: number;
	running: number;
	completed: number;
	failed: number;
	lastTaskId?: string;
	lastUpdatedAt?: string;
}

export interface USVDSV11GateReport {
	projectId: string;
	baseline: USVDSV11Baseline;
	gates: USVDSV11Gate[];
	workflows: USVDSV11Workflow[];
	summary: USVDSV11GateSummary;
	generation: USVDSV11GenerationState;
}

export const usvdsV11GateKey = (projectId: string) => projectPath(projectId, "/usvds-v11/gates");

export const getUSVDSV11Gates = async (projectId: string) => {
	const response = await httpClient.get<USVDSV11GateReport>(usvdsV11GateKey(projectId));
	return response.data;
};
