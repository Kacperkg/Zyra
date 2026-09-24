package repository

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// entity supplies private shared persistence mechanics; callers use domain interfaces.
type entity[T any] struct{ db *gorm.DB }

func (s *entity[T]) Get(ctx context.Context, out *T, id string, lock bool) error {
	q := s.db.WithContext(ctx)
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	return translate(q.First(out, "id = ?", id).Error)
}
func (s *entity[T]) Find(ctx context.Context, out *[]T, q Query) error {
	return translate(query(s.db.WithContext(ctx), q).Find(out).Error)
}
func (s *entity[T]) Count(ctx context.Context, q Query) (int64, error) {
	var n int64
	q.Limit = 0
	q.Offset = 0
	q.Order = ""
	err := query(s.db.WithContext(ctx).Model(new(T)), q).Count(&n).Error
	return n, translate(err)
}
func (s *entity[T]) Create(ctx context.Context, m *T) error {
	return translate(s.db.WithContext(ctx).Create(m).Error)
}
func (s *entity[T]) Save(ctx context.Context, m *T) error {
	return translate(s.db.WithContext(ctx).Save(m).Error)
}
