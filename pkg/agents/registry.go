package agents

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/cloudwego/eino/adk"
	einomodel "github.com/cloudwego/eino/components/model"

	"github.com/opencodehq/opencode/pkg/models"
	"github.com/opencodehq/opencode/pkg/tools"
)

var ErrAgentNotFound = errors.New("agent not found")

type Factory func(context.Context, *models.Manager, *tools.Registry) (adk.Agent, error)

type registry struct {
	mu        sync.RWMutex
	factories map[string]Factory
}

var globalRegistry = &registry{factories: make(map[string]Factory)}

func Register(name string, factory Factory) {
	if name == "" {
		return
	}
	if factory == nil {
		return
	}
	globalRegistry.mu.Lock()
	globalRegistry.factories[name] = factory
	globalRegistry.mu.Unlock()
}

func Get(ctx context.Context, name string, modelMgr *models.Manager, toolReg *tools.Registry) (adk.Agent, error) {
	if name == "" {
		return nil, ErrAgentNotFound
	}
	globalRegistry.mu.RLock()
	factory := globalRegistry.factories[name]
	globalRegistry.mu.RUnlock()
	if factory == nil {
		return nil, ErrAgentNotFound
	}
	return factory(ctx, modelMgr, toolReg)
}

func List() []string {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	if len(globalRegistry.factories) == 0 {
		return nil
	}
	names := make([]string, 0, len(globalRegistry.factories))
	for name := range globalRegistry.factories {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func RegisterDefaults(modelMgr *models.Manager, toolReg *tools.Registry) {
	Register("code-writer", func(ctx context.Context, mm *models.Manager, tr *tools.Registry) (adk.Agent, error) {
		return CreateCodeWriterAgent(ctx, mm, tr)
	})
	Register("code-reviewer", func(ctx context.Context, mm *models.Manager, tr *tools.Registry) (adk.Agent, error) {
		return CreateCodeReviewerAgent(ctx, mm, tr)
	})
	Register("code-pipeline", func(ctx context.Context, mm *models.Manager, tr *tools.Registry) (adk.Agent, error) {
		return CreateCodePipelineAgent(ctx, mm, tr)
	})
	if modelMgr == nil {
		return
	}
	modelMgr.Register("claude-3-5-sonnet", func(ctx context.Context) (einomodel.ToolCallingChatModel, error) {
		return models.NewStaticModel("claude-3-5-sonnet"), nil
	})
	modelMgr.Register("gpt-4.1", func(ctx context.Context) (einomodel.ToolCallingChatModel, error) {
		return models.NewStaticModel("gpt-4.1"), nil
	})
	if toolReg == nil {
		return
	}
	_ = toolReg
}
