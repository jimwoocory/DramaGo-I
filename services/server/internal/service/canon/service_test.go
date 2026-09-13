package canon

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/mediago-dev/mediago-drama/services/server/internal/domain"
	"github.com/mediago-dev/mediago-drama/services/server/internal/repository"
	"github.com/mediago-dev/mediago-drama/services/server/internal/service/model"
	"github.com/mediago-dev/mediago-drama/services/server/internal/testutil"
	"gorm.io/gorm"
)

func TestEnsureCoreFromSourceCreatesAndVersions(t *testing.T) {
	service, repo := newTestService(t)

	first, err := service.EnsureCoreFromSource(SourceResource{
		ProjectID:        "project-a",
		ResourceType:     "character",
		ResourceID:       "section-lintong",
		SourceDocumentID: "characters",
		Name:             "林书彤",
		SpecJSON:         `{"hair":"black"}`,
		PromptText:       "黑色长发，鹅蛋脸",
	})
	if err != nil {
		t.Fatalf("EnsureCoreFromSource(first) error = %v", err)
	}
	if !first.Created || !first.Changed || first.Asset.Version != 1 || first.Asset.Status != StatusDraft {
		t.Fatalf("first result = %+v, want created draft v1", first)
	}

	same, err := service.EnsureCoreFromSource(SourceResource{
		ProjectID:        "project-a",
		ResourceType:     "character",
		ResourceID:       "section-lintong",
		SourceDocumentID: "characters",
		Name:             "林书彤",
		SpecJSON:         `{"hair":"black"}`,
		PromptText:       "黑色长发，鹅蛋脸",
	})
	if err != nil {
		t.Fatalf("EnsureCoreFromSource(same) error = %v", err)
	}
	if same.Created || same.Changed || same.Asset.Version != 1 {
		t.Fatalf("same result = %+v, want unchanged v1", same)
	}

	changed, err := service.EnsureCoreFromSource(SourceResource{
		ProjectID:        "project-a",
		ResourceType:     "character",
		ResourceID:       "section-lintong",
		SourceDocumentID: "characters",
		Name:             "林书彤",
		SpecJSON:         `{"hair":"black","eyes":"brown"}`,
		PromptText:       "黑色长发，棕色眼睛，鹅蛋脸",
	})
	if err != nil {
		t.Fatalf("EnsureCoreFromSource(changed) error = %v", err)
	}
	if !changed.Changed || changed.Asset.Version != 2 {
		t.Fatalf("changed result = %+v, want changed v2", changed)
	}

	persisted, err := repo.GetCanonAsset("project-a", first.Asset.ID)
	if err != nil {
		t.Fatalf("GetCanonAsset() error = %v", err)
	}
	if persisted.PromptText != "黑色长发，棕色眼睛，鹅蛋脸" {
		t.Fatalf("persisted prompt = %q", persisted.PromptText)
	}
}

func TestEnsureCoreFromSourceDoesNotOverwriteLockedCanon(t *testing.T) {
	service, repo := newTestService(t)
	first, err := service.EnsureCoreFromSource(SourceResource{
		ProjectID:        "project-a",
		ResourceType:     "scene",
		ResourceID:       "section-cafeteria",
		SourceDocumentID: "scenes",
		Name:             "食堂",
		PromptText:       "室内食堂，绿色卷帘门",
	})
	if err != nil {
		t.Fatalf("EnsureCoreFromSource(first) error = %v", err)
	}
	if _, err := repo.UpdateCanonAsset("project-a", first.Asset.ID, map[string]any{"status": StatusLocked}); err != nil {
		t.Fatalf("locking canon: %v", err)
	}

	locked, err := service.EnsureCoreFromSource(SourceResource{
		ProjectID:        "project-a",
		ResourceType:     "scene",
		ResourceID:       "section-cafeteria",
		SourceDocumentID: "scenes",
		Name:             "食堂",
		PromptText:       "室外食堂，银灰卷帘门",
	})
	if err != nil {
		t.Fatalf("EnsureCoreFromSource(locked) error = %v", err)
	}
	if !locked.Locked || locked.Changed {
		t.Fatalf("locked result = %+v, want locked unchanged", locked)
	}
	if locked.Asset.PromptText != "室内食堂，绿色卷帘门" || locked.Asset.Version != 1 {
		t.Fatalf("locked canon was overwritten: %+v", locked.Asset)
	}
}

func TestCreateVariantInheritsSourceIdentity(t *testing.T) {
	service, _ := newTestService(t)
	core, err := service.EnsureCoreFromSource(SourceResource{
		ProjectID:        "project-a",
		ResourceType:     "character",
		ResourceID:       "section-lintong",
		SourceDocumentID: "characters",
		Name:             "林书彤",
		PromptText:       "固定脸型与黑色长发",
	})
	if err != nil {
		t.Fatalf("EnsureCoreFromSource() error = %v", err)
	}

	variant, err := service.CreateVariant(VariantInput{
		ProjectID:   "project-a",
		ParentID:    core.Asset.ID,
		VariantKind: "look",
		Name:        "婚礼造型",
		SpecJSON:    `{"wardrobe":"white wedding dress"}`,
		PromptText:  "白色婚纱，盘发",
		Status:      StatusApproved,
	})
	if err != nil {
		t.Fatalf("CreateVariant() error = %v", err)
	}
	if domain.StringValue(variant.ParentID) != core.Asset.ID ||
		variant.ResourceType != core.Asset.ResourceType ||
		variant.ResourceID != core.Asset.ResourceID ||
		variant.SourceDocumentID != core.Asset.SourceDocumentID ||
		variant.VariantKind != "look" {
		t.Fatalf("variant = %+v, want inherited source identity", variant)
	}
}

func TestSyncDocumentResourcesMaterializesNestedVariantsWithoutPollutingCore(t *testing.T) {
	service, repo := newTestService(t)
	markdown := strings.Join([]string{
		"## 林书彤",
		"固定鹅蛋脸，黑色长发，23岁。",
		"### 造型变体：婚礼",
		"白色婚纱，盘发。",
		"### 造型变体：工作",
		"深灰西装，低马尾。",
	}, "\n")
	resource := model.WorkspaceDocumentResourceRecord{
		Type:       "character",
		Title:      "林书彤",
		Prompt:     markdown,
		PlainText:  markdown,
		Markdown:   markdown,
		DocumentID: "characters",
		SectionID:  "section-lintong",
	}

	first, err := service.SyncDocumentResources("project-a", []model.WorkspaceDocumentResourceRecord{resource})
	if err != nil {
		t.Fatalf("SyncDocumentResources(first) error = %v", err)
	}
	if first.Created != 3 || first.Updated != 0 {
		t.Fatalf("first summary = %+v, want core + 2 variants created", first)
	}
	core, err := repo.FindCanonCoreBySource("project-a", "character", "characters", "section-lintong")
	if err != nil {
		t.Fatalf("FindCanonCoreBySource() error = %v", err)
	}
	if !strings.Contains(core.PromptText, "固定鹅蛋脸") {
		t.Fatalf("core prompt = %q, want identity facts", core.PromptText)
	}
	for _, forbidden := range []string{"婚礼", "白色婚纱", "工作", "深灰西装"} {
		if strings.Contains(core.PromptText, forbidden) {
			t.Fatalf("core prompt leaked variant %q: %s", forbidden, core.PromptText)
		}
	}
	variants, err := repo.ListCanonVariants("project-a", core.ID)
	if err != nil {
		t.Fatalf("ListCanonVariants() error = %v", err)
	}
	if len(variants) != 2 {
		t.Fatalf("variants = %+v, want 2", variants)
	}
	byName := map[string]domain.CanonAssetModel{}
	for _, variant := range variants {
		byName[variant.Name] = variant
		if variant.VariantKind != "look" {
			t.Fatalf("variant %q kind = %q, want look", variant.Name, variant.VariantKind)
		}
	}
	if !strings.Contains(byName["婚礼"].PromptText, "白色婚纱") || !strings.Contains(byName["工作"].PromptText, "深灰西装") {
		t.Fatalf("variant prompts = %+v", byName)
	}
	if core.Version != 1 || byName["婚礼"].Version != 1 || byName["工作"].Version != 1 {
		t.Fatalf("initial versions core=%d wedding=%d work=%d", core.Version, byName["婚礼"].Version, byName["工作"].Version)
	}

	resource.Markdown = strings.Replace(markdown, "白色婚纱，盘发。", "白色婚纱，盘发，增加珍珠头纱。", 1)
	resource.Prompt = resource.Markdown
	resource.PlainText = resource.Markdown
	second, err := service.SyncDocumentResources("project-a", []model.WorkspaceDocumentResourceRecord{resource})
	if err != nil {
		t.Fatalf("SyncDocumentResources(second) error = %v", err)
	}
	if second.Created != 0 || second.Updated != 1 {
		t.Fatalf("second summary = %+v, want only wedding variant updated", second)
	}
	coreAfter, err := repo.GetCanonAsset("project-a", core.ID)
	if err != nil {
		t.Fatalf("GetCanonAsset(core after) error = %v", err)
	}
	if coreAfter.Version != 1 {
		t.Fatalf("core version = %d, want unchanged 1 when only variant changed", coreAfter.Version)
	}
	variantsAfter, err := repo.ListCanonVariants("project-a", core.ID)
	if err != nil {
		t.Fatalf("ListCanonVariants(after) error = %v", err)
	}
	byName = map[string]domain.CanonAssetModel{}
	for _, variant := range variantsAfter {
		byName[variant.Name] = variant
	}
	if byName["婚礼"].Version != 2 || byName["工作"].Version != 1 {
		t.Fatalf("variant versions after = wedding:%d work:%d", byName["婚礼"].Version, byName["工作"].Version)
	}
}

func TestSplitSourceResourceRecognizesSceneZonesAndPropStates(t *testing.T) {
	scene := splitSourceResource("scene", strings.Join([]string{
		"## 食堂",
		"老厂房食堂前厅，绿色卷帘门。",
		"### 区域：取餐口",
		"不锈钢台面与保温槽。",
		"### 场景变体：夜间",
		"主灯关闭，仅保留应急灯。",
	}, "\n"), "")
	if strings.Contains(scene.CorePrompt, "取餐口") || strings.Contains(scene.CorePrompt, "夜间") {
		t.Fatalf("scene core leaked variants: %s", scene.CorePrompt)
	}
	if len(scene.Variants) != 2 || scene.Variants[0].Kind != "zone" || scene.Variants[0].Name != "取餐口" || scene.Variants[1].Kind != "scene" || scene.Variants[1].Name != "夜间" {
		t.Fatalf("scene variants = %+v", scene.Variants)
	}

	prop := splitSourceResource("prop", strings.Join([]string{
		"## 木勺",
		"深棕色旧木勺。",
		"### 状态变体：断裂",
		"勺柄中段断开。",
	}, "\n"), "")
	if len(prop.Variants) != 1 || prop.Variants[0].Kind != "state" || prop.Variants[0].Name != "断裂" {
		t.Fatalf("prop variants = %+v", prop.Variants)
	}
	if strings.Contains(prop.CorePrompt, "断裂") {
		t.Fatalf("prop core leaked state variant: %s", prop.CorePrompt)
	}
}

func TestEnsureCoreRejectsStoryboardAsCanonResource(t *testing.T) {
	service, _ := newTestService(t)
	if _, err := service.EnsureCoreFromSource(SourceResource{
		ProjectID:        "project-a",
		ResourceType:     "storyboard",
		ResourceID:       "shot-1",
		SourceDocumentID: "storyboard-ep1",
	}); err == nil {
		t.Fatal("EnsureCoreFromSource(storyboard) error = nil, want validation error")
	}
}

func TestSyncDocumentResourcesSkipsStoryboardAndCreatesCanonCores(t *testing.T) {
	service, repo := newTestService(t)
	summary, err := service.SyncDocumentResources("project-a", []model.WorkspaceDocumentResourceRecord{
		{
			Type:       "character",
			Title:      "林书彤",
			Summary:    "女主角",
			Prompt:     "固定脸型，黑色长发",
			PlainText:  "林书彤，23岁。",
			Markdown:   "## 林书彤\n林书彤，23岁。",
			DocumentID: "characters",
			SectionID:  "section-lintong",
		},
		{
			Type:       "scene",
			Title:      "食堂",
			Prompt:     "室内食堂，绿色卷帘门",
			PlainText:  "食堂内景。",
			Markdown:   "## 食堂\n食堂内景。",
			DocumentID: "scenes",
			SectionID:  "section-cafeteria",
		},
		{
			Type:       "storyboard",
			Title:      "第1组",
			Prompt:     "镜头推进",
			Markdown:   "## 第1组\n镜头推进",
			DocumentID: "storyboard-ep1",
			SectionID:  "section-shot-1",
		},
	})
	if err != nil {
		t.Fatalf("SyncDocumentResources() error = %v", err)
	}
	if summary.Created != 2 || summary.Skipped != 1 || len(summary.Assets) != 2 {
		t.Fatalf("summary = %+v, want 2 created and 1 skipped", summary)
	}
	assets, err := repo.ListCanonAssets("project-a", "")
	if err != nil {
		t.Fatalf("ListCanonAssets() error = %v", err)
	}
	if len(assets) != 2 {
		t.Fatalf("canon assets = %+v, want 2", assets)
	}
}

func TestEnsureReferenceFromSelectedAssetReusesPhysicalMedia(t *testing.T) {
	service, _, db := newTestServiceWithDB(t)
	core, err := service.EnsureCoreFromSource(SourceResource{
		ProjectID:        "project-a",
		ResourceType:     "character",
		ResourceID:       "section-lintong",
		SourceDocumentID: "characters",
		Name:             "林书彤",
		PromptText:       "固定脸型，黑色长发",
	})
	if err != nil {
		t.Fatalf("EnsureCoreFromSource() error = %v", err)
	}

	asset := domain.AssetModel{
		ID:        "asset-selected-lintong",
		ProjectID: domain.StringPtr("project-a"),
		Kind:      "image",
		Filename:  "lintong.png",
		MIMEType:  "image/png",
		SizeBytes: 42,
		RelPath:   "projects/project-a/library/lintong.png",
		Source:    "generated",
	}
	if err := db.Create(&asset).Error; err != nil {
		t.Fatalf("creating media asset: %v", err)
	}
	selected := domain.ProjectSelectedAssetModel{
		ID:               "selected-lintong",
		ProjectID:        "project-a",
		ResourceType:     "character",
		ResourceID:       domain.StringPtr("section-lintong"),
		ResourceTitle:    domain.StringPtr("林书彤"),
		AssetID:          asset.ID,
		SourceDocumentID: domain.StringPtr("characters"),
	}

	reference, created, err := service.EnsureReferenceFromSelectedAsset(selected)
	if err != nil {
		t.Fatalf("EnsureReferenceFromSelectedAsset() error = %v", err)
	}
	if !created || reference.CanonAssetID != core.Asset.ID || reference.AssetID != asset.ID || reference.Role != "identity" {
		t.Fatalf("reference = %+v created=%v", reference, created)
	}
	if reference.Asset.ID != asset.ID {
		t.Fatalf("reference asset preload = %+v, want original physical asset", reference.Asset)
	}

	again, createdAgain, err := service.EnsureReferenceFromSelectedAsset(selected)
	if err != nil {
		t.Fatalf("EnsureReferenceFromSelectedAsset(again) error = %v", err)
	}
	if createdAgain || again.ID != reference.ID {
		t.Fatalf("again = %+v created=%v, want idempotent reuse", again, createdAgain)
	}
}

func TestBindReferenceSupportsVariantAnchorsAndRejectsCrossProjectAssets(t *testing.T) {
	service, repo, db := newTestServiceWithDB(t)
	core, err := service.EnsureCoreFromSource(SourceResource{
		ProjectID:        "project-a",
		ResourceType:     ResourceTypeCharacter,
		ResourceID:       "section-lintong",
		SourceDocumentID: "characters",
		Name:             "林书彤",
		PromptText:       "固定身份",
	})
	if err != nil {
		t.Fatalf("EnsureCoreFromSource() error = %v", err)
	}
	variant, err := service.CreateVariant(VariantInput{
		ProjectID:   "project-a",
		ParentID:    core.Asset.ID,
		VariantKind: "look",
		Name:        "婚礼",
		PromptText:  "白色婚纱，盘发",
	})
	if err != nil {
		t.Fatalf("CreateVariant() error = %v", err)
	}
	asset := domain.AssetModel{
		ID: "asset-wedding", ProjectID: domain.StringPtr("project-a"), Kind: "image",
		Filename: "wedding.png", MIMEType: "image/png", SizeBytes: 42,
		RelPath: "projects/project-a/library/wedding.png", Source: "generated",
	}
	if err := db.Create(&asset).Error; err != nil {
		t.Fatalf("creating wedding asset: %v", err)
	}

	bound, err := service.BindReference(ReferenceInput{
		ProjectID: "project-a", CanonAssetID: variant.ID, AssetID: asset.ID, Locked: true,
	})
	if err != nil {
		t.Fatalf("BindReference() error = %v", err)
	}
	if bound.Role != "look" || bound.Priority != 900 || !bound.Locked {
		t.Fatalf("bound reference = %+v, want default look/900/locked", bound)
	}
	updated, err := service.BindReference(ReferenceInput{
		ProjectID: "project-a", CanonAssetID: variant.ID, AssetID: asset.ID,
		Role: "full_body", Priority: 950, Locked: false,
	})
	if err != nil {
		t.Fatalf("BindReference(update) error = %v", err)
	}
	if updated.ID != bound.ID || updated.Role != "full_body" || updated.Priority != 950 || updated.Locked {
		t.Fatalf("updated reference = %+v, want idempotent policy update", updated)
	}
	refs, err := repo.ListCanonReferences("project-a", variant.ID)
	if err != nil || len(refs) != 1 {
		t.Fatalf("variant references = %+v error=%v, want one", refs, err)
	}

	if err := db.Create(&domain.WorkspaceProjectModel{
		ID: "project-b", Name: "Other", Category: "drama", Status: "active", RelativeDir: "project-b",
	}).Error; err != nil {
		t.Fatalf("creating other project: %v", err)
	}
	otherAsset := domain.AssetModel{
		ID: "asset-other", ProjectID: domain.StringPtr("project-b"), Kind: "image",
		Filename: "other.png", MIMEType: "image/png", SizeBytes: 12,
		RelPath: "projects/project-b/library/other.png", Source: "generated",
	}
	if err := db.Create(&otherAsset).Error; err != nil {
		t.Fatalf("creating other asset: %v", err)
	}
	if _, err := service.BindReference(ReferenceInput{
		ProjectID: "project-a", CanonAssetID: variant.ID, AssetID: otherAsset.ID,
	}); err == nil {
		t.Fatal("BindReference(cross-project) error = nil, want rejection")
	}
}

func newTestService(t *testing.T) (*Service, *repository.CanonRepository) {
	t.Helper()
	service, repo, _ := newTestServiceWithDB(t)
	return service, repo
}

func newTestServiceWithDB(t *testing.T) (*Service, *repository.CanonRepository, *gorm.DB) {
	t.Helper()
	db, err := repository.OpenWorkspaceDB(filepath.Join(t.TempDir(), "workspace.db"))
	if err != nil {
		t.Fatalf("OpenWorkspaceDB() error = %v", err)
	}
	testutil.CloseDB(t, db)
	if err := db.Create(&domain.WorkspaceProjectModel{
		ID:          "project-a",
		Name:        "Test Project",
		Category:    "drama",
		Status:      "active",
		RelativeDir: "project-a",
	}).Error; err != nil {
		t.Fatalf("creating project: %v", err)
	}
	repo := repository.NewCanonRepositoryFromDB(db)
	return NewService(repo, nil), repo, db
}
