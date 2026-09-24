package services

import (
	"context"
	"zyra-api/internal/repository"
)

type DashboardSummary struct {
	Oracle struct {
		OpenIssues int64 `json:"open_issues"`
	} `json:"oracle"`
	SQL struct {
		Available bool `json:"available"`
	} `json:"sql"`
}

func (s *Service) Dashboard(ctx context.Context) (DashboardSummary, error) {
	var summary DashboardSummary
	count, err := s.Store.Tickets().Count(ctx, repository.Query{Where: map[string]any{"status": "open"}})
	summary.Oracle.OpenIssues = count
	summary.SQL.Available = false
	return summary, err
}
