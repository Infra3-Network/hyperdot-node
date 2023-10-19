package system

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"infra-3.xyz/hyperdot-node/internal/apis/base"
	"infra-3.xyz/hyperdot-node/internal/common"
	"infra-3.xyz/hyperdot-node/internal/datamodel"
)

type Service struct {
	redisClient *redis.Client
	// bboltStore  *store.BoltStore
}

func New(cfg *common.Config) *Service {
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.Redis.Addr,
	})

	return &Service{
		redisClient: redisClient,
		// bboltStore:  bboltStore,
	}
}

func (s *Service) listEnginesHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cmd := s.redisClient.HGetAll(context.Background(), datamodel.HyperdotQueryEnginesKey)
		if cmd.Err() != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, cmd.Err().Error())
		}

		var data []datamodel.QueryEngine
		for _, v := range cmd.Val() {
			var engine datamodel.QueryEngine
			if err := json.Unmarshal([]byte(v), &engine); err != nil {
				base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
			}
			data = append(data, engine)
		}

		base.ResponseWithData(ctx, data)

	}
}

func (s *Service) getBigqueryDataset() (map[string]interface{}, error) {
	cmds := s.redisClient.HGetAll(context.Background(), datamodel.BigQueryRawPolkadotChainsKey)
	if cmds.Err() != nil {
		return nil, cmds.Err()
	}

	res := make(map[string]interface{})

	chains := make(map[string]interface{})
	for k, v := range cmds.Val() {
		var chain datamodel.ChainModel
		if err := json.Unmarshal([]byte(v), &chain); err != nil {
			return nil, err
		}

		chains[k] = chain
	}

	cmds = s.redisClient.HGetAll(context.Background(), datamodel.BigQueryRawPolkadotRelayKey)
	if cmds.Err() != nil {
		return nil, cmds.Err()
	}

	relayChains := make(map[string]interface{})
	for k, v := range cmds.Val() {
		var relayChain datamodel.RelayChainMetadata
		if err := json.Unmarshal([]byte(v), &relayChain); err != nil {
			return nil, err
		}

		relayChains[k] = relayChain
	}

	cmds = s.redisClient.HGetAll(context.Background(), datamodel.BigQueryRawPolkadotTablesKey)
	if cmds.Err() != nil {
		return nil, cmds.Err()
	}

	chainTables := make(map[int]interface{})
	for k, v := range cmds.Val() {
		var tables []datamodel.Table
		if err := json.Unmarshal([]byte(v), &tables); err != nil {
			return nil, err
		}

		chainId, err := strconv.Atoi(k)
		if err != nil {
			log.Printf("Error convert chainId to int: %v", err)
			continue
		}

		chainTables[chainId] = tables
	}

	res["chains"] = chains
	res["relayChains"] = relayChains
	res["chainTables"] = chainTables

	return res, nil

}

func (s *Service) getQueryEngineDatasetHandle() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		engineId := ctx.Param("engineId")
		if len(engineId) == 0 {
			ctx.JSON(200, base.BaseResponse{
				ErrorCode:    base.Err,
				ErrorMessage: "engineId is empty",
			})
			return
		}

		switch engineId {
		case "bigquery":
			dataset, err := s.getBigqueryDataset()
			if err != nil {
				base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
				return
			}

			base.ResponseWithData(ctx, dataset)
		default:
			base.ResponseErr(ctx, http.StatusBadRequest, "engineId is invalid")
			return
		}

	}
}

func (s *Service) Name() string {
	return "system"
}

func (s *Service) RouteTables() []base.RouteTable {
	group := "system"
	return []base.RouteTable{
		{
			Method:  "GET",
			Path:    group + "/engines",
			Handler: s.listEnginesHandler(),
		},
		{
			Method:  "GET",
			Path:    group + "/engines/:engineId",
			Handler: s.getQueryEngineDatasetHandle(),
		},
	}
}
