package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"infra-3.xyz/hyperdot-node/internal/apis/base"
	"infra-3.xyz/hyperdot-node/internal/dataengine"
	"infra-3.xyz/hyperdot-node/internal/datamodel"
	"infra-3.xyz/hyperdot-node/internal/utils"

	"infra-3.xyz/hyperdot-node/internal/store"

	"infra-3.xyz/hyperdot-node/internal/apis"
	"infra-3.xyz/hyperdot-node/internal/clients"
	"infra-3.xyz/hyperdot-node/internal/common"
	"infra-3.xyz/hyperdot-node/internal/jobs"
)

var (
	config = flag.String("config", "", "Path to config file")
)

func initialSystemConfig() *common.Config {
	cfg := ""
	if config == nil || *config == "" {
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

func initRedisClient(cfg *common.Config) (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.Redis.Addr,
	})

	return redisClient, nil
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

	if err := db.AutoMigrate(&datamodel.DashboardModel{}); err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&datamodel.DashboardPanelModel{}); err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&datamodel.UserDashboardFavorites{}); err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&datamodel.UserQueryFavorites{}); err != nil {
		return nil, err
	}

	if err := datamodel.HackAutoMigrate(db); err != nil {
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

func createTestUser(db *gorm.DB) error {
	password, err := utils.GeneratePassword("test")
	if err != nil {
		return err
	}
	user := &datamodel.UserModel{
		UserBasic: datamodel.UserBasic{
			Provider:          "password",
			Username:          "test",
			Email:             "test",
			EncryptedPassword: password,
		},
	}

	if err := db.Save(user).Error; err != nil {
		return err
	}

	return nil
}

func MakeTestUserToken() (string, error) {
	password, err := utils.GeneratePassword("test")
	if err != nil {
		return "", err
	}
	user := &datamodel.UserModel{
		ID: 1,
		UserBasic: datamodel.UserBasic{
			Provider:          "password",
			Username:          "test",
			Email:             "test",
			EncryptedPassword: password,
		},
	}

	expireAt := base.TokenDefaultExpireTime()
	signing, err := base.GenerateJwtToken(user.ToClaims(), expireAt)
	if err != nil {
		return "", err
	}

	return url.QueryEscape(signing), nil
}

func MakeTokenRequest(method string, path string, body any) (*http.Request, error) {
	var reader io.Reader

	// make body as json
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = strings.NewReader(string(data))
	}

	req, _ := http.NewRequest(method, path, reader)
	token, err := MakeTestUserToken()
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", token)
	return req, nil
}

func MarshalResponseBody(body *bytes.Buffer, to interface{}) error {
	if err := json.Unmarshal(body.Bytes(), to); err != nil {
		return err
	}

	return nil
}

func setupGCloud() {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	gcloud_cred := pwd + "/hyperdot-gcloud-iam.test.json"
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", gcloud_cred)
}

func SetupApiServer() *apis.ApiServer {
	os.Setenv("HYPETDOT_NODE_CONFIG", "hyperdot.test.json")

	setupGCloud()

	_ = context.Background()
	cfg := initialSystemConfig()
	boltStore, err := store.NewBoltStore(cfg)
	if err != nil {
		log.Fatalf("Error initial bolt store: %v", err)
	}

	_, err = initRedisClient(cfg)
	if err != nil {
		log.Fatalf("Error initial redis client: %v", err)
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

	if err := createTestUser(db); err != nil {
		log.Fatalf("Error creating test user: %v", err)
	}

	apiserver, err := apis.NewApiServer(boltStore, cfg, db, engines, s3Client)
	if err != nil {
		log.Fatalf("Error creating apiserver: %v", err)
	}

	return apiserver
}
