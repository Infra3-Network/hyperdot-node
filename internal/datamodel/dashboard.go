package datamodel

import "time"

type DashboardPanelModel struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	UserID      uint      `json:"user_id" gorm:"index:idx_user_query_user_id"`
	DashboardID uint      `json:"dashboard_id" gorm:"index:idx_dashboard_panel_dashboard_id"`
	QueryID     uint      `json:"query_id"`
	ChartID     uint      `json:"chart_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_At"`
	UpdatedAt   time.Time `json:"updated_At"`
}

func (DashboardPanelModel) TableName() string {
	return "hyperdot_dashboard_panels"
}

type DashboardModel struct {
	ID          uint                  `json:"id" gorm:"primarykey"`
	UserID      uint                  `json:"user_id" gorm:"index:idx_user_query_user_id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	IsPrivacy   bool                  `json:"is_privacy"`
	Stars       uint                  `json:"stars"`
	Panels      []DashboardPanelModel `json:"panels" gorm:"-"`
}

func (DashboardModel) TableName() string {
	return "hyperdot_dashboards"
}
