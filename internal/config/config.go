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