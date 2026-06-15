package config

import "time"

type Config struct {
	App         AppConfig
	SMSC        SMSCConfig
	SourceAddr  AddressConfig
	DefaultDest AddressConfig
	TLS         TLSConfig
}

type AppConfig struct {
	Env          string
	EnquireLink  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type SMSCConfig struct {
	Addr       string
	SystemID   string
	Password   string
	SystemType string
}

type AddressConfig struct {
	Address string
	Ton     uint8
	Npi     uint8
}

type TLSConfig struct {
	Enabled    bool
	SkipVerify bool
}

func (c *Config) IsProd() bool { return c.App.Env == "production" }
func (c *Config) IsDev() bool  { return c.App.Env == "development" }
