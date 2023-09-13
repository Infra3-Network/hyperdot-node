package base

import (
	"cloud.google.com/go/bigquery"
	"fmt"
	"github.com/gin-gonic/gin"
	"infra-3.xyz/hyperdot-node/internal/datamodel"
)

const (
	_                = iota
	Ok               // Response code for success
	Err              // Response code for error
	InvalidArguments // Response code for invalid arguments
)

type BaseResponse struct {
	Success      bool   `json:"success"`
	ErrorMessage string `json:"errorMessage"`
	ErrorCode    int    `json:"errorCode"`
}

func ResponseOk() BaseResponse {
	return BaseResponse{
		Success: true,
	}
}

func ResponseErr(ctx *gin.Context, code int, format string, args ...any) {
	ctx.JSON(code, BaseResponse{
		Success:      false,
		ErrorCode:    Err,
		ErrorMessage: fmt.Sprintf(format, args...),
	})
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
}

type QueryRunResponseData struct {
	Schemas []datamodel.TableSchema     `json:"schemas"`
	Rows    []map[string]bigquery.Value `json:"rows"`
}

type QueryRunResponse struct {
	BaseResponse
	Data QueryRunResponseData `json:"data"`
}
