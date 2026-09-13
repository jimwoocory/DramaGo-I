import { describe, expect, it } from "vitest";
import {
	parseStoryboardShots,
	readStoryboardLaneSources,
} from "@/domains/episode/lib/storyboard-shots";
import { createSectionBlockId } from "@/domains/documents/lib/sections";

describe("parseStoryboardShots", () => {
	it("keeps body text for h2 storyboard groups without structured fields", () => {
		const shots = parseStoryboardShots(
			["## 开场落水", "", "陈远推开门。", "林书彤回头。"].join("\n"),
		);

		expect(shots).toEqual([
			{
				prompt: "陈远推开门。\n林书彤回头。",
				text: "陈远推开门。\n林书彤回头。",
				title: "开场落水",
			},
		]);
	});

	it("parses duration metadata from one h2 storyboard group", () => {
		const shots = parseStoryboardShots(
			["## 第 01 组 总时长：00:08", "", "**时长**：4.00-7.50秒", "", "**动作**：陈远靠近。"].join(
				"\n",
			),
		);

		expect(shots[0]).toEqual(
			expect.objectContaining({
				durationLabel: "4.00-7.50秒",
				durationSeconds: 3.5,
				text: "陈远靠近。",
			}),
		);
		expect(shots).toHaveLength(1);
	});

	it("splits timed bullets inside one h2 group into production shots", () => {
		const shots = parseStoryboardShots(
			[
				"## 第 02 组 · 泔水醒转（上）",
				"",
				"- 0.00–3.20：韩三和从两只蓝桶之间撑起身；画面：低机位近景；景别：中近景；",
				"- 3.20–6.40：韩三和抓起木勺回头；画面：侧后方跟拍；景别：中景；",
			].join("\n"),
		);

		expect(shots).toHaveLength(2);
		expect(shots[0]).toEqual(
			expect.objectContaining({
				title: "第 02 组 · 泔水醒转（上） · 镜头 01",
				durationLabel: "0.00-3.20秒",
				durationSeconds: 3.2,
				startSeconds: 0,
				endSeconds: 3.2,
				shotSize: "中近景",
				text: "低机位近景",
			}),
		);
		expect(shots[0]?.prompt).not.toContain("韩三和抓起木勺回头");
		expect(shots[1]?.prompt).toContain("韩三和抓起木勺回头");
	});
});

describe("readStoryboardLaneSources", () => {
	it("returns one h2 lane with the group body as its shot summary", () => {
		const lanes = readStoryboardLaneSources(
			["# 分镜脚本", "", "## 开场落水", "", "陈远站在校门口。"].join("\n"),
			{ documentId: "story-doc" },
		);

		expect(lanes).toHaveLength(1);
		expect(lanes[0]).toEqual(
			expect.objectContaining({
				blockId: createSectionBlockId("story-doc", 2, 1, "开场落水"),
				headingLevel: 2,
				headingOccurrence: 1,
				title: "开场落水",
			}),
		);
		expect(lanes[0]?.shots).toEqual([
			{
				prompt: "陈远站在校门口。",
				text: "陈远站在校门口。",
				title: "开场落水",
			},
		]);
	});

	it("prefers explicit section ids before storyboard lane headings", () => {
		const lanes = readStoryboardLaneSources(
			[
				"# 分镜脚本",
				"",
				"<!-- section-id: section_reel_01 -->",
				"## 开场落水",
				"",
				"陈远站在校门口。",
			].join("\n"),
			{ documentId: "story-doc" },
		);

		expect(lanes[0]?.blockId).toBe("section_reel_01");
	});

	it("keeps the h2 as a lane while exposing timed bullets as separate shots", () => {
		const lanes = readStoryboardLaneSources(
			[
				"# 分镜脚本",
				"",
				"## 第 01 组",
				"",
				"- 0.00-2.00：人物推门。",
				"- 2.00-5.50：人物转身。",
			].join("\n"),
			{ documentId: "story-doc" },
		);

		expect(lanes).toHaveLength(1);
		expect(lanes[0]?.shots).toHaveLength(2);
		expect(lanes[0]?.shots.map((shot) => shot.durationSeconds)).toEqual([2, 3.5]);
	});
});
