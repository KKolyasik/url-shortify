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

var Addr NetAddress = NetAddress{
	Host: "localhost",
	Port: 8080,
}
var URLAddr string = "http://localhost:8080"

func parseURLAddr(s string) error {
	addr, err := url.ParseRequestURI(s)
	if err != nil {
		return err
	}

	if addr.Scheme == "" || addr.Host == "" {
		return errors.New("base url must include scheme and host")
	}

	URLAddr = addr.String()
	return nil
}

func AddFlags() {
	flag.Var(&Addr, "a", "Net address host:port")
	flag.Func("b", "Base URL for shortened links, including scheme (e.g. http://localhost:8080)", parseURLAddr)
	flag.Parse()
}
