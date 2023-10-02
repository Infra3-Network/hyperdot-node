package user

import (
	"time"

	"infra-3.xyz/hyperdot-node/internal/apis/base"
	"infra-3.xyz/hyperdot-node/internal/datamodel"
)

type GetUserResponseData struct {
	ID                string     `json:"id"`
	UID               string     `json:"uid"`
	Username          string     `json:"username"`
	EncryptedPassword string     `json:"encrypted_password"`
	Email             string     `json:"email"`
	Bio               string     `json:"bio"`
	IconUrl           string     `json:"icon_url"`
	Twitter           string     `json:"twitter"`
	Github            string     `json:"github"`
	Telgram           string     `json:"telgram"`
	Discord           string     `json:"discord"`
	Location          string     `json:"location"`
	ConfirmedAt       *time.Time `json:"confirmed_at"`
	CreatedAt         *time.Time `json:"created_at"`
	UpdatedAt         *time.Time `json:"updated_at"`
	Stars             uint       `json:"stars"`
	Queries           uint       `json:"queries"`
	Dashboards        uint       `json:"dashboards"`
}

type GetUserResponse struct {
	Data GetUserResponseData `json:"data"`
	base.BaseResponse
}

type UpdateUserResponse struct {
	Data datamodel.UserModel `json:"data"`
	base.BaseResponse
}

type CreateAccountResponse struct {
	base.BaseResponse
}

type LoginResponseData struct {
	Algorithm string `json:"algorithm"`
	Token     string `json:"token"`
}

type LoginResponse struct {
	Data LoginResponseData `json:"data"`
	base.BaseResponse
}

type QueryResponse struct {
	Data datamodel.UserQueryModel `json:"data"`
	base.BaseResponse
}
