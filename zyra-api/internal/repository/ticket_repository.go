package repository

import (
	"context"
	"gorm.io/gorm"
	"time"
	"zyra-api/internal/models"
)

type TicketListQuery struct {
	ClosureActor  string
	Search        string
	Filters       map[string]any
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	Sort          string
	Order         string
	Limit         int
	Offset        int
}

type TicketSummaryRow struct {
	ID             string     `json:"id"`
	Number         int64      `json:"number"`
	CheckType      string     `json:"check_type"`
	Title          string     `json:"title"`
	Status         string     `json:"status"`
	AssessmentType string     `json:"assessment_type"`
	ClientID       string     `json:"client_id"`
	ClientName     string     `json:"client_name"`
	DatabaseID     string     `json:"database_id"`
	DatabaseName   string     `json:"database_name"`
	CreatedAt      time.Time  `json:"created_at"`
	ClosedAt       *time.Time `json:"closed_at,omitempty"`
}

type SimilarTicketRow struct {
	ID        string    `json:"id"`
	Number    int64     `json:"number"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type TicketRepository interface {
	Get(context.Context, *models.Ticket, string, bool) error
	Find(context.Context, *[]models.Ticket, Query) error
	Count(context.Context, Query) (int64, error)
	Create(context.Context, *models.Ticket) error
	Save(context.Context, *models.Ticket) error
	ListSummaries(context.Context, TicketListQuery) ([]TicketSummaryRow, int64, error)
	FindSimilar(context.Context, models.Ticket, int) ([]SimilarTicketRow, error)
}
type ticketRepository struct{ entity[models.Ticket] }

func NewTicketRepository(db *gorm.DB) TicketRepository {
	return &ticketRepository{entity[models.Ticket]{db: db}}
}
func (s *store) Tickets() TicketRepository { return NewTicketRepository(s.db) }

var ticketSortColumns = map[string]string{
	"number":          "tickets.number",
	"check_type":      "tickets.check_type",
	"created_at":      "tickets.created_at",
	"client":          "clients.name",
	"client_name":     "clients.name",
	"database":        "databases.name",
	"database_name":   "databases.name",
	"assessment_type": "tickets.assessment_type",
	"status":          "tickets.status",
}

func (r *ticketRepository) summaryQuery(ctx context.Context, q TicketListQuery) *gorm.DB {
	db := r.db.WithContext(ctx).Table("tickets").
		Joins("JOIN clients ON clients.id = tickets.client_id").
		Joins("JOIN databases ON databases.id = tickets.database_id")
	if q.ClosureActor != "" {
		db = db.Where("EXISTS (SELECT 1 FROM ticket_events WHERE ticket_events.ticket_id = tickets.id AND ticket_events.user_id = ? AND ticket_events.type IN ?)", q.ClosureActor, []string{"close", "comment_and_close"})
	}
	for column, value := range q.Filters {
		db = db.Where("tickets."+column+" = ?", value)
	}
	if q.CreatedAfter != nil {
		db = db.Where("tickets.created_at >= ?", *q.CreatedAfter)
	}
	if q.CreatedBefore != nil {
		db = db.Where("tickets.created_at < ?", *q.CreatedBefore)
	}
	if q.Search != "" {
		like := "%" + q.Search + "%"
		db = db.Where("tickets.title ILIKE ? OR clients.name ILIKE ? OR databases.name ILIKE ?", like, like, like)
	}
	return db
}

func (r *ticketRepository) ListSummaries(ctx context.Context, q TicketListQuery) ([]TicketSummaryRow, int64, error) {
	base := r.summaryQuery(ctx, q)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, translate(err)
	}
	sortColumn, ok := ticketSortColumns[q.Sort]
	if !ok {
		sortColumn = "tickets.created_at"
	}
	order := "DESC"
	if q.Order == "asc" {
		order = "ASC"
	}
	rows := []TicketSummaryRow{}
	err := base.Select(`tickets.id, tickets.number, tickets.check_type, tickets.title,
		tickets.status, tickets.assessment_type, tickets.client_id, clients.name AS client_name,
		tickets.database_id, databases.name AS database_name, tickets.created_at, tickets.closed_at`).
		Order(sortColumn + " " + order + ", tickets.id " + order).
		Limit(q.Limit).Offset(q.Offset).Scan(&rows).Error
	return rows, total, translate(err)
}

func (r *ticketRepository) FindSimilar(ctx context.Context, ticket models.Ticket, limit int) ([]SimilarTicketRow, error) {
	rows := []SimilarTicketRow{}
	err := r.db.WithContext(ctx).Model(&models.Ticket{}).
		Select("id, number, title, status, created_at").
		Where("client_id = ? AND database_id = ? AND check_type = ? AND id <> ?", ticket.ClientID, ticket.DatabaseID, ticket.CheckType, ticket.ID).
		Order("created_at DESC, id DESC").Limit(limit).Scan(&rows).Error
	return rows, translate(err)
}
