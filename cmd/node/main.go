package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"infra-3.xyz/hyperdot-node/internal/dataengine"
	"infra-3.xyz/hyperdot-node/internal/datamodel"

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

func initDB(cfg *common.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=%s",
		cfg.Postgres.Host,
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.DBName,
		cfg.Postgres.Port,
		cfg.Postgres.TimeZone,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, err
	}

	// execute auto migratre
	if err := db.AutoMigrate(&datamodel.UserModel{}); err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&datamodel.UserStatistics{}); err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&datamodel.QueryModel{}); err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&datamodel.ChartModel{}); err != nil {
		return nil, err
	}

	return db, nil

}

func initEngines(cfg *common.Config) (map[string]dataengine.QueryEngine, error) {
	res := make(map[string]dataengine.QueryEngine)
	bigquery, err := dataengine.Make(dataengine.BigQueryName, &dataengine.BigQueryEngineConfig{
		ProjectId: cfg.Bigquery.ProjectId,
	})

	if err != nil {
		return nil, err
	}

	res[dataengine.BigQueryName] = bigquery
	return res, nil
}

func initS3Client(cfg *common.Config) (*clients.SimpleS3Cliet, error) {
	return clients.NewSimpleS3Client(cfg.S3.Endpoint, cfg.S3.AccessKey, cfg.S3.SecretKey, cfg.S3.UseSSL), nil
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

	db, err := initDB(cfg)
	if err != nil {
		log.Fatalf("Error initial database: %v", err)
	}

	engines, err := initEngines(cfg)
	if err != nil {
		log.Fatalf("Error initial query engines: %v", err)
	}

	s3Client, err := initS3Client(cfg)
	if err != nil {
		log.Fatalf("Error initial s3 client: %v", err)
	}

	apiserver, err := apis.NewApiServer(boltStore, cfg, db, engines, s3Client)
	if err != nil {
		log.Fatalf("Error creating apiserver: %v", err)
	}

	if err := apiserver.Start(); err != nil {
		log.Fatalf("Error starting apiserver: %v", err)
	}
}
