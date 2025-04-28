package ticket

import (
	"github.com/google/uuid"
	"log/slog"
	"ticket-bot/internal/domain/entity"
)

type TicketRepository interface {
	Save(ticket *entity.Ticket) error
	FindAllOpen() ([]*entity.Ticket, error)
}

type Service struct {
	repo   TicketRepository
	logger *slog.Logger
}

func NewService(repo TicketRepository, logger *slog.Logger) *Service {
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
