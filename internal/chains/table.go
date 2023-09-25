package chains

type TableSchema struct {
	Mode string `json:"mode"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type Table struct {
	Name    string        `json:"name"`
	Cols    []string      `json:"cols"`
	Schemas []TableSchema `json:"schemas"`
}
