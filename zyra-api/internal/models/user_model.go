package models

import (
	"time"
)

type Role string

const (
	RoleAdmin   Role = "admin"
	RoleTrusted Role = "trusted"
	RoleNormal  Role = "normal"
)

func (r Role) Valid() bool        { return r == RoleAdmin || r == RoleTrusted || r == RoleNormal }
func (r Role) CanConfigure() bool { return r == RoleAdmin || r == RoleTrusted }

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null"`
	Name         string    `json:"name"`
	Role         Role      `json:"role"`
	Theme        string    `json:"theme"`
	Disabled     bool      `json:"disabled"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}
