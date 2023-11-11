package query

// RequestRunQuery is the request body for the RunQuery endpoint
type RequestRunQuery struct {
	Query  string `json:"query"`
	Engine string `json:"engine"`
}
