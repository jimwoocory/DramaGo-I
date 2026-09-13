import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { MarkdownHybridEditor } from "./MarkdownHybridEditor";

describe("MarkdownHybridEditor", () => {
	afterEach(() => {
		cleanup();
		delete window.mediagoDesktop;
		vi.unstubAllEnvs();
	});

	it("does not persist parser normalization when an existing markdown document is only opened", async () => {
		const onChange = vi.fn();
		const legacyMarkdown = "# 第 02 集 铁瓮入灶\n\n正文。\n\n";

		render(
			<MarkdownHybridEditor
				documentId="doc-legacy-screenplay"
				value={legacyMarkdown}
				onChange={onChange}
			/>,
		);

		const editor = await screen.findByLabelText("Markdown 编辑器");
		await new Promise((resolve) => window.setTimeout(resolve, 250));
		fireEvent.blur(editor);
		await new Promise((resolve) => window.setTimeout(resolve, 50));

		expect(onChange).not.toHaveBeenCalled();
	});

	it("renders markdown image URLs against the packaged desktop server", async () => {
		vi.stubEnv("DEV", false);
		window.mediagoDesktop = {
			isElectron: true,
			sidecarOrigin: "http://127.0.0.1:48273",
		} as typeof window.mediagoDesktop;

		render(
			<MarkdownHybridEditor
				documentId="doc-character"
				value="![角色图](</api/v1/media-assets/character/content>)"
				onChange={vi.fn()}
			/>,
		);

		await waitFor(() => {
			expect(screen.getByRole("img", { name: "角色图" })).toHaveAttribute(
				"src",
				"http://127.0.0.1:48273/api/v1/media-assets/character/content",
			);
		});
	});
});
