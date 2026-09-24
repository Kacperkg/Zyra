package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	Address, DatabaseURL, JWTSecret, BootstrapEmail, BootstrapPassword string
	Migrate                                                            bool
}

func Load() (Config, error) {
	c := Config{Address: os.Getenv("ADDRESS"), DatabaseURL: os.Getenv("DATABASE_URL"), JWTSecret: os.Getenv("JWT_SECRET"), BootstrapEmail: os.Getenv("BOOTSTRAP_ADMIN_EMAIL"), BootstrapPassword: os.Getenv("BOOTSTRAP_ADMIN_PASSWORD"), Migrate: os.Getenv("AUTO_MIGRATE") == "true"}
	if c.Address == "" {
		c.Address = "127.0.0.1:8080"
	}
	if c.DatabaseURL == "" {
		return c, errors.New("DATABASE_URL is required")
	}
	if len(strings.TrimSpace(c.JWTSecret)) < 32 {
		return c, errors.New("JWT_SECRET must contain at least 32 characters")
	}
	if (c.BootstrapEmail == "") != (c.BootstrapPassword == "") {
		return c, errors.New("both bootstrap admin variables are required together")
	}
	return c, nil
}
