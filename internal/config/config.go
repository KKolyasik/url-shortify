package config

type Config struct {
	Addr    NetAddress
	URLAddr string
}

func NewConfig() Config {
	return Config{
		Addr: NetAddress{
			Host: "localhost",
			Port: 8080,
		},
		URLAddr: "http://localhost:8080",
	}
}