package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// ThinkingParserType defines how thinking/reasoning is extracted from model responses.
type ThinkingParserType string

const (
	// ThinkingParserNone indicates no thinking extraction.
	ThinkingParserNone ThinkingParserType = "none"
	// ThinkingParserInBand indicates thinking is embedded in response text with delimiters.
	ThinkingParserInBand ThinkingParserType = "in_band"
	// ThinkingParserOutOfBand indicates thinking is in a separate API response field.
	ThinkingParserOutOfBand ThinkingParserType = "out_of_band"
)

// ThinkingParserConfig defines how to extract thinking from a model's response.
type ThinkingParserConfig struct {
	Type ThinkingParserType `toml:"type"`

	// For in_band parsers: delimiters that wrap thinking text
	StartDelimiter string `toml:"start_delimiter,omitempty"`
	EndDelimiter   string `toml:"end_delimiter,omitempty"`

	// For out_of_band parsers: JSONPath-like field path
	FieldPath string `toml:"field_path,omitempty"`
}

// Model represents a language model configuration.
type Model struct {
	Name         string  `toml:"name"`          // API model identifier (e.g., "claude-3-5-sonnet-20241022")
	Provider     string  `toml:"provider"`      // Reference to provider name from providers.toml
	ThinkingParser *ThinkingParserConfig `toml:"thinking_parser,omitempty"` // Optional: auto-detected if nil
}

// NewModel creates an empty Model configuration.
func NewModel() *Model {
	return &Model{}
}

// LoadModel creates and populates a Model from TOML data.
// It performs auto-detection of thinking parser configuration based on model name patterns.
func LoadModel(data []byte) (*Model, error) {
	m := NewModel()
	if err := toml.Unmarshal(data, m); err != nil {
		return nil, err
	}

	// Auto-detect thinking parser if not explicitly configured
	if m.ThinkingParser == nil {
		m.ThinkingParser = autoDetectThinkingParser(m.Name)
	}

	return m, nil
}

// LoadModelFromFile loads a model configuration from a file path.
func LoadModelFromFile(path string) (*Model, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadModel(data)
}

// LoadModelsFromDir loads all model configurations from a directory.
// Returns a map of model name (without .toml extension) -> Model.
func LoadModelsFromDir(dirPath string) (map[string]*Model, error) {
	models := make(map[string]*Model)

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read models directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".toml") {
			continue
		}

		modelPath := filepath.Join(dirPath, entry.Name())
		model, err := LoadModelFromFile(modelPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load model from %s: %w", entry.Name(), err)
		}

		// Use the model's configured name as the key
		if model.Name == "" {
			return nil, fmt.Errorf("model in %s has no name configured", entry.Name())
		}

		models[model.Name] = model
	}

	return models, nil
}

// autoDetectThinkingParser determines the appropriate thinking parser based on model name patterns.
func autoDetectThinkingParser(modelName string) *ThinkingParserConfig {
	lower := strings.ToLower(modelName)

	// Anthropic Claude models
	if strings.HasPrefix(lower, "claude-") {
		return &ThinkingParserConfig{
			Type:      ThinkingParserOutOfBand,
			FieldPath: "thinking",
		}
	}

	// OpenAI reasoning models (o1, o3 series)
	if strings.HasPrefix(lower, "o1-") || strings.HasPrefix(lower, "o3-") {
		return &ThinkingParserConfig{
			Type:      ThinkingParserOutOfBand,
			FieldPath: "reasoning.summary",
		}
	}

	// Qwen reasoning models (QwQ)
	if strings.Contains(lower, "qwq") || strings.HasPrefix(lower, "qwen") {
		return &ThinkingParserConfig{
			Type:           ThinkingParserInBand,
			StartDelimiter: "<think>",
			EndDelimiter:   "</think>",
		}
	}

	// DeepSeek reasoning models
	if strings.Contains(lower, "deepseek-r1") {
		return &ThinkingParserConfig{
			Type:           ThinkingParserInBand,
			StartDelimiter: "<think>",
			EndDelimiter:   "</think>",
		}
	}

	// Default: no thinking extraction
	return &ThinkingParserConfig{
		Type: ThinkingParserNone,
	}
}

// Validate checks if the model configuration is valid.
func (m *Model) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("model name is required")
	}
	if m.Provider == "" {
		return fmt.Errorf("model provider is required")
	}
	if m.ThinkingParser != nil {
		if err := m.ThinkingParser.Validate(); err != nil {
			return fmt.Errorf("invalid thinking parser config: %w", err)
		}
	}
	return nil
}

// Validate checks if the thinking parser configuration is valid.
func (t *ThinkingParserConfig) Validate() error {
	switch t.Type {
	case ThinkingParserNone:
		return nil
	case ThinkingParserInBand:
		if t.StartDelimiter == "" || t.EndDelimiter == "" {
			return fmt.Errorf("in_band parser requires both start_delimiter and end_delimiter")
		}
	case ThinkingParserOutOfBand:
		if t.FieldPath == "" {
			return fmt.Errorf("out_of_band parser requires field_path")
		}
	default:
		return fmt.Errorf("unknown parser type: %s", t.Type)
	}
	return nil
}
