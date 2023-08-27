package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
	"infra-3.xyz/hyperdot-node/internal/common"
	"infra-3.xyz/hyperdot-node/internal/datamodel"
)

type SimpleBigQueryClinet struct {
	bigquery *bigquery.Client
}

func initBigQuery(ctx context.Context, cfg *common.BigQueryConfig) (*bigquery.Client, error) {
	return bigquery.NewClient(ctx, cfg.ProjectId)

}

// NewSimpleBigQueryClient creates a new simple client for bigquery
func NewSimpleBigQueryClient(ctx context.Context, cfg *common.Config) (*SimpleBigQueryClinet, error) {
	client, err := initBigQuery(ctx, &cfg.Bigquery)
	if err != nil {
		return nil, err
	}
	return &SimpleBigQueryClinet{
		bigquery: client,
	}, nil
}

func (s *SimpleBigQueryClinet) Query(ctx context.Context, q string) (*bigquery.RowIterator, error) {
	query := s.bigquery.Query(q)
	job, err := query.Run(ctx)
	if err != nil {
		return nil, err
	}
	status, err := job.Wait(ctx)
	if err != nil {
		return nil, err
	}
	if err := status.Err(); err != nil {
		return nil, err
	}
	return job.Read(ctx)

}

// Query polkadot relaychain table schemes
func (s *SimpleBigQueryClinet) QueryCryptoPolkadotTableScheme(ctx context.Context) ([]datamodel.Table, error) {
	iter, err := s.Query(ctx, "select * from `bigquery-public-data.crypto_polkadot.AAA_tableschema`")
	if err != nil {
		log.Fatalf("Failed to run query: %v", err)
		return nil, err
	}
	return iterTable(ctx, iter)
}

// Query kusama relaychain table schemes
func (s *SimpleBigQueryClinet) QueryCryptoKusamaTableScheme(ctx context.Context) ([]datamodel.Table, error) {
	iter, err := s.Query(ctx, "select * from `bigquery-public-data.crypto_kusama.AAA_tableschema`")
	if err != nil {
		log.Fatalf("Failed to run query: %v", err)
		return nil, err
	}

	return iterTable(ctx, iter)
}

func iterTable(ctx context.Context, iter *bigquery.RowIterator) ([]datamodel.Table, error) {
	var tables []datamodel.Table
	for {
		tableRaw := datamodel.TableRaw{}
		err := iter.Next(&tableRaw)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var tableSchemes []datamodel.TableSchema
		if err := json.Unmarshal([]byte(tableRaw.TableSchemaJSON), &tableSchemes); err != nil {
			return nil, fmt.Errorf("error decoding column schema JSON: %v", err)
		}

		tables = append(tables, datamodel.Table{
			TableID: tableRaw.TableID,
			// TimePartitioningField: tableRaw.TimePartitioningField,
			Cols:    tableRaw.TableCols,
			Schemas: tableSchemes,
		})
	}
	return tables, nil
}

// Close bigquery client
func (s *SimpleBigQueryClinet) Close() error {
	return s.bigquery.Close()
}
