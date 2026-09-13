package shotmanifest

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mediago-dev/mediago-drama/services/server/internal/domain"
	"github.com/mediago-dev/mediago-drama/services/server/internal/repository"
	servicecanon "github.com/mediago-dev/mediago-drama/services/server/internal/service/canon"
	"github.com/mediago-dev/mediago-drama/services/server/internal/service/model"
	"github.com/mediago-dev/mediago-drama/services/server/internal/testutil"
)

func TestServiceAutoInheritsPreviousResolvedState(t *testing.T) {
	service, _, canonService := newTestShotService(t)
	character := mustCanonCore(t, canonService, servicecanon.SourceResource{
		ProjectID:        "project-a",
		ResourceType:     servicecanon.ResourceTypeCharacter,
		ResourceID:       "section-han",
		SourceDocumentID: "characters",
		Name:             "韩三河",
		PromptText:       "固定脸型，黑色束发",
	})

	first, err := service.Upsert(UpsertInput{
		ProjectID:  "project-a",
		DocumentID: "storyboard-ep1",
		SectionID:  "section-shot-1",
		ShotKey:    "shot-1",
		Sequence:   1,
		Bindings: Bindings{Characters: []CharacterBinding{{
			CanonID: character.ID,
		}}},
		ActionText:       "韩三河把外套拉到肩下。",
		StateChangesJSON: `{"characters":{"` + character.ID + `":{"look":{"outerwear":"half_removed","hair":"tied_messy"},"injury":{"forehead_wound":"present"}}}}`,
	})
	if err != nil {
		t.Fatalf("Upsert(first) error = %v", err)
	}

	second, err := service.Upsert(UpsertInput{
		ProjectID:  "project-a",
		DocumentID: "storyboard-ep1",
		SectionID:  "section-shot-2",
		ShotKey:    "shot-2",
		Sequence:   2,
		Bindings: Bindings{Characters: []CharacterBinding{{
			CanonID: character.ID,
		}}},
		ActionText:       "韩三河继续向前走。",
		StateChangesJSON: `{}`,
	})
	if err != nil {
		t.Fatalf("Upsert(second) error = %v", err)
	}
	if second.InheritsFromID != first.ID {
		t.Fatalf("second.InheritsFromID = %q, want %q", second.InheritsFromID, first.ID)
	}
	if !jsonEquivalent(t, second.ResolvedStateJSON, first.ResolvedStateJSON) {
		t.Fatalf("second state = %s, want inherited %s", second.ResolvedStateJSON, first.ResolvedStateJSON)
	}
}

func TestServiceExplicitNestedChangeDoesNotResetSiblingState(t *testing.T) {
	service, _, canonService := newTestShotService(t)
	character := mustCanonCore(t, canonService, servicecanon.SourceResource{
		ProjectID:        "project-a",
		ResourceType:     servicecanon.ResourceTypeCharacter,
		ResourceID:       "section-han",
		SourceDocumentID: "characters",
		Name:             "韩三河",
	})
	first, err := service.Upsert(UpsertInput{
		ProjectID:        "project-a",
		DocumentID:       "storyboard-ep1",
		SectionID:        "section-shot-1",
		ShotKey:          "shot-1",
		Sequence:         1,
		Bindings:         Bindings{Characters: []CharacterBinding{{CanonID: character.ID}}},
		StateChangesJSON: `{"characters":{"` + character.ID + `":{"look":{"outerwear":"half_removed","hair":"tied_messy"},"injury":{"forehead_wound":"present"}}}}`,
	})
	if err != nil {
		t.Fatalf("Upsert(first) error = %v", err)
	}
	second, err := service.Upsert(UpsertInput{
		ProjectID:        "project-a",
		DocumentID:       "storyboard-ep1",
		SectionID:        "section-shot-2",
		ShotKey:          "shot-2",
		Sequence:         2,
		Bindings:         Bindings{Characters: []CharacterBinding{{CanonID: character.ID}}},
		StateChangesJSON: `{"characters":{"` + character.ID + `":{"look":{"outerwear":"restored"}}}}`,
	})
	if err != nil {
		t.Fatalf("Upsert(second) error = %v", err)
	}
	var state map[string]any
	if err := json.Unmarshal([]byte(second.ResolvedStateJSON), &state); err != nil {
		t.Fatalf("decoding resolved state: %v", err)
	}
	characters := state["characters"].(map[string]any)
	han := characters[character.ID].(map[string]any)
	look := han["look"].(map[string]any)
	injury := han["injury"].(map[string]any)
	if look["outerwear"] != "restored" || look["hair"] != "tied_messy" || injury["forehead_wound"] != "present" {
		t.Fatalf("resolved state = %#v, sibling continuity was reset", state)
	}
	if second.InheritsFromID != first.ID {
		t.Fatalf("second.InheritsFromID = %q, want %q", second.InheritsFromID, first.ID)
	}
}

func TestServiceValidatesVariantOwnership(t *testing.T) {
	service, _, canonService := newTestShotService(t)
	firstCore := mustCanonCore(t, canonService, servicecanon.SourceResource{
		ProjectID:        "project-a",
		ResourceType:     servicecanon.ResourceTypeCharacter,
		ResourceID:       "section-a",
		SourceDocumentID: "characters",
		Name:             "角色A",
	})
	secondCore := mustCanonCore(t, canonService, servicecanon.SourceResource{
		ProjectID:        "project-a",
		ResourceType:     servicecanon.ResourceTypeCharacter,
		ResourceID:       "section-b",
		SourceDocumentID: "characters",
		Name:             "角色B",
	})
	variant, err := canonService.CreateVariant(servicecanon.VariantInput{
		ProjectID:   "project-a",
		ParentID:    secondCore.ID,
		VariantKind: "look",
		Name:        "角色B婚礼造型",
	})
	if err != nil {
		t.Fatalf("CreateVariant() error = %v", err)
	}

	_, err = service.Upsert(UpsertInput{
		ProjectID:  "project-a",
		DocumentID: "storyboard-ep1",
		SectionID:  "section-shot-1",
		Sequence:   1,
		Bindings: Bindings{Characters: []CharacterBinding{{
			CanonID:   firstCore.ID,
			VariantID: variant.ID,
		}}},
	})
	if err == nil || !strings.Contains(err.Error(), "does not belong") {
		t.Fatalf("Upsert(wrong variant) error = %v, want ownership conflict", err)
	}
}

func TestServiceRejectsLockedCanonIdentityMutation(t *testing.T) {
	service, _, canonService := newTestShotService(t)
	character := mustCanonCore(t, canonService, servicecanon.SourceResource{
		ProjectID:        "project-a",
		ResourceType:     servicecanon.ResourceTypeCharacter,
		ResourceID:       "section-locked",
		SourceDocumentID: "characters",
		Name:             "锁定角色",
	})
	if _, err := canonService.UpdateStatus("project-a", character.ID, servicecanon.StatusLocked); err != nil {
		t.Fatalf("locking Canon: %v", err)
	}
	_, err := service.Upsert(UpsertInput{
		ProjectID:        "project-a",
		DocumentID:       "storyboard-ep1",
		SectionID:        "section-shot-1",
		Sequence:         1,
		Bindings:         Bindings{Characters: []CharacterBinding{{CanonID: character.ID}}},
		StateChangesJSON: `{"characters":{"` + character.ID + `":{"identity":{"face":"different"}}}}`,
	})
	if err == nil || !strings.Contains(err.Error(), "locked character Canon") {
		t.Fatalf("Upsert(identity mutation) error = %v, want locked Canon conflict", err)
	}
}

func TestServiceUpsertVersionsOnlyMaterialChanges(t *testing.T) {
	service, _, _ := newTestShotService(t)
	input := UpsertInput{
		ProjectID:        "project-a",
		DocumentID:       "storyboard-ep1",
		SectionID:        "section-shot-1",
		ShotKey:          "shot-1",
		Sequence:         1,
		ActionText:       "走向门口。",
		StateChangesJSON: `{}`,
	}
	first, err := service.Upsert(input)
	if err != nil {
		t.Fatalf("Upsert(first) error = %v", err)
	}
	second, err := service.Upsert(input)
	if err != nil {
		t.Fatalf("Upsert(same) error = %v", err)
	}
	if second.Version != first.Version {
		t.Fatalf("same version = %d, want %d", second.Version, first.Version)
	}
	input.ActionText = "快速走向门口。"
	third, err := service.Upsert(input)
	if err != nil {
		t.Fatalf("Upsert(changed) error = %v", err)
	}
	if third.Version != first.Version+1 {
		t.Fatalf("changed version = %d, want %d", third.Version, first.Version+1)
	}
}

func TestInheritBindingsCarriesPropsUntilExplicitlyCleared(t *testing.T) {
	previous := Bindings{Props: []ResourceBinding{{CanonID: "prop-envelope", VariantID: "prop-sealed", ReferenceAssetIDs: []string{"asset-envelope"}}}}
	inherited, err := inheritBindings(Bindings{}, previous, `{}`, true)
	if err != nil {
		t.Fatalf("inheritBindings() error = %v", err)
	}
	if len(inherited.Props) != 1 || inherited.Props[0].CanonID != "prop-envelope" || inherited.Props[0].VariantID != "prop-sealed" || len(inherited.Props[0].ReferenceAssetIDs) != 1 {
		t.Fatalf("inherited props = %+v", inherited.Props)
	}
	cleared, err := inheritBindings(Bindings{}, previous, `{"props":{"prop-envelope":null}}`, true)
	if err != nil {
		t.Fatalf("inheritBindings(clear) error = %v", err)
	}
	if len(cleared.Props) != 0 {
		t.Fatalf("cleared props = %+v, want none", cleared.Props)
	}
}

func TestServiceInheritsCharacterAndSceneVariantsAcrossShots(t *testing.T) {
	service, _, canonService := newTestShotService(t)
	character := mustCanonCore(t, canonService, servicecanon.SourceResource{
		ProjectID:        "project-a",
		ResourceType:     servicecanon.ResourceTypeCharacter,
		ResourceID:       "section-char",
		SourceDocumentID: "characters",
		Name:             "林书彤",
		PromptText:       "固定鹅蛋脸，黑色长发",
	})
	wedding, err := canonService.CreateVariant(servicecanon.VariantInput{
		ProjectID:   "project-a",
		ParentID:    character.ID,
		VariantKind: "look",
		Name:        "婚礼",
		PromptText:  "白色婚纱，盘发",
	})
	if err != nil {
		t.Fatalf("CreateVariant(wedding) error = %v", err)
	}
	work, err := canonService.CreateVariant(servicecanon.VariantInput{
		ProjectID:   "project-a",
		ParentID:    character.ID,
		VariantKind: "look",
		Name:        "工作",
		PromptText:  "深灰西装，低马尾",
	})
	if err != nil {
		t.Fatalf("CreateVariant(work) error = %v", err)
	}
	scene := mustCanonCore(t, canonService, servicecanon.SourceResource{
		ProjectID:        "project-a",
		ResourceType:     servicecanon.ResourceTypeScene,
		ResourceID:       "section-hall",
		SourceDocumentID: "scenes",
		Name:             "婚礼大厅",
		PromptText:       "固定大厅空间结构",
	})
	night, err := canonService.CreateVariant(servicecanon.VariantInput{
		ProjectID:   "project-a",
		ParentID:    scene.ID,
		VariantKind: "scene",
		Name:        "夜间",
		PromptText:  "夜间暖白灯",
	})
	if err != nil {
		t.Fatalf("CreateVariant(night) error = %v", err)
	}

	first, err := service.Upsert(UpsertInput{
		ProjectID:  "project-a",
		DocumentID: "storyboard-ep1",
		SectionID:  "section-shot-1",
		Sequence:   1,
		Bindings: Bindings{
			Characters: []CharacterBinding{{CanonID: character.ID, VariantID: wedding.ID}},
			Scene:      &ResourceBinding{CanonID: scene.ID, VariantID: night.ID},
		},
		StateChangesJSON: `{}`,
	})
	if err != nil {
		t.Fatalf("Upsert(first) error = %v", err)
	}
	second, err := service.Upsert(UpsertInput{
		ProjectID:  "project-a",
		DocumentID: "storyboard-ep1",
		SectionID:  "section-shot-2",
		Sequence:   2,
		Bindings: Bindings{
			Characters: []CharacterBinding{{CanonID: character.ID}},
			Scene:      nil,
		},
		StateChangesJSON: `{}`,
	})
	if err != nil {
		t.Fatalf("Upsert(second) error = %v", err)
	}
	if second.InheritsFromID != first.ID || len(second.Bindings.Characters) != 1 {
		t.Fatalf("second = %+v, want inherited shot and character", second)
	}
	if second.Bindings.Characters[0].VariantID != wedding.ID {
		t.Fatalf("second character variant = %q, want wedding %q", second.Bindings.Characters[0].VariantID, wedding.ID)
	}
	if second.Bindings.Scene == nil || second.Bindings.Scene.CanonID != scene.ID || second.Bindings.Scene.VariantID != night.ID {
		t.Fatalf("second scene = %+v, want inherited night scene", second.Bindings.Scene)
	}

	third, err := service.Upsert(UpsertInput{
		ProjectID:  "project-a",
		DocumentID: "storyboard-ep1",
		SectionID:  "section-shot-3",
		Sequence:   3,
		Bindings: Bindings{
			Characters: []CharacterBinding{{CanonID: character.ID, VariantID: work.ID}},
			Scene:      &ResourceBinding{CanonID: scene.ID},
		},
		StateChangesJSON: `{}`,
	})
	if err != nil {
		t.Fatalf("Upsert(third) error = %v", err)
	}
	if third.Bindings.Characters[0].VariantID != work.ID {
		t.Fatalf("third character variant = %q, want explicit work %q", third.Bindings.Characters[0].VariantID, work.ID)
	}
	if third.Bindings.Scene == nil || third.Bindings.Scene.VariantID != night.ID {
		t.Fatalf("third scene = %+v, want same-core scene variant inherited", third.Bindings.Scene)
	}
}

func TestLegacyBindingInferenceSelectsMentionedVariantThenContinuityCarriesItForward(t *testing.T) {
	service, _, canonService := newTestShotService(t)
	resources := fakeShotDocumentResources{resources: []model.WorkspaceDocumentResourceRecord{
		{
			Type:       "character",
			Title:      "林书彤",
			Prompt:     "## 林书彤\n固定鹅蛋脸，黑色长发。\n### 造型变体：婚礼\n白色婚纱，盘发。",
			PlainText:  "林书彤 固定鹅蛋脸 黑色长发 婚礼 白色婚纱 盘发",
			Markdown:   "## 林书彤\n固定鹅蛋脸，黑色长发。\n### 造型变体：婚礼\n白色婚纱，盘发。",
			DocumentID: "characters",
			SectionID:  "section-char",
		},
		{
			Type:       "storyboard",
			Title:      "第1组",
			Prompt:     "林书彤婚礼造型走入大厅。",
			PlainText:  "林书彤婚礼造型走入大厅。",
			Markdown:   "## 第1组\n林书彤@[林书彤](mention://characters/section-char)以婚礼造型走入大厅。",
			DocumentID: "storyboard-ep1",
			SectionID:  "section-shot-1",
		},
		{
			Type:       "storyboard",
			Title:      "第2组",
			Prompt:     "林书彤继续向前走。",
			PlainText:  "林书彤继续向前走。",
			Markdown:   "## 第2组\n林书彤@[林书彤](mention://characters/section-char)继续向前走。",
			DocumentID: "storyboard-ep1",
			SectionID:  "section-shot-2",
		},
	}}
	canonService.SetDocumentResourceProvider(resources)
	service.SetDocumentResourceProvider(resources)

	summary, err := service.SyncProject("project-a")
	if err != nil {
		t.Fatalf("SyncProject() error = %v", err)
	}
	if summary.Created != 2 {
		t.Fatalf("summary = %+v, want two storyboard shots", summary)
	}
	shots, err := service.ListDocument("project-a", "storyboard-ep1")
	if err != nil {
		t.Fatalf("ListDocument() error = %v", err)
	}
	if len(shots) != 2 || len(shots[0].Bindings.Characters) != 1 || len(shots[1].Bindings.Characters) != 1 {
		t.Fatalf("shots = %+v", shots)
	}
	variantID := shots[0].Bindings.Characters[0].VariantID
	if variantID == "" {
		t.Fatalf("first shot did not infer mentioned wedding variant: %+v", shots[0].Bindings)
	}
	if shots[1].Bindings.Characters[0].VariantID != variantID {
		t.Fatalf("second variant = %q, want inherited %q", shots[1].Bindings.Characters[0].VariantID, variantID)
	}
	compiled, err := service.Compile(shots[1])
	if err != nil {
		t.Fatalf("Compile(second) error = %v", err)
	}
	if !strings.Contains(compiled.Prompt, "白色婚纱") {
		t.Fatalf("compiled second prompt lost wedding variant:\n%s", compiled.Prompt)
	}
}

type fakeShotDocumentResources struct {
	resources []model.WorkspaceDocumentResourceRecord
}

func (fake fakeShotDocumentResources) ListWorkspaceDocumentResources(projectID string) (model.WorkspaceDocumentResourcesResponse, error) {
	return model.WorkspaceDocumentResourcesResponse{ProjectID: projectID, Resources: fake.resources}, nil
}

func TestSyncProjectLazilyCreatesLegacyShotsAndDoesNotOverwriteExistingManifest(t *testing.T) {
	service, repo, canonService := newTestShotService(t)
	resources := fakeShotDocumentResources{resources: []model.WorkspaceDocumentResourceRecord{
		{
			Type:       "character",
			Title:      "韩三河",
			Prompt:     "固定脸型，黑色束发",
			PlainText:  "韩三河角色设定",
			Markdown:   "## 韩三河\n固定脸型，黑色束发",
			DocumentID: "characters",
			SectionID:  "section-han",
		},
		{
			Type:       "scene",
			Title:      "食堂",
			Prompt:     "室内食堂，绿色卷帘门",
			PlainText:  "食堂内景",
			Markdown:   "## 食堂\n室内食堂，绿色卷帘门",
			DocumentID: "scenes",
			SectionID:  "section-cafeteria",
		},
		{
			Type:       "prop",
			Title:      "木勺",
			Prompt:     "旧木勺",
			PlainText:  "一把旧木勺",
			Markdown:   "## 木勺\n一把旧木勺",
			DocumentID: "props",
			SectionID:  "section-spoon",
		},
		{
			Type:       "storyboard",
			Title:      "第1组",
			Prompt:     "韩三河进入食堂，腰间挂着木勺。",
			PlainText:  "韩三河进入食堂，腰间挂着木勺。",
			Markdown:   "## 第1组\n韩三河@[韩三河](mention://characters/section-han?kind=section&category=character)进入食堂，腰间挂着木勺。",
			DocumentID: "storyboard-ep1",
			SectionID:  "section-shot-1",
		},
	}}
	canonService.SetDocumentResourceProvider(resources)
	service.SetDocumentResourceProvider(resources)

	summary, err := service.SyncProject("project-a")
	if err != nil {
		t.Fatalf("SyncProject() error = %v", err)
	}
	if summary.Created != 1 || summary.Reused != 0 || summary.Skipped != 0 {
		t.Fatalf("summary = %+v, want one created legacy shot", summary)
	}
	shots, err := service.ListDocument("project-a", "storyboard-ep1")
	if err != nil {
		t.Fatalf("ListDocument() error = %v", err)
	}
	if len(shots) != 1 {
		t.Fatalf("shots = %+v, want 1", shots)
	}
	shot := shots[0]
	if len(shot.Bindings.Characters) != 1 || shot.Bindings.Scene == nil || len(shot.Bindings.Props) != 1 {
		t.Fatalf("legacy bindings = %+v, want character/scene/prop inferred", shot.Bindings)
	}
	if shot.Bindings.Characters[0].CanonID == "" || shot.Bindings.Scene.CanonID == "" || shot.Bindings.Props[0].CanonID == "" {
		t.Fatalf("legacy bindings missing Canon IDs: %+v", shot.Bindings)
	}

	manual, err := service.Upsert(UpsertInput{
		ProjectID:        "project-a",
		DocumentID:       "storyboard-ep1",
		SectionID:        "section-shot-1",
		Sequence:         1,
		ActionText:       "用户手工调整后的镜头动作",
		Bindings:         shot.Bindings,
		StateChangesJSON: `{}`,
		Status:           StatusReady,
	})
	if err != nil {
		t.Fatalf("manual Upsert() error = %v", err)
	}
	if manual.ActionText != "用户手工调整后的镜头动作" {
		t.Fatalf("manual action = %q", manual.ActionText)
	}

	secondSummary, err := service.SyncProject("project-a")
	if err != nil {
		t.Fatalf("SyncProject(second) error = %v", err)
	}
	if secondSummary.Created != 0 || secondSummary.Reused != 1 {
		t.Fatalf("second summary = %+v, want reuse without overwrite", secondSummary)
	}
	persisted, err := repo.FindBySource("project-a", "storyboard-ep1", "section-shot-1", "")
	if err != nil {
		t.Fatalf("FindBySource() error = %v", err)
	}
	if persisted.ActionText != "用户手工调整后的镜头动作" || persisted.Status != StatusReady {
		t.Fatalf("legacy resync overwrote existing manifest: %+v", persisted)
	}
}

func TestSyncProjectSplitsTimedStoryboardGroupIntoProductionShots(t *testing.T) {
	service, _, canonService := newTestShotService(t)
	resources := fakeShotDocumentResources{resources: []model.WorkspaceDocumentResourceRecord{
		{
			Type:       "character",
			Title:      "韩三河",
			Prompt:     "固定脸型，黑色束发",
			PlainText:  "韩三河角色设定",
			Markdown:   "## 韩三河\n固定脸型，黑色束发",
			DocumentID: "characters",
			SectionID:  "section-han",
		},
		{
			Type:       "scene",
			Title:      "二号食堂后厨",
			Prompt:     "对应场次：第 02 场「泔水醒转」。\n核心物件：泔水桶、灶台。",
			PlainText:  "二号食堂后厨",
			Markdown:   "## 二号食堂后厨\n泔水桶、灶台",
			DocumentID: "scenes",
			SectionID:  "scene-kitchen",
		},
		{
			Type:       "prop",
			Title:      "陷阵木勺",
			Prompt:     "深色全木长柄勺",
			PlainText:  "陷阵木勺",
			Markdown:   "## 陷阵木勺\n深色全木长柄勺",
			DocumentID: "props",
			SectionID:  "prop-spoon",
		},
		{
			Type:      "storyboard",
			Title:     "第 02 组 · 泔水醒转（上）",
			Prompt:    "整组旧 prompt 不应成为 Production Shot 的动作文本。",
			PlainText: "整组旧 plain text。",
			Markdown: strings.Join([]string{
				"## 第 02 组 · 泔水醒转（上）",
				"",
				"- 0.00–3.20：韩三河在二号食堂后厨撑身坐起，腰间挂着陷阵木勺；画面：低角度手持推近；音频：排风扇低鸣；台词：无。",
				"- 3.20–6.40：韩三河抓起陷阵木勺转身查看灶台；画面：中景侧后方跟拍；音频：脚步声；台词：韩三河“谁人吹角？”。",
			}, "\n"),
			DocumentID: "storyboard-ep1",
			SectionID:  "section-shot-2",
		},
	}}
	canonService.SetDocumentResourceProvider(resources)
	service.SetDocumentResourceProvider(resources)

	first, err := service.SyncProject("project-a")
	if err != nil {
		t.Fatalf("SyncProject(first) error = %v", err)
	}
	if first.Created != 2 || first.Updated != 0 || first.Reused != 0 {
		t.Fatalf("first summary = %+v, want two created Production Shots", first)
	}
	shots, err := service.ListDocument("project-a", "storyboard-ep1")
	if err != nil {
		t.Fatalf("ListDocument() error = %v", err)
	}
	if len(shots) != 2 {
		t.Fatalf("shots = %+v, want 2 Production Shots", shots)
	}
	if shots[0].ShotKey != "beat-001" || shots[1].ShotKey != "beat-002" {
		t.Fatalf("shot keys = %q/%q, want stable beat keys", shots[0].ShotKey, shots[1].ShotKey)
	}
	if shots[0].StartSeconds != 0 || shots[0].EndSeconds != 3.2 || shots[0].DurationSeconds != 3.2 {
		t.Fatalf("first timing = %.2f-%.2f duration=%.2f", shots[0].StartSeconds, shots[0].EndSeconds, shots[0].DurationSeconds)
	}
	if shots[1].InheritsFromID != shots[0].ID {
		t.Fatalf("second inheritsFrom = %q, want %q", shots[1].InheritsFromID, shots[0].ID)
	}
	if strings.Contains(shots[0].ActionText, "3.20") || strings.Contains(shots[0].ActionText, "谁人吹角") {
		t.Fatalf("first action leaked the next beat: %q", shots[0].ActionText)
	}
	if shots[0].CameraText != "低角度手持推近" || !strings.Contains(shots[1].AudioText, "谁人吹角") {
		t.Fatalf("parsed camera/audio wrong: first=%q second=%q", shots[0].CameraText, shots[1].AudioText)
	}
	if len(shots[0].Bindings.Characters) != 1 || shots[0].Bindings.Scene == nil || len(shots[0].Bindings.Props) != 1 {
		t.Fatalf("first bindings = %+v, want beat-scoped character/scene/prop", shots[0].Bindings)
	}

	second, err := service.SyncProject("project-a")
	if err != nil {
		t.Fatalf("SyncProject(second) error = %v", err)
	}
	if second.Created != 0 || second.Updated != 0 || second.Reused != 2 {
		t.Fatalf("second summary = %+v, want idempotent reuse", second)
	}

	resources.resources[3].Markdown = strings.Replace(
		resources.resources[3].Markdown,
		"韩三河在二号食堂后厨撑身坐起",
		"韩三河在二号食堂后厨猛地撑身坐起",
		1,
	)
	third, err := service.SyncProject("project-a")
	if err != nil {
		t.Fatalf("SyncProject(third) error = %v", err)
	}
	if third.Created != 0 || third.Updated != 1 || third.Reused != 1 {
		t.Fatalf("third summary = %+v, want only changed auto beat updated", third)
	}
	updatedShots, err := service.ListDocument("project-a", "storyboard-ep1")
	if err != nil {
		t.Fatalf("ListDocument(updated) error = %v", err)
	}
	if !strings.Contains(updatedShots[0].ActionText, "猛地撑身坐起") {
		t.Fatalf("updated first action = %q", updatedShots[0].ActionText)
	}
	fourth, err := service.SyncProject("project-a")
	if err != nil {
		t.Fatalf("SyncProject(fourth) error = %v", err)
	}
	if fourth.Created != 0 || fourth.Updated != 0 || fourth.Reused != 2 {
		t.Fatalf("fourth summary = %+v, want updated auto beats to remain idempotent", fourth)
	}
}

func TestSyncProjectDoesNotOverwriteManualProductionShot(t *testing.T) {
	service, repo, canonService := newTestShotService(t)
	resources := fakeShotDocumentResources{resources: []model.WorkspaceDocumentResourceRecord{
		{
			Type:       "storyboard",
			Title:      "第 01 组",
			Prompt:     "group prompt",
			PlainText:  "group text",
			Markdown:   "## 第 01 组\n\n- 0.00–2.00：人物推门；画面：中景。\n- 2.00–4.00：人物回头；画面：近景。",
			DocumentID: "storyboard-ep1",
			SectionID:  "section-shot-1",
		},
	}}
	canonService.SetDocumentResourceProvider(resources)
	service.SetDocumentResourceProvider(resources)

	if _, err := service.SyncProject("project-a"); err != nil {
		t.Fatalf("SyncProject(first) error = %v", err)
	}
	firstModel, err := repo.FindBySource("project-a", "storyboard-ep1", "section-shot-1", "beat-001")
	if err != nil {
		t.Fatalf("FindBySource(first beat) error = %v", err)
	}
	manual, err := service.Upsert(UpsertInput{
		ProjectID:        "project-a",
		DocumentID:       "storyboard-ep1",
		SectionID:        "section-shot-1",
		ShotKey:          "beat-001",
		Sequence:         1,
		StartSeconds:     0,
		EndSeconds:       2,
		DurationSeconds:  2,
		ActionText:       "用户手工改成单一近景动作",
		CameraText:       "固定近景",
		StateChangesJSON: `{}`,
		SourceHash:       firstModel.SourceHash,
		Status:           StatusReady,
	})
	if err != nil {
		t.Fatalf("manual Upsert() error = %v", err)
	}
	if manual.Status != StatusReady {
		t.Fatalf("manual status = %q", manual.Status)
	}

	second, err := service.SyncProject("project-a")
	if err != nil {
		t.Fatalf("SyncProject(second) error = %v", err)
	}
	if second.Created != 0 || second.Updated != 0 || second.Reused != 2 {
		t.Fatalf("second summary = %+v, want both beats reused", second)
	}
	persisted, err := repo.FindBySource("project-a", "storyboard-ep1", "section-shot-1", "beat-001")
	if err != nil {
		t.Fatalf("FindBySource(persisted) error = %v", err)
	}
	if persisted.ActionText != "用户手工改成单一近景动作" || persisted.CameraText != "固定近景" || persisted.Status != StatusReady {
		t.Fatalf("manual Production Shot was overwritten: %+v", persisted)
	}
}

func TestInferLegacyBindingsUsesStoryboardTitleAndSpecificSceneEvidence(t *testing.T) {
	assets := []servicecanon.AssetRecord{
		{ID: "scene-front", ResourceType: servicecanon.ResourceTypeScene, Name: "二号食堂打饭窗口与就餐区", PromptText: "对应场次：第 01 场「敲盆」、第 05 场「分餐」。\n核心物件：打饭窗口台面、卷帘门、铁皮餐盘。"},
		{ID: "scene-kitchen", ResourceType: servicecanon.ResourceTypeScene, Name: "二号食堂后厨", PromptText: "对应场次：第 02 场「泔水醒转」、第 04 场「陷阵分粮术」、第 07 场「电饭煲彩蛋」。\n核心物件：泔水桶、冰柜、调味罐、大铁锅。"},
		{ID: "scene-bridge", ResourceType: servicecanon.ResourceTypeScene, Name: "二号食堂前厅与后厨贯通区", PromptText: "对应场次：第 03 场「十五分钟」。\n核心物件：传菜口台面、塑料门帘。"},
		{ID: "scene-stove", ResourceType: servicecanon.ResourceTypeScene, Name: "后厨灶位与分餐台", PromptText: "对应场次：第 04 场「陷阵分粮术」。\n核心物件：灶眼、大铁锅、菜刀、砧板、分餐台。"},
		{ID: "scene-close", ResourceType: servicecanon.ResourceTypeScene, Name: "二号食堂就餐区与后厨门边", PromptText: "对应场次：第 06 场「收场与木勺」。\n核心物件：厨房门、折叠桌、餐盘、塑料椅。"},
		{ID: "scene-rice", ResourceType: servicecanon.ResourceTypeScene, Name: "后厨傍晚电饭煲彩蛋处", PromptText: "对应场次：第 07 场「电饭煲彩蛋」。\n核心物件：立式电饭煲、保温灯、旧木勺。"},
	}

	cases := []struct {
		name  string
		title string
		text  string
		want  string
	}{
		{name: "bridge by stage title", title: "第 05 组 · 十五分钟", text: "前厅通道与后厨门口折角，穿过塑料门帘进入后厨。", want: "scene-bridge"},
		{name: "stove by core objects", title: "第 06 组 · 传令生火", text: "后厨灶台前，王小五点燃灶眼，韩三和转身去砧板抽菜刀。", want: "scene-stove"},
		{name: "serving window by stage and place", title: "第 10 组 · 分餐（上）", text: "打饭窗口与就餐区，卷帘门升起，铁皮餐盘排开。", want: "scene-front"},
		{name: "close by stage title", title: "第 13 组 · 收场（上）", text: "打饭窗口内侧，最后一份餐打完，冉冬青数签收单。", want: "scene-close"},
		{name: "rice cooker specific zone beats broad kitchen", title: "第 15 组 · 电饭煲彩蛋", text: "傍晚的二号食堂后厨，韩三和蹲在电饭煲前端详，保温灯亮起。", want: "scene-rice"},
		{name: "rice cooker ending stays specific", title: "第 16 组 · 电饭煲彩蛋（结尾）", text: "电饭煲前，夜色进入后厨窗边，韩三和捧着旧木勺。", want: "scene-rice"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bindings := inferLegacyBindings(tc.title, tc.text, assets)
			if bindings.Scene == nil || bindings.Scene.CanonID != tc.want {
				t.Fatalf("scene binding = %+v, want %s", bindings.Scene, tc.want)
			}
		})
	}
}

func TestInferLegacyBindingsMatchesUniqueQualifiedPropAlias(t *testing.T) {
	assets := []servicecanon.AssetRecord{
		{ID: "char-han", ResourceType: servicecanon.ResourceTypeCharacter, Name: "韩三和"},
		{ID: "prop-spoon", ResourceType: servicecanon.ResourceTypeProp, Name: "陷阵木勺", PromptText: "深色旧木勺，勺柄刻陷阵二字"},
		{ID: "prop-bucket", ResourceType: servicecanon.ResourceTypeProp, Name: "蒸饭木桶", PromptText: "旧木桶"},
	}

	bindings := inferLegacyBindings("第 10 组 · 分餐（上）", "韩三和从腰间抽出木勺，把菜压平后继续分餐。", assets)
	if len(bindings.Props) != 1 || bindings.Props[0].CanonID != "prop-spoon" {
		t.Fatalf("prop bindings = %+v, want unique qualified prop alias prop-spoon", bindings.Props)
	}
}

func TestInferLegacyBindingsDoesNotGuessAmbiguousPropAlias(t *testing.T) {
	assets := []servicecanon.AssetRecord{
		{ID: "prop-old-spoon", ResourceType: servicecanon.ResourceTypeProp, Name: "旧木勺"},
		{ID: "prop-large-spoon", ResourceType: servicecanon.ResourceTypeProp, Name: "大木勺"},
	}

	bindings := inferLegacyBindings("第 1 组", "人物拿起木勺。", assets)
	if len(bindings.Props) != 0 {
		t.Fatalf("ambiguous prop bindings = %+v, want none", bindings.Props)
	}
}

func TestInheritBindingsCanSkipImplicitLegacyPropCarry(t *testing.T) {
	previous := Bindings{Props: []ResourceBinding{{CanonID: "prop-old"}}}
	current := Bindings{Props: []ResourceBinding{{CanonID: "prop-current"}}}

	resolved, err := inheritBindings(current, previous, `{}`, false)
	if err != nil {
		t.Fatalf("inheritBindings() error = %v", err)
	}
	if len(resolved.Props) != 1 || resolved.Props[0].CanonID != "prop-current" {
		t.Fatalf("legacy migration props = %+v, want only current explicit prop", resolved.Props)
	}
}

func TestSyncProjectReconcilesUntouchedLegacyDraftButRemainsIdempotent(t *testing.T) {
	service, repo, canonService := newTestShotService(t)
	resources := fakeShotDocumentResources{resources: []model.WorkspaceDocumentResourceRecord{
		{Type: "scene", Title: "二号食堂后厨", Prompt: "对应场次：第 02 场「泔水醒转」。\n核心物件：泔水桶、冰柜。", PlainText: "二号食堂后厨", Markdown: "## 二号食堂后厨\n泔水桶、冰柜", DocumentID: "scenes", SectionID: "scene-kitchen"},
		{Type: "scene", Title: "二号食堂打饭窗口与就餐区", Prompt: "对应场次：第 05 场「分餐」。\n核心物件：卷帘门、铁皮餐盘。", PlainText: "打饭窗口与就餐区", Markdown: "## 二号食堂打饭窗口与就餐区\n卷帘门、铁皮餐盘", DocumentID: "scenes", SectionID: "scene-front"},
		{Type: "prop", Title: "卷帘门", Prompt: "绿色旧卷帘门", PlainText: "卷帘门", Markdown: "## 卷帘门", DocumentID: "props", SectionID: "prop-shutter"},
		{Type: "storyboard", Title: "第 10 组 · 分餐（上）", Prompt: "打饭窗口与就餐区，卷帘门升起。", PlainText: "打饭窗口与就餐区，卷帘门升起。", Markdown: "## 第 10 组 · 分餐（上）\n打饭窗口与就餐区，卷帘门升起。", DocumentID: "storyboard-ep1", SectionID: "shot-10"},
	}}
	canonService.SetDocumentResourceProvider(resources)
	service.SetDocumentResourceProvider(resources)

	first, err := service.SyncProject("project-a")
	if err != nil {
		t.Fatalf("SyncProject(first) error = %v", err)
	}
	if first.Created != 1 {
		t.Fatalf("first summary = %+v, want one created", first)
	}
	shots, err := service.ListDocument("project-a", "storyboard-ep1")
	if err != nil || len(shots) != 1 || shots[0].Bindings.Scene == nil {
		t.Fatalf("first shot = %+v, err=%v", shots, err)
	}
	correctSceneID := shots[0].Bindings.Scene.CanonID

	allScenes, err := canonService.List("project-a", servicecanon.ResourceTypeScene)
	if err != nil {
		t.Fatalf("List(scene) error = %v", err)
	}
	wrongSceneID := ""
	for _, scene := range allScenes {
		if scene.Name == "二号食堂后厨" {
			wrongSceneID = scene.ID
		}
	}
	if wrongSceneID == "" || wrongSceneID == correctSceneID {
		t.Fatalf("scene ids wrong=%q correct=%q", wrongSceneID, correctSceneID)
	}

	model, err := repo.FindBySource("project-a", "storyboard-ep1", "shot-10", "")
	if err != nil {
		t.Fatalf("FindBySource() error = %v", err)
	}
	wrongBindings, _ := json.Marshal(Bindings{Scene: &ResourceBinding{CanonID: wrongSceneID}})
	model.BindingsJSON = string(wrongBindings)
	if err := repo.Upsert(model); err != nil {
		t.Fatalf("injecting legacy wrong binding: %v", err)
	}

	second, err := service.SyncProject("project-a")
	if err != nil {
		t.Fatalf("SyncProject(second) error = %v", err)
	}
	if second.Updated != 1 {
		t.Fatalf("second summary = %+v, want one safely reconciled draft", second)
	}
	reconciled, err := service.ListDocument("project-a", "storyboard-ep1")
	if err != nil || len(reconciled) != 1 || reconciled[0].Bindings.Scene == nil || reconciled[0].Bindings.Scene.CanonID != correctSceneID {
		t.Fatalf("reconciled shot = %+v, err=%v", reconciled, err)
	}

	third, err := service.SyncProject("project-a")
	if err != nil {
		t.Fatalf("SyncProject(third) error = %v", err)
	}
	if third.Updated != 0 || third.Reused != 1 {
		t.Fatalf("third summary = %+v, want idempotent reuse", third)
	}
}

func newTestShotService(t *testing.T) (*Service, *repository.ShotManifestRepository, *servicecanon.Service) {
	t.Helper()
	db, err := repository.OpenWorkspaceDB(filepath.Join(t.TempDir(), "workspace.db"))
	if err != nil {
		t.Fatalf("OpenWorkspaceDB() error = %v", err)
	}
	testutil.CloseDB(t, db)
	if err := db.Create(&domain.WorkspaceProjectModel{
		ID:          "project-a",
		Name:        "Test",
		Category:    "drama",
		Status:      "active",
		RelativeDir: "project-a",
	}).Error; err != nil {
		t.Fatalf("creating project: %v", err)
	}
	canonRepo := repository.NewCanonRepositoryFromDB(db)
	canonService := servicecanon.NewService(canonRepo, nil)
	shotRepo := repository.NewShotManifestRepositoryFromDB(db)
	return NewService(shotRepo, canonService, nil), shotRepo, canonService
}

func mustCanonCore(t *testing.T, service *servicecanon.Service, input servicecanon.SourceResource) servicecanon.AssetRecord {
	t.Helper()
	result, err := service.EnsureCoreFromSource(input)
	if err != nil {
		t.Fatalf("EnsureCoreFromSource() error = %v", err)
	}
	record, err := service.Get(input.ProjectID, result.Asset.ID)
	if err != nil {
		t.Fatalf("Canon Get() error = %v", err)
	}
	return record
}

func jsonEquivalent(t *testing.T, first string, second string) bool {
	t.Helper()
	var a any
	var b any
	if err := json.Unmarshal([]byte(first), &a); err != nil {
		t.Fatalf("decode first JSON: %v", err)
	}
	if err := json.Unmarshal([]byte(second), &b); err != nil {
		t.Fatalf("decode second JSON: %v", err)
	}
	firstBytes, _ := json.Marshal(a)
	secondBytes, _ := json.Marshal(b)
	return string(firstBytes) == string(secondBytes)
}
