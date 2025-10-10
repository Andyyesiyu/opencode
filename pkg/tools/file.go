package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type fileWriter struct{}

var _ einotool.InvokableTool = (*fileWriter)(nil)

type fileWriteRequest struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func FileWriter() *fileWriter {
	return &fileWriter{}
}

func (t *fileWriter) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "file-write",
		Desc: "Write file contents to disk",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"path":    {Type: schema.String, Desc: "Filesystem path to write", Required: true},
			"content": {Type: schema.String, Desc: "Contents to write", Required: true},
		}),
	}, nil
}

func (t *fileWriter) InvokableRun(_ context.Context, argumentsInJSON string, _ ...einotool.Option) (string, error) {
	if strings.TrimSpace(argumentsInJSON) == "" {
		return "", errors.New("file-write tool requires JSON input")
	}
	var req fileWriteRequest
	if err := json.Unmarshal([]byte(argumentsInJSON), &req); err != nil {
		return "", err
	}
	if strings.TrimSpace(req.Path) == "" {
		return "", errors.New("file-write tool requires a path")
	}
	if err := os.WriteFile(req.Path, []byte(req.Content), 0o644); err != nil {
		return "", err
	}
	return fmt.Sprintf("wrote %d bytes to %s", len(req.Content), req.Path), nil
}
