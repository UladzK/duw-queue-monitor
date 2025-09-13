package handlers

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type ReplyHandler interface {
	GetReplyPatterns() []string
	HandleReply(ctx context.Context, b *bot.Bot, update *models.Update)
}

type ReplyRegistry interface {
	RegisterReplyHandler(handler ReplyHandler)
	FindHandler(replyText string) ReplyHandler
}
