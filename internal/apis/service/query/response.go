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
	Data datamodel.QueryModel `json:"data"`
	base.BaseResponse
}

type ListResponse struct {
	Data []map[string]interface{} `json:"data"`
	base.BaseResponse
}
