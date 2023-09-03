package apis

import (
	"cloud.google.com/go/bigquery"
	"infra-3.xyz/hyperdot-node/internal/datamodel"
)

const (
	_                = iota
	Ok               // Response code for success
	Err              // Response code for error
	InvalidArguments // Response code for invalid arguments
)

type BaseResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func ResponseOk() BaseResponse {
	return BaseResponse{
		Code:    Ok,
		Message: "ok",
	}
}

func ResponseOkWithMsg(msg string) BaseResponse {
	return BaseResponse{
		Code:    Ok,
		Message: msg,
	}
}

type ListEngineResponse struct {
	BaseResponse
	Data []datamodel.QueryEngine `json:"data"`
}

type GetQueryEngineDatasetResponse struct {
	BaseResponse
	Data struct {
		Id          string                                   `json:"id"`
		Chains      map[int]datamodel.Chain                  `json:"chains"`
		RelayChains map[string]*datamodel.RelayChainMetadata `json:"relayChains"`
		ChainTables map[int][]string                         `json:"chainTables"`
	} `json:"data"`
	//Data *datamodel.QueryEngineDatasetInfo
}

type QueryRunResponse struct {
	BaseResponse
	Rows []map[string]bigquery.Value `json:"rows"`
}
