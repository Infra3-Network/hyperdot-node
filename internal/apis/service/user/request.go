package user

import "infra-3.xyz/hyperdot-node/internal/datamodel"

type CreateAccountRequest struct {
	Provider string `json:"provider"`
	UserId   string `json:"userId"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Provider string `json:"provider"`
	UserId   string `json:"userId"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type GetQueryRequest struct {
	ID string `json:"id"`
}

type CreateQueryRequest struct {
	Data datamodel.UserQueryModel
}

type UpdateQueryRequest struct {
	Data datamodel.UserQueryModel
}
