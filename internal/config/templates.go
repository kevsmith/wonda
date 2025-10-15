package config

import (
	"github.com/poiesic/wonda/internal/config/templates"
)

// GetTemplate retrieves a configuration template by name.
// The name should not include the "_template.toml" suffix - it will be added automatically.
//
// Available templates:
//   - "scenario" - Scenario definition template
//   - "character" - Character definition template
//   - "model" - Model configuration template
//   - "embeddings" - Embeddings configuration template
//
// Example:
//   content, err := config.GetTemplate("scenario")  // reads scenario_template.toml
func GetTemplate(name string) (string, error) {
	return templates.GetTemplate(name)
}
