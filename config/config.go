package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// YAMLConfig represents the structure of config.yaml
type YAMLConfig struct {
	General struct {
		OutputDir string `yaml:"output_dir"`
		Timeout   string `yaml:"timeout"`
	} `yaml:"general"`

	Modules struct {
		Feroxbuster bool `yaml:"feroxbuster"`
		Dirsearch   bool `yaml:"dirsearch"`
		Crawling    bool `yaml:"crawling"`
	} `yaml:"modules"`

	Feroxbuster struct {
		Wordlist       string   `yaml:"wordlist"`
		StatusCodes    []int    `yaml:"status_codes"`
		OutputFile     string   `yaml:"output_file"`
		Threads        int      `yaml:"threads"`
		RecursionDepth int      `yaml:"recursion_depth"`
		SkipSSL        bool     `yaml:"skip_ssl"`
		NoState        bool     `yaml:"no_state"`
	} `yaml:"feroxbuster"`

	Dirsearch struct {
		Wordlist    string   `yaml:"wordlist"`
		Extensions  []string `yaml:"extensions"`
		StatusCodes []int    `yaml:"status_codes"`
		OutputFile  string   `yaml:"output_file"`
		FullURL     bool     `yaml:"full_url"`
		Recursive   bool     `yaml:"recursive"`
		Threads     int      `yaml:"threads"`
	} `yaml:"dirsearch"`

	Crawling struct {
		OutputFile string `yaml:"output_file"`
		Katana     struct {
			Depth       int `yaml:"depth"`
			Concurrency int `yaml:"concurrency"`
			RateLimit   int `yaml:"rate_limit"`
		} `yaml:"katana"`
	} `yaml:"crawling"`
}

// Config holds the runtime configuration for the recon tool
type Config struct {
	Target         string
	OutputDir      string
	Timeout        time.Duration
	EnableModules  map[string]bool
	YAMLConfig     *YAMLConfig
	CrawlingConfig *CrawlingConfig
}

// extractSubdomain extracts the subdomain/hostname from a URL and includes port if non-standard
func extractSubdomain(targetURL string) (string, error) {
	// Add scheme if missing
	if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
		targetURL = "https://" + targetURL
	}

	// Parse URL
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %v", err)
	}

	// Get hostname (subdomain.domain.com or IP)
	hostname := parsedURL.Hostname()
	if hostname == "" {
		return "", fmt.Errorf("could not extract hostname from URL")
	}

	// Get port
	port := parsedURL.Port()

	// Include port in directory name if it's non-standard
	// Standard ports: 80 for http, 443 for https
	if port != "" {
		scheme := parsedURL.Scheme
		isStandardPort := (scheme == "http" && port == "80") || (scheme == "https" && port == "443")

		if !isStandardPort {
			// Replace colon with underscore to avoid filesystem issues
			hostname = hostname + "_" + port
		}
	}

	return hostname, nil
}

// LoadConfig loads configuration from config.yaml and merges with target
func LoadConfig(target string, configPath string, outputDirOverride string) (*Config, error) {
	// Read YAML file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	// Parse YAML
	var yamlConfig YAMLConfig
	if err := yaml.Unmarshal(data, &yamlConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	// Parse timeout
	timeout, err := time.ParseDuration(yamlConfig.General.Timeout)
	if err != nil {
		return nil, fmt.Errorf("invalid timeout format '%s': %v", yamlConfig.General.Timeout, err)
	}

	// Determine output directory:
	//   1. CLI override (-o flag) takes precedence
	//   2. Fall back to the default single output directory
	outputDir := "./fuzzing-output"
	if outputDirOverride != "" {
		outputDir = outputDirOverride
	}

	// Create config
	cfg := &Config{
		Target:     target,
		OutputDir:  outputDir,
		Timeout:    timeout,
		YAMLConfig: &yamlConfig,
		EnableModules: map[string]bool{
			"feroxbuster": yamlConfig.Modules.Feroxbuster,
			"dirsearch":   yamlConfig.Modules.Dirsearch,
			"crawling":    yamlConfig.Modules.Crawling,
		},
	}

	return cfg, nil
}

// FeroxbusterConfig holds configuration specific to Feroxbuster
type FeroxbusterConfig struct {
	Wordlist       string
	StatusCodes    []string
	OutputFile     string
	Threads        int
	RecursionDepth int
	SkipSSL        bool
	NoState        bool
}

// NewFeroxbusterConfig returns feroxbuster config from YAML
func (c *Config) NewFeroxbusterConfig() *FeroxbusterConfig {
	// Convert int status codes to strings
	statusCodes := make([]string, len(c.YAMLConfig.Feroxbuster.StatusCodes))
	for i, code := range c.YAMLConfig.Feroxbuster.StatusCodes {
		statusCodes[i] = fmt.Sprintf("%d", code)
	}

	return &FeroxbusterConfig{
		Wordlist:       c.YAMLConfig.Feroxbuster.Wordlist,
		StatusCodes:    statusCodes,
		OutputFile:     c.OutputDir + "/" + c.YAMLConfig.Feroxbuster.OutputFile,
		Threads:        c.YAMLConfig.Feroxbuster.Threads,
		RecursionDepth: c.YAMLConfig.Feroxbuster.RecursionDepth,
		SkipSSL:        c.YAMLConfig.Feroxbuster.SkipSSL,
		NoState:        c.YAMLConfig.Feroxbuster.NoState,
	}
}

// DirsearchConfig holds configuration specific to Dirsearch
type DirsearchConfig struct {
	Wordlist    string
	Extensions  []string
	StatusCodes []string
	OutputFile  string
	FullURL     bool
	Recursive   bool
	Threads     int
}

// NewDirsearchConfig returns dirsearch config from YAML
func (c *Config) NewDirsearchConfig() *DirsearchConfig {
	// Convert int status codes to strings
	statusCodes := make([]string, len(c.YAMLConfig.Dirsearch.StatusCodes))
	for i, code := range c.YAMLConfig.Dirsearch.StatusCodes {
		statusCodes[i] = fmt.Sprintf("%d", code)
	}

	return &DirsearchConfig{
		Wordlist:    c.YAMLConfig.Dirsearch.Wordlist,
		Extensions:  c.YAMLConfig.Dirsearch.Extensions,
		StatusCodes: statusCodes,
		OutputFile:  c.OutputDir + "/" + c.YAMLConfig.Dirsearch.OutputFile,
		FullURL:     c.YAMLConfig.Dirsearch.FullURL,
		Recursive:   c.YAMLConfig.Dirsearch.Recursive,
		Threads:     c.YAMLConfig.Dirsearch.Threads,
	}
}

// CrawlingConfig holds configuration specific to web crawling
type CrawlingConfig struct {
	OutputFile string
	Katana     struct {
		Depth       int
		Concurrency int
		RateLimit   int
	}
}

// NewCrawlingConfig returns crawling config from YAML
func (c *Config) NewCrawlingConfig() *CrawlingConfig {
	return &CrawlingConfig{
		OutputFile: c.OutputDir + "/" + c.YAMLConfig.Crawling.OutputFile,
		Katana: struct {
			Depth       int
			Concurrency int
			RateLimit   int
		}{
			Depth:       c.YAMLConfig.Crawling.Katana.Depth,
			Concurrency: c.YAMLConfig.Crawling.Katana.Concurrency,
			RateLimit:   c.YAMLConfig.Crawling.Katana.RateLimit,
		},
	}
}
