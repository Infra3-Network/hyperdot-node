package query

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"infra-3.xyz/hyperdot-node/internal/dataengine"
	"log"
	"net/http"
	"strings"

	"cloud.google.com/go/bigquery"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"

	"infra-3.xyz/hyperdot-node/internal/apis/base"
	"infra-3.xyz/hyperdot-node/internal/cache"
	"infra-3.xyz/hyperdot-node/internal/clients"
	"infra-3.xyz/hyperdot-node/internal/common"
	"infra-3.xyz/hyperdot-node/internal/datamodel"
	"infra-3.xyz/hyperdot-node/internal/store"
)

const Name = "query"

type Service struct {
	group          string
	db             *gorm.DB
	bboltStore     *store.BoltStore
	bigqueryClient *clients.SimpleBigQueryClient
	engines        map[string]dataengine.QueryEngine
}

func New(bboltStore *store.BoltStore, cfg *common.Config, db *gorm.DB, engines map[string]dataengine.QueryEngine) *Service {
	ctx := context.Background()
	bigqueryClient, err := clients.NewSimpleBigQueryClient(ctx, cfg)
	if err != nil {
		panic(err)
	}
	return &Service{
		group:          "/query",
		db:             db,
		bboltStore:     bboltStore,
		bigqueryClient: bigqueryClient,
		engines:        engines,
	}
}

func (s *Service) ListEnginesHandle() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// get query engines from bolt
		data, err := s.bboltStore.GetQueryEngines()
		if err != nil {
			ctx.JSON(200, base.BaseResponse{
				ErrorCode:    base.Err,
				ErrorMessage: err.Error(),
			})
			return
		}

		ctx.JSON(200, base.ListEngineResponse{
			BaseResponse: base.ResponseOk(),
			Data:         data,
		})
	}
}

func (s *Service) GetQueryEngineDatasetHandle() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// get :engineId
		engineId := ctx.Param("engineId")
		if len(engineId) == 0 {
			ctx.JSON(200, base.BaseResponse{
				ErrorCode:    base.Err,
				ErrorMessage: "engineId is empty",
			})
			return
		}

		// get :datasetId
		datasetId := ctx.Param("datasetId")
		if len(datasetId) == 0 {
			ctx.JSON(200, base.BaseResponse{
				ErrorCode:    base.Err,
				ErrorMessage: "datasetId is empty",
			})
			return
		}

		engineId = strings.ToLower(engineId)
		datasetId = strings.ToLower(datasetId)

		var err error
		var data *datamodel.QueryEngineDatasetInfo
		data, err = cache.GlobalDataEngine.GetDatasets(engineId, datasetId)
		if err != nil || data == nil {
			data, err = s.bboltStore.GetDataset(engineId, datasetId)
			if err != nil {
				ctx.JSON(200, base.BaseResponse{
					ErrorCode:    base.Err,
					ErrorMessage: err.Error(),
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
		ctx.JSON(200, base.GetQueryEngineDatasetResponse{
			BaseResponse: base.ResponseOk(),
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

// @BasePath /apis/v1/

// @Summary run query
// @Schemes
// @Description run query
// @Accept json
// @Produce json
// @Success 200 {QueryRunResponseData} QueryRunResponseData
// @Router /apis/v1/query/run [post]
func (s *Service) RunHandle() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// extract QueryRunRequest
		var req base.QueryRunRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(200, base.BaseResponse{
				ErrorCode:    base.Err,
				ErrorMessage: err.Error(),
			})
			return
		}

		if len(req.Query) == 0 {
			ctx.JSON(200, base.BaseResponse{
				ErrorCode:    base.Err,
				ErrorMessage: "Query is empty",
			})
			return
		}

		if len(req.Engine) == 0 {
			ctx.JSON(200, base.BaseResponse{
				ErrorCode:    base.Err,
				ErrorMessage: "Engine is empty",
			})
			return
		}

		if strings.ToLower(req.Engine) != "bigquery" {
			ctx.JSON(200, base.BaseResponse{
				ErrorCode:    base.Err,
				ErrorMessage: fmt.Sprintf("Engine %s is not supported", req.Engine),
			})
			return
		}

		// run query
		iter, err := s.bigqueryClient.Query(ctx, req.Query)
		if err != nil {
			ctx.JSON(200, base.BaseResponse{
				ErrorCode:    base.Err,
				ErrorMessage: err.Error(),
			})
			return
		}

		var schemas []datamodel.TableSchema
		for _, filed := range iter.Schema {
			mode := ""
			if filed.Repeated {
				mode = "REPEATED"
			} else if filed.Required {
				mode = "REQUIRED"
			} else {
				mode = "NULLABLE"
			}

			schemas = append(schemas, datamodel.TableSchema{
				Mode: mode,
				Name: filed.Name,
				Type: string(filed.Type),
			})
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

		ctx.JSON(200, base.QueryRunResponse{
			BaseResponse: base.ResponseOk(),
			Data:         base.QueryRunResponseData{Schemas: schemas, Rows: results},
		})
	}
}

func (s *Service) runHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		query := ctx.Query("q")
		engineName := ctx.Query("engine")
		_, err := base.CurrentUserId(ctx)
		if err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}

		if len(query) == 0 {
			base.ResponseErr(ctx, http.StatusBadRequest, "query engine is required")
			return
		}
		if len(engineName) == 0 {
			base.ResponseErr(ctx, http.StatusBadRequest, "query is required")
			return
		}

		engine, ok := s.engines[engineName]
		if !ok {
			base.ResponseErr(ctx, http.StatusBadRequest, "The %s query engine unsupported now", engineName)
			return
		}

		iter, err := engine.Run(context.Background(), query)
		if err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, "query error: %v", err)
			return
		}

		var rows []map[string]interface{}
		for {
			row, err := iter.Next()
			if err != nil {
				if errors.Is(err, dataengine.IterDone) {
					break
				}
				base.ResponseErr(ctx, http.StatusBadRequest, "query error: %v", err)
				return
			}
			rows = append(rows, row)
		}

		schemas := iter.Schema()

		ctx.JSON(http.StatusOK, RunResponse{
			BaseResponse: base.BaseResponse{
				Success: true,
			},
			Data: RunResponseData{
				Rows:    rows,
				Schemas: schemas,
			},
		})
	}
}

func (s *Service) checkUserQueryModelRequest(ctx *gin.Context, model *datamodel.UserQueryModel) bool {
	currentUserId, err := base.CurrentUserId(ctx)
	if err != nil {
		base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
		return false
	}

	if model.UserID != 0 && model.UserID != currentUserId {
		base.ResponseErr(ctx, http.StatusUnauthorized, "Unauthorized")
		return false
	}

	model.UserID = currentUserId

	if len(model.QueryEngine) == 0 {
		base.ResponseErr(ctx, http.StatusBadRequest, "query engine is required")
		return false
	}
	if len(model.Query) == 0 {
		base.ResponseErr(ctx, http.StatusBadRequest, "query is required")
		return false
	}

	if !model.Unsaved {
		if len(model.Name) == 0 {
			base.ResponseErr(ctx, http.StatusBadRequest, "name is required")
			return false
		}
	} else {
		model.Name = "unsaved"
	}

	return true
}

func (s *Service) updateHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var request datamodel.UserQueryModel
		if err := ctx.ShouldBindJSON(&request); err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, "bind error: %v", err)
			return
		}

		if !s.checkUserQueryModelRequest(ctx, &request) {
			return
		}

		result := s.db.Save(&request)
		if result.Error != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, result.Error.Error())
			return
		}

		ctx.JSON(http.StatusOK, Response{
			BaseResponse: base.BaseResponse{
				Success: true,
			},
			Data: request,
		})
	}
}

func (s *Service) Name() string {
	return Name
}

func (s *Service) RouteTables() []base.RouteTable {
	return []base.RouteTable{
		{
			Method:  "GET",
			Path:    s.group + "/engines",
			Handler: s.ListEnginesHandle(),
		},
		{
			Method:  "GET",
			Path:    s.group + "/engines/:engineId/datasets/:datasetId",
			Handler: s.GetQueryEngineDatasetHandle(),
		},
		{
			Method:  "GET",
			Path:    s.group + "/run",
			Handler: s.runHandler(),
		},
		{
			Method:  "PUT",
			Path:    s.group,
			Handler: s.updateHandler(),
		},
	}
}
