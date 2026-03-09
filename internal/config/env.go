package config

import (
	"crypto/rand"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
)

type Settings struct {
	ServerAddr        string        `env:"SERVER_ADDRESS"`
	BaseURL           string        `env:"BASE_URL"`
	FileStorage       string        `env:"FILE_STORAGE_PATH"`
	ReadHeaderTimeout time.Duration `env:"READ_HEADER_TIMEOUT"`
	ReadTimeout       time.Duration `env:"READ_TIMEOUT"`
	WriteTimeout      time.Duration `env:"WRITE_TIMEOUT"`
	IdleTimeout       time.Duration `env:"IDLE_TIMEOUT"`
	DBURL             string        `env:"DATABASE_DSN"`
	SecretKey         string        `env:"SECRET_KEY"`
	TokenTTL          time.Duration `env:"TOKEN_TTL"`
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

	if _, ok := os.LookupEnv("READ_HEADER_TIMEOUT"); ok {
		cfg.ReadHeaderTimeout = s.ReadHeaderTimeout
	}

	if _, ok := os.LookupEnv("READ_TIMEOUT"); ok {
		cfg.ReadTimeout = s.ReadTimeout
	}

	if _, ok := os.LookupEnv("WRITE_TIMEOUT"); ok {
		cfg.WriteTimeout = s.WriteTimeout
	}

	if _, ok := os.LookupEnv("IDLE_TIMEOUT"); ok {
		cfg.IdleTimeout = s.IdleTimeout
	}

	if _, ok := os.LookupEnv("DATABASE_DSN"); ok {
		cfg.DBURL = s.DBURL
	}

	if _, ok := os.LookupEnv("SECRET_KEY"); ok {
		cfg.SecretKey = s.SecretKey
	} else {
		secretKey, err := genereteSecretKey(16)
		if err != nil {
			return err
		}
		cfg.SecretKey = secretKey
	}

	if _, ok := os.LookupEnv("TOKEN_TTL"); ok {
		cfg.TokenTTL = s.TokenTTL
	}

	return nil
}

func genereteSecretKey(size int) (string, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
