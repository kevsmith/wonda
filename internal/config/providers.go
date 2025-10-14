package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// validProviderName is the regex pattern for validating provider names.
// Provider names must start with an alphabetic character and contain only
// alphanumeric characters, dashes, and underscores.
var validProviderName = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)

// Provider represents a single LLM provider configuration.
type Provider struct {
	Name    string  `toml:"-"`
	BaseURL string  `toml:"base_url"` // Base URL for the provider's API endpoint
	APIKey  *string `toml:"api_key"`  // Optional: If nil, falls back to <PROVIDER_NAME>_API_KEY env var (uppercase, dashes/spaces → underscores)
}

// LoadFromEnvironment validates the provider name and loads the API key from
// environment variables if not already set in the configuration.
//
// Provider names must:
//   - Start with an alphabetic character (a-z, A-Z)
//   - Contain only alphanumeric characters, dashes (-), and underscores (_)
//
// Environment variable lookup:
//   - Only performed if APIKey is nil
//   - Name transformation: uppercase, dashes/spaces → underscores, append "_API_KEY"
//   - Example: "ollama-local" → OLLAMA_LOCAL_API_KEY
func (p *Provider) LoadFromEnvironment() error {
	// Validate provider name
	if !validProviderName.MatchString(p.Name) {
		return fmt.Errorf("invalid provider name '%s': must start with alphabetic character and contain only alphanumeric, dash, or underscore characters", p.Name)
	}

	// Only fetch from environment if APIKey is not already set
	if p.APIKey != nil {
		return nil
	}

	// Transform name to environment variable name
	// 1. Convert to uppercase
	// 2. Replace dashes with underscores
	// 3. Replace spaces with underscores
	// 4. Append _API_KEY suffix
	envName := strings.ToUpper(p.Name)
	envName = strings.ReplaceAll(envName, "-", "_")
	envName = strings.ReplaceAll(envName, " ", "_")
	envName = envName + "_API_KEY"

	// Fetch from environment
	if value := os.Getenv(envName); value != "" {
		p.APIKey = &value
	}

	return nil
}

// Providers represents the top-level providers configuration.
// Provider names from [providers.{name}] map to {NAME}_API_KEY environment variables.
//
// Transformation rules:
//  1. Convert to uppercase
//  2. Replace dashes (-) with underscores (_)
//  3. Replace spaces with underscores (_)
//
// Examples:
//   - [providers.anthropic] → ANTHROPIC_API_KEY
//   - [providers.ollama-local] → OLLAMA_LOCAL_API_KEY
//   - [providers."my provider"] → MY_PROVIDER_API_KEY
//
// Naming requirements:
//   - Must start with an alphabetic character (a-z, A-Z)
//   - Can contain alphanumeric characters, dashes, and underscores
type Providers struct {
	Providers map[string]*Provider `toml:"providers"`
}

// NewProviders creates an empty Providers configuration.
func NewProviders() *Providers {
	return &Providers{
		Providers: make(map[string]*Provider),
	}
}

// LoadProviders creates and populates a Providers configuration from TOML.
func LoadProviders(data []byte) (*Providers, error) {
	p := NewProviders()
	if err := toml.Unmarshal(data, p); err != nil {
		return nil, err
	}
	for name, provider := range p.Providers {
		provider.Name = name
		if err := provider.LoadFromEnvironment(); err != nil {
			return nil, err
		}
	}
	return p, nil
}

// LoadProvidersFromFile loads providers configuration from a file path.
func LoadProvidersFromFile(path string) (*Providers, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadProviders(data)
}
