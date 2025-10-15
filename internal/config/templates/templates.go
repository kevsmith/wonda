package templates

import (
	"embed"
	"fmt"
)

// FS contains all template files embedded at build time.
// All .toml files are embedded, including templates and starter configs.
//
//go:embed *.toml
var FS embed.FS

// GetTemplate retrieves a template by name.
// The name should not include the "_template.toml" suffix - it will be added automatically.
//
// Example:
//
//	content, err := templates.GetTemplate("scenario")  // reads scenario_template.toml
func GetTemplate(name string) (string, error) {
	filename := name + "_template.toml"
	content, err := FS.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read template '%s': %w", name, err)
	}
	return string(content), nil
}
