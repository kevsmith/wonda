package config

import "fmt"

// ConfigVersion is the current version for all configuration files.
// This should match the version field in TOML templates.
const ConfigVersion = "1.0.0"

// ValidateVersion checks if the provided version matches the expected config version.
// Returns an error if versions don't match.
func ValidateVersion(fileType string, version string) error {
	if version == "" {
		return fmt.Errorf("%s missing version field (expected version %s)", fileType, ConfigVersion)
	}
	if version != ConfigVersion {
		return fmt.Errorf("%s version mismatch: got %s, expected %s", fileType, version, ConfigVersion)
	}
	return nil
}
