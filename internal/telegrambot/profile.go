package telegrambot

import (
	"context"

	"github.com/go-telegram/bot"
)

type Profile struct {
	bot      *bot.Bot
	registry *HandlerRegistry
}

func NewProfile(b *bot.Bot, registry *HandlerRegistry) *Profile {
	return &Profile{bot: b, registry: registry}
}

func (p *Profile) SetProfile(ctx context.Context) error {
	p.bot.SetMyName(ctx, &bot.SetMyNameParams{
		Name: "DUW Kolejka Checker Bot",
	})

	p.bot.SetMyDescription(ctx, &bot.SetMyDescriptionParams{
		Description: "DUW Kolejka Checker Bot - Get queue status and send feedback",
	})

	p.bot.SetMyShortDescription(ctx, &bot.SetMyShortDescriptionParams{
		ShortDescription: "DUW Kolejka Checker Bot",
	})

	p.bot.SetMyCommands(ctx, &bot.SetMyCommandsParams{
		Commands: p.registry.GetAvailableCommands(),
	})

	return nil
}
