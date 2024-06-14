package server

type Config struct {
	Addr string
}

func NewServerConfig(addr string) *Config {
	return &Config{
		Addr: addr,
	}
}
