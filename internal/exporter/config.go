package exporter

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Web struct {
		ListenAddress string `yaml:"listen_address"`
	} `yaml:"web"`
	Redfish struct {
		Host        string `yaml:"host"`
		Username    string `yaml:"username"`
		Password    string `yaml:"password"`
		InsecureTLS bool   `yaml:"insecure_tls"`
		ChassisID   string `yaml:"chassis_id"`
		TimeoutSec  int    `yaml:"timeout_sec"`
	} `yaml:"redfish"`
}

func LoadConfig(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
