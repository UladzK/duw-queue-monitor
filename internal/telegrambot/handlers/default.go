package handlers

import (
	"context"
	"uladzk/duw_kolejka_checker/internal/logger"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type DefaultHandler struct {
	replyRegistry ReplyRegistry
	log           *logger.Logger
}

func NewDefaultHandler(log *logger.Logger, replyRegistry ReplyRegistry) *DefaultHandler {
	return &DefaultHandler{
		log:           log,
		replyRegistry: replyRegistry,
	}
}

func (d *DefaultHandler) Register(b *bot.Bot, replyRegistry ReplyRegistry) {
}

func (d *DefaultHandler) HandleUpdate(ctx context.Context, b *bot.Bot, update *models.Update) {

	if update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.Text != "" {
		handlerInterface := d.replyRegistry.FindHandler(update.Message.ReplyToMessage.Text)
		if handlerInterface != nil {
			handlerInterface.HandleReply(ctx, b, update)
			return
		} else {
			d.log.Warn("No handler found for reply: " + update.Message.ReplyToMessage.Text)
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
		d.log.Error("Failed to send menu message: ", err)
	} else {
		d.log.Info("Menu message sent to user: " + msg.Text)
	}
}
