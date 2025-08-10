package handlers

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// TODO: these interfaces can be moved closer to the structs that use them
type ReplyHandler interface {
	GetReplyPatterns() []string
	HandleReply(ctx context.Context, b *bot.Bot, update *models.Update)
}

type ReplyRegistry interface {
	RegisterReplyHandler(handler ReplyHandler)
	FindHandler(replyText string) ReplyHandler
}
