package datamodel

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
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
	ConfirmedAt       *time.Time `json:"confirmed_at"`
}

// UserModel auth identity session model
type UserModel struct {
	ID uint `gorm:"primarykey" json:"id"`
	UserBasic
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
	gorm.Model
	UserID      int `gorm:"uniqueIndex:idx_user_query_user_id"`
	Name        string
	Query       string
	QueryEngine string
	Charts      JSON `gorm:"type:json"`
}

func (UserQueryModel) TableName() string {
	return "hyperdot_user_query"
}
