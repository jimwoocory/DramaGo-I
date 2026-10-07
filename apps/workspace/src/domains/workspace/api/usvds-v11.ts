import type { MarkdownDocument } from "@/domains/documents/stores";
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
	episodeArchitectureDocumentId?: string;
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

export interface USVDSV11ApprovalState {
	gate: USVDSV11GateId;
	documentId?: string;
	documentTitle?: string;
	documentVersion?: number;
	contentDigest?: string;
	approved: boolean;
	stale: boolean;
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
	approvals: USVDSV11ApprovalState[];
	workflows: USVDSV11Workflow[];
	nextWorkflow?: USVDSV11Workflow;
	summary: USVDSV11GateSummary;
	generation: USVDSV11GenerationState;
}

export const usvdsV11GateKey = (projectId: string) => projectPath(projectId, "/usvds-v11/gates");

export interface USVDSV11GateMutationResult {
	document: MarkdownDocument;
	report: USVDSV11GateReport;
}

const mutateUSVDSV11Gate = async (
	projectId: string,
	gate: USVDSV11GateId,
	action: "approve" | "revoke",
	documentId: string,
	expectedVersion: number,
) => {
	const response = await httpClient.post<USVDSV11GateMutationResult>(
		projectPath(projectId, `/usvds-v11/gates/${encodeURIComponent(gate)}/${action}`),
		{ documentId, expectedVersion },
	);
	return response.data;
};

export const approveUSVDSV11Gate = (
	projectId: string,
	gate: USVDSV11GateId,
	documentId: string,
	expectedVersion: number,
) => mutateUSVDSV11Gate(projectId, gate, "approve", documentId, expectedVersion);

export const revokeUSVDSV11Gate = (
	projectId: string,
	gate: USVDSV11GateId,
	documentId: string,
	expectedVersion: number,
) => mutateUSVDSV11Gate(projectId, gate, "revoke", documentId, expectedVersion);

export const getUSVDSV11Gates = async (projectId: string) => {
	const response = await httpClient.get<USVDSV11GateReport>(usvdsV11GateKey(projectId));
	return response.data;
};
