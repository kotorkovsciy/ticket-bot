package ticket

import (
	"github.com/google/uuid"
	"log/slog"
	"ticket-bot/internal/domain/entity"
	"ticket-bot/internal/domain/repository"
)

type Service struct {
	repo   repository.TicketRepository
	logger *slog.Logger
}

func NewService(repo repository.TicketRepository, logger *slog.Logger) *Service {
	return &Service{repo: repo, logger: logger}
}

func (s *Service) CreateTicket(userID string) (*entity.Ticket, error) {
	allTickets, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	ticketNumber := len(allTickets) + 1
	ticket := entity.NewTicket(generateID(), userID, ticketNumber)
	if err := s.repo.Save(ticket); err != nil {
		return nil, err
	}
	s.logger.Info("ticket created", "ticket_id", ticket.ID, "user_id", userID)
	return ticket, nil
}

func (s *Service) GetOpenTickets() ([]*entity.Ticket, error) {
	allTickets, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	var openTickets []*entity.Ticket
	for _, t := range allTickets {
		if t.Status == "OPEN" {
			openTickets = append(openTickets, t)
		}
	}
	return openTickets, nil
}

func (s *Service) UpdateTicketChannelID(ticketID, channelID string) error {
	ticket, err := s.repo.Get(ticketID)
	if err != nil {
		return err
	}
	ticket.ChannelID = channelID
	return s.repo.Save(ticket)
}

func generateID() string {
	return "TICKET-" + uuid.New().String()
}
