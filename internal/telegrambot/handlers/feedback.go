package handlers

import (
	"context"
	"fmt"
	"uladzk/duw_kolejka_checker/internal/logger"
	"uladzk/duw_kolejka_checker/internal/notifications"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	thankYouText          = "Dziękujemy za Twoją opinię! Twoja wiadomość została wysłana do nas."
	feedbackInfoText      = "Możesz wysłać swoją opinię na temat działania bota. Twoja wiadomość będzie anonimowa i nie będzie publikowana."
	feedbackReplyText     = "Aby wysłać opinię, proszę odpowiedz na tę wiadomość swoją opinią:"
	feedbackAdminTemplate = "💬 <b>Nowa opinia od użytkownika</b>\n\n📝 Treść:\n%s"
)

type FeedbackHandler struct {
	log              *logger.Logger
	telegramNotifier *notifications.TelegramNotifier
	adminChatID      string
}

func NewFeedbackHandler(log *logger.Logger, telegramNotifier *notifications.TelegramNotifier, adminChatID string) *FeedbackHandler {
	return &FeedbackHandler{
		log:              log,
		telegramNotifier: telegramNotifier,
		adminChatID:      adminChatID,
	}
}

func (f *FeedbackHandler) GetReplyPatterns() []string {
	return []string{feedbackReplyText}
}

func (f *FeedbackHandler) HandleReply(ctx context.Context, b *bot.Bot, update *models.Update) {
	feedbackText := update.Message.Text

	if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      thankYouText,
		ParseMode: models.ParseModeHTML,
	}); err != nil {
		f.log.Error("Failed to send thank you message for feedback: ", err)
	}

	adminMessage := fmt.Sprintf(feedbackAdminTemplate, feedbackText)
	if err := f.telegramNotifier.SendMessage(f.adminChatID, adminMessage); err != nil {
		f.log.Error("Failed to forward feedback to admin: ", err)
	} else {
		f.log.Info("Feedback forwarded to feedback chat successfully")
	}
}

func (f *FeedbackHandler) Register(b *bot.Bot, replyRegistry ReplyRegistry) {
	b.RegisterHandler(bot.HandlerTypeMessageText, "feedback", bot.MatchTypeCommand, func(ctx context.Context, b *bot.Bot, update *models.Update) {
		f.HandleUpdate(ctx, b, update)
	})

	replyRegistry.RegisterReplyHandler(f)
}

func (f *FeedbackHandler) HandleUpdate(ctx context.Context, b *bot.Bot, update *models.Update) {
	if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      feedbackInfoText,
		ParseMode: models.ParseModeHTML,
	}); err != nil {
		f.log.Error("Failed to send feedback info message: ", err)
		return
	}

	if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   feedbackReplyText,
		ReplyMarkup: &models.ForceReply{
			ForceReply:            true,
			InputFieldPlaceholder: "Napisz swoją opinię tutaj...",
			Selective:             true,
		},
	}); err != nil {
		f.log.Error("Failed to send feedback reply prompt: ", err)
	}
}
