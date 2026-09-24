package models

import (
	"time"
)

type Ticket struct {
	ID             string     `json:"id"`
	Number         int64      `json:"number" gorm:"uniqueIndex;autoIncrement"`
	ClientID       string     `json:"client_id" gorm:"index:idx_ticket_similar,priority:1"`
	DatabaseID     string     `json:"database_id" gorm:"index:idx_ticket_similar,priority:2;index:idx_ticket_database_created,priority:1"`
	AssessmentID   string     `json:"assessment_id,omitempty" gorm:"index"`
	AssessmentType string     `json:"assessment_type" gorm:"index:idx_ticket_assessment_created,priority:1"`
	CheckType      string     `json:"check_type" gorm:"index:idx_ticket_similar,priority:3;index:idx_ticket_check_created,priority:1"`
	Resource       string     `json:"resource,omitempty"`
	Title          string     `json:"title"`
	Status         string     `json:"status" gorm:"index:idx_ticket_status_created,priority:1"`
	Hostname       string     `json:"hostname"`
	IP             string     `json:"ip"`
	Evidence       string     `json:"evidence"`
	CreatedAt      time.Time  `json:"created_at" gorm:"index;index:idx_ticket_status_created,priority:2;index:idx_ticket_similar,priority:4;index:idx_ticket_database_created,priority:2;index:idx_ticket_check_created,priority:2;index:idx_ticket_assessment_created,priority:2"`
	UpdatedAt      time.Time  `json:"updated_at"`
	ClosedAt       *time.Time `json:"closed_at,omitempty"`
	ClosedBy       string     `json:"closed_by,omitempty" gorm:"index"`
}
