package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the agent configuration
type Config struct {
	Agent   AgentConfig   `yaml:"agent"`
	Nexus   NexusConfig   `yaml:"nexus"`
	Limits  LimitsConfig  `yaml:"limits"`
	Logging LoggingConfig `yaml:"logging"`
}

// AgentConfig represents agent identity
type AgentConfig struct {
	ID    string `yaml:"id"`
	Name  string `yaml:"name"`
	Token string `yaml:"token"`
}

// NexusConfig represents Nexus Core connection settings
type NexusConfig struct {
	CoreURL           string        `yaml:"core_url"`
	ReconnectInterval time.Duration `yaml:"reconnect_interval"`
	HeartbeatInterval time.Duration `yaml:"heartbeat_interval"`
}

// LimitsConfig represents query limits
type LimitsConfig struct {
	MaxRows              int           `yaml:"max_rows"`
	QueryTimeout         time.Duration `yaml:"query_timeout"`
	MaxConcurrentQueries int           `yaml:"max_concurrent_queries"`
}

// LoggingConfig represents logging settings
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// Load reads configuration from a YAML file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Set defaults
	if cfg.Limits.MaxRows == 0 {
		cfg.Limits.MaxRows = 100000
	}
	if cfg.Limits.QueryTimeout == 0 {
		cfg.Limits.QueryTimeout = 10 * time.Minute
	}
	if cfg.Limits.MaxConcurrentQueries == 0 {
		cfg.Limits.MaxConcurrentQueries = 10
	}
	if cfg.Nexus.ReconnectInterval == 0 {
		cfg.Nexus.ReconnectInterval = 5 * time.Second
	}
	if cfg.Nexus.HeartbeatInterval == 0 {
		cfg.Nexus.HeartbeatInterval = 30 * time.Second
	}

	return &cfg, nil
}
