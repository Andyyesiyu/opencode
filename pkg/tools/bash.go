package tools

import (
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"strings"

	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type bashTool struct{}

var _ einotool.InvokableTool = (*bashTool)(nil)

type bashRequest struct {
	Command string `json:"command"`
}

func Bash() *bashTool {
	return &bashTool{}
}

func (t *bashTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "bash",
		Desc: "Execute shell commands using bash -lc",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"command": {Type: schema.String, Desc: "Shell command to execute", Required: true},
		}),
	}, nil
}

func (t *bashTool) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...einotool.Option) (string, error) {
	var req bashRequest
	if strings.TrimSpace(argumentsInJSON) == "" {
		return "", errors.New("bash tool requires a command")
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &req); err != nil {
		req.Command = strings.TrimSpace(argumentsInJSON)
	}
	command := strings.TrimSpace(req.Command)
	if command == "" {
		return "", errors.New("bash tool requires a command")
	}
	cmd := exec.CommandContext(ctx, "bash", "-lc", command)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
