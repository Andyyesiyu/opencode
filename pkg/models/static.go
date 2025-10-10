package models

import (
	"context"
	"fmt"
	"strings"

	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type StaticModel struct {
	name string
}

var _ einomodel.ToolCallingChatModel = (*StaticModel)(nil)

func NewStaticModel(name string) *StaticModel {
	return &StaticModel{name: name}
}

func (m *StaticModel) Generate(_ context.Context, messages []*schema.Message, _ ...einomodel.Option) (*schema.Message, error) {
	lastContent := ""
	if len(messages) > 0 {
		lastMessage := messages[len(messages)-1]
		if lastMessage != nil {
			lastContent = strings.TrimSpace(lastMessage.Content)
		}
	}
	response := fmt.Sprintf("[%s] echo: %s", m.name, lastContent)
	return schema.AssistantMessage(response, nil), nil
}

func (m *StaticModel) Stream(ctx context.Context, messages []*schema.Message, _ ...einomodel.Option) (*schema.StreamReader[*schema.Message], error) {
	msg, err := m.Generate(ctx, messages)
	if err != nil {
		return nil, err
	}
	return schema.StreamReaderFromArray([]*schema.Message{msg}), nil
}

func (m *StaticModel) WithTools(_ []*schema.ToolInfo) (einomodel.ToolCallingChatModel, error) {
	return m, nil
}
