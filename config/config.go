package config

import (
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Overall policy for a path
type Policy struct {
	Path string

	// If configured, only the allowed time range will be permitted to access this website
	// Otherwise, it will be always allowed unless reaches the MaxAllowed duration
	AllowedRange []TimeRange

	// Max duration allowed for this website.
	MaxAllowed time.Duration
}

type LogsConfig struct {
	Provider string
	Config   map[string]string
	// Don't log these garbage hosts/path (or host + / + path)
	SkipLogging []string `yaml:"skip-logging"`
}

type Config struct {
	Policies []Policy
	Logs     LogsConfig
	// Hosts that have pinned certificates, e.g., icloud
	SkipProxy []string `yaml:"skip-proxy"`
	// Compeletely blocked sites
	Blocked []string
}

func NewConfig() *Config {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatal("Unable to open config file config.yaml")
	}
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Unable to parse config: %s", err)
	}
	log.Printf("%v", config.SkipProxy)
	return &config
}
