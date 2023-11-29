package dataengine

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

// BigQueryEngineConfig is the config for bigquery engine.
type BigQueryEngineConfig struct {
	ProjectId string `json:"projectId"`
}

// BigQueryEngine is the query engine for bigquery.
type BigQueryEngine struct {
	ctx    context.Context
	client *bigquery.Client
}

// NewBigQueryEngine creates a new bigquery engine.
func NewBigQueryEngine(cfg interface{}) (*BigQueryEngine, error) {
	v, ok := cfg.(*BigQueryEngineConfig)
	if !ok {
		return nil, fmt.Errorf("config type incompatible")
	}

	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, v.ProjectId)
	if err != nil {
		return nil, err
	}

	return &BigQueryEngine{ctx: ctx, client: client}, nil
}

// Run executes a query and return a row iterator.
func (bq *BigQueryEngine) Run(ctx context.Context, query string) (RowIterator, error) {
	q := bq.client.Query(query)
	job, err := q.Run(ctx)
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

	iter, err := job.Read(ctx)
	if err != nil {
		return nil, err
	}

	return &BigQueryEngineRowIter{iter: iter}, nil
}

// BigQueryEngineRowIter is the row iterator for bigquery.
type BigQueryEngineRowIter struct {
	iter *bigquery.RowIterator
}

// Schema returns the schema of the rows.
func (b BigQueryEngineRowIter) Schema() []*FieldSchema {
	res := make([]*FieldSchema, len(b.iter.Schema))
	for i, filed := range b.iter.Schema {
		res[i] = &FieldSchema{
			Name:        filed.Name,
			Description: filed.Description,
			Repeated:    filed.Repeated,
			Required:    filed.Required,
			Type:        string(filed.Type),
		}
	}
	return res
}

// Next returns the next row.  If there are no more rows, it returns IterDone.
func (b BigQueryEngineRowIter) Next() (map[string]interface{}, error) {
	var row map[string]bigquery.Value
	err := b.iter.Next(&row)
	if err != nil {
		if errors.Is(err, iterator.Done) {
			return nil, IterDone
		}
		return nil, err
	}

	res := make(map[string]interface{}, len(row))
	for k, v := range row {
		res[k] = v
	}
	return res, nil
}

// TotalRows returns the total number of rows in the iterator.
func (b BigQueryEngineRowIter) TotalRows() uint64 {
	return b.iter.TotalRows
}
