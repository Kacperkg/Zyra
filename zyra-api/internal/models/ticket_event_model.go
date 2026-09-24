package models

import (
	"time"
)

type TicketEvent struct {
	ID        string    `json:"id"`
	TicketID  string    `json:"ticket_id" gorm:"index:idx_event_timeline,priority:1;index:idx_event_closure,priority:3;index:idx_event_participant,priority:1"`
	UserID    string    `json:"user_id" gorm:"index:idx_event_closure,priority:1;index:idx_event_participant,priority:2"`
	Type      string    `json:"type" gorm:"index:idx_event_closure,priority:2"`
	Comment   string    `json:"comment,omitempty"`
	CreatedAt time.Time `json:"created_at" gorm:"index:idx_event_timeline,priority:2"`
}
