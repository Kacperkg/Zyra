package services

import (
	"context"
	"strings"
	"zyra-api/internal/apperrors"
	"zyra-api/internal/auth"
	"zyra-api/internal/models"
	"zyra-api/internal/repository"
)

type UserInput struct {
	Email    *string      `json:"email"`
	Name     *string      `json:"name"`
	Password *string      `json:"password"`
	Role     *models.Role `json:"role"`
	Disabled *bool        `json:"disabled"`
	Theme    *string      `json:"theme"`
}

func (s *Service) SaveUser(ctx context.Context, actor models.User, id string, in UserInput) (models.User, error) {
	var u models.User
	self := id != "" && actor.ID == id
	if !self {
		if err := requireAdmin(actor); err != nil {
			return u, err
		}
	}
	if actor.Role != models.RoleAdmin && (in.Role != nil || in.Disabled != nil) {
		return u, apperrors.ErrForbidden
	}
	if id != "" {
		if err := s.Store.Users().Get(ctx, &u, id, false); err != nil {
			return u, err
		}
	} else {
		u.ID = auth.Random()
		u.CreatedAt = s.Now().UTC()
		u.Role = models.RoleNormal
		u.Theme = "light"
	}
	if in.Email != nil {
		e, err := emailAddress(*in.Email)
		if err != nil {
			return u, err
		}
		u.Email = e
	}
	if in.Name != nil {
		u.Name = strings.TrimSpace(*in.Name)
	}
	if in.Role != nil {
		if !in.Role.Valid() {
			return u, invalid("invalid role")
		}
		u.Role = *in.Role
	}
	if in.Disabled != nil {
		u.Disabled = *in.Disabled
	}
	if self && (u.Disabled || u.Role != actor.Role) {
		return u, invalid("cannot disable or change your own role")
	}
	if in.Theme != nil {
		if *in.Theme != "light" && *in.Theme != "dark" {
			return u, invalid("invalid theme")
		}
		u.Theme = *in.Theme
	}
	if in.Password != nil {
		if id != "" {
			return u, invalid("use password recovery to change an existing password")
		}
		hash, err := passwordHash(*in.Password)
		if err != nil {
			return u, err
		}
		u.PasswordHash = hash
	}
	if u.Email == "" || u.Name == "" || u.PasswordHash == "" {
		return u, invalid("email, name and password are required")
	}
	if id == "" {
		return u, s.Store.Users().Create(ctx, &u)
	}
	err := s.Store.Transaction(ctx, func(tx repository.Store) error {
		if err := tx.Users().Save(ctx, &u); err != nil {
			return err
		}
		if u.Disabled {
			return revokeSessions(ctx, tx, u.ID)
		}
		return nil
	})
	return u, err
}
