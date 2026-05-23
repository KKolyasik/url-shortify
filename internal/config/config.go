package config

import (
	"time"
)

type Config struct {
	Addr              NetAddress
	URLAddr           string
	FileStorage       string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	DBURL             string
	Audit             AuditConfig
}

type AuditConfig struct {
	AuditFile string
	AuditURL  string
}

func NewConfig() Config {
	return Config{
		Addr: NetAddress{
			Host: "localhost",
			Port: 8080,
		},
		URLAddr:           "http://localhost:8080",
		FileStorage:       "",
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
}

func Load() (Config, error) {
	cfg := NewConfig()

	err := ParseFlags(&cfg)
	if err != nil {
		return Config{}, err
	}

	err = ParseEnv(&cfg)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}
