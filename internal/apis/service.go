package apis

import (
	"cloud.google.com/go/bigquery"
	"context"
	"errors"
	"google.golang.org/api/iterator"
	"infra-3.xyz/hyperdot-node/internal/cache"
	"infra-3.xyz/hyperdot-node/internal/datamodel"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"infra-3.xyz/hyperdot-node/internal/clients"
	"infra-3.xyz/hyperdot-node/internal/common"
	"infra-3.xyz/hyperdot-node/internal/store"
)

type QueryService struct {
	group          string
	bboltStore     *store.BoltStore
	bigqueryClient *clients.SimpleBigQueryClient
}

func NewQueryService(bboltStore *store.BoltStore, cfg *common.Config) (*QueryService, error) {
	ctx := context.Background()
	bigqueryClient, err := clients.NewSimpleBigQueryClient(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &QueryService{
		group:          "/query",
		bboltStore:     bboltStore,
		bigqueryClient: bigqueryClient,
	}, nil
}

func (q *QueryService) ListEnginesHandle() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// get query engines from bolt
		data, err := q.bboltStore.GetQueryEngines()
		if err != nil {
			ctx.JSON(200, BaseResponse{
				Code:    Err,
				Message: err.Error(),
			})
			return
		}

		ctx.JSON(200, ListEngineResponse{
			BaseResponse: ResponseOk(),
			Data:         data,
		})
	}
}

func (q *QueryService) GetQueryEngineDatasetHandle() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// get :engineId
		engineId := ctx.Param("engineId")
		if len(engineId) == 0 {
			ctx.JSON(200, BaseResponse{
				Code:    Err,
				Message: "engineId is empty",
			})
			return
		}

		// get :datasetId
		datasetId := ctx.Param("datasetId")
		if len(datasetId) == 0 {
			ctx.JSON(200, BaseResponse{
				Code:    Err,
				Message: "datasetId is empty",
			})
			return
		}

		engineId = strings.ToLower(engineId)
		datasetId = strings.ToLower(datasetId)

		var err error
		var data *datamodel.QueryEngineDatasetInfo
		data, err = cache.GlobalDataEngine.GetDatasets(engineId, datasetId)
		if err != nil || data == nil {
			data, err = q.bboltStore.GetDataset(engineId, datasetId)
			if err != nil {
				ctx.JSON(200, BaseResponse{
					Code:    Err,
					Message: err.Error(),
				})
				return
			}
		}

		chainTables := make(map[int][]string)
		for chainID, tables := range data.ChainTables {
			for _, table := range tables {
				chainTables[chainID] = append(chainTables[chainID], table.TableID)
			}
		}

		ctx.JSON(200, GetQueryEngineDatasetResponse{
			BaseResponse: ResponseOk(),
			Data: struct {
				Id          string                                   `json:"id"`
				Chains      map[int]datamodel.Chain                  `json:"chains"`
				RelayChains map[string]*datamodel.RelayChainMetadata `json:"relayChains"`
				ChainTables map[int][]string                         `json:"chainTables"`
			}(struct {
				Id          string
				Chains      map[int]datamodel.Chain
				RelayChains map[string]*datamodel.RelayChainMetadata
				ChainTables map[int][]string
			}{Id: data.Id, Chains: data.Chains, RelayChains: data.RelayChains, ChainTables: chainTables}),
		})
	}
}

func (q *QueryService) RunHandle() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// extract QueryRunRequest
		var req QueryRunRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(200, BaseResponse{
				Code:    Err,
				Message: err.Error(),
			})
			return
		}

		if len(req.Query) == 0 {
			ctx.JSON(200, BaseResponse{
				Code:    Err,
				Message: "Query is empty",
			})
			return
		}

		// run query
		iter, err := q.bigqueryClient.Query(ctx, req.Query)
		if err != nil {
			ctx.JSON(200, BaseResponse{
				Code:    Err,
				Message: err.Error(),
			})
			return
		}

		// extract result
		var results []map[string]bigquery.Value
		for {
			var row map[string]bigquery.Value
			err := iter.Next(&row)
			if errors.Is(err, iterator.Done) {
				break
			} else if err != nil {
				log.Fatalf("Error iterating through results: %v", err)
			}
			results = append(results, row)
		}

		ctx.JSON(200, QueryRunResponse{
			BaseResponse: ResponseOk(),
			Rows:         results,
		})
	}
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
		{
			Method:  "GET",
			Group:   q.group,
			Path:    q.group + "/engines/:engineId/datasets/:datasetId",
			Handler: q.GetQueryEngineDatasetHandle(),
		},

		{
			Method:  "POST",
			Group:   q.group,
			Path:    q.group + "/run",
			Handler: q.RunHandle(),
		},
	}
}
