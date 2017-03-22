package config

import (
	"encoding/json"
	"io/ioutil"
	"strings"
)

// Environment types.
const (
	EnvTesting    = "testing"
	EnvLocal      = "local"
	EnvProduction = "production"
)

// Configuration contains all configuration parameters.
type Configuration struct {
	Hostname string `json:"hostname"`
	Port     uint32 `json:"port"`
	Database string `json:"database"`
	LogLevel string `json:"log_level"`
	Username string `json:"username"`
	Password string `json:"password"`

	Environment string `json:"environment"`
}

// IsEnvTesting Returns whether we are in this environment.
func (c *Configuration) IsEnvTesting() bool {
	return c.Environment == EnvTesting
}

// IsEnvLocal Returns whether we are in this environment.
func (c *Configuration) IsEnvLocal() bool {
	return c.Environment == EnvLocal
}

// IsEnvProduction Returns whether we are in this environment.
func (c *Configuration) IsEnvProduction() bool {
	return c.Environment == EnvProduction
}

// Config is the global configuration instance.
var Config = &Configuration{}

// NewConfigFromFile creates a new configuration from a JSON file.
func NewConfigFromFile(file string) (config *Configuration, err error) {
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}

	config = &Configuration{}
	err = json.Unmarshal(raw, config)
	if err != nil {
		return
	}

	err = cleanUp(config)
	if err != nil {
		return
	}

	return
}

func cleanUp(config *Configuration) error {
	config.Hostname = strings.TrimSpace(config.Hostname)

	if config.Port == 0 {
		config.Port = 8080
	}

	if len(config.Database) == 0 {
		config.Database = "file::memory:?mode=memory&cache=shared"
	}

	config.LogLevel = strings.ToLower(config.LogLevel)

	switch config.LogLevel {
	case "debug":
	case "info":
	case "warn":
	case "error":
	default:
		config.LogLevel = "info"
	}

	config.Environment = strings.ToLower(config.Environment)
	switch config.Environment {
	case EnvProduction:
	case EnvTesting:
	case EnvLocal:
	default:
		config.Environment = EnvLocal
	}

	if len(config.Username) == 0 {
		config.Username = "admin"
		config.Password = ""
	}

	return nil
}
