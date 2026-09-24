package repository

import (
	"context"
	"gorm.io/gorm"
	"zyra-api/internal/models"
)

type AssessmentRepository interface {
	Get(context.Context, *models.Assessment, string, bool) error
	Find(context.Context, *[]models.Assessment, Query) error
	Count(context.Context, Query) (int64, error)
	Create(context.Context, *models.Assessment) error
	Save(context.Context, *models.Assessment) error
}
type assessmentRepository struct{ entity[models.Assessment] }

func NewAssessmentRepository(db *gorm.DB) AssessmentRepository {
	return &assessmentRepository{entity[models.Assessment]{db: db}}
}
func (s *store) Assessments() AssessmentRepository { return NewAssessmentRepository(s.db) }
