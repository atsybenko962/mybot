package adapters

import (
	"context"
	"errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

var ErrChatNotFound = errors.New("chat session not found or removed")

// TelegramAdapter реализует интерфейс TelegramSender для отправки сообщений в Telegram.
type TelegramAdapter struct {
	bot      *tgbotapi.BotAPI
	registry *SessionRegistry
}

func NewTelegramAdapter(botToken string, registry *SessionRegistry) (*TelegramAdapter, error) {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, err
	}
	return &TelegramAdapter{bot: bot, registry: registry}, nil
}

// SendTo отправляет сообщение в чат по уникальному ключу сессии (например, "tg:-1001234567890")
func (t *TelegramAdapter) SendTo(ctx context.Context, sessionKey, message string) error {
	session, ok := t.registry.GetActive(sessionKey)
	if !ok {
		return ErrChatNotFound
	}

	msg := tgbotapi.NewMessage(session.ChatID, message)
	_, err := t.bot.Send(msg)
	return err
}

// SendToAll — для broadcast (например, alert во все группы)
func (t *TelegramAdapter) SendToAll(ctx context.Context, message string) error {
	sessions := t.registry.GetAllActive()
	for key, session := range sessions {
		if err := t.SendTo(ctx, key, message); err != nil {
			log.Printf("⚠️ Не удалось отправить в %s (%d): %v", session.Title, session.ChatID, err)
			continue
		}
	}
	return nil
}
