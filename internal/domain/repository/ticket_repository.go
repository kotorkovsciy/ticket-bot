package repository

import "ticket-bot/internal/domain/entity"

type TicketRepository interface {
	Save(ticket *entity.Ticket) error
	FindAllOpen() ([]*entity.Ticket, error)
}
