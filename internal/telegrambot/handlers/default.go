package handlers

import (
	"context"
	"uladzk/duw_kolejka_checker/internal/logger"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func RegisterDefaultHandler(b *bot.Bot, log *logger.Logger, replyRegistry ReplyRegistry) {
	// Default handler is set during bot creation with bot.WithDefaultHandler option
	// This function exists for consistency but actual registration happens in buildBot
}

func DefaultHandler(ctx context.Context, b *bot.Bot, update *models.Update, log *logger.Logger, replyRegistry ReplyRegistry) {
	log.Info("Processing message in default handler")

	if update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.Text != "" {
		handlerInterface := replyRegistry.FindHandler(update.Message.ReplyToMessage.Text)
		if handlerInterface != nil {
			// Import the ReplyHandler interface from telegrambot package
			if handler, ok := handlerInterface.(interface {
				HandleReply(ctx context.Context, b *bot.Bot, update *models.Update, log *logger.Logger)
			}); ok {
				handler.HandleReply(ctx, b, update, log)
				return
			}
		}
	}

	chatID := update.Message.Chat.ID
	showMenuText := "Welcome to the DUW Kolejka Checker Bot!\n\n" +
		"<b>Available commands</b>\n" +
		"/feedback - Send feedback about the bot\n" +
		"Use /start to see this menu again\n"

	if msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      showMenuText,
		ParseMode: models.ParseModeHTML,
	}); err != nil {
		log.Error("Failed to send menu message: ", err)
	} else {
		log.Info("Menu message sent to user: " + msg.Text)
	}
}
