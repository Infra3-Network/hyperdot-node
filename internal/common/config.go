package common

type PolkaholicConfig struct {
	ApiKey  string `json:"apiKey"`
	BaseUrl string `json:"baseUrl"`
}

type ApiServerConfig struct {
	Addr string `json:"addr"`
}

type BigQueryConfig struct {
	ProjectId string `json:"projectId"`
}

type Config struct {
	Polkaholic PolkaholicConfig `json:"polkaholic"`
	ApiServer  ApiServerConfig  `json:"apiServer"`
	Bigquery   BigQueryConfig   `json:"bigquery"`
}
