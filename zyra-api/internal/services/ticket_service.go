package services

import (
	"context"
	"strings"
	"zyra-api/internal/apperrors"
	"zyra-api/internal/auth"
	"zyra-api/internal/models"
	"zyra-api/internal/repository"
)

type TicketDetail struct {
	Ticket       models.Ticket                 `json:"ticket"`
	Client       models.Client                 `json:"client"`
	Database     models.Database               `json:"database"`
	Similar      []repository.SimilarTicketRow `json:"similar"`
	Participants []models.User                 `json:"participants"`
}

func (s *Service) TicketDetail(ctx context.Context, id string) (TicketDetail, error) {
	d := TicketDetail{Similar: []repository.SimilarTicketRow{}, Participants: []models.User{}}
	if err := s.Store.Tickets().Get(ctx, &d.Ticket, id, false); err != nil {
		return d, err
	}
	if err := s.Store.Clients().Get(ctx, &d.Client, d.Ticket.ClientID, false); err != nil {
		return d, err
	}
	if err := s.Store.Databases().Get(ctx, &d.Database, d.Ticket.DatabaseID, false); err != nil {
		return d, err
	}
	var err error
	d.Similar, err = s.Store.Tickets().FindSimilar(ctx, d.Ticket, 5)
	if err != nil {
		return d, err
	}
	participantIDs, err := s.Store.TicketEvents().ParticipantUserIDs(ctx, id)
	if err != nil {
		return d, err
	}
	d.Participants, err = s.Store.Users().FindByIDs(ctx, participantIDs)
	if err != nil {
		return d, err
	}
	return d, nil
}

type TicketEventPage struct {
	Items []models.TicketEvent `json:"items"`
	Total int64                `json:"total"`
	Page  int                  `json:"page"`
	Limit int                  `json:"limit"`
}

func (s *Service) TicketEvents(ctx context.Context, id string, page int) (TicketEventPage, error) {
	result := TicketEventPage{Items: []models.TicketEvent{}, Page: page, Limit: 50}
	var ticket models.Ticket
	if err := s.Store.Tickets().Get(ctx, &ticket, id, false); err != nil {
		return result, err
	}
	q := repository.Query{Where: map[string]any{"ticket_id": ticket.ID}, Order: "created_at asc, id asc", Limit: result.Limit, Offset: (page - 1) * result.Limit}
	var err error
	result.Total, err = s.Store.TicketEvents().Count(ctx, q)
	if err != nil {
		return result, err
	}
	err = s.Store.TicketEvents().Find(ctx, &result.Items, q)
	return result, err
}

func (s *Service) TicketAction(ctx context.Context, u models.User, id, action, comment string) (models.Ticket, error) {
	var t models.Ticket
	// Initial policy follows the product permission matrix: authenticated users may comment/close/reopen.
	if !u.Role.Valid() || u.Disabled {
		return t, apperrors.ErrForbidden
	}
	comment = strings.TrimSpace(comment)
	if len(comment) > 20000 {
		return t, invalid("comment exceeds 20000 bytes")
	}
	if (action == "comment" || action == "comment_and_close") && comment == "" {
		return t, invalid("comment required")
	}
	err := s.Store.Transaction(ctx, func(tx repository.Store) error {
		if err := tx.Tickets().Get(ctx, &t, id, true); err != nil {
			return err
		}
		switch action {
		case "close", "comment_and_close":
			if t.Status == "closed" {
				return apperrors.ErrConflict
			}
			now := s.Now().UTC()
			t.Status = "closed"
			t.ClosedAt = &now
			t.ClosedBy = u.ID
		case "reopen":
			if t.Status != "closed" {
				return apperrors.ErrConflict
			}
			t.Status = "open"
			t.ClosedAt = nil
			t.ClosedBy = ""
		case "comment":
		default:
			return invalid("invalid ticket action")
		}
		t.UpdatedAt = s.Now().UTC()
		if err := tx.Tickets().Save(ctx, &t); err != nil {
			return err
		}
		return tx.TicketEvents().Create(ctx, &models.TicketEvent{ID: auth.Random(), TicketID: id, UserID: u.ID, Type: action, Comment: comment, CreatedAt: t.UpdatedAt})
	})
	return t, err
}
