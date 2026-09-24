package models

type EmailSource struct {
	ID         string     `json:"id"`
	DatabaseID string     `json:"database_id" gorm:"index"`
	Sender     string     `json:"sender"`
	Subject    string     `json:"subject"`
	Timezone   string     `json:"timezone"`
	Enabled    bool       `json:"enabled" gorm:"index"`
	Schedules  []Schedule `json:"schedules" gorm:"serializer:json;type:jsonb"`
}

type Schedule struct {
	Name  string `json:"name"`
	Start string `json:"start"`
	End   string `json:"end"`
}
