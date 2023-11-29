package common

// PolkaholicConfig is the config for polkaholic.
type PolkaholicConfig struct {
	// ApiKey is the api key for polkaholic.
	// See https://polkaholic.io/login for more details.
	ApiKey string `json:"apiKey"`
	// BaseUrl is the base api address for polkaholic.
	BaseUrl string `json:"baseUrl"`
}

// ApiServerConfig is the config for hyperdot-node
type ApiServerConfig struct {
	// Addr is the listen address of hyperdot-node server.
	Addr string `json:"addr"`
}

// BigQueryConfig is the config for google bigquery.
type BigQueryConfig struct {
	// ProjectId is the project id for google bigquery.
	ProjectId string `json:"projectId"`
}

// BoltStoreConfig is the config for bblot store.
type BoltStoreConfig struct {
	Path string `json:"path"`
}

// LocalStoreConfig is the config for manager multiple storage in local.
type LocalStoreConfig struct {
	// Refer to BoltStoreConfig
	Bolt BoltStoreConfig `json:"bolt"`
}

// PostgresConfig is the config for postgres.
type PostgresConfig struct {
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"db"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	TimeZone string `json:"tz"`
}

// S3Config is the config for s3.
type S3Config struct {
	Endpoint  string `json:"endpoint"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
	UseSSL    bool   `json:"useSSL"`
}

// RedisConfig is the config for redis.
type RedisConfig struct {
	Addr string `json:"addr"`
}

// Config is the config for hyperdot-node.
type Config struct {
	// Refer to PolkaholicConfig
	Polkaholic PolkaholicConfig `json:"polkaholic"`
	// Refer to ApiServerConfig
	ApiServer ApiServerConfig `json:"apiServer"`
	// Refer to BigQueryConfig
	Bigquery BigQueryConfig `json:"bigquery"`
	// Refer to LocalStoreConfig
	LocalStore LocalStoreConfig `json:"localStore"`
	// Refer to PostgresConfig
	Postgres PostgresConfig `json:"postgres"`
	// Refer to S3Config
	S3 S3Config `json:"s3"`
	// Refer to RedisConfig
	Redis RedisConfig `json:"redis"`
}
