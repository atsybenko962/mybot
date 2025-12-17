package config

import (
	"errors"
	"log/slog"
	"os"
)

type Config struct {
	BotToken string // Токен Telegram-бота
}

// LoadConfig загружает конфигурацию из переменных окружения.
func LoadConfig(logger *slog.Logger) (*Config, error) {
	// Чтение переменных окружения
	botToken := os.Getenv("BOT_TOKEN")

	// Проверка наличия обязательных переменных
	if botToken == "" {
		logger.Error("Необходимые переменные окружения отсутствуют",
			"BOT_TOKEN", maskToken(botToken))
		return nil, errors.New("необходимые переменные окружения отсутствуют")
	}

	//TODO: Сделать валидацию токена и чатid

	return &Config{
		BotToken: botToken,
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
