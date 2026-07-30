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

	if configPath != "" {
		v.AddConfigPath(configPath)
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, err
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
