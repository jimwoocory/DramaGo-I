package usvdsv11

// WorkflowDescriptor is a user-facing entry into the embedded V11 skill pack.
// It does not create a second workflow runtime; DramaGo's existing Agent/Skill
// system executes the referenced skill.
type WorkflowDescriptor struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Skill       string `json:"skill"`
	Stage       string `json:"stage"`
}

// WorkflowCatalog returns the primary USVDS V11 creative flow exposed by DramaGo.
func WorkflowCatalog() []WorkflowDescriptor {
	return []WorkflowDescriptor{
		{
			ID:          "controller",
			Label:       "V11 总控",
			Description: "识别当前材料和 Gate，选择正确的 V11 阶段。",
			Skill:       "usvd-v10-controller",
			Stage:       "controller",
		},
		{
			ID:          "intake",
			Label:       "改编 / 立项",
			Description: "从想法、现有大纲或源材料建立 Project Brief 与改编方向。",
			Skill:       "usvd-v10-00-intake-adaptation",
			Stage:       "intake",
		},
		{
			ID:          "story",
			Label:       "故事架构",
			Description: "建立完整 Story Package、人物关系、故事引擎与季弧。",
			Skill:       "usvd-v10-01-story-architect",
			Stage:       "story",
		},
		{
			ID:          "episode",
			Label:       "分集架构",
			Description: "把已审故事包展开为逐集 Episode Architecture。",
			Skill:       "usvd-v10-02-episode-architect",
			Stage:       "episode",
		},
		{
			ID:          "screenplay",
			Label:       "剧本草稿",
			Description: "从已审分集架构起草中英对照的非投产逐场剧本。",
			Skill:       "usvd-v10-03-creator-script-draft",
			Stage:       "screenplay",
		},
		{
			ID:          "review",
			Label:       "独立审核",
			Description: "独立审核故事、分集或剧本，并输出连续性与 Gate 结论。",
			Skill:       "usvd-v10-04-review-continuity",
			Stage:       "review",
		},
		{
			ID:          "continuity",
			Label:       "连续性",
			Description: "维护人物、关系、道具、伤势、信息差与承诺兑现连续性。",
			Skill:       "us-vertical-drama-continuity-editor",
			Stage:       "continuity",
		},
		{
			ID:          "storyboard",
			Label:       "分镜导演",
			Description: "把通过 Gate 的剧本转为资产锁定、ShotManifest 与视频提示词执行包。",
			Skill:       "us-vertical-drama-storyboard-director",
			Stage:       "storyboard",
		},
	}
}

// recommendedWorkflow derives the next creative action from authoritative
// DramaGo artifacts and revision-bound approvals. It is guidance only; it does
// not mutate Gate state or create a second workflow state machine.
func recommendedWorkflow(report ProjectGateReport) *WorkflowDescriptor {
	if report.Summary.StoryDocumentID == "" {
		return workflowByID("story")
	}
	if !approvalForGate(report.Approvals, GateStoryApproved).Approved {
		return workflowByID("review")
	}
	if report.Summary.EpisodeArchitectureDocumentID == "" {
		return workflowByID("episode")
	}
	if report.Summary.ScreenplayDocumentID == "" {
		return workflowByID("screenplay")
	}
	if !approvalForGate(report.Approvals, GateScreenplayReviewed).Approved {
		return workflowByID("review")
	}
	for _, gate := range report.Gates {
		if gate.Gate == GateGenerationReady && !gate.Ready {
			return workflowByID("storyboard")
		}
	}
	return nil
}

func workflowByID(id string) *WorkflowDescriptor {
	for _, workflow := range WorkflowCatalog() {
		if workflow.ID == id {
			copy := workflow
			return &copy
		}
	}
	return nil
}

func approvalForGate(approvals []ApprovalState, gate GateID) ApprovalState {
	for _, approval := range approvals {
		if approval.Gate == gate {
			return approval
		}
	}
	return ApprovalState{Gate: gate}
}
