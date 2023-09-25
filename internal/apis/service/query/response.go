package query

import (
	"infra-3.xyz/hyperdot-node/internal/apis/base"
	"infra-3.xyz/hyperdot-node/internal/dataengine"
	"infra-3.xyz/hyperdot-node/internal/datamodel"
)

type RunResponseData struct {
	Rows    []map[string]interface{}  `json:"rows"`
	Schemas []*dataengine.FieldSchema `json:"schemas"`
}

type RunResponse struct {
	Data RunResponseData `json:"data"`
	base.BaseResponse
}

type Response struct {
	Data datamodel.UserQueryModel `json:"data"`
	base.BaseResponse
}

type ListResponseData struct {
	datamodel.UserQueryModel
	Username string `json:"username"`
	Uid      string `json:"uid"`
	Email    string `json:"email"`
}

type ListResponse struct {
	Data []ListResponseData `json:"data"`
	base.BaseResponse
}
