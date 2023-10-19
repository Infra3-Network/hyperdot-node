package datamodel

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"infra-3.xyz/hyperdot-node/internal/common"
)

var (
	HyperdotQueryEnginesKey      = "hyperdot:queryengines"
	BigQueryRawPolkadotChainsKey = "bigquery:polkadot:raw:chains"
	BigQueryRawPolkadotRelayKey  = "bigquery:polkadot:raw:relaychains"
	BigQueryRawPolkadotTablesKey = "bigquery:polkadot:raw:tables"
)

type RelayChainMetadata struct {
	ChainID      uint   `json:"chainID"`
	Name         string `json:"name"`
	ShowColor    string `json:"showColor"`
	ParaChainIDs []uint `json:"paraChainIDs"`
}

type QueryEngineDatasetInfo struct {
	Id          string                         `json:"id"`
	Chains      map[uint]ChainModel            `json:"chains"`
	RelayChains map[string]*RelayChainMetadata `json:"relayChains"`
	ChainTables map[int][]Table                `json:"chainTables"`
}

func (q *QueryEngineDatasetInfo) WriteToRedis(ctx context.Context, cfg *common.RedisConfig, engine string) error {
	// set raw
	raw := q
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.Addr,
		DB:   0,
	})

	defer redisClient.Close()

	_, err := redisClient.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		hMapKey := fmt.Sprintf("%s:polkadot:raw:chains", engine)
		for _, chain := range raw.Chains {
			key := fmt.Sprintf("%d", chain.ChainID)
			value, err := json.Marshal(chain)
			if err != nil {
				return err
			}
			cmd := pipe.HSet(ctx, hMapKey, key, string(value))
			if cmd.Err() != nil {
				return err
			}

			log.Printf("HMap [%s] set redis key: %s", hMapKey, key)
		}

		hMapKey = fmt.Sprintf("%s:polkadot:raw:relaychains", engine)
		for _, chain := range raw.RelayChains {
			key := chain.Name
			value, err := json.Marshal(chain)
			if err != nil {
				return err
			}

			cmd := pipe.HSet(ctx, hMapKey, key, string(value))
			if cmd.Err() != nil {
				return err
			}
			log.Printf("HMap [%s] set redis key: %s", hMapKey, key)
		}

		hMapKey = fmt.Sprintf("%s:polkadot:raw:tables", engine)
		for chainId, tables := range raw.ChainTables {
			key := fmt.Sprintf("%d", chainId)
			value, err := json.Marshal(tables)
			if err != nil {
				return err
			}

			cmd := pipe.HSet(ctx, hMapKey, key, string(value))
			if cmd.Err() != nil {
				return err
			}

			log.Printf("HMap [%s] set redis key: %s", hMapKey, key)
		}
		return nil
	})

	if err != nil {
		log.Printf("Error set raw data: %v", err)
		return err
	}

	return nil
}

type QueryEngineDatasets struct {
	// Raw is raw blockchain data
	Raw *QueryEngineDatasetInfo `json:"raw"`
}

type QueryEngineDatasetMetadata struct {
	Id          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type QueryEngine struct {
	Name     string                                `json:"name"`
	Datasets map[string]QueryEngineDatasetMetadata `json:"datasets"`
}

func InitQueryEngineMetadata(ctx context.Context, redisClient *redis.Client) error {
	// write to redis
	engines := []QueryEngine{
		{
			Name: "Bigquery",
			Datasets: map[string]QueryEngineDatasetMetadata{
				"Bigquery": {
					Id:          "raw",
					Title:       "Raw",
					Description: "Raw blockchain crypto data",
				},
			},
		},
	}

	for _, engine := range engines {
		key := engine.Name
		value, err := json.Marshal(engine)
		if err != nil {
			return err
		}

		cmd := redisClient.HSet(ctx, HyperdotQueryEnginesKey, key, string(value))
		if cmd.Err() != nil {
			return err
		}
		log.Printf("HMap [%s] set redis key: %s", HyperdotQueryEnginesKey, key)
	}

	return nil
}
