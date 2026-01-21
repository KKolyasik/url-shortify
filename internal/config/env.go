package config

import (
	"errors"
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
		if err := cfg.Addr.Set(ua.ServerAddr); err != nil {
			return errors.New("incorrect net address. Should be host:port")
		}
	}
	if _, ok := os.LookupEnv("BASE_URL"); ok {
		baseURL, err := parseURLAddr(ua.BaseURL)
		if err != nil {
			return errors.New("base url must include scheme and host")
		}
		cfg.URLAddr = baseURL
	}

	return nil
}
