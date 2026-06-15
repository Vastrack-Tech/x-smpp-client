package config

import (
	"fmt"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()
	v.SetEnvPrefix("SMPP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	if configPath != "" {
		v.AddConfigPath(configPath)
	}
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("/etc/x-smpp-client")

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg, func(c *mapstructure.DecoderConfig) {
		c.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
		)
	}); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.env", "development")
	v.SetDefault("app.enquire_link", "5s")
	v.SetDefault("app.read_timeout", "10s")
	v.SetDefault("app.write_timeout", "5s")
	v.SetDefault("app.pool_size", 1)

	v.SetDefault("smsc.addr", "smscsim.smpp.org:2775")
	v.SetDefault("smsc.system_id", "SYSTEMID")
	v.SetDefault("smsc.password", "PASSWORD")
	v.SetDefault("smsc.system_type", "")

	v.SetDefault("source_addr.address", "MelroseLabs")
	v.SetDefault("source_addr.ton", 5)
	v.SetDefault("source_addr.npi", 0)

	v.SetDefault("default_dest.address", "447712345678")
	v.SetDefault("default_dest.ton", 1)
	v.SetDefault("default_dest.npi", 1)

	v.SetDefault("tls.enabled", false)
	v.SetDefault("tls.skip_verify", false)

	v.SetDefault("encoding", "gsm")

	v.SetDefault("server.listen_addr", ":8080")
	v.SetDefault("server.queue_size", 1000)
}
