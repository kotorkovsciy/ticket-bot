package persistence

import (
	"fmt"
	"log/slog"
	"sync"
	"ticket-bot/internal/domain/entity"
	"ticket-bot/internal/usecase/ticket"
)

type InMemoryTicketRepository struct {
	tickets map[string]*entity.Ticket
	mu      sync.RWMutex
	logger  *slog.Logger
}

func NewInMemoryTicketRepository(logger *slog.Logger) ticket.TicketRepository {
	return &InMemoryTicketRepository{
		tickets: make(map[string]*entity.Ticket),
		logger:  logger,
	}
}

func (r *InMemoryTicketRepository) FindAllOpen() ([]*entity.Ticket, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var openTickets []*entity.Ticket
	for _, t := range r.tickets {
		if t.Status == "OPEN" {
			openTickets = append(openTickets, t)
		}
	}

	if len(openTickets) == 0 {
		r.logger.Debug("no open tickets found")
	}
	return openTickets, nil
}

func (r *InMemoryTicketRepository) Save(ticket *entity.Ticket) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if ticket == nil {
		r.logger.Error("attempt to save nil ticket")
		return fmt.Errorf("cannot save nil ticket")
	}

	r.tickets[ticket.ID] = ticket
	r.logger.Info("ticket saved", "ticket_id", ticket.ID, "user_id", ticket.UserID)
	return nil
}
