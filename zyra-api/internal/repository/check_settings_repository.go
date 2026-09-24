package repository

import (
	"context"
	"gorm.io/gorm"
	"zyra-api/internal/models"
)

type DatabaseSettingsRepository interface {
	Get(context.Context, *models.DatabaseSettings, string, bool) error
	Find(context.Context, *[]models.DatabaseSettings, Query) error
	Count(context.Context, Query) (int64, error)
	Create(context.Context, *models.DatabaseSettings) error
	Save(context.Context, *models.DatabaseSettings) error
}
type checksettingsRepository struct {
	entity[models.DatabaseSettings]
}

func NewDatabaseSettingsRepository(db *gorm.DB) DatabaseSettingsRepository {
	return &checksettingsRepository{entity[models.DatabaseSettings]{db: db}}
}
func (s *store) Settings() DatabaseSettingsRepository { return NewDatabaseSettingsRepository(s.db) }
