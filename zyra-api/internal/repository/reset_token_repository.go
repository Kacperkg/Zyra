package repository

import (
	"context"
	"gorm.io/gorm"
	"zyra-api/internal/models"
)

type ResetTokenRepository interface {
	Get(context.Context, *models.ResetToken, string, bool) error
	Find(context.Context, *[]models.ResetToken, Query) error
	Count(context.Context, Query) (int64, error)
	Create(context.Context, *models.ResetToken) error
	Save(context.Context, *models.ResetToken) error
}
type resettokenRepository struct{ entity[models.ResetToken] }

func NewResetTokenRepository(db *gorm.DB) ResetTokenRepository {
	return &resettokenRepository{entity[models.ResetToken]{db: db}}
}
func (s *store) ResetTokens() ResetTokenRepository { return NewResetTokenRepository(s.db) }
