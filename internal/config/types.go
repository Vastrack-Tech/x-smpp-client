package config

type Config struct {
	AppEnv         string `mapstructure:"APP_ENV"`
	EnquireLink    int    `mapstructure:"ENQUIRE_LINK"`
	ReadTimeout    int    `mapstructure:"READ_TIMEOUT"`
	WriteTimeout   int    `mapstructure:"WRITE_TIMEOUT"`
	PoolSize       int    `mapstructure:"POOL_SIZE"`

	SMSCAddr       string `mapstructure:"SMSC_ADDR"`
	SMSCSystemID   string `mapstructure:"SMSC_SYSTEM_ID"`
	SMSCPassword   string `mapstructure:"SMSC_PASSWORD"`
	SMSCSystemType string `mapstructure:"SMSC_SYSTEM_TYPE"`

	SourceAddr     string `mapstructure:"SOURCE_ADDR"`
	SourceTon      uint8  `mapstructure:"SOURCE_TON"`
	SourceNpi      uint8  `mapstructure:"SOURCE_NPI"`

	TLSEnabled    bool `mapstructure:"TLS_ENABLED"`
	TLSSkipVerify bool `mapstructure:"TLS_SKIP_VERIFY"`

	Encoding string `mapstructure:"ENCODING"`

	ServerListenAddr string `mapstructure:"SERVER_LISTEN_ADDR"`
	ServerQueueSize  int    `mapstructure:"SERVER_QUEUE_SIZE"`

	DatabaseDSN string `mapstructure:"DATABASE_DSN"`

	CacheAddr     string `mapstructure:"CACHE_ADDR"`
	CachePassword string `mapstructure:"CACHE_PASSWORD"`
	CacheDB       int    `mapstructure:"CACHE_DB"`

	JWTSecret string `mapstructure:"JWT_SECRET"`
}

func (c *Config) IsProd() bool { return c.AppEnv == "production" }
func (c *Config) IsDev() bool  { return c.AppEnv == "development" }
