package dataengine

import (
	"context"
	"errors"
	"fmt"
)

const (
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

var IterDone = errors.New("no more items in iterator")

type RowIterator interface {
	Schema() []*FieldSchema

	Next() (map[string]interface{}, error)

	TotalRows() uint64
}

type QueryEngine interface {
	Run(ctx context.Context, query string) (RowIterator, error)
}

func Make(engine string, cfg interface{}) (QueryEngine, error) {
	switch engine {
	case BigQueryName:
		return NewBigQueryEngine(cfg)
	default:
		return nil, fmt.Errorf("unsupported %s data engine", engine)
	}
}
