package repository

import (
	"context"
	"gorm.io/gorm"
	"zyra-api/internal/models"
)

type SessionRepository interface {
	Get(context.Context, *models.Session, string, bool) error
	Find(context.Context, *[]models.Session, Query) error
	Count(context.Context, Query) (int64, error)
	Create(context.Context, *models.Session) error
	Save(context.Context, *models.Session) error
}
type sessionRepository struct{ entity[models.Session] }

func NewSessionRepository(db *gorm.DB) SessionRepository {
	return &sessionRepository{entity[models.Session]{db: db}}
}
func (s *store) Sessions() SessionRepository { return NewSessionRepository(s.db) }
