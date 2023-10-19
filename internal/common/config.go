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

type BoltStoreConfig struct {
	Path string `json:"path"`
}

type LocalStoreConfig struct {
	Bolt BoltStoreConfig `json:"bolt"`
}

type PostgresConfig struct {
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"db"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	TimeZone string `json:"tz"`
}

type S3Config struct {
	Endpoint  string `json:"endpoint"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
	UseSSL    bool   `json:"useSSL"`
}

type RedisConfig struct {
	Addr string `json:"addr"`
}

type Config struct {
	Polkaholic PolkaholicConfig `json:"polkaholic"`
	ApiServer  ApiServerConfig  `json:"apiServer"`
	Bigquery   BigQueryConfig   `json:"bigquery"`
	LocalStore LocalStoreConfig `json:"localStore"`
	Postgres   PostgresConfig   `json:"postgres"`
	S3         S3Config         `json:"s3"`
	Redis      RedisConfig      `json:"redis"`
}
