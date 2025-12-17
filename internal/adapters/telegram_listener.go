package adapters

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"time"
)

type TelegramListener struct {
	bot      *tgbotapi.BotAPI
	registry *SessionRegistry
}

func NewTelegramListener(botToken string, registry *SessionRegistry) (*TelegramListener, error) {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, err
	}
	return &TelegramListener{bot: bot, registry: registry}, nil
}

// Start начинает прослушивание событий (блокирующий вызов)
func (l *TelegramListener) Start(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	u.AllowedUpdates = []string{"my_chat_member"}

	updates := l.bot.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Telegram listener остановлен")
			return ctx.Err()
		case update, ok := <-updates:
			if !ok {
				return nil // канал закрыт
			}
			if update.MyChatMember != nil {
				l.handleMyChatMember(update.MyChatMember)
			}
		}
	}
}

func (l *TelegramListener) handleMyChatMember(update *tgbotapi.ChatMemberUpdated) {
	chat := update.Chat
	oldStatus := update.OldChatMember.Status
	newStatus := update.NewChatMember.Status

	log.Printf("🔄 Chat %d (%s): %s → %s", chat.ID, chat.Title, oldStatus, newStatus)

	key := "tg:" + fmt.Sprintf("%d", chat.ID)

	switch {
	// ✅ Бот добавлен в чат (или стал админом)
	case (oldStatus == "left" || oldStatus == "kicked") &&
		(newStatus == "member" || newStatus == "administrator"):
		session := &ChatSession{
			ChatID:   chat.ID,
			Title:    chat.Title,
			Type:     chat.Type,
			JoinedAt: time.Now(),
		}
		l.registry.Upsert(key, session)

		// Отправляем приветствие
		msg := tgbotapi.NewMessage(chat.ID, "✅ Бот подключён. ID сессии: <code>"+key+"</code>")
		msg.ParseMode = "HTML"
		msg.DisableWebPagePreview = true
		_, _ = l.bot.Send(msg)

	// ❌ Бот удалён из чата
	case newStatus == "left" || newStatus == "kicked":
		l.registry.Remove(key)
	}
}
