package telegrambot

import (
	"context"
	"uladzk/duw_kolejka_checker/internal/logger"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type ReplyHandler interface {
	GetReplyPatterns() []string
	HandleReply(ctx context.Context, b *bot.Bot, update *models.Update, log *logger.Logger)
}