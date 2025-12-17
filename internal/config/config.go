package config

import (
	"errors"
	"log/slog"
	"os"
	"strconv"
)

type Config struct {
	BotToken string // Токен Telegram-бота
	ChatID   int64  // Идентификатор Telegram-канала
}

// LoadConfig загружает конфигурацию из переменных окружения.
func LoadConfig(logger *slog.Logger) (*Config, error) {
	// Чтение переменных окружения
	botToken := os.Getenv("BOT_TOKEN")
	chatIDStr := os.Getenv("CHAT_ID")

	// Проверка наличия обязательных переменных
	if botToken == "" || chatIDStr == "" {
		logger.Error("Необходимые переменные окружения отсутствуют",
			"BOT_TOKEN", maskToken(botToken), "CHAT_ID", chatIDStr)
		return nil, errors.New("необходимые переменные окружения отсутствуют")
	}

	// Преобразование CHAT_ID в int64
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		logger.Error("Ошибка преобразования CHAT_ID в int64", "error", err)
		return nil, err
	}

	//TODO: Сделать валидацию токена и чатid

	return &Config{
		BotToken: botToken,
		ChatID:   chatID,
	}, nil
}

// maskToken маскирует токен для безопасного логирования
func maskToken(token string) string {
	if token == "" {
		return "***"
	}
	if len(token) <= 8 {
		return "***"
	}
	return token[:4] + "***" + token[len(token)-4:]
}
