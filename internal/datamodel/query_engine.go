package datamodel

type RelayChainMetadata struct {
	ChainID      int    `json:"chainID"`
	Name         string `json:"name"`
	ShowColor    string `json:"showColor"`
	ParaChainIDs []int  `json:"paraChainIDs"`
}

type QueryEngineDatasetInfo struct {
	Id          string                         `json:"id"`
	Chains      map[int]Chain                  `json:"chains"`
	RelayChains map[string]*RelayChainMetadata `json:"relayChains"`
	ChainTables map[int][]Table                `json:"chainTables"`
}

type QueryEngineDatasets struct {
	Raw *QueryEngineDatasetInfo `json:"raw"`
}

type QueryEngineDatasetMetadata struct {
	Id          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type QueryEngine struct {
	Name     string                                `json:"name"`
	Datasets map[string]QueryEngineDatasetMetadata `json:"datasets"`
}
