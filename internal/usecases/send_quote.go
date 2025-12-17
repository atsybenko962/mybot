package usecases

import (
	"context"
	"fmt"
	"mybot/internal/interfaces"
)

// SendQuoteService предоставляет методы для отправки цитат в Telegram-канал.
type SendQuoteService struct {
	telegram interfaces.TelegramSender // Интерфейс для отправки сообщений в Telegram
}

// NewSendQuoteService создаёт новый экземпляр SendQuoteService.
// Принимает интерфейс TelegramSender для отправки сообщений в Telegram.
func NewSendQuoteService(telegram interfaces.TelegramSender) *SendQuoteService {
	return &SendQuoteService{telegram: telegram}
}

func (s *SendQuoteService) SendQuote(ctx context.Context) error {
	message := "Вы все петушары!"

	err := s.telegram.SendMessage(ctx, message)
	if err != nil {
		// Если произошла ошибка при отправке, возвращаем её с описанием
		return fmt.Errorf("не удалось отправить сообщение: %w", err)
	}

	return nil
}
