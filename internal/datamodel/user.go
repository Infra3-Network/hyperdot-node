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
	EncryptedPassword string     `json:"encrypted_password"`
	Username          string     `json:"username"`
	Email             string     `json:"email"`
	IconUrl           string     `json:"icon_url"`
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

// UserQueryModel user query model
type UserQueryModel struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	UserID      uint      `json:"user_id" gorm:"index:idx_user_query_user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Query       string    `json:"query"`
	QueryEngine string    `json:"query_engine"`
	IsPrivacy   bool      `json:"is_privacy"`
	Unsaved     bool      `json:"unsaved"`
	Stars       uint      `json:"stars"`
	Charts      JSON      `json:"charts" gorm:"type:json"`
	CreatedAt   time.Time `json:"created_At"`
	UpdatedAt   time.Time `json:"updated_At"`
}

func (UserQueryModel) TableName() string {
	return "hyperdot_user_query"
}
