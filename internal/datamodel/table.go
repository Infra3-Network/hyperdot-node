package datamodel

// TableSchema represents a table schema for dataengine.
type TableSchema struct {
	Mode string `json:"mode"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// TableRaw represents a table schema for dataengine.
type TableRaw struct {
	TableID string `bigquery:"table_id"`
	// TimePartitioningField string   `bigquery:"time_partitioning_field"`
	TableCols       []string `bigquery:"table_cols"`
	TableSchemaJSON string   `bigquery:"table_schema"`
}

// Table represents a table for dataengine.
type Table struct {
	TableID string `json:"table_id"`
	// TimePartitioningField string        `json:"time_partitioning_field"`
	Cols    []string      `json:"cols"`
	Schemas []TableSchema `json:"schemas"`
}
