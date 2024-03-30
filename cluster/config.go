package cluster

type Config struct {
	Name    string
	Address string
	Port    int
}

func Configure(name, addr string, port int) *Config {
	return &Config{
		Name:    name,
		Address: addr,
		Port:    port,
	}
}
