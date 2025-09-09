package telegram

import (
	"context"
	"fmt"
	"time"

	gobotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/mackb/releaseradar/pkg/logger"
	"github.com/mackb/releaseradar/pkg/retry"
)

type telegramClient struct {
	bot *gobotapi.BotAPI
}

func NewTelegramClient(botToken string) (Client, error) {
	bot, err := gobotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot api: %w", err)
	}

	bot.Debug = false // Set to true for debugging telegram API calls

	return &telegramClient{bot: bot}, nil
}

func (t *telegramClient) SendMessage(ctx context.Context, chatID string, text string) error {
	msg := gobotapi.NewMessageToChannel(chatID, text)
	msg.ParseMode = gobotapi.ModeHTML

	var err error
	err = retry.Do(3, 2*time.Second, func() error {
		_, sendErr := t.bot.Send(msg)
		if sendErr != nil {
			logger.L().Sugar().Errorf("failed to send telegram message to chat %s: %v", chatID, sendErr)
			return fmt.Errorf("telegram client error: %w", sendErr)
		}
		return nil
	})
	return err
}
