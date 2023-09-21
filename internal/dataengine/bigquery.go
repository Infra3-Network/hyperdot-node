package dataengine

import (
	"cloud.google.com/go/bigquery"
	"context"
	"errors"
	"fmt"
	"google.golang.org/api/iterator"
)

type BigQueryEngineConfig struct {
	ProjectId string `json:"projectId"`
}

type BigQueryEngine struct {
	ctx    context.Context
	client *bigquery.Client
}

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

type BigQueryEngineRowIter struct {
	iter *bigquery.RowIterator
}

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

func (b BigQueryEngineRowIter) TotalRows() uint64 {
	return b.iter.TotalRows
}
