package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"log/slog"
	"mybot/internal/adapters"
	"mybot/internal/config"
	"mybot/internal/usecases"
	"os"
	"os/signal"
	"syscall"
)

// setupLogger логгер
func setupLogger() *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	return logger
}

func main() {
	// Настройка логгера
	logger := setupLogger()

	// Загрузка .env файла
	if err := godotenv.Load(); err != nil {
		logger.Warn("Файл .env не найден или не загружен")
	}

	// Создаем контекст с обработкой сигналов для graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Загрузка конфигурации
	cfg, err := config.LoadConfig(logger)
	if err != nil {
		logger.Error("Ошибка загрузки конфигурации", "error", err)
		os.Exit(1)
	}
	// Инициализация адаптора
	telegramAdapter, err := adapters.NewTelegramAdapter(cfg.BotToken, cfg.ChatID)
	if err != nil {
		logger.Error("Не удалось инициализировать TelegramAdapter", "error", err)
		os.Exit(1)
	}

	// Инициализация сервисов
	sendQuoteService := usecases.NewSendQuoteService(telegramAdapter)

	err = sendQuoteService.SendQuote(ctx)
	if err != nil {
		fmt.Println(err)
	}
}
