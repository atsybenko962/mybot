package main

import (
	"context"
	"github.com/joho/godotenv"
	"log"
	"log/slog"
	"mybot/internal/adapters"
	"mybot/internal/config"
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

	registry := adapters.NewSessionRegistry(1000)

	listener, err := adapters.NewTelegramListener(cfg.BotToken, registry)
	if err != nil {
		log.Fatal(err)
	}

	// Инициализация адаптора -1003439532972
	telegramAdapter, err := adapters.NewTelegramAdapter(cfg.BotToken, registry)
	if err != nil {
		logger.Error("Не удалось инициализировать TelegramAdapter", "error", err)
		os.Exit(1)
	}

	// Запускаем listener в отдельной горутине
	go func() {
		if err := listener.Start(ctx); err != nil {
			log.Printf("Listener error: %v", err)
		}
	}()

}
