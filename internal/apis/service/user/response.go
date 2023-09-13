package user

import (
	"infra-3.xyz/hyperdot-node/internal/apis/base"
)

type CreateAccountResponse struct {
	base.BaseResponse
}

type LoginResponseData struct {
	Algorithm string `json:"algorithm"`
	Secret    string `json:"secret"`
	Token     string `json:"token"`
}

type LoginResponse struct {
	Data LoginResponseData `json:"data"`
	base.BaseResponse
}
