package main

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"

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

func initialGlobalData(cfg *common.Config) error {
	// Fetch parachain data and initialize global cache
	ctx := context.Background()
	client, err := clients.NewSimpleBigQueryClient(ctx, cfg)
	if err != nil {
		return err
	}

	defer client.Close()

	bigQueryDataEngine, err := jobs.BuildBigQueryEngine(ctx, client, &cfg.Polkaholic)
	if err != nil {
		return err
	}
	cache.GlobalDataEngine.SetBigQuery(bigQueryDataEngine)

	return nil
}

func initJobs(jm *jobs.JobManager) error {
	if err := jm.Init(); err != nil {
		return err
	}

	go func() {
		<-jm.Start()
		jm.Stop()
	}()

	return nil
}

func main() {
	flag.Parse()

	cfg := initialSystemConfig()
	if err := initialGlobalData(cfg); err != nil {
		log.Fatalf("Error initial global data: %v", err)
	}

	jobManager := jobs.NewJobManager(cfg)
	if err := initJobs(jobManager); err != nil {
		log.Fatalf("Error initial jobs: %v", err)
	}

	apiserver, err := apis.NewApiServer(cfg)
	if err != nil {
		log.Fatalf("Error creating apiserver: %v", err)
	}

	if err := apiserver.Start(); err != nil {
		log.Fatalf("Error starting apiserver: %v", err)
	}
}
