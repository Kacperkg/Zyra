package repository

import (
	"context"
	"gorm.io/gorm"
	"zyra-api/internal/models"
)

type EmailSourceRepository interface {
	Get(context.Context, *models.EmailSource, string, bool) error
	Find(context.Context, *[]models.EmailSource, Query) error
	Count(context.Context, Query) (int64, error)
	Create(context.Context, *models.EmailSource) error
	Save(context.Context, *models.EmailSource) error
}
type emailsourceRepository struct{ entity[models.EmailSource] }

func NewEmailSourceRepository(db *gorm.DB) EmailSourceRepository {
	return &emailsourceRepository{entity[models.EmailSource]{db: db}}
}
func (s *store) EmailSources() EmailSourceRepository { return NewEmailSourceRepository(s.db) }
