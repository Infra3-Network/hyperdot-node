package user

import (
	"infra-3.xyz/hyperdot-node/internal/apis/base"
	"infra-3.xyz/hyperdot-node/internal/datamodel"
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

type CreateQueryResponse struct {
	Data datamodel.UserQueryModel `json:"data"`
	base.BaseResponse
}
