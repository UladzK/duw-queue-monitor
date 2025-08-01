package handlers

import (
	"context"
	"fmt"
	"uladzk/duw_kolejka_checker/internal/logger"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type ReplyHandlerRegistry interface {
	RegisterReplyHandler(handler interface{})
}

type FeedbackHandler struct{}

func (f *FeedbackHandler) GetReplyPatterns() []string {
	return []string{"Please reply to this message with your general feedback:"}
}

func (f *FeedbackHandler) HandleReply(ctx context.Context, b *bot.Bot, update *models.Update, log *logger.Logger) {
	chatID := update.Message.Chat.ID
	feedbackText := update.Message.Text
	user := update.Message.From

	log.Info("General feedback received. No specific type to process.")
	log.Info(fmt.Sprintf(
		"Feedback (feedback_general) from @%s (userID=%d, chatID=%d): %q",
		user.Username, user.ID, chatID, feedbackText,
	))
	log.Info(update.Message.Text)

	thankYouText := "Thank you for your feedback! We appreciate your input and will consider it for future improvements."

	if msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      thankYouText,
		ParseMode: models.ParseModeHTML,
	}); err != nil {
		log.Error("Failed to send thank you message: ", err)
	} else {
		log.Info("Thank you message sent to user: " + msg.Text)
	}
}


func RegisterFeedbackHandler(b *bot.Bot, log *logger.Logger, replyRegistry ReplyHandlerRegistry) {
	b.RegisterHandler(bot.HandlerTypeMessageText, "feedback", bot.MatchTypeCommand, func(ctx context.Context, b *bot.Bot, update *models.Update) {
		feedbackCommandHandler(ctx, b, update, log)
	})

	feedbackHandler := &FeedbackHandler{}
	replyRegistry.RegisterReplyHandler(feedbackHandler)
}

func feedbackCommandHandler(ctx context.Context, b *bot.Bot, update *models.Update, log *logger.Logger) {
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