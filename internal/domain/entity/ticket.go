package entity

import "time"

type Ticket struct {
	ID           string
	UserID       string
	ChannelID    string
	TicketNumber int
	Message      string
	Status       string
	CreatedAt    time.Time
}

func NewTicket(id, userID string, ticketNumber int) *Ticket {
	return &Ticket{
		ID:           id,
		UserID:       userID,
		TicketNumber: ticketNumber,
		Status:       "OPEN",
		CreatedAt:    time.Now(),
	}
}
