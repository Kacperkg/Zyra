package models

// ResourceRule is keyed by the exact resource name from a report.
// A nil threshold is different from a configured zero threshold.
type ResourceRule struct {
	Ignore         bool     `json:"ignore"`
	MinFreePercent *float64 `json:"min_free_percent,omitempty"`
	MaxUsedPercent *float64 `json:"max_used_percent,omitempty"`
}

type CheckSettings struct {
	Selected  map[string]bool                    `json:"selected"`
	Resources map[string]map[string]ResourceRule `json:"resources"`
}

type DatabaseSettings struct {
	ID       string        `json:"-" gorm:"primaryKey"`
	Settings CheckSettings `json:"settings" gorm:"serializer:json;type:jsonb"`
}
