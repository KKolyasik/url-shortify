package config

import (
	"flag"
	"strconv"
)

type NetAddress struct {
	Host string
	Port int
}

func (na *NetAddress) String() string {
	return na.Host + ":" + strconv.Itoa(na.Port)
}

func (na *NetAddress) Set(s string) error {
	parsed, err := ParseNetAddress(s)
	if err != nil {
		return err
	}
	*na = parsed
	return nil
}

func ParseFlags(cfg *Config) error {
	flag.Var(&cfg.Addr, "a", "Net address host:port")
	flag.Func("b", "Base URL for shortened links", func(s string) error {
		parsed, err := ParseBaseURL(s)
		if err != nil {
			return err
		}
		cfg.URLAddr = parsed
		return nil
	})
	flag.Func("f", "File storage path", func(s string) error {
		filename, err := ParseFileName(s)
		if err != nil {
			return err
		}
		cfg.FileStorage = filename
		return nil
	})

	flag.Parse()
	return nil
}
