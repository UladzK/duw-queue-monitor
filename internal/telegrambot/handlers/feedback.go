package handlers

import (
	"context"
	"fmt"
	"uladzk/duw_kolejka_checker/internal/logger"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
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
	return []string{"Please reply to this message with your general feedback:"}
}

func (f *FeedbackHandler) HandleReply(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID
	feedbackText := update.Message.Text
	user := update.Message.From

	f.log.Info("General feedback received. No specific type to process.")
	f.log.Info(fmt.Sprintf(
		"Feedback (feedback_general) from @%s (userID=%d, chatID=%d): %q",
		user.Username, user.ID, chatID, feedbackText,
	))
	f.log.Info(update.Message.Text)

	thankYouText := "Thank you for your feedback! We appreciate your input and will consider it for future improvements."

	if msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      thankYouText,
		ParseMode: models.ParseModeHTML,
	}); err != nil {
		f.log.Error("Failed to send thank you message: ", err)
	} else {
		f.log.Info("Thank you message sent to user: " + msg.Text)
	}
}

func (f *FeedbackHandler) Register(b *bot.Bot, replyRegistry ReplyRegistry) {
	b.RegisterHandler(bot.HandlerTypeMessageText, "feedback", bot.MatchTypeCommand, func(ctx context.Context, b *bot.Bot, update *models.Update) {
		f.HandleUpdate(ctx, b, update)
	})

	replyRegistry.RegisterReplyHandler(f)
}

func (f *FeedbackHandler) HandleUpdate(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID
	promptForFeedbackText(ctx, b, chatID, update.Message.ID, "general", "💬")
}

func promptForFeedbackText(ctx context.Context, b *bot.Bot, chatID int64, messageID int, feedbackType, emoji string) {
	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    chatID,
		MessageID: messageID,
		Text:      fmt.Sprintf("%s You selected: %s feedback", emoji, feedbackType),
	})

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   fmt.Sprintf("Please reply to this message with your %s feedback:", feedbackType),
		ReplyMarkup: &models.ForceReply{
			ForceReply:            true,
			InputFieldPlaceholder: fmt.Sprintf("Type your %s feedback here...", feedbackType),
			Selective:             true,
		},
	})
}
