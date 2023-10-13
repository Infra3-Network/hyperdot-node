package datamodel

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

// UserClaims is jwt auth claims
type UserClaims struct {
	UserID                           uint           `json:"user_id"`
	Provider                         string         `json:"provider,omitempty"`
	Username                         string         `json:"username,omitempty"`
	LastLoginAt                      *time.Time     `json:"last_login,omitempty"`
	LastActiveAt                     *time.Time     `json:"last_active,omitempty"`
	LongestDistractionSinceLastLogin *time.Duration `json:"distraction_time,omitempty"`
	jwt.RegisteredClaims
}

// UserSignLog sign log
type UserSignLog struct {
	UserAgent string     `json:"user_agent"`
	At        *time.Time `json:"at"`
	IP        string     `json:"ip"`
}

// SignLogs record sign in logs
type UserSignLogs struct {
	Log         string        `sql:"-" json:"log"`
	SignInCount uint          `json:"sign_in_count"`
	Logs        []UserSignLog `json:"logs"`
}

// Scan scan data into sign logs
func (signLogs *UserSignLogs) Scan(data interface{}) (err error) {
	switch values := data.(type) {
	case []byte:
		if string(values) != "" {
			return json.Unmarshal(values, signLogs)
		}
	case string:
		return signLogs.Scan([]byte(values))
	case []string:
		for _, str := range values {
			if err := signLogs.Scan(str); err != nil {
				return err
			}
		}
	default:
		err = errors.New("unsupported driver -> Scan pair for SignLogs")
	}
	return
}

// Value return struct's Value
func (signLogs UserSignLogs) Value() (driver.Value, error) {
	results, err := json.Marshal(signLogs)
	return string(results), err
}

// UserBasic basic information about auth identity
type UserBasic struct {
	Provider          string     `json:"provider"`
	UID               string     `gorm:"column:uid" json:"uid"`
	EncryptedPassword string     `json:"-"`
	Username          string     `json:"username"`
	Email             string     `json:"email"`
	Bio               string     `json:"bio"`
	IconUrl           string     `json:"icon_url"`
	Twitter           string     `json:"twitter"`
	Github            string     `json:"github"`
	Telgram           string     `json:"telgram"`
	Discord           string     `json:"discord"`
	Location          string     `json:"location"`
	ConfirmedAt       *time.Time `json:"confirmed_at"`
}

// UserModel auth identity session model
type UserModel struct {
	ID uint `gorm:"primarykey" json:"id"`
	UserBasic
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at"`

	UserSignLogs
}

// ToClaims convert to auth Claims
func (model UserModel) ToClaims() *UserClaims {
	claims := &UserClaims{}
	claims.Provider = model.Provider
	claims.Username = model.Username
	claims.UserID = model.ID
	return claims
}

func (UserModel) TableName() string {
	return "hyperdot_user"
}

type UserStatistics struct {
	ID         uint `json:"id" gorm:"primarykey"`
	UserId     uint `json:"user_id" gorm:"index:idx_user_statistics_user_id"`
	Stars      uint `json:"stars"`
	Queries    uint `json:"queries"`
	Dashboards uint `json:"dashboards"`
}

func (UserStatistics) TableName() string {
	return "hyperdot_user_statistics"
}

type UserDashboardFavorites struct {
	UserID          uint      `json:"user_id" gorm:"index:idx_user_dashboard_favorites_user_id"`
	DashboardID     uint      `json:"dashboard_id" gorm:"index:idx_user_dashboard_favorites_dashboard_id"`
	DashboardUserID uint      `json:"dashboard_user_id" gorm:"-"`
	Stared          bool      `json:"stared"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (UserDashboardFavorites) TableName() string {
	return "hyperdot_user_dashboard_favorites"
}

type UserQueryFavorites struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	UserID    uint      `json:"user_id" gorm:"index:idx_user_query_favorites_user_id"`
	QueryID   uint      `json:"query_id" gorm:"index:idx_user_query_favorites_query_id"`
	Stared    bool      `json:"stared"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (UserQueryFavorites) TableName() string {
	return "hyperdot_user_query_favorites"
}
