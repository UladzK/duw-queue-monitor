package telegrambot

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Profile struct {
	bot *bot.Bot
}

func NewProfile(b *bot.Bot) *Profile {
	return &Profile{bot: b}
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
		Commands: []models.BotCommand{
			{
				Command:     "feedback",
				Description: "Send feedback about the bot",
			},
		},
	})

	return nil
}
