package repository

import (
	"context"
	"gorm.io/gorm"
	"zyra-api/internal/models"
)

type DatabaseRepository interface {
	Get(context.Context, *models.Database, string, bool) error
	Find(context.Context, *[]models.Database, Query) error
	Count(context.Context, Query) (int64, error)
	Create(context.Context, *models.Database) error
	Save(context.Context, *models.Database) error
}
type databaseRepository struct{ entity[models.Database] }

func NewDatabaseRepository(db *gorm.DB) DatabaseRepository {
	return &databaseRepository{entity[models.Database]{db: db}}
}
func (s *store) Databases() DatabaseRepository { return NewDatabaseRepository(s.db) }
