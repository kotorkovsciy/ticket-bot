package repository

import "ticket-bot/internal/domain/entity"

type TicketRepository interface {
	Save(ticket *entity.Ticket) error
	Get(id string) (*entity.Ticket, error)
	GetAll() ([]*entity.Ticket, error)
}
