package repository

import (
	"context"
	"gorm.io/gorm"
	"zyra-api/internal/models"
)

type ClientRepository interface {
	Get(context.Context, *models.Client, string, bool) error
	Find(context.Context, *[]models.Client, Query) error
	Count(context.Context, Query) (int64, error)
	Create(context.Context, *models.Client) error
	Save(context.Context, *models.Client) error
}
type clientRepository struct{ entity[models.Client] }

func NewClientRepository(db *gorm.DB) ClientRepository {
	return &clientRepository{entity[models.Client]{db: db}}
}
func (s *store) Clients() ClientRepository { return NewClientRepository(s.db) }
