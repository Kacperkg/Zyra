package repository

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"
	"zyra-api/internal/database"
	"zyra-api/internal/models"
)

func TestPostgresSimilarTickets(t *testing.T) {
	dsn := os.Getenv("ZYRA_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set ZYRA_TEST_DATABASE_URL to a disposable PostgreSQL connection (keyword DSN)")
	}
	db, err := database.Connect(dsn)
	if err != nil {
		t.Fatal(err)
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	// Keep the table and fixture isolated in a rollback-only transaction.
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	defer tx.Rollback()
	if err := tx.Exec("CREATE TEMP TABLE tickets (id text PRIMARY KEY, number bigint, title text, status text, created_at timestamptz, client_id text, database_id text, check_type text)").Error; err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)
	current := models.Ticket{ID: "current", ClientID: "client", DatabaseID: "db", CheckType: "tablespace"}
	insert := func(id, status string, age int, client, databaseID, check string) {
		t.Helper()
		if err := tx.Exec("INSERT INTO tickets VALUES (?, ?, ?, ?, ?, ?, ?, ?)", id, age+100, "Issue "+id, status, now.Add(-time.Duration(age)*time.Hour), client, databaseID, check).Error; err != nil {
			t.Fatal(err)
		}
	}
	insert("current", "open", -1, "client", "db", "tablespace")
	for i := 1; i <= 6; i++ {
		insert(fmt.Sprintf("recent-%d", i), "closed", i, "client", "db", "tablespace")
	}
	insert("open-old", "open", 10, "client", "db", "tablespace")
	insert("open-older", "open", 11, "client", "db", "tablespace")
	insert("other-client", "open", 0, "other", "db", "tablespace")
	insert("other-db", "open", 0, "client", "other", "tablespace")
	insert("other-check", "open", 0, "client", "db", "backups")
	repo := NewTicketRepository(tx)
	assertIDs := func(ticket models.Ticket, limit int, want ...string) {
		t.Helper()
		rows, err := repo.FindSimilar(context.Background(), ticket, limit)
		if err != nil {
			t.Fatal(err)
		}
		ids := []string{}
		for _, row := range rows {
			ids = append(ids, row.ID)
			if row.Title == "" || row.CreatedAt.IsZero() || row.Number == 0 || row.Status == "" {
				t.Fatalf("missing link metadata: %+v", row)
			}
		}
		if want == nil {
			want = []string{}
		}
		if !reflect.DeepEqual(ids, want) {
			t.Fatalf("got %v; want %v", ids, want)
		}
	}
	assertIDs(current, 5, "open-old", "recent-1", "recent-2", "recent-3", "recent-4")
	// An older viewed ticket can link to newer matches; only one open match is pinned.
	older := current
	older.ID = "open-older"
	assertIDs(older, 5, "current", "recent-1", "recent-2", "recent-3", "recent-4")
	if err := tx.Exec("UPDATE tickets SET status = 'open' WHERE id = 'recent-3'").Error; err != nil {
		t.Fatal(err)
	}
	assertIDs(current, 5, "recent-3", "recent-1", "recent-2", "recent-4", "recent-5")
	assertIDs(current, 1, "recent-3")
	assertIDs(current, 0)
	if err := tx.Exec("UPDATE tickets SET status = 'closed'").Error; err != nil {
		t.Fatal(err)
	}
	assertIDs(current, 5, "recent-1", "recent-2", "recent-3", "recent-4", "recent-5")
	insert("tie-a", "open", 0, "client", "db", "tablespace")
	insert("tie-z", "open", 0, "client", "db", "tablespace")
	assertIDs(current, 5, "tie-z", "tie-a", "recent-1", "recent-2", "recent-3")
	if err := tx.Exec("UPDATE tickets SET status = 'closed' WHERE id IN ('tie-a', 'tie-z')").Error; err != nil {
		t.Fatal(err)
	}
	assertIDs(current, 5, "tie-z", "tie-a", "recent-1", "recent-2", "recent-3")
	empty := current
	empty.CheckType = "filesystem"
	assertIDs(empty, 5)
	empty.CheckType = "backups"
	assertIDs(empty, 5, "other-check")
}
