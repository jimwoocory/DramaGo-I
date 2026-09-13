package productionqa

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

func TestProject679RegressionFixtureDetectsKnownLegacyFailures(t *testing.T) {
	// Minimal, portable fixture extracted from the real project
	// project-679e2eab7f3a12cc. It preserves only the facts needed by QA.
	screenplay := strings.Join([]string{
		"# 第一集 陷阵木勺",
		"> 类型：都市穿越短剧 · 本集时长约十二分钟 · 主线第一阶段第一集",
		"冉冬青独自站在灶台前。她把手机贴在耳边。",
		"手机从指缝滑下去，磕在灶台上。",
	}, "\n")
	scene := strings.Join([]string{
		"## 二号食堂打饭窗口与就餐区",
		"空间结构：老厂房改造的食堂前厅，窗台上方挂卷帘门。",
		"材质：不锈钢台面、绿色老式卷帘门、灰白水泥地。",
	}, "\n")
	props := strings.Join([]string{
		"## 手机",
		"可成像细节：深色直板智能手机。",
		"使用关系：刘铁柱与工人举着拍罢工现场，王小五用手机录韩三和掂锅，把视频传出食堂。",
	}, "\n")

	groups := make([]string, 0, 16)
	for index := 1; index <= 16; index++ {
		visual := "普通镜头"
		if index == 1 {
			visual = "二号食堂门外，老厂房铁皮外墙、半闭的银灰卷帘门"
		}
		groups = append(groups, fmt.Sprintf(
			"## 第 %02d 组\n- 0.00–14.80：%s。%s",
			index,
			visual,
			strings.Repeat("禁用：无字幕、无水印、无 BGM。", 5),
		))
	}
	storyboard := strings.Join(groups, "\n")

	report := AuditLegacyDocuments(screenplay, scene, props, storyboard)
	if report.TargetDurationSeconds != 720 {
		t.Fatalf("target duration = %.1f, want 720", report.TargetDurationSeconds)
	}
	if math.Abs(report.StoryboardDurationSeconds-236.8) > 0.001 {
		t.Fatalf("storyboard duration = %.2f, want 236.8", report.StoryboardDurationSeconds)
	}
	if report.StoryboardGroupCount != 16 {
		t.Fatalf("group count = %d, want 16", report.StoryboardGroupCount)
	}
	if math.Abs(report.DurationRatio-(236.8/720.0)) > 0.000001 {
		t.Fatalf("duration ratio = %f", report.DurationRatio)
	}

	codes := map[string]bool{}
	for _, finding := range report.Findings {
		codes[finding.Code] = true
	}
	for _, code := range []string{
		CodeDurationCompression,
		CodeSceneCanonConflict,
		CodeRepeatedGlobalBoilerplate,
		CodeUnsupportedPropUsage,
	} {
		if !codes[code] {
			t.Fatalf("findings = %+v, missing %s", report.Findings, code)
		}
	}
}

func TestParseTargetDurationSecondsSupportsChineseAndArabicMinutes(t *testing.T) {
	for _, test := range []struct {
		text string
		want float64
	}{
		{text: "本集时长约十二分钟", want: 720},
		{text: "目标时长：8分钟", want: 480},
		{text: "时长大约二十五分钟", want: 1500},
		{text: "时长约一百零二分钟", want: 6120},
	} {
		if got := ParseTargetDurationSeconds(test.text); got != test.want {
			t.Fatalf("ParseTargetDurationSeconds(%q) = %.1f, want %.1f", test.text, got, test.want)
		}
	}
}

func TestStoryboardDurationUsesMaximumEndPerGroup(t *testing.T) {
	storyboard := strings.Join([]string{
		"## 第 01 组",
		"- 0.00–2.80：A",
		"- 2.80–6.40：B",
		"- 6.40–14.80：C",
		"## 第 02 组",
		"- 0.00-3.20：D",
		"- 3.20-9.50：E",
	}, "\n")
	got, groups := StoryboardDurationSeconds(storyboard)
	if groups != 2 || math.Abs(got-24.3) > 0.001 {
		t.Fatalf("duration=%.2f groups=%d, want 24.3/2", got, groups)
	}
}

func TestAuditDoesNotFlagMatchingSceneColorOrSupportedPropUsage(t *testing.T) {
	screenplay := "时长约1分钟。王小五用手机录下现场。"
	scene := "## 食堂\n材质：绿色老式卷帘门。"
	props := "## 手机\n使用关系：王小五用手机录下现场。"
	storyboard := "## 第 01 组\n- 0.00–14.80：绿色卷帘门前。"
	report := AuditLegacyDocuments(screenplay, scene, props, storyboard)
	for _, finding := range report.Findings {
		if finding.Code == CodeSceneCanonConflict || finding.Code == CodeUnsupportedPropUsage {
			t.Fatalf("unexpected finding: %+v", finding)
		}
	}
}
