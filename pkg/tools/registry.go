package tools

import (
	"sync"

	einotool "github.com/cloudwego/eino/components/tool"
)

type Factory func() einotool.BaseTool

type Registry struct {
	mu        sync.RWMutex
	factories map[string]Factory
}

func NewRegistry() *Registry {
	return &Registry{factories: make(map[string]Factory)}
}

func (r *Registry) Register(name string, factory Factory) {
	if name == "" {
		return
	}
	if factory == nil {
		return
	}
	r.mu.Lock()
	r.factories[name] = factory
	r.mu.Unlock()
}

func (r *Registry) Get(name string) einotool.BaseTool {
	if name == "" {
		return nil
	}
	r.mu.RLock()
	factory := r.factories[name]
	r.mu.RUnlock()
	if factory == nil {
		return nil
	}
	return factory()
}

func (r *Registry) List() []einotool.BaseTool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.factories) == 0 {
		return nil
	}
	tools := make([]einotool.BaseTool, 0, len(r.factories))
	for _, factory := range r.factories {
		tool := factory()
		if tool == nil {
			continue
		}
		tools = append(tools, tool)
	}
	return tools
}

func RegisterDefaults(registry *Registry) {
	if registry == nil {
		return
	}
	registry.Register("bash", func() einotool.BaseTool { return Bash() })
	registry.Register("file-write", func() einotool.BaseTool { return FileWriter() })
}
