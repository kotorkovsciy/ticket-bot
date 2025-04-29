package discord

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log/slog"
	"ticket-bot/internal/domain/entity"
	"ticket-bot/internal/domain/repository"
	"ticket-bot/internal/usecase/ticket"
)

const (
	TicketChannelName   = "тикет-"
	ControlMessageTitle = "Создать новый тикет"
)

type TicketHandler struct {
	ticketService  *ticket.Service
	ticketRepo     repository.TicketRepository
	guildID        string
	controlChannel string
	modRoleID      string
	categoryID     string
	logger         *slog.Logger
}

func NewTicketHandler(
	ts *ticket.Service,
	tr repository.TicketRepository,
	guildID,
	controlChannel,
	modRoleID,
	categoryID string,
	logger *slog.Logger,
) *TicketHandler {
	return &TicketHandler{
		ticketService:  ts,
		ticketRepo:     tr,
		guildID:        guildID,
		controlChannel: controlChannel,
		modRoleID:      modRoleID,
		categoryID:     categoryID,
		logger:         logger,
	}
}

func (h *TicketHandler) InitializeTicketSystem(s *discordgo.Session) {
	messages, err := s.ChannelMessages(h.controlChannel, 100, "", "", "")
	if err != nil {
		h.logger.Error("failed to fetch channel messages", "error", err)
		return
	}

	for _, msg := range messages {
		if err := s.ChannelMessageDelete(h.controlChannel, msg.ID); err != nil {
			h.logger.Warn("failed to delete message", "message_id", msg.ID, "error", err)
		}
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Создать тикет",
					Style:    discordgo.PrimaryButton,
					CustomID: "create_ticket",
					Emoji:    &discordgo.ComponentEmoji{Name: "📨"},
				},
			},
		},
	}

	_, err = s.ChannelMessageSendComplex(h.controlChannel, &discordgo.MessageSend{
		Content:    fmt.Sprintf("**%s**\nНажмите кнопку ниже, чтобы создать новый тикет", ControlMessageTitle),
		Components: components,
	})

	if err != nil {
		h.logger.Error("failed to create control message", "error", err)
	}
}

func (h *TicketHandler) HandleInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}

	switch i.MessageComponentData().CustomID {
	case "create_ticket":
		h.handleCreateTicket(s, i)
	case "close_ticket":
		h.handleCloseTicket(s, i)
	case "delete_ticket":
		h.handleDeleteTicket(s, i)
	}
}

func (h *TicketHandler) handleCreateTicket(s *discordgo.Session, i *discordgo.InteractionCreate) {
	user := i.Member.User
	if user.Bot {
		return
	}

	newTicket, err := h.ticketService.CreateTicket(user.ID)
	if err != nil {
		h.logger.Error("failed to create ticket", "error", err)
		_ = h.sendErrorResponse(s, i, "❌ Произошла ошибка при создании тикета")
		return
	}

	channel, err := h.createPrivateChannel(s, user.ID, newTicket)
	if err != nil {
		h.logger.Error("failed to create private channel", "error", err)
		_ = h.sendErrorResponse(s, i, "❌ Не удалось создать канал для тикета")
		return
	}

	newTicket.ChannelID = channel.ID
	if err := h.ticketRepo.Save(newTicket); err != nil {
		h.logger.Error("failed to save ticket", "error", err)
	}

	if err := h.sendSuccessResponse(s, i, fmt.Sprintf("Тикет создан: <#%s>", channel.ID)); err != nil {
		h.logger.Error("failed to send interaction response", "error", err)
	}

	if _, err := s.ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
		Content: fmt.Sprintf(
			"**Тикет #%d**\nСоздатель: <@%s>\n\nОпишите вашу проблему здесь.",
			newTicket.TicketNumber,
			user.ID,
		),
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Закрыть тикет",
						Style:    discordgo.DangerButton,
						CustomID: "close_ticket",
						Emoji:    &discordgo.ComponentEmoji{Name: "🔒"},
					},
				},
			},
		},
	}); err != nil {
		h.logger.Error("failed to send welcome message", "error", err)
	}
}

func (h *TicketHandler) handleCloseTicket(s *discordgo.Session, i *discordgo.InteractionCreate) {
	channel, err := s.Channel(i.ChannelID)
	if err != nil {
		h.logger.Error("failed to get channel", "error", err)
		return
	}

	isModerator := false
	for _, roleID := range i.Member.Roles {
		if roleID == h.modRoleID {
			isModerator = true
			break
		}
	}

	overwrites := []*discordgo.PermissionOverwrite{
		{
			Type: discordgo.PermissionOverwriteTypeRole,
			ID:   h.guildID,
			Deny: discordgo.PermissionViewChannel,
		},
		{
			ID:    h.modRoleID,
			Type:  discordgo.PermissionOverwriteTypeRole,
			Allow: discordgo.PermissionViewChannel | discordgo.PermissionReadMessageHistory,
			Deny:  discordgo.PermissionSendMessages,
		},
	}

	if !isModerator {
		overwrites = append(overwrites, &discordgo.PermissionOverwrite{
			ID:   i.Member.User.ID,
			Type: discordgo.PermissionOverwriteTypeMember,
			Deny: discordgo.PermissionSendMessages | discordgo.PermissionViewChannel,
		})
	}

	_, err = s.ChannelEditComplex(channel.ID, &discordgo.ChannelEdit{
		PermissionOverwrites: overwrites,
	})

	if err != nil {
		h.logger.Error("failed to close ticket", "error", err)
		_ = h.sendErrorResponse(s, i, "❌ Не удалось закрыть тикет")
		return
	}

	if err := h.updateInteractionResponse(s, i, "🔒 Тикет закрыт."); err != nil {
		h.logger.Error("failed to update interaction response", "error", err)
	}

	if _, err := s.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{
		Content: "🔒 Тикет закрыт.",
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Удалить тикет",
						Style:    discordgo.DangerButton,
						CustomID: "delete_ticket",
						Emoji:    &discordgo.ComponentEmoji{Name: "🗑️"},
					},
				},
			},
		},
	}); err != nil {
		h.logger.Error("failed to send close message", "error", err)
	}
}

func (h *TicketHandler) createPrivateChannel(s *discordgo.Session, userID string, ticket *entity.Ticket) (*discordgo.Channel, error) {
	if h.guildID == "" || h.categoryID == "" || h.modRoleID == "" || userID == "" {
		return nil, fmt.Errorf("missing required IDs (guild: %s, category: %s, modRole: %s, user: %s)",
			h.guildID, h.categoryID, h.modRoleID, userID)
	}

	overwrites := []*discordgo.PermissionOverwrite{
		{
			Type: discordgo.PermissionOverwriteTypeRole,
			ID:   h.guildID,
			Deny: discordgo.PermissionViewChannel,
		},
		{
			Type:  discordgo.PermissionOverwriteTypeMember,
			ID:    userID,
			Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages,
		},
		{
			Type:  discordgo.PermissionOverwriteTypeRole,
			ID:    h.modRoleID,
			Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionManageChannels,
		},
	}

	channelData := discordgo.GuildChannelCreateData{
		Name:                 fmt.Sprintf("%s%d", TicketChannelName, ticket.TicketNumber),
		Type:                 discordgo.ChannelTypeGuildText,
		ParentID:             h.categoryID,
		PermissionOverwrites: overwrites,
	}

	channel, err := s.GuildChannelCreateComplex(h.guildID, channelData)
	if err != nil {
		h.logger.Error("channel creation failed", "error", err)
		return nil, fmt.Errorf("channel creation failed: %w", err)
	}

	h.logger.Info("channel created", "channel_id", channel.ID)
	return channel, nil
}

func (h *TicketHandler) handleDeleteTicket(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if _, err := s.ChannelDelete(i.ChannelID); err != nil {
		h.logger.Error("failed to delete ticket channel", "error", err)
		_ = h.sendErrorResponse(s, i, "❌ Не удалось удалить тикет")
		return
	}
	h.logger.Info("ticket channel deleted", "channel_id", i.ChannelID)
}

func (h *TicketHandler) sendErrorResponse(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (h *TicketHandler) sendSuccessResponse(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (h *TicketHandler) updateInteractionResponse(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    msg,
			Components: []discordgo.MessageComponent{},
		},
	})
}
