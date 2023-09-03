package main

import (
	"context"
	"encoding/json"
	"flag"
	"infra-3.xyz/hyperdot-node/internal/datamodel"
	"io"
	"log"
	"os"

	"infra-3.xyz/hyperdot-node/internal/store"

	"infra-3.xyz/hyperdot-node/internal/apis"
	"infra-3.xyz/hyperdot-node/internal/cache"
	"infra-3.xyz/hyperdot-node/internal/clients"
	"infra-3.xyz/hyperdot-node/internal/common"
	"infra-3.xyz/hyperdot-node/internal/jobs"
)

var (
	config = flag.String("config", "config.json", "Path to config file")
)

func initialSystemConfig() *common.Config {
	cfg := ""
	if config == nil {
		cfg = os.Getenv("HYPETDOT_NODE_CONFIG")
	} else {
		cfg = *config
	}

	if cfg == "" {
		log.Fatalf("Config file not found")
	}

	configFile, err := os.Open(cfg)
	if err != nil {
		log.Fatalf("Error opening config file: %v", err)
	}
	defer configFile.Close()

	data, err := io.ReadAll(configFile)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	config := new(common.Config)
	if err := json.Unmarshal(data, config); err != nil {
		log.Fatalf("Error unmarshalling config JSON: %v", err)
	}

	return config
}

func initialGlobalData(cfg *common.Config, boltStore *store.BoltStore) error {
	// Fetch chain metadata and initialize global cache
	ctx := context.Background()
	client, err := clients.NewSimpleBigQueryClient(ctx, cfg)
	if err != nil {
		return err
	}

	defer client.Close()

	raw, err := jobs.BuildBigQueryEngineRawDataset(ctx, client, &cfg.Polkaholic)
	if err != nil {
		return err
	}

	cache.GlobalDataEngine.SetDatasets("bigquery", &datamodel.QueryEngineDatasets{Raw: raw})
	if err := boltStore.SetDatasets("bigquery", &datamodel.QueryEngineDatasets{Raw: raw}); err != nil {
		return err
	}
	return nil
}

func initJobs(jobManager *jobs.JobManager, store *store.BoltStore) error {
	if err := jobManager.Init(store); err != nil {
		return err
	}

	go func() {
		<-jobManager.Start()
		jobManager.Stop()
	}()

	return nil
}

func main() {
	flag.Parse()

	cfg := initialSystemConfig()
	boltStore, err := store.NewBoltStore(cfg)
	if err != nil {
		log.Fatalf("Error initial bolt store: %v", err)
	}

	if err := initialGlobalData(cfg, boltStore); err != nil {
		log.Fatalf("Error initial global data: %v", err)
	}

	jobManager := jobs.NewJobManager(cfg)

	if err := initJobs(jobManager, boltStore); err != nil {
		log.Fatalf("Error initial jobs: %v", err)
	}

	apiserver, err := apis.NewApiServer(boltStore, cfg)
	if err != nil {
		log.Fatalf("Error creating apiserver: %v", err)
	}

	if err := apiserver.Start(); err != nil {
		log.Fatalf("Error starting apiserver: %v", err)
	}
}
