package telegram

import (
	"context"
)

type Client interface {
	SendMessage(ctx context.Context, chatID string, text string) error
}
