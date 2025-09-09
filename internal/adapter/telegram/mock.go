package telegram

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockTelegramClient struct {
	mock.Mock
}

func (m *MockTelegramClient) SendMessage(ctx context.Context, chatID string, text string) error {
	args := m.Called(ctx, chatID, text)
	return args.Error(0)
}
