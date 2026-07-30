package config

import (
	"os"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()
	v.AddConfigPath(".")
	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AutomaticEnv()

	v.BindEnv("DATABASE_DSN")
	v.BindEnv("ENQUIRE_LINK")
    v.BindEnv("READ_TIMEOUT")
	v.BindEnv("WRITE_TIMEOUT")
    v.BindEnv("POOL_SIZE")

    v.BindEnv("SMSC_ADDR")
    v.BindEnv("SMSC_SYSTEM_ID")
    v.BindEnv("SMSC_PASSWORD")
    v.BindEnv("SMSC_SYSTEM_TYPE")

    v.BindEnv("SOURCE_ADDR")
    v.BindEnv("SOURCE_TON")
    v.BindEnv("SOURCE_NPI")

    v.BindEnv("TLS_ENABLED")
    v.BindEnv("TLS_SKIP_VERIFY")

    v.BindEnv("ENCODING")

    v.BindEnv("SERVER_LISTEN_ADDR")
    v.BindEnv("SERVER_QUEUE_SIZE")

    v.BindEnv("DATABASE_DSN")

    v.BindEnv("CACHE_ADDR")
    v.BindEnv("CACHE_PASSWORD")
    v.BindEnv("CACHE_DB")

    v.BindEnv("JWT_SECRET")

	if err := v.BindEnv("DATABASE_DSN"); err != nil {
        return nil, err
    }
	if err := v.BindEnv("CACHE_ADDR"); err != nil {
        return nil, err
    }
    if err := v.BindEnv("CACHE_PASSWORD"); err != nil {
        return nil, err
    }
    if err := v.BindEnv("CACHE_DB"); err != nil {
        return nil, err
    }
	if err := v.BindEnv("ENQUIRE_LINK"); err != nil {
        return nil, err
    }

    if err := v.BindEnv("READ_TIMEOUT"); err != nil {
        return nil, err
    }

	if configPath != "" {
		v.AddConfigPath(configPath)
	}

	if err := v.ReadInConfig(); err != nil {
     // Ignore missing .env on production
    }

	var cfg Config
	if err := v.Unmarshal(&cfg, func(c *mapstructure.DecoderConfig) {
		c.DecodeHook = mapstructure.StringToBasicTypeHookFunc()
	}); err != nil {
		return nil, err
	}
// Allow Render (and other cloud platforms) to override the listen port.
    if port := os.Getenv("PORT"); port != "" {
        cfg.ServerListenAddr = ":" + port
   }  
	return &cfg, nil
}
