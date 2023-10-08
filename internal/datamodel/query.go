package datamodel

import "time"

type ChartModel struct {
	ID      uint   `json:"id" gorm:"primarykey"`
	Index   uint32 `json:"index"`
	QueryID uint   `json:"query_id" gorm:"index:idx_charts_id"`
	UserID  uint   `json:"user_id" gorm:"index:idx_charts_id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Config  JSON   `json:"config" gorm:"type:json"`
}

func (ChartModel) TableName() string {
	return "hyperdot_charts"
}

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
	CreatedAt   time.Time    `json:"created_At"`
	UpdatedAt   time.Time    `json:"updated_At"`
}

func (QueryModel) TableName() string {
	return "hyperdot_queries"
}
