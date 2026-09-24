package models

import (
	"time"
)

type Database struct {
	ID        string    `json:"id"`
	ClientID  string    `json:"client_id"`
	Name      string    `json:"name"`
	Hostname  string    `json:"hostname"`
	IP        string    `json:"ip"`
	Notes     string    `json:"notes"`
	Archived  bool      `json:"archived"`
	CreatedAt time.Time `json:"created_at"`
}
