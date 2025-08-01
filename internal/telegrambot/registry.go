package telegrambot

import (
	"context"
	"uladzk/duw_kolejka_checker/internal/logger"
	"uladzk/duw_kolejka_checker/internal/telegrambot/handlers"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Handler interface {
	Register(b *bot.Bot, replyRegistry handlers.ReplyRegistry)
	HandleUpdate(ctx context.Context, b *bot.Bot, update *models.Update)
}

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
	logger        *logger.Logger
	replyRegistry *ReplyHandlerRegistry
	handlersMap   map[string]Handler
}

func NewHandlerRegistry(log *logger.Logger) *HandlerRegistry {
	handlersMap := map[string]Handler{
		"feedback": handlers.NewFeedbackHandler(log),
	}

	return &HandlerRegistry{
		logger:        log,
		replyRegistry: NewReplyHandlerRegistry(),
		handlersMap:   handlersMap,
	}
}

func (hr *HandlerRegistry) GetDefaultHandler() func(context.Context, *bot.Bot, *models.Update) {
	return handlers.NewDefaultHandler(hr.logger, hr.replyRegistry).HandleUpdate
}

func (hr *HandlerRegistry) RegisterAllHandlers(b *bot.Bot) {
	for _, handler := range hr.handlersMap {
		handler.Register(b, hr.replyRegistry)
	}
}
