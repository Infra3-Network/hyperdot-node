package user

import (
	"infra-3.xyz/hyperdot-node/internal/apis/base"
	"infra-3.xyz/hyperdot-node/internal/datamodel"
)

type GetUserResponse struct {
	datamodel.UserModel `json:"data"`
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

type CreateQueryResponse struct {
	Data datamodel.UserQueryModel `json:"data"`
	base.BaseResponse
}
