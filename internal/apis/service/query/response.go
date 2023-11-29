package query

import (
	"infra-3.xyz/hyperdot-node/internal/apis/base"
	"infra-3.xyz/hyperdot-node/internal/dataengine"
	"infra-3.xyz/hyperdot-node/internal/datamodel"
)

// RequestCreateQuery is response of POST /query/run
type ResponseRunData struct {
	Rows    []map[string]interface{}  `json:"rows"`
	Schemas []*dataengine.FieldSchema `json:"schemas"`
}

// ResponseCreateQuery is response of POST /query/run
type ResponseRun struct {
	Data ResponseRunData `json:"data"`
	base.BaseResponse
}

// ResponseCreateQuery is response of GET /query
type Response struct {
	base.BaseResponse
	Data datamodel.QueryModel `json:"data"`
}
