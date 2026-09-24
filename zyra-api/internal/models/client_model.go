package models

import (
	"time"
)

type Client struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Notes     string    `json:"notes"`
	Archived  bool      `json:"archived"`
	CreatedAt time.Time `json:"created_at"`
}
