package models

import (
	"context"
	"errors"
	"sync"

	einomodel "github.com/cloudwego/eino/components/model"
)

var ErrModelNotFound = errors.New("model not found")

type Provider func(context.Context) (einomodel.ToolCallingChatModel, error)

type Manager struct {
	mu        sync.RWMutex
	providers map[string]Provider
}

func NewManager() *Manager {
	return &Manager{providers: make(map[string]Provider)}
}

func (m *Manager) Register(name string, provider Provider) {
	if name == "" {
		return
	}
	if provider == nil {
		return
	}
	m.mu.Lock()
	m.providers[name] = provider
	m.mu.Unlock()
}

func (m *Manager) Get(ctx context.Context, name string) (einomodel.ToolCallingChatModel, error) {
	if name == "" {
		return nil, ErrModelNotFound
	}
	m.mu.RLock()
	provider := m.providers[name]
	m.mu.RUnlock()
	if provider == nil {
		return nil, ErrModelNotFound
	}
	return provider(ctx)
}

func (m *Manager) MustGet(ctx context.Context, name string) einomodel.ToolCallingChatModel {
	model, err := m.Get(ctx, name)
	if err != nil {
		return nil
	}
	return model
}
