package adapters

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"mybot/internal/factApi"
	"time"
)

const defaultFactInterval = 30 * time.Second

type TelegramListener struct {
	bot      *tgbotapi.BotAPI
	registry *SessionRegistry
	ctx      context.Context
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
	l.ctx = ctx

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

// StartWorkersForExistingSessions запускает воркеры фактов для уже известных активных сессий.
func (l *TelegramListener) StartWorkersForExistingSessions(ctx context.Context) {
	l.ctx = ctx

	sessions := l.registry.GetAllActive()
	for key, s := range sessions {
		l.startFactsWorker(key, s.ChatID)
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

		// Запускаем воркер, который периодически шлёт интересные факты
		l.startFactsWorker(key, chat.ID)

	// ❌ Бот удалён из чата
	case newStatus == "left" || newStatus == "kicked":
		l.registry.Remove(key) // Remove также остановит воркер
	}
}

// startFactsWorker запускает воркер, который с заданной периодичностью присылает факт в чат.
func (l *TelegramListener) startFactsWorker(sessionKey string, chatID int64) {
	baseCtx := l.ctx
	if baseCtx == nil {
		baseCtx = context.Background()
	}

	ctx, cancel := context.WithCancel(baseCtx)
	l.registry.AttachWorkerCancel(sessionKey, cancel)

	client := factApi.NewFactAPI()

	go func() {
		log.Printf("▶️ Запуск воркера фактов для %s", sessionKey)

		fact, err := client.GetRandomFact(ctx)
		if err != nil {
			log.Printf("⚠️ Не удалось получить факт для %s: %v", sessionKey, err)
		}
		text := "Интересный факт:\n" + fact.Text
		msg := tgbotapi.NewMessage(chatID, text)
		if _, err := l.bot.Send(msg); err != nil {
			log.Printf("⚠️ Не удалось отправить факт в %s: %v", sessionKey, err)
		}

		ticker := time.NewTicker(defaultFactInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Printf("⏹ Остановлен воркер фактов для %s", sessionKey)
				return
			case <-ticker.C:
				fact, err := client.GetRandomFact(ctx)
				if err != nil {
					log.Printf("⚠️ Не удалось получить факт для %s: %v", sessionKey, err)
					continue
				}

				text := "Интересный факт:\n" + fact.Text
				msg := tgbotapi.NewMessage(chatID, text)
				if _, err := l.bot.Send(msg); err != nil {
					log.Printf("⚠️ Не удалось отправить факт в %s: %v", sessionKey, err)
				}
			}
		}
	}()
}
