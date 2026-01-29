package config

import (
	"os"

	"github.com/caarlos0/env/v6"
)

type Settings struct {
	ServerAddr  string `env:"SERVER_ADDRESS"`
	BaseURL     string `env:"BASE_URL"`
	FileStorage string `env:"FILE_STORAGE_PATH"`
}

func ParseEnv(cfg *Config) error {
	var s Settings
	if err := env.Parse(&s); err != nil {
		return err
	}

	if _, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		parsed, err := ParseNetAddress(s.ServerAddr)
		if err != nil {
			return err
		}
		cfg.Addr = parsed
	}
	if _, ok := os.LookupEnv("BASE_URL"); ok {
		parsed, err := ParseBaseURL(s.BaseURL)
		if err != nil {
			return err
		}
		cfg.URLAddr = parsed
	}

	if _, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		parsed, err := ParseFileName(s.FileStorage)
		if err != nil {
			return err
		}
		cfg.FileStorage = parsed
	}

	return nil
}
