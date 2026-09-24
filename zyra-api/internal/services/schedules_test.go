package services

import (
	"testing"
	"time"
	"zyra-api/internal/models"
)

func TestScheduleOvernightAndDST(t *testing.T) {
	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		t.Fatal(err)
	}
	day := time.Date(2026, 10, 24, 0, 0, 0, 0, loc)
	start, end, err := scheduleBounds(day, models.Schedule{Name: "night", Start: "22:00", End: "06:00"})
	if err != nil || end.Sub(start) != 9*time.Hour {
		t.Fatalf("DST bounds: %v %v %v", start, end, err)
	}
}
