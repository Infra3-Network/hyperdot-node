package apis

import (
	"github.com/gin-gonic/gin"
	"infra-3.xyz/hyperdot-node/internal/clients"
	"infra-3.xyz/hyperdot-node/internal/common"
)

type QueryService struct {
	group          string
	bigqueryClient *clients.SimpleBigQueryClinet
}

func NewQueryService(cfg *common.Config) (*QueryService, error) {
	bigqueryClient, err := clients.NewSimpleBigQueryClient(cfg)
	if err != nil {
		return nil, err
	}
	return &QueryService{
		group:          "/query",
		bigqueryClient: bigqueryClient,
	}, nil
}

func (q *QueryService) ListEnginesHandle() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.JSON(200, ListEngineResponse{
			BaseResponse: ResponseOk(),
			Engines: []EngineModel{
				{
					Name:   "bigquery",
					Chains: common.GlobalParaChainCache.GetChains(),
				},
			},
		})
	}
}

func (q *QueryService) RunHandle() gin.HandlerFunc {
	return func(ctx *gin.Context) {}
}

func (q *QueryService) Name() string {
	return "query"
}

func (q *QueryService) RouteTables() []RouteTable {
	return []RouteTable{
		{
			Method:  "GET",
			Group:   q.group,
			Path:    q.group + "/engines",
			Handler: q.ListEnginesHandle(),
		},
	}
}
