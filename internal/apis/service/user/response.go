package user

import (
	"time"

	"infra-3.xyz/hyperdot-node/internal/apis/base"
	"infra-3.xyz/hyperdot-node/internal/datamodel"
)

// ResponseGetUserData is response of GET /user
type ResponseGetUserData struct {
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

// ResponseGetUser is response of GET /user or GET /user/:id
type ResponseGetUser struct {
	base.BaseResponse
	Data ResponseGetUserData `json:"data"`
}

// ResponseGetUsers is response of PUT /user
type ResponseUpdateUser struct {
	Data datamodel.UserModel `json:"data"`
	base.BaseResponse
}

// ResponseCreateAccount is response of POST /user/auth/createAccount
type ResponseCreateAccount struct {
}

// ResponseLogin is response of POST /user/auth/login
type ResponseLogin struct {
	Algorithm string `json:"algorithm"`
	Token     string `json:"token"`
}

// ResponseUploadAvatarData is data of response of POST /user/avatar/upload
type ResponseUploadAvatarData struct {
	Key     string `json:"key"`
	Filsize int64  `json:"filesize"`
}

// ResponseUploadAvatar is response of POST /user/avatar/upload
type ResponseUploadAvatar struct {
	base.BaseResponse
	Data ResponseUploadAvatarData `json:"data"`
}
