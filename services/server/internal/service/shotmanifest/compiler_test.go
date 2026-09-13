package shotmanifest

import (
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/mediago-dev/mediago-drama/services/server/internal/domain"
	"github.com/mediago-dev/mediago-drama/services/server/internal/repository"
	servicecanon "github.com/mediago-dev/mediago-drama/services/server/internal/service/canon"
	"github.com/mediago-dev/mediago-drama/services/server/internal/testutil"
)

type fakeCompilerCanon struct {
	records map[string]servicecanon.AssetRecord
}

type fakeContinuityAssets struct {
	assetID      string
	found        bool
	err          error
	shotAssetID  string
	shotFound    bool
	shotErr      error
	project      string
	doc          string
	section      string
	shotProject  string
	shotManifest string
}

func (fake *fakeContinuityAssets) LatestImageAssetIDForSection(projectID string, documentID string, sectionID string) (string, bool, error) {
	fake.project = projectID
	fake.doc = documentID
	fake.section = sectionID
	return fake.assetID, fake.found, fake.err
}

func (fake *fakeContinuityAssets) LatestImageAssetIDForShotManifest(projectID string, shotManifestID string) (string, bool, error) {
	fake.shotProject = projectID
	fake.shotManifest = shotManifestID
	return fake.shotAssetID, fake.shotFound, fake.shotErr
}

func (fake fakeCompilerCanon) Get(_ string, id string) (servicecanon.AssetRecord, error) {
	record, ok := fake.records[id]
	if !ok {
		return servicecanon.AssetRecord{}, repository.ErrRecordNotFound
	}
	return record, nil
}

func (fake fakeCompilerCanon) List(_ string, resourceType string) ([]servicecanon.AssetRecord, error) {
	result := []servicecanon.AssetRecord{}
	for _, record := range fake.records {
		if strings.TrimSpace(resourceType) != "" && record.ResourceType != resourceType {
			continue
		}
		result = append(result, record)
	}
	return result, nil
}

func TestCompileLegacyDraftEnrichesUniqueQualifiedPropAlias(t *testing.T) {
	canon := fakeCompilerCanon{records: map[string]servicecanon.AssetRecord{
		"prop-spoon": {
			ID:           "prop-spoon",
			ProjectID:    "project-a",
			ResourceType: servicecanon.ResourceTypeProp,
			Name:         "陷阵木勺",
			PromptText:   "深色全木长柄勺，勺柄刻陷阵二字，器型固定",
			References: []servicecanon.ReferenceRecord{{
				ID: "ref-spoon", AssetID: "asset-spoon", Role: "prop_master", Priority: 100,
			}},
		},
		"prop-bucket": {
			ID: "prop-bucket", ProjectID: "project-a", ResourceType: servicecanon.ResourceTypeProp, Name: "蒸饭木桶",
		},
	}}
	service := &Service{canon: canon}

	compiled, err := service.Compile(Record{
		ProjectID:  "project-a",
		DocumentID: "storyboard-ep1",
		SectionID:  "shot-10",
		Status:     StatusDraft,
		ActionText: "韩三和从腰间抽出木勺，把菜压平后继续分餐。",
		Bindings:   Bindings{},
	})
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	if !strings.Contains(compiled.Prompt, "道具「陷阵木勺」") {
		t.Fatalf("compiled prompt missing enriched Prop Canon:\n%s", compiled.Prompt)
	}
	if !reflect.DeepEqual(compiled.ReferenceAssetIDs, []string{"asset-spoon"}) {
		t.Fatalf("reference assets = %#v, want asset-spoon", compiled.ReferenceAssetIDs)
	}
}

func TestCompileAndPersistIsDeterministicAndMapsCanonReferences(t *testing.T) {
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

	canon := fakeCompilerCanon{records: map[string]servicecanon.AssetRecord{
		"char-core": {
			ID:               "char-core",
			ProjectID:        "project-a",
			ResourceType:     servicecanon.ResourceTypeCharacter,
			ResourceID:       "section-char",
			SourceDocumentID: "characters",
			Name:             "林书彤",
			PromptText:       "23岁中国女性，固定鹅蛋脸，黑色长发",
			Status:           servicecanon.StatusLocked,
			References: []servicecanon.ReferenceRecord{{
				ID:       "ref-char",
				AssetID:  "asset-char",
				Role:     "identity",
				Priority: 100,
				Locked:   true,
			}},
		},
		"char-look-wedding": {
			ID:               "char-look-wedding",
			ProjectID:        "project-a",
			ResourceType:     servicecanon.ResourceTypeCharacter,
			ResourceID:       "section-char",
			SourceDocumentID: "characters",
			ParentID:         "char-core",
			VariantKind:      "look",
			Name:             "婚礼造型",
			PromptText:       "白色婚纱，盘发",
			Status:           servicecanon.StatusApproved,
			References: []servicecanon.ReferenceRecord{{
				ID:       "ref-look",
				AssetID:  "asset-look",
				Role:     "look",
				Priority: 90,
				Locked:   true,
			}},
		},
		"scene-core": {
			ID:               "scene-core",
			ProjectID:        "project-a",
			ResourceType:     servicecanon.ResourceTypeScene,
			ResourceID:       "section-scene",
			SourceDocumentID: "scenes",
			Name:             "婚礼大厅",
			PromptText:       "室内婚礼大厅，暖白顶灯，固定空间结构",
			Status:           servicecanon.StatusApproved,
			References: []servicecanon.ReferenceRecord{{
				ID:       "ref-scene",
				AssetID:  "asset-scene",
				Role:     "scene_master",
				Priority: 80,
				Locked:   true,
			}},
		},
		"prop-core": {
			ID:               "prop-core",
			ProjectID:        "project-a",
			ResourceType:     servicecanon.ResourceTypeProp,
			ResourceID:       "section-prop",
			SourceDocumentID: "props",
			Name:             "捧花",
			PromptText:       "白玫瑰与浅绿色叶材组成的圆形捧花",
			Status:           servicecanon.StatusApproved,
			References: []servicecanon.ReferenceRecord{{
				ID:       "ref-prop",
				AssetID:  "asset-prop",
				Role:     "prop_master",
				Priority: 1000,
			}},
		},
	}}

	repo := repository.NewShotManifestRepositoryFromDB(db)
	service := NewService(repo, canon, nil)
	record, err := service.Upsert(UpsertInput{
		ProjectID:  "project-a",
		DocumentID: "storyboard-ep1",
		SectionID:  "section-shot-1",
		ShotKey:    "shot-1",
		Sequence:   1,
		Bindings: Bindings{
			Characters: []CharacterBinding{{CanonID: "char-core", VariantID: "char-look-wedding"}},
			Scene:      &ResourceBinding{CanonID: "scene-core"},
			Props:      []ResourceBinding{{CanonID: "prop-core"}},
		},
		StateChangesJSON: `{"characters":{"char-core":{"look":{"wardrobe":"wedding","hair":"updo"},"props":{"bouquet":"held"}}},"scene":{"zone":"main_hall"}}`,
		ActionText:       "林书彤双手持捧花，向前走两步后停下。",
		CameraText:       "中景稳定推进到近景。",
		AudioText:        "无台词，仅保留轻微脚步声。",
		StyleProfileID:   "cinematic-clean-v1",
	})
	if err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	first, err := service.CompileAndPersist("project-a", record.ID)
	if err != nil {
		t.Fatalf("CompileAndPersist() error = %v", err)
	}
	second, err := service.Compile(record)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("compiler is not deterministic:\nfirst=%+v\nsecond=%+v", first, second)
	}

	for _, fragment := range []string{
		"角色「林书彤」：23岁中国女性，固定鹅蛋脸，黑色长发；白色婚纱，盘发",
		"场景「婚礼大厅」：室内婚礼大厅，暖白顶灯，固定空间结构",
		"道具「捧花」：白玫瑰与浅绿色叶材组成的圆形捧花",
		"道具一致性约束：已绑定道具的形状、材质、颜色和识别细节以对应 Prop Master 参考为视觉权威",
		`连续性状态：{"characters":{"char-core":{"look":{"hair":"updo","wardrobe":"wedding"},"props":{"bouquet":"held"}}},"scene":{"zone":"main_hall"}}`,
		"动作：林书彤双手持捧花，向前走两步后停下。",
		"镜头：中景稳定推进到近景。",
		"音频：无台词，仅保留轻微脚步声。",
		"风格配置：cinematic-clean-v1",
	} {
		if !strings.Contains(first.Prompt, fragment) {
			t.Fatalf("compiled prompt missing %q:\n%s", fragment, first.Prompt)
		}
	}
	wantAssetIDs := []string{"asset-char", "asset-look", "asset-scene", "asset-prop"}
	if !reflect.DeepEqual(first.ReferenceAssetIDs, wantAssetIDs) {
		t.Fatalf("ReferenceAssetIDs = %#v, want %#v", first.ReferenceAssetIDs, wantAssetIDs)
	}
	if len(first.References) != 4 {
		t.Fatalf("References = %#v, want 4", first.References)
	}
	seen := map[string]CompiledReference{}
	for _, reference := range first.References {
		seen[reference.AssetID] = reference
	}
	if seen["asset-char"].DocumentID != "characters" || seen["asset-char"].BlockID != "section-char" || seen["asset-char"].Role != "identity" {
		t.Fatalf("character reference = %+v", seen["asset-char"])
	}
	if seen["asset-scene"].DocumentID != "scenes" || seen["asset-scene"].BlockID != "section-scene" {
		t.Fatalf("scene reference = %+v", seen["asset-scene"])
	}

	persisted, err := repo.Get("project-a", record.ID)
	if err != nil {
		t.Fatalf("repo.Get() error = %v", err)
	}
	if persisted.CompiledPrompt != first.Prompt {
		t.Fatalf("persisted CompiledPrompt = %q, want compiler output", persisted.CompiledPrompt)
	}
}

func TestCompileAddsPreviousShotAfterIdentityAndLookBeforeSceneReferences(t *testing.T) {
	db, err := repository.OpenWorkspaceDB(filepath.Join(t.TempDir(), "workspace.db"))
	if err != nil {
		t.Fatalf("OpenWorkspaceDB() error = %v", err)
	}
	testutil.CloseDB(t, db)
	if err := db.Create(&domain.WorkspaceProjectModel{
		ID: "project-a", Name: "Test", Category: "drama", Status: "active", RelativeDir: "project-a",
	}).Error; err != nil {
		t.Fatalf("creating project: %v", err)
	}
	canon := fakeCompilerCanon{records: map[string]servicecanon.AssetRecord{
		"char-core": {
			ID: "char-core", ResourceType: servicecanon.ResourceTypeCharacter, SourceDocumentID: "characters", ResourceID: "char",
			Name: "角色", PromptText: "固定脸", References: []servicecanon.ReferenceRecord{{AssetID: "identity", Role: "identity", Priority: 1000, Locked: true}},
		},
		"look": {
			ID: "look", ParentID: "char-core", ResourceType: servicecanon.ResourceTypeCharacter, SourceDocumentID: "characters", ResourceID: "char",
			Name: "婚礼", PromptText: "婚礼造型", References: []servicecanon.ReferenceRecord{{AssetID: "look-anchor", Role: "look", Priority: 900, Locked: true}},
		},
		"scene-core": {
			ID: "scene-core", ResourceType: servicecanon.ResourceTypeScene, SourceDocumentID: "scenes", ResourceID: "hall",
			Name: "走廊", PromptText: "固定走廊", References: []servicecanon.ReferenceRecord{{AssetID: "scene-master", Role: "scene_master", Priority: 700, Locked: true}},
		},
		"scene-night": {
			ID: "scene-night", ParentID: "scene-core", ResourceType: servicecanon.ResourceTypeScene, SourceDocumentID: "scenes", ResourceID: "hall",
			Name: "夜间", PromptText: "夜间灯光", References: []servicecanon.ReferenceRecord{{AssetID: "scene-night", Role: "scene_variant", Priority: 800, Locked: true}},
		},
	}}
	repo := repository.NewShotManifestRepositoryFromDB(db)
	service := NewService(repo, canon, nil)
	continuity := &fakeContinuityAssets{assetID: "previous-shot-image", found: true}
	service.SetContinuityAssetProvider(continuity)
	previous, err := service.Upsert(UpsertInput{
		ProjectID: "project-a", DocumentID: "storyboard", SectionID: "shot-1", Sequence: 1,
		Bindings:         Bindings{Characters: []CharacterBinding{{CanonID: "char-core", VariantID: "look"}}, Scene: &ResourceBinding{CanonID: "scene-core", VariantID: "scene-night"}},
		StateChangesJSON: `{}`,
	})
	if err != nil {
		t.Fatalf("Upsert(previous) error = %v", err)
	}
	current, err := service.Upsert(UpsertInput{
		ProjectID: "project-a", DocumentID: "storyboard", SectionID: "shot-2", Sequence: 2,
		Bindings:         Bindings{Characters: []CharacterBinding{{CanonID: "char-core"}}, Scene: &ResourceBinding{CanonID: "scene-core"}},
		StateChangesJSON: `{}`,
	})
	if err != nil {
		t.Fatalf("Upsert(current) error = %v", err)
	}
	compiled, err := service.Compile(current)
	if err != nil {
		t.Fatalf("Compile(current) error = %v", err)
	}
	want := []string{"identity", "look-anchor", "previous-shot-image", "scene-night"}
	if !reflect.DeepEqual(compiled.ReferenceAssetIDs, want) {
		t.Fatalf("ReferenceAssetIDs = %#v, want %#v", compiled.ReferenceAssetIDs, want)
	}
	if continuity.project != "project-a" || continuity.doc != previous.DocumentID || continuity.section != previous.SectionID {
		t.Fatalf("continuity lookup = %q/%q/%q, want previous shot source", continuity.project, continuity.doc, continuity.section)
	}
	var previousRef CompiledReference
	for _, reference := range compiled.References {
		if reference.AssetID == "previous-shot-image" {
			previousRef = reference
			break
		}
	}
	if previousRef.Role != "previous_shot" || previousRef.Scope != "continuity" {
		t.Fatalf("previous shot reference = %+v", previousRef)
	}
}

func TestCompileProductionShotUsesExactPreviousShotManifestAsset(t *testing.T) {
	db, err := repository.OpenWorkspaceDB(filepath.Join(t.TempDir(), "workspace.db"))
	if err != nil {
		t.Fatalf("OpenWorkspaceDB() error = %v", err)
	}
	testutil.CloseDB(t, db)
	if err := db.Create(&domain.WorkspaceProjectModel{
		ID: "project-a", Name: "Test", Category: "drama", Status: "active", RelativeDir: "project-a",
	}).Error; err != nil {
		t.Fatalf("creating project: %v", err)
	}
	canon := fakeCompilerCanon{records: map[string]servicecanon.AssetRecord{
		"char-core": {
			ID: "char-core", ResourceType: servicecanon.ResourceTypeCharacter, SourceDocumentID: "characters", ResourceID: "char",
			Name: "角色", PromptText: "固定脸", References: []servicecanon.ReferenceRecord{{AssetID: "identity", Role: "identity", Priority: 1000, Locked: true}},
		},
	}}
	repo := repository.NewShotManifestRepositoryFromDB(db)
	service := NewService(repo, canon, nil)
	continuity := &fakeContinuityAssets{
		assetID: "wrong-section-image", found: true,
		shotAssetID: "exact-beat-image", shotFound: true,
	}
	service.SetContinuityAssetProvider(continuity)
	previous, err := service.Upsert(UpsertInput{
		ProjectID: "project-a", DocumentID: "storyboard", SectionID: "group-1", ShotKey: "beat-001", Sequence: 1,
		Bindings: Bindings{Characters: []CharacterBinding{{CanonID: "char-core"}}}, StateChangesJSON: `{}`,
	})
	if err != nil {
		t.Fatalf("Upsert(previous) error = %v", err)
	}
	current, err := service.Upsert(UpsertInput{
		ProjectID: "project-a", DocumentID: "storyboard", SectionID: "group-1", ShotKey: "beat-002", Sequence: 2,
		Bindings: Bindings{Characters: []CharacterBinding{{CanonID: "char-core"}}}, StateChangesJSON: `{}`,
	})
	if err != nil {
		t.Fatalf("Upsert(current) error = %v", err)
	}
	compiled, err := service.Compile(current)
	if err != nil {
		t.Fatalf("Compile(current) error = %v", err)
	}
	if !containsString(compiled.ReferenceAssetIDs, "exact-beat-image") {
		t.Fatalf("ReferenceAssetIDs = %#v, want exact previous Production Shot", compiled.ReferenceAssetIDs)
	}
	if containsString(compiled.ReferenceAssetIDs, "wrong-section-image") {
		t.Fatalf("ReferenceAssetIDs = %#v, section-level image must not leak into Production Shot continuity", compiled.ReferenceAssetIDs)
	}
	if continuity.shotProject != "project-a" || continuity.shotManifest != previous.ID {
		t.Fatalf("shot continuity lookup = %q/%q, want project-a/%s", continuity.shotProject, continuity.shotManifest, previous.ID)
	}
	if continuity.doc != "" || continuity.section != "" {
		t.Fatalf("section fallback was unexpectedly used: %q/%q", continuity.doc, continuity.section)
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func TestCompileRejectsReferenceAssetOutsideCanon(t *testing.T) {
	canon := fakeCompilerCanon{records: map[string]servicecanon.AssetRecord{
		"char-core": {
			ID:               "char-core",
			ResourceType:     servicecanon.ResourceTypeCharacter,
			ResourceID:       "section-char",
			SourceDocumentID: "characters",
			Name:             "角色",
			PromptText:       "固定角色",
			References:       []servicecanon.ReferenceRecord{{AssetID: "asset-allowed", Role: "identity"}},
		},
	}}
	service := &Service{canon: canon}
	compiled, err := service.Compile(Record{
		ProjectID: "project-a",
		Bindings: Bindings{Characters: []CharacterBinding{{
			CanonID:           "char-core",
			ReferenceAssetIDs: []string{"asset-not-in-canon"},
		}}},
	})
	if err != nil && !errors.Is(err, repository.ErrRecordNotFound) {
		t.Fatalf("Compile() unexpected error = %v", err)
	}
	if len(compiled.ReferenceAssetIDs) != 0 || len(compiled.References) != 0 {
		t.Fatalf("compiler accepted reference outside Canon: %+v", compiled)
	}
}
