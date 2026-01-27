package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func FuzzConfigParsing(f *testing.F) {
	f.Add([]byte("service_name: gateway"))
	f.Fuzz(func(t *testing.T, data []byte) {
		var cfg GatewayConfig
		_ = loadYAMLBytes(data, &cfg)
	})
}

func loadYAMLBytes(data []byte, target any) error {
	return yaml.Unmarshal(data, target)
}
