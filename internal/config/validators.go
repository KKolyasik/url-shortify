package config

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var fileNameRe = regexp.MustCompile(
    `^[a-zA-Z0-9_-]+\.(txt|log|json)$`,
)

func ParseNetAddress(s string) (NetAddress, error) {
	raw := strings.Split(s, ":")
	if len(raw) != 2 {
		return NetAddress{}, ErrInvalidNetAddress
	}

	host := strings.TrimSpace(raw[0])
	if host == "" {
		return NetAddress{}, ErrInvalidNetAddress
	}

	port, err := strconv.Atoi(raw[1])
	if err != nil || port <= 0 || port > 65535 {
		return NetAddress{}, ErrInvalidNetAddress
	}

	return NetAddress{Host: host, Port: port}, nil
}

func ParseBaseURL(s string) (string, error) {
	u, err := url.ParseRequestURI(strings.TrimSpace(s))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidBaseURL, err)
	}
	if u.Scheme == "" || u.Host == "" {
		return "", ErrInvalidBaseURL
	}
	return u.String(), nil
}

func ParseFileName(filename string) (string, error) {
    name := strings.ToLower(strings.TrimSpace(filename))
    if name == "" || !fileNameRe.MatchString(name) {
		return "", ErrInvalidFileName
	}
    return name, nil
}