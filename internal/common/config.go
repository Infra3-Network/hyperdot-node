package common

type Polkaholic struct {
	ApiKey  string `json:"apiKey"`
	BaseUrl string `json:"baseUrl"`
}

type Config struct {
	Polkaholic Polkaholic `json:"polkaholic"`
}
