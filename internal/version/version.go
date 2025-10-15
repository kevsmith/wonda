package version

import "fmt"

// These variables are set at build time via -ldflags
var (
	Version   = "dev"     // Version from git tag
	Commit    = "unknown" // Git commit SHA
	BuildTime = "unknown" // Build timestamp
)

// Info returns a formatted version string with all build information
func Info() string {
	return fmt.Sprintf("wonda %s (commit: %s, built: %s)", Version, Commit, BuildTime)
}

// Short returns just the version number
func Short() string {
	return Version
}
