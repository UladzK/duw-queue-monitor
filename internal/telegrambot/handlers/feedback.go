package handlers

import (
	"context"
	"uladzk/duw_kolejka_checker/internal/logger"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	thankYouText      = "Dziękujemy za Twoją opinię! Twoja wiadomość została wysłana do nas."
	feedbackInfoText  = "Możesz wysłać swoją opinię na temat działania bota. Twoja wiadomość będzie anonimowa i nie będzie publikowana."
	feedbackReplyText = "Aby wysłać opinię, proszę odpowiedz na tę wiadomość swoją opinią:"
)

type FeedbackHandler struct {
	log *logger.Logger
}

func NewFeedbackHandler(log *logger.Logger) *FeedbackHandler {
	return &FeedbackHandler{
		log: log,
	}
}

func (f *FeedbackHandler) GetReplyPatterns() []string {
	return []string{feedbackReplyText}
}

func (f *FeedbackHandler) HandleReply(ctx context.Context, b *bot.Bot, update *models.Update) {
	if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      thankYouText,
		ParseMode: models.ParseModeHTML,
	}); err != nil {
		f.log.Error("Failed to send thank you message for feedback: ", err)
	}
}

func (f *FeedbackHandler) Register(b *bot.Bot, replyRegistry ReplyRegistry) {
	b.RegisterHandler(bot.HandlerTypeMessageText, "feedback", bot.MatchTypeCommand, func(ctx context.Context, b *bot.Bot, update *models.Update) {
		f.HandleUpdate(ctx, b, update)
	})

	replyRegistry.RegisterReplyHandler(f)
}

func (f *FeedbackHandler) HandleUpdate(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      feedbackInfoText,
		ParseMode: models.ParseModeHTML,
	})

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   feedbackReplyText,
		ReplyMarkup: &models.ForceReply{
			ForceReply:            true,
			InputFieldPlaceholder: "Napisz swoją opinię tutaj...",
			Selective:             true,
		},
	})
}
