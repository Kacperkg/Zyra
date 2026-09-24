package repository

import (
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
	"zyra-api/internal/apperrors"
)

type Query struct {
	ClosureActor  string
	TimeField     string
	After, Before *time.Time
	Where         map[string]any
	Search        string
	SearchColumns []string
	Order         string
	Limit, Offset int
}

// Store keeps persistence details out of services. Query fields are built by services,
// never populated directly from untrusted request parameters.

func translate(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperrors.ErrNotFound
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return apperrors.ErrConflict
	}
	return err
}

func query(db *gorm.DB, q Query) *gorm.DB {
	if q.ClosureActor != "" {
		db = db.Where("id IN (SELECT ticket_id FROM ticket_events WHERE user_id = ? AND type IN ?)", q.ClosureActor, []string{"close", "comment_and_close"})
	}
	if q.TimeField != "" {
		if q.After != nil {
			db = db.Where(clause.Gte{Column: clause.Column{Name: q.TimeField}, Value: *q.After})
		}
		if q.Before != nil {
			db = db.Where(clause.Lt{Column: clause.Column{Name: q.TimeField}, Value: *q.Before})
		}
	}
	if len(q.Where) > 0 {
		db = db.Where(q.Where)
	}
	if q.Search != "" && len(q.SearchColumns) > 0 {
		expr := []clause.Expression{}
		for _, col := range q.SearchColumns {
			expr = append(expr, clause.Expr{SQL: "? ILIKE ?", Vars: []any{clause.Column{Name: col}, "%" + q.Search + "%"}})
		}
		db = db.Where(clause.Or(expr...))
	}
	if q.Order != "" {
		db = db.Order(q.Order)
	}
	if q.Limit > 0 {
		db = db.Limit(q.Limit)
	}
	if q.Offset > 0 {
		db = db.Offset(q.Offset)
	}
	return db
}
