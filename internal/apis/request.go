package apis

type QueryRunRequest struct {
	Engine string `json:"engine"`
	Query  string `json:"query"`
}
