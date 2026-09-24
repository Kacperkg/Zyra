package services

import (
	"context"
	"zyra-api/internal/models"
)

func (s *Service) Client(ctx context.Context, id string) (models.Client, error) {
	var client models.Client
	err := s.Store.Clients().Get(ctx, &client, id, false)
	return client, err
}

func (s *Service) Database(ctx context.Context, id string) (models.Database, error) {
	var database models.Database
	err := s.Store.Databases().Get(ctx, &database, id, false)
	return database, err
}

func (s *Service) Assessment(ctx context.Context, id string) (models.Assessment, error) {
	var assessment models.Assessment
	err := s.Store.Assessments().Get(ctx, &assessment, id, false)
	return assessment, err
}
