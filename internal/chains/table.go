package chains

// TableSchema represents a table schema for dataengine.
type TableSchema struct {
	Mode string `json:"mode"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// Table represents a table for dataengine.
type Table struct {
	Name    string        `json:"name"`
	Cols    []string      `json:"cols"`
	Schemas []TableSchema `json:"schemas"`
}
