package dataengine

import (
	"context"
	"errors"
	"fmt"
)

const (
	// BigQueryName is the name of bigquery engine.
	BigQueryName = "bigquery"
)

// FieldSchema describes a single field.
type FieldSchema struct {
	// The field name.
	// Must contain only letters (a-z, A-Z), numbers (0-9), or underscores (_),
	// and must start with a letter or underscore.
	// The maximum length is 128 characters.
	Name string `json:"name"`

	// A description of the field. The maximum length is 16,384 characters.
	Description string `json:"description"`

	// Whether the field may contain multiple values.
	Repeated bool `json:"repeated"`
	// Whether the field is required.  Ignored if Repeated is true.
	Required bool `json:"required"`

	// The field data type.  If Type is Record, then this field contains a nested schema,
	// which is described by Schema.
	Type string `json:"type"`
}

// IterDone is returned by RowIterator.Next when there are no more items.
var IterDone = errors.New("no more items in iterator")

// RowIterator iterates over a set of rows.
type RowIterator interface {
	// Schema returns the schema of the rows.
	Schema() []*FieldSchema

	// Next returns the next row.  If there are no more rows, it returns IterDone.
	Next() (map[string]interface{}, error)

	// TotalRows returns the total number of rows in the iterator.
	TotalRows() uint64
}

// QueryEngine is the interface for query engine.
type QueryEngine interface {
	// Run executes a query and return a row iterator.
	Run(ctx context.Context, query string) (RowIterator, error)
}

// Make creates a new query engine by given engine name and config.
func Make(engine string, cfg interface{}) (QueryEngine, error) {
	switch engine {
	case BigQueryName:
		return NewBigQueryEngine(cfg)
	default:
		return nil, fmt.Errorf("unsupported %s data engine", engine)
	}
}
