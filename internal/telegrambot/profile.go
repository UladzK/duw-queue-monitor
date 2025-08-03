package telegrambot

import (
	"context"
	"fmt"
	"uladzk/duw_kolejka_checker/internal/logger"

	"github.com/go-telegram/bot"
)

const (
	botName             = "DUW Kolejka Info Bot"
	botDescription      = "DUW Kolejka Info Bot - Sprawdź status kolejki i wyślij opinię"
	botShortDescription = "DUW Kolejka Info Bot"
)

type Profile struct {
	bot      *bot.Bot
	registry *HandlerRegistry
	logger   *logger.Logger
}

func NewProfile(b *bot.Bot, registry *HandlerRegistry, logger *logger.Logger) *Profile {
	return &Profile{bot: b, registry: registry, logger: logger}
}

func (p *Profile) SetProfile(ctx context.Context) error {
	if _, err := p.bot.SetMyName(ctx, &bot.SetMyNameParams{
		Name: botName,
	}); err != nil {
		p.logger.Error("Failed to set bot name: ", err)
		return fmt.Errorf("failed to set bot name: %w", err)
	}

	if _, err := p.bot.SetMyDescription(ctx, &bot.SetMyDescriptionParams{
		Description: botDescription,
	}); err != nil {
		p.logger.Error("Failed to set bot description: ", err)
		return fmt.Errorf("failed to set bot description: %w", err)
	}

	if _, err := p.bot.SetMyShortDescription(ctx, &bot.SetMyShortDescriptionParams{
		ShortDescription: botShortDescription,
	}); err != nil {
		p.logger.Error("Failed to set bot short description: ", err)
		return fmt.Errorf("failed to set bot short description: %w", err)
	}

	if _, err := p.bot.SetMyCommands(ctx, &bot.SetMyCommandsParams{
		Commands: p.registry.GetAvailableCommands(),
	}); err != nil {
		p.logger.Error("Failed to set bot commands: ", err)
		return fmt.Errorf("failed to set bot commands: %w", err)
	}

	return nil
}
