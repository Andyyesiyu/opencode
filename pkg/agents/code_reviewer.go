package agents

import (
	"context"

	"github.com/cloudwego/eino/adk"
	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"

	"github.com/opencodehq/opencode/pkg/models"
	"github.com/opencodehq/opencode/pkg/tools"
)

func CreateCodeReviewerAgent(ctx context.Context, modelMgr *models.Manager, toolReg *tools.Registry) (adk.Agent, error) {
	model := modelMgr.MustGet(ctx, "claude-3-5-sonnet")
	if model == nil {
		model = models.NewStaticModel("claude-3-5-sonnet")
	}
	var toolList []einotool.BaseTool
	if toolReg != nil {
		toolList = toolReg.List()
	}
	cfg := &adk.ChatModelAgentConfig{
		Name:        "code-reviewer",
		Description: "Review generated code and provide feedback",
		Instruction: "You are a meticulous code reviewer",
		Model:       model,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{Tools: toolList},
		},
		MaxIterations: 10,
	}
	return adk.NewChatModelAgent(ctx, cfg)
}
