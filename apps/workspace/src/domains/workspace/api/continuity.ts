import httpClient from "@/shared/lib/http";

const projectPath = (projectId: string, suffix: string) => {
	const id = projectId.trim();
	if (!id) throw new Error("projectId is required");
	return `/projects/${encodeURIComponent(id)}${suffix}`;
};

export interface CanonReferenceRecord {
	id: string;
	assetId: string;
	role: string;
	priority: number;
	locked: boolean;
	url?: string;
	posterUrl?: string;
}

export interface CanonAssetRecord {
	id: string;
	projectId: string;
	resourceType: "character" | "scene" | "prop" | (string & {});
	resourceId: string;
	sourceDocumentId: string;
	parentId?: string;
	variantKind?: string;
	name: string;
	specJson: string;
	promptText: string;
	status: "draft" | "approved" | "locked" | "deprecated" | (string & {});
	version: number;
	sourceHash?: string;
	references: CanonReferenceRecord[];
	createdAt: string;
	updatedAt: string;
}

export interface CanonProjectSyncSummary {
	resources: {
		created: number;
		updated: number;
		locked: number;
		skipped: number;
	};
	referencesCreated: number;
	referencesReused: number;
	selectionsSkipped: number;
}

export interface ShotCharacterBinding {
	canonId: string;
	variantId?: string;
	referenceAssetIds?: string[];
}

export interface ShotResourceBinding {
	canonId: string;
	variantId?: string;
	referenceAssetIds?: string[];
}

export interface ShotBindings {
	characters?: ShotCharacterBinding[];
	scene?: ShotResourceBinding;
	props?: ShotResourceBinding[];
}

export interface ShotManifestRecord {
	id: string;
	projectId: string;
	documentId: string;
	sectionId: string;
	shotKey: string;
	sequence: number;
	startSeconds?: number;
	endSeconds?: number;
	durationSeconds?: number;
	inheritsFromId?: string;
	actionText: string;
	cameraText: string;
	audioText: string;
	styleProfileId?: string;
	bindings: ShotBindings;
	stateChangesJson: string;
	resolvedStateJson: string;
	compiledPrompt?: string;
	sourceHash?: string;
	version: number;
	status: "draft" | "ready" | "conflict" | (string & {});
	createdAt: string;
	updatedAt: string;
}

export interface ShotProjectSyncSummary {
	created: number;
	updated: number;
	reused: number;
	skipped: number;
}

export const projectCanonKey = (projectId: string) => projectPath(projectId, "/canon");

export const getProjectCanon = async (projectId: string) => {
	const response = await httpClient.get<CanonAssetRecord[]>(projectCanonKey(projectId));
	return response.data;
};

export const syncProjectCanon = async (projectId: string) => {
	const response = await httpClient.post<CanonProjectSyncSummary>(
		projectPath(projectId, "/canon/sync"),
	);
	return response.data;
};

export const projectShotManifestsKey = (projectId: string, documentIds: readonly string[]) =>
	`${projectPath(projectId, "/shot-manifests")}?documents=${documentIds
		.map((id) => id.trim())
		.filter(Boolean)
		.sort()
		.join(",")}`;

export const getProjectShotManifests = async (
	projectId: string,
	documentIds: readonly string[],
): Promise<ShotManifestRecord[]> => {
	const ids = Array.from(new Set(documentIds.map((id) => id.trim()).filter(Boolean))).sort();
	if (ids.length === 0) return [];
	const lists = await Promise.all(
		ids.map(async (documentId) => {
			const response = await httpClient.get<ShotManifestRecord[]>(
				projectPath(projectId, "/shot-manifests"),
				{ params: { documentId } },
			);
			return response.data;
		}),
	);
	return lists.flat().sort((first, second) => {
		if (first.documentId !== second.documentId)
			return first.documentId.localeCompare(second.documentId);
		if (first.sequence !== second.sequence) return first.sequence - second.sequence;
		return first.id.localeCompare(second.id);
	});
};

export const syncProjectShotManifests = async (projectId: string) => {
	const response = await httpClient.post<ShotProjectSyncSummary>(
		projectPath(projectId, "/shot-manifests/sync"),
	);
	return response.data;
};
