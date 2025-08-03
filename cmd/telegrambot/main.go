package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"uladzk/duw_kolejka_checker/internal/logger"
	"uladzk/duw_kolejka_checker/internal/telegrambot"

	"github.com/caarlos0/env/v11"
	"github.com/go-telegram/bot"
)

var log *logger.Logger

func main() {
	ctx, cancel := signal.NotifyContext(context.Background())
	defer cancel()

	var err error
	log, err = buildLogger()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	log.Info("Building bot with handlers...")
	b, handlerRegistry, err := buildBotWithHandlers()
	if err != nil {
		panic(err)
	}

	log.Info("Configuring Telegram bot profile...")
	if err := setProfile(ctx, b, handlerRegistry); err != nil {
		panic("failed to set bot profile: " + err.Error())
	}
	log.Info("Bot profile set successfully")

	log.Info("Starting Telegram bot...")
	go b.Start(ctx)
	log.Info("Telegram bot started. Waiting for shutdown signal...")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	log.Info("Received shutdown signal, stopping Telegram bot...")
	cancel()

	log.Info("Telegram bot stopped")
}

func setProfile(ctx context.Context, b *bot.Bot, registry *telegrambot.HandlerRegistry) error {
	profile := telegrambot.NewProfile(b, registry, log)
	if err := profile.SetProfile(ctx); err != nil {
		return err
	}
	return nil
}

func buildBotWithHandlers() (*bot.Bot, *telegrambot.HandlerRegistry, error) {
	var cfg telegrambot.Config
	if err := env.Parse(&cfg); err != nil {
		return nil, nil, err
	}

	handlerRegistry := telegrambot.NewHandlerRegistry(log)

	opts := []bot.Option{
		bot.WithDefaultHandler(handlerRegistry.GetDefaultHandler()),
	}

	bot, err := bot.New(cfg.NotificationTelegram.BotToken, opts...)
	if err != nil {
		return nil, nil, err
	}

	handlerRegistry.RegisterAllHandlers(bot)

	return bot, handlerRegistry, nil
}

func buildLogger() (*logger.Logger, error) {
	var cfg logger.Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return logger.NewLogger(&cfg), nil
}
