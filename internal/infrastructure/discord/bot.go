package discord

import (
	"github.com/bwmarrin/discordgo"
	"log/slog"
)

type Bot struct {
	Session *discordgo.Session
	logger  *slog.Logger
}

func NewBot(token string, logger *slog.Logger) (*Bot, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		logger.Error("failed to create discord session", "error", err)
		return nil, err
	}

	logger.Info("discord bot initialized successfully")
	return &Bot{
		Session: session,
		logger:  logger,
	}, nil
}
