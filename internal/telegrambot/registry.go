package telegrambot

import (
	"context"
	"uladzk/duw_kolejka_checker/internal/logger"
	"uladzk/duw_kolejka_checker/internal/telegrambot/handlers"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type ReplyHandlerRegistry struct {
	patternToHandler map[string]handlers.ReplyHandler
}

func NewReplyHandlerRegistry() *ReplyHandlerRegistry {
	return &ReplyHandlerRegistry{
		patternToHandler: make(map[string]handlers.ReplyHandler),
	}
}

func (r *ReplyHandlerRegistry) RegisterReplyHandler(handler handlers.ReplyHandler) {
	for _, pattern := range handler.GetReplyPatterns() {
		r.patternToHandler[pattern] = handler
	}
}

func (r *ReplyHandlerRegistry) FindHandler(replyText string) handlers.ReplyHandler {
	return r.patternToHandler[replyText]
}

type HandlerRegistry struct {
	bot           *bot.Bot
	logger        *logger.Logger
	replyRegistry *ReplyHandlerRegistry
}

func NewHandlerRegistry(b *bot.Bot, log *logger.Logger) *HandlerRegistry {
	return &HandlerRegistry{
		bot:           b,
		logger:        log,
		replyRegistry: NewReplyHandlerRegistry(),
	}
}

func (hr *HandlerRegistry) GetReplyRegistry() *ReplyHandlerRegistry {
	return hr.replyRegistry
}

func (hr *HandlerRegistry) GetDefaultHandler() func(context.Context, *bot.Bot, *models.Update) {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		handlers.DefaultHandler(ctx, b, update, hr.logger, hr.replyRegistry)
	}
}

func (hr *HandlerRegistry) UpdateBot(newBot *bot.Bot) {
	hr.bot = newBot
}

func (hr *HandlerRegistry) RegisterAllHandlers() {
	handlers.RegisterFeedbackHandler(hr.bot, hr.logger, hr.replyRegistry)
	handlers.RegisterDefaultHandler(hr.bot, hr.logger, hr.replyRegistry)
}
