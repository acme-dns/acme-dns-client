package client

import "github.com/cpu/goacmedns"

type AcmednsClient struct {
	Config *Config
	Storage goacmedns.Storage
}

type Config struct {
	Verbose bool
	Debug bool
	Domain string
	Server string
	AllowList string
	DNSServer string
	Dangerous bool
}

func NewAcmednsConfig() *Config {
	return &Config{
		Verbose: false,
		Debug: false,
		Domain: "",
		Server: "",
		AllowList: "",
	}
}

func NewAcmednsClient(storagepath string) *AcmednsClient {
	return &AcmednsClient{
		Config: NewAcmednsConfig(),
		Storage: goacmedns.NewFileStorage(storagepath, 0600),
	}
}

func (c *AcmednsClient) Debug(input string) {
	if c.Config.Debug {
		PrintDebug(input, 0)
	}
}

func (c *AcmednsClient) Verbose(input string) {
	if c.Config.Verbose || c.Config.Debug {
		PrintDebug(input, 0)
	}
}