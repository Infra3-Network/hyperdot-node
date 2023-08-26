package clients

import (
	"context"

	"cloud.google.com/go/bigquery"
	"infra-3.xyz/hyperdot-node/internal/common"
)

type SimpleBigQueryClinet struct {
	bigquery *bigquery.Client
}

func initBigQuery(cfg *common.BigQueryConfig) (*bigquery.Client, error) {
	ctx := context.Background()
	return bigquery.NewClient(ctx, cfg.ProjectId)

}

// NewSimpleBigQueryClient creates a new simple client for bigquery
func NewSimpleBigQueryClient(cfg *common.Config) (*SimpleBigQueryClinet, error) {
	client, err := initBigQuery(&cfg.Bigquery)
	if err != nil {
		return nil, err
	}
	return &SimpleBigQueryClinet{
		bigquery: client,
	}, nil
}
