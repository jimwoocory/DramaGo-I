package prompt

import (
	"fmt"
	"strings"
	"testing"
)

func TestCreateReferenceSectionBlockIDMatchesFrontendHash(t *testing.T) {
	tests := []struct {
		name       string
		documentID string
		level      int
		occurrence int
		title      string
		want       string
	}{
		{
			name:       "character",
			documentID: "doc-1",
			level:      2,
			occurrence: 1,
			title:      "沈阎",
			want:       "section-2iwmqt",
		},
		{
			name:       "scene",
			documentID: "scene-doc",
			level:      2,
			occurrence: 1,
			title:      "雨夜巷口",
			want:       "section-xeu5uu",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := createReferenceSectionBlockID(test.documentID, test.level, test.occurrence, test.title)
			if got != test.want {
				t.Fatalf("createReferenceSectionBlockID() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestBuildReferenceIndexItemsUsesStableSectionIDs(t *testing.T) {
	items := buildReferenceIndexItems(AgentRunRequest{
		Documents: []AgentDocumentContext{
			{
				ID:       "character-doc",
				Title:    "角色设定",
				Category: "character",
				Content: strings.Join([]string{
					"<!-- section-id: section_shenyan -->",
					"## 沈阎",
					"",
					"男主角设定正文不应该进入索引。",
					"",
					"## 林晚",
				}, "\n"),
			},
		},
	})

	if len(items) != 2 {
		t.Fatalf("items = %d, want 2: %#v", len(items), items)
	}
	if items[0].Title != "沈阎" || items[0].BlockID != "section_shenyan" {
		t.Fatalf("first item = %#v, want persisted section id for 沈阎", items[0])
	}
	if items[0].MentionMarkdown != "@[沈阎](mention://character-doc/section_shenyan)" {
		t.Fatalf("mention = %q", items[0].MentionMarkdown)
	}
	if strings.Contains(items[0].MentionMarkdown, "男主角设定正文") {
		t.Fatalf("mention index leaked document body: %q", items[0].MentionMarkdown)
	}
	if items[1].Title != "林晚" || !strings.HasPrefix(items[1].BlockID, "section-") {
		t.Fatalf("second item = %#v, want fallback section id for 林晚", items[1])
	}
}

func TestBuildReferenceIndexItemsSkipsDocumentRootHeadingWhenChildSectionsExist(t *testing.T) {
	items := buildReferenceIndexItems(AgentRunRequest{
		Documents: []AgentDocumentContext{
			{
				ID:       "character-book-doc",
				Title:    "角色册 第一章",
				Category: "character",
				Content: strings.Join([]string{
					"# 角色册 第一章",
					"",
					"视觉风格：3DCG动漫",
					"",
					"<!-- section-id: section_chenyuan -->",
					"## 陈远",
					"",
					"陈远，21岁男大学生。",
					"",
					"<!-- section-id: section_linshutong -->",
					"## 林书彤",
					"",
					"林书彤，21岁女大学生。",
				}, "\n"),
			},
		},
	})

	if len(items) != 2 {
		t.Fatalf("items = %d, want 2: %#v", len(items), items)
	}
	if items[0].Title != "陈远" || items[0].MentionMarkdown != "@[陈远](mention://character-book-doc/section_chenyuan)" {
		t.Fatalf("first item = %#v, want 陈远 section mention", items[0])
	}
	if items[1].Title != "林书彤" || items[1].MentionMarkdown != "@[林书彤](mention://character-book-doc/section_linshutong)" {
		t.Fatalf("second item = %#v, want 林书彤 section mention", items[1])
	}
}

func TestBuildReferenceIndexItemsOnlyUsesSecondLevelHeadings(t *testing.T) {
	items := buildReferenceIndexItems(AgentRunRequest{
		Documents: []AgentDocumentContext{
			{
				ID:       "story-doc",
				Title:    "第一集剧本",
				Category: "storyboard",
				Content: strings.Join([]string{
					"# 第一集",
					"",
					"## 1-1 日 外景 湖大校门口",
					"",
					"### 镜头细节",
					"",
					"## 1-2 日 外景 湖大校门口",
				}, "\n"),
			},
		},
	})

	if len(items) != 2 {
		t.Fatalf("items = %d, want only h2 sections: %#v", len(items), items)
	}
	for _, forbidden := range []string{"第一集", "镜头细节"} {
		for _, item := range items {
			if item.Title == forbidden {
				t.Fatalf("items = %#v, want %q ignored", items, forbidden)
			}
		}
	}
	if items[0].Title != "1-1 日 外景 湖大校门口" || items[1].Title != "1-2 日 外景 湖大校门口" {
		t.Fatalf("items = %#v, want h2 section titles", items)
	}
}

func TestBuildReferenceIndexItemsIncludesReferenceDocuments(t *testing.T) {
	items := buildReferenceIndexItems(AgentRunRequest{
		Documents: []AgentDocumentContext{
			{
				ID:       "reference-doc",
				Title:    "原始素材",
				Category: "reference",
				Content:  "没有标题的原始素材正文。",
			},
		},
	})

	if len(items) != 1 {
		t.Fatalf("items = %d, want 1: %#v", len(items), items)
	}
	if items[0].Title != "原始素材" || items[0].CategoryLabel != "资料" {
		t.Fatalf("item = %#v, want reference document labeled as 资料", items[0])
	}
	if items[0].MentionMarkdown != "@[原始素材](mention://reference-doc)" {
		t.Fatalf("mention = %q", items[0].MentionMarkdown)
	}
}

func TestBuildReferenceIndexItemsIgnoresAssetReferences(t *testing.T) {
	items := buildReferenceIndexItems(AgentRunRequest{
		References: []AgentReference{
			{
				Kind:       "asset",
				DocumentID: "asset-1",
				AssetID:    "asset-1",
				AssetKind:  "image",
				Category:   "reference",
				Title:      "参考图.png",
				URL:        "/api/media/assets/asset-1/content",
			},
		},
	})

	if len(items) != 0 {
		t.Fatalf("items = %#v, want asset references ignored", items)
	}
}

func TestStoryboardReferenceIndexCompactsLargeProjectsByRelevanceAndKeepsExplicitReferences(t *testing.T) {
	var characterBook strings.Builder
	for index := 1; index <= 80; index++ {
		fmt.Fprintf(&characterBook, "<!-- section-id: section_%03d -->\n## 角色%03d\n角色设定正文 %03d。\n\n", index, index, index)
	}
	items := buildReferenceIndexItems(AgentRunRequest{
		Prompt: "优化这一组分镜，保持人物连续性",
		Document: &AgentDocumentContext{
			ID:       "storyboard-ep1",
			Title:    "第一集分镜",
			Category: "storyboard",
			Content:  "## 第12组\n角色079推门进入走廊。",
		},
		Documents: []AgentDocumentContext{{
			ID:       "character-doc",
			Title:    "角色设定",
			Category: "character",
			Content:  characterBook.String(),
		}},
		References: []AgentReference{{
			Kind:       "section",
			DocumentID: "character-doc",
			BlockID:    "section_080",
			Title:      "角色080",
			Category:   "character",
		}},
	})

	if len(items) != maxStoryboardReferenceIndexItems {
		t.Fatalf("items = %d, want compact storyboard limit %d", len(items), maxStoryboardReferenceIndexItems)
	}
	byTitle := map[string]referenceIndexItem{}
	for _, item := range items {
		byTitle[item.Title] = item
	}
	if _, ok := byTitle["角色079"]; !ok {
		t.Fatalf("items = %#v, want resource mentioned by current storyboard context", items)
	}
	if explicit, ok := byTitle["角色080"]; !ok || !explicit.Explicit {
		t.Fatalf("items = %#v, want explicit @ reference retained", items)
	}
	if _, ok := byTitle["角色078"]; ok {
		t.Fatalf("items = %#v, want unrelated high-index resource compacted out", items)
	}
}

func TestStoryboardReferenceIndexExpandsWhenExplicitReferencesExceedCompactLimit(t *testing.T) {
	references := make([]AgentReference, 0, 40)
	for index := 1; index <= 40; index++ {
		references = append(references, AgentReference{
			Kind:       "section",
			DocumentID: "character-doc",
			BlockID:    fmt.Sprintf("section_explicit_%03d", index),
			Title:      fmt.Sprintf("明确角色%03d", index),
			Category:   "character",
		})
	}
	items := buildReferenceIndexItems(AgentRunRequest{
		Prompt: "生成分镜",
		Document: &AgentDocumentContext{
			ID:       "storyboard-ep1",
			Title:    "第一集分镜",
			Category: "storyboard",
		},
		References: references,
	})
	if len(items) != len(references) {
		t.Fatalf("items = %d, want all %d explicit references retained", len(items), len(references))
	}
	for _, item := range items {
		if !item.Explicit {
			t.Fatalf("item = %#v, want explicit reference", item)
		}
	}
}
