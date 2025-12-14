package config

import "time"

// Config holds the configuration for the recon tool
type Config struct {
	Target         string
	OutputDir      string
	WordlistPath   string
	Timeout        time.Duration
	EnableModules  map[string]bool
}

// NewConfig creates a new configuration with defaults
func NewConfig(target string) *Config {
	return &Config{
		Target:        target,
		OutputDir:     "./output",
		WordlistPath:  "/root/myLists/all.txt",
		Timeout:       5 * time.Hour, // 5 hours timeout
		EnableModules: map[string]bool{
			"content_discovery": true,
		},
	}
}

// ModuleConfig holds configuration specific to Content Discovery
type ContentDiscoveryConfig struct {
	StatusCodes []string
	OutputFile  string
}

// NewContentDiscoveryConfig returns default content discovery config
func NewContentDiscoveryConfig(outputDir string) *ContentDiscoveryConfig {
	return &ContentDiscoveryConfig{
		StatusCodes: []string{"200", "403", "301", "302", "307"},
		OutputFile:  outputDir + "/dirs.txt",
	}
}
