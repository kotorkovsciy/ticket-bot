package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"ticket-bot/internal/infrastructure/discord"
	"ticket-bot/internal/infrastructure/persistence"
	discordHandler "ticket-bot/internal/interfaces/discord"
	"ticket-bot/internal/usecase/ticket"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	logger.Info("starting application")

	ticketRepo := persistence.NewInMemoryTicketRepository(logger)
	ticketService := ticket.NewService(ticketRepo, logger)
	guildID := os.Getenv("DISCORD_GUILD_ID")
	controlChannelID := os.Getenv("CONTROL_CHANNEL_ID")
	modRoleID := os.Getenv("MOD_ROLE_ID")
	categoryID := os.Getenv("TICKET_CATEGORY_ID")

	ticketHandler := discordHandler.NewTicketHandler(
		ticketService,
		ticketRepo,
		guildID,
		controlChannelID,
		modRoleID,
		categoryID,
		logger,
	)

	bot, err := discord.NewBot(os.Getenv("DISCORD_TOKEN"), logger)
	if err != nil {
		logger.Error("failed to create bot", "error", err)
		os.Exit(1)
	}

	bot.Session.AddHandler(ticketHandler.HandleInteractionCreate)
	ticketHandler.InitializeTicketSystem(bot.Session)

	if err := bot.Session.Open(); err != nil {
		logger.Error("failed to open session", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := bot.Session.Close(); err != nil {
			logger.Error("failed to close session", "error", err)
		}
	}()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	logger.Info("shutting down")
}
