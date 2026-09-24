package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"zyra-api/internal/models"
)

func Connect(url string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(url), &gorm.Config{TranslateError: true})
}

// Migrate is explicitly enabled by the operator for this development implementation.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.User{}, &models.Client{}, &models.Database{}, &models.DatabaseSettings{}, &models.Session{}, &models.ResetToken{}, &models.EmailSource{}, &models.Assessment{}, &models.Ticket{}, &models.TicketEvent{})
}
