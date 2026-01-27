package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

type GatewayConfig struct {
	ServiceName string `yaml:"service_name"`
	Listener    struct {
		Address string `yaml:"address"`
		TLS     struct {
			Enabled  bool   `yaml:"enabled"`
			CertFile string `yaml:"cert_file"`
			KeyFile  string `yaml:"key_file"`
		} `yaml:"tls"`
		Timeouts struct {
			ReadHeaderSeconds int `yaml:"read_header_seconds"`
			ReadSeconds       int `yaml:"read_seconds"`
			WriteSeconds      int `yaml:"write_seconds"`
		} `yaml:"timeouts"`
	} `yaml:"listener"`
	Routes []RoutePolicy `yaml:"routes"`
	Riskd  struct {
		Endpoint  string `yaml:"endpoint"`
		TimeoutMs int    `yaml:"timeout_ms"`
		Salt      string `yaml:"salt"`
	} `yaml:"riskd"`
	Redis struct {
		Addr string `yaml:"addr"`
	} `yaml:"redis"`
	JWT struct {
		PublicKey string `yaml:"public_key"`
		Secret    string `yaml:"secret"`
	} `yaml:"jwt"`
	BlockedResponse struct {
		Message           string `yaml:"message"`
		RetryAfterSeconds int    `yaml:"retry_after_seconds"`
	} `yaml:"blocked_response"`
	TLSUpstream struct {
		Enabled  bool   `yaml:"enabled"`
		CAFile   string `yaml:"ca_file"`
		CertFile string `yaml:"cert_file"`
		KeyFile  string `yaml:"key_file"`
	} `yaml:"tls_upstream"`
	Observability ObservabilityConfig `yaml:"observability"`
}

type RiskdConfig struct {
	ServiceName string `yaml:"service_name"`
	Address     string `yaml:"address"`
	Redis       struct {
		Addr string `yaml:"addr"`
	} `yaml:"redis"`
	Thresholds struct {
		Block     int `yaml:"block"`
		Challenge int `yaml:"challenge"`
		Monitor   int `yaml:"monitor"`
	} `yaml:"thresholds"`
	Observability ObservabilityConfig `yaml:"observability"`
}

type SocialConfig struct {
	ServiceName string `yaml:"service_name"`
	Address     string `yaml:"address"`
	TLS         struct {
		Enabled  bool   `yaml:"enabled"`
		CertFile string `yaml:"cert_file"`
		KeyFile  string `yaml:"key_file"`
		CAFile   string `yaml:"ca_file"`
	} `yaml:"tls"`
	Database    struct {
		DSN string `yaml:"dsn"`
	} `yaml:"database"`
	Redis struct {
		Addr string `yaml:"addr"`
	} `yaml:"redis"`
	JWT struct {
		Secret string `yaml:"secret"`
	} `yaml:"jwt"`
	CSRF struct {
		Secret string `yaml:"secret"`
	} `yaml:"csrf"`
	Observability ObservabilityConfig `yaml:"observability"`
}

type ObservabilityConfig struct {
	OTLPEndpoint string `yaml:"otel_endpoint"`
	MetricsAddr  string `yaml:"metrics_address"`
}

type RoutePolicy struct {
	ID                 string          `yaml:"id"`
	Host               string          `yaml:"host"`
	PathPrefix         string          `yaml:"path_prefix"`
	UpstreamURL        string          `yaml:"upstream_url"`
	AllowedMethods     []string        `yaml:"allowed_methods"`
	AuthRequired       bool            `yaml:"auth_required"`
	RolesAllowed       []string        `yaml:"roles_allowed"`
	MaxBodyBytes       int64           `yaml:"max_body_bytes"`
	AllowedContentType []string        `yaml:"allowed_content_types"`
	RateLimits         RateLimitConfig `yaml:"rate_limits"`
	Risk               RiskConfig      `yaml:"risk"`
}

type RateLimitConfig struct {
	PerIP    Limit `yaml:"per_ip"`
	PerUser  Limit `yaml:"per_user"`
	PerRoute Limit `yaml:"per_route"`
}

type Limit struct {
	RPS   int `yaml:"rps"`
	Burst int `yaml:"burst"`
}

type RiskConfig struct {
	Enabled     bool           `yaml:"enabled"`
	Thresholds  RiskThresholds `yaml:"thresholds"`
	ActionMap   RiskActions    `yaml:"action_map"`
	AuthHint    bool           `yaml:"auth_hint"`
	MaxBodyHint int64          `yaml:"max_body_hint"`
}

type RiskThresholds struct {
	Block     int `yaml:"block"`
	Challenge int `yaml:"challenge"`
	Monitor   int `yaml:"monitor"`
}

type RiskActions struct {
	Block     string `yaml:"block"`
	Challenge string `yaml:"challenge"`
	Monitor   string `yaml:"monitor"`
}

func LoadGatewayConfig(path string) (GatewayConfig, error) {
	var cfg GatewayConfig
	if err := loadYAML(path, &cfg); err != nil {
		return cfg, err
	}
	if env := os.Getenv("GATEWAY_LISTEN_ADDR"); env != "" {
		cfg.Listener.Address = env
	}
	if env := os.Getenv("REDIS_ADDR"); env != "" {
		cfg.Redis.Addr = env
	}
	if env := os.Getenv("RISKD_ENDPOINT"); env != "" {
		cfg.Riskd.Endpoint = env
	}
	return cfg, nil
}

func LoadRiskdConfig(path string) (RiskdConfig, error) {
	var cfg RiskdConfig
	if err := loadYAML(path, &cfg); err != nil {
		return cfg, err
	}
	if env := os.Getenv("RISKD_LISTEN_ADDR"); env != "" {
		cfg.Address = env
	}
	if env := os.Getenv("REDIS_ADDR"); env != "" {
		cfg.Redis.Addr = env
	}
	return cfg, nil
}

func LoadSocialConfig(path string) (SocialConfig, error) {
	var cfg SocialConfig
	if err := loadYAML(path, &cfg); err != nil {
		return cfg, err
	}
	if env := os.Getenv("SOCIALD_LISTEN_ADDR"); env != "" {
		cfg.Address = env
	}
	if env := os.Getenv("POSTGRES_DSN"); env != "" {
		cfg.Database.DSN = env
	}
	if env := os.Getenv("REDIS_ADDR"); env != "" {
		cfg.Redis.Addr = env
	}
	return cfg, nil
}

func loadYAML(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return errors.New("empty config file")
	}
	return yaml.Unmarshal(data, target)
}
