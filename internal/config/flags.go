package config

import (
	"errors"
	"flag"
	"net/url"
	"strconv"
	"strings"
)

type NetAddress struct {
	Host string
	Port int
}

func (na *NetAddress) String() string {
	return na.Host + ":" + strconv.Itoa(na.Port)
}

func (na *NetAddress) Set(s string) error {
	raw := strings.Split(s, ":")
	if len(raw) != 2 {
		return errors.New("incorrect net address. Should be host:port")
	}
	host := raw[0]
	port, err := strconv.Atoi(raw[1])
	if err != nil {
		return err
	}

	na.Host = host
	na.Port = port
	return nil
}

func parseURLAddr(s string) (string, error) {
	addr, err := url.ParseRequestURI(s)
	if err != nil {
		return "", err
	}

	if addr.Scheme == "" || addr.Host == "" {
		return "", errors.New("base url must include scheme and host")
	}

	return addr.String(), nil
}

func ParseFlags(cfg *Config) error {
	flag.Var(&cfg.Addr, "a", "Net address host:port")
	flag.Func("b", "Base URL for shortened links", func(s string) error {
		parsed, err := parseURLAddr(s)
		if err != nil {
			return err
		}
		cfg.URLAddr = parsed
		return nil
	})

	flag.Parse()
	return nil
}
