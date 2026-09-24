package models

import (
	"time"
)

type ResetToken struct {
	ID        string `gorm:"primaryKey"`
	UserID    string
	TokenHash string `gorm:"uniqueIndex"`
	ExpiresAt time.Time
	Used      bool
}
