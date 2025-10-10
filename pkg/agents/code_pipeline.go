package agents

import (
	"context"

	"github.com/cloudwego/eino/adk"

	"github.com/opencodehq/opencode/pkg/models"
	"github.com/opencodehq/opencode/pkg/tools"
)

func CreateCodePipelineAgent(ctx context.Context, modelMgr *models.Manager, toolReg *tools.Registry) (adk.Agent, error) {
	model := modelMgr.MustGet(ctx, "claude-3-5-sonnet")
	if model == nil {
		model = models.NewStaticModel("claude-3-5-sonnet")
	}
	coordinator, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:          "coordinator",
		Description:   "Coordinate code development workflow",
		Instruction:   "You coordinate the development workflow",
		Model:         model,
		MaxIterations: 30,
	})
	if err != nil {
		return nil, err
	}
	writer, err := CreateCodeWriterAgent(ctx, modelMgr, toolReg)
	if err != nil {
		return nil, err
	}
	reviewer, err := CreateCodeReviewerAgent(ctx, modelMgr, toolReg)
	if err != nil {
		return nil, err
	}
	return adk.SetSubAgents(ctx, coordinator, []adk.Agent{writer, reviewer})
}
