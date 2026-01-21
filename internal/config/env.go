package config

import (
	"os"

	"github.com/caarlos0/env/v6"
)

type URLAddresses struct {
	ServerAddr string `env:"SERVER_ADDRESS"`
	BaseURL    string `env:"BASE_URL"`
}

func ParseEnv(cfg *Config) error {
	var ua URLAddresses
	if err := env.Parse(&ua); err != nil {
		return err
	}

	if _, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		parsed, err := ParseNetAddress(ua.ServerAddr)
		if err != nil {
			return err
		}
		cfg.Addr = parsed
	}
	if _, ok := os.LookupEnv("BASE_URL"); ok {
		parsed, err := ParseBaseURL(ua.BaseURL)
		if err != nil {
			return err
		}
		cfg.URLAddr = parsed
	}

	return nil
}
