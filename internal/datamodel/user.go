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
	Provider                         string         `json:"provider,omitempty"`
	UserID                           string         `json:"userid,omitempty"`
	LastLoginAt                      *time.Time     `json:"last_login,omitempty"`
	LastActiveAt                     *time.Time     `json:"last_active,omitempty"`
	LongestDistractionSinceLastLogin *time.Duration `json:"distraction_time,omitempty"`
	jwt.RegisteredClaims
}

// UserSignLog sign log
type UserSignLog struct {
	UserAgent string
	At        *time.Time
	IP        string
}

// SignLogs record sign in logs
type UserSignLogs struct {
	Log         string `sql:"-"`
	SignInCount uint
	Logs        []UserSignLog
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
	Provider          string // phone, email, wechat, github...
	UID               string `gorm:"column:uid"`
	EncryptedPassword string
	UserID            string
	Email             string
	ConfirmedAt       *time.Time
}

// UserModel auth identity session model
type UserModel struct {
	gorm.Model
	UserBasic
	UserSignLogs
}

// ToClaims convert to auth Claims
func (basic UserBasic) ToClaims() *UserClaims {
	claims := &UserClaims{}
	claims.Provider = basic.Provider
	claims.UserID = basic.UserID
	return claims
}

func (UserModel) TableName() string {
	return "hyperdot_user"
}
