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

func (s *Service) CreateTicket(userID string, ticketNumber int) (*entity.Ticket, error) {
	ticket := entity.NewTicket(generateID(), userID, ticketNumber)
	if err := s.repo.Save(ticket); err != nil {
		return nil, err
	}
	s.logger.Info("ticket created", "ticket_id", ticket.ID, "user_id", userID)
	return ticket, nil
}

func generateID() string {
	return "TICKET-" + uuid.New().String()
}
