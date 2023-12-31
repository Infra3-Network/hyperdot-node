package datamodel

import "time"

// ChartModel represents a chart model ant the config is a json string.
type ChartModel struct {
	ID        uint   `json:"id" gorm:"primarykey"`
	QueryID   uint   `json:"query_id" gorm:"index:idx_charts_id"`
	UserID    uint   `json:"user_id" gorm:"index:idx_charts_id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Closeable bool   `json:"closeable"`
	Config    JSON   `json:"config" gorm:"type:json"`
}

func (ChartModel) TableName() string {
	return "hyperdot_charts"
}

// DashboardModel represents a dashboard model ant the config is a json string.
type QueryModel struct {
	ID          uint         `json:"id" gorm:"primarykey"`
	UserID      uint         `json:"user_id" gorm:"index:idx_user_query_user_id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Query       string       `json:"query"`
	QueryEngine string       `json:"query_engine"`
	IsPrivacy   bool         `json:"is_privacy"`
	Unsaved     bool         `json:"unsaved"`
	Stars       uint         `json:"stars"`
	Charts      []ChartModel `json:"charts" gorm:"-"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

func (QueryModel) TableName() string {
	return "hyperdot_queries"
}
