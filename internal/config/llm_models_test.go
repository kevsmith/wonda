package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadModel(t *testing.T) {
	t.Run("loads minimal model configuration", func(t *testing.T) {
		tomlData := `
name = "claude-3-5-sonnet-20241022"
provider = "anthropic"
`
		model, err := LoadModel([]byte(tomlData))
		require.NoError(t, err)
		assert.Equal(t, "claude-3-5-sonnet-20241022", model.Name)
		assert.Equal(t, "anthropic", model.Provider)
		assert.NotNil(t, model.ThinkingParser)
	})

	t.Run("auto-detects Anthropic Claude thinking parser", func(t *testing.T) {
		tomlData := `
name = "claude-3-5-sonnet-20241022"
provider = "anthropic"
`
		model, err := LoadModel([]byte(tomlData))
		require.NoError(t, err)
		assert.Equal(t, ThinkingParserOutOfBand, model.ThinkingParser.Type)
		assert.Equal(t, "thinking", model.ThinkingParser.FieldPath)
	})

	t.Run("auto-detects OpenAI o1 thinking parser", func(t *testing.T) {
		tomlData := `
name = "o1-preview"
provider = "openai"
`
		model, err := LoadModel([]byte(tomlData))
		require.NoError(t, err)
		assert.Equal(t, ThinkingParserOutOfBand, model.ThinkingParser.Type)
		assert.Equal(t, "reasoning.summary", model.ThinkingParser.FieldPath)
	})

	t.Run("auto-detects OpenAI o3 thinking parser", func(t *testing.T) {
		tomlData := `
name = "o3-mini"
provider = "openai"
`
		model, err := LoadModel([]byte(tomlData))
		require.NoError(t, err)
		assert.Equal(t, ThinkingParserOutOfBand, model.ThinkingParser.Type)
		assert.Equal(t, "reasoning.summary", model.ThinkingParser.FieldPath)
	})

	t.Run("auto-detects Qwen QwQ thinking parser", func(t *testing.T) {
		tomlData := `
name = "qwq-32b-preview"
provider = "ollama"
`
		model, err := LoadModel([]byte(tomlData))
		require.NoError(t, err)
		assert.Equal(t, ThinkingParserInBand, model.ThinkingParser.Type)
		assert.Equal(t, "<think>", model.ThinkingParser.StartDelimiter)
		assert.Equal(t, "</think>", model.ThinkingParser.EndDelimiter)
	})

	t.Run("auto-detects DeepSeek R1 thinking parser", func(t *testing.T) {
		tomlData := `
name = "deepseek-r1-distill-qwen-32b"
provider = "ollama"
`
		model, err := LoadModel([]byte(tomlData))
		require.NoError(t, err)
		assert.Equal(t, ThinkingParserInBand, model.ThinkingParser.Type)
		assert.Equal(t, "<think>", model.ThinkingParser.StartDelimiter)
		assert.Equal(t, "</think>", model.ThinkingParser.EndDelimiter)
	})

	t.Run("defaults to no thinking parser for unknown models", func(t *testing.T) {
		tomlData := `
name = "gpt-4-turbo"
provider = "openai"
`
		model, err := LoadModel([]byte(tomlData))
		require.NoError(t, err)
		assert.Equal(t, ThinkingParserNone, model.ThinkingParser.Type)
	})

	t.Run("respects explicit thinking parser configuration", func(t *testing.T) {
		tomlData := `
name = "custom-model"
provider = "ollama"

[thinking_parser]
type = "in_band"
start_delimiter = "<reasoning>"
end_delimiter = "</reasoning>"
`
		model, err := LoadModel([]byte(tomlData))
		require.NoError(t, err)
		assert.Equal(t, ThinkingParserInBand, model.ThinkingParser.Type)
		assert.Equal(t, "<reasoning>", model.ThinkingParser.StartDelimiter)
		assert.Equal(t, "</reasoning>", model.ThinkingParser.EndDelimiter)
	})

	t.Run("explicit config overrides auto-detection", func(t *testing.T) {
		tomlData := `
name = "claude-3-5-sonnet-20241022"
provider = "anthropic"

[thinking_parser]
type = "none"
`
		model, err := LoadModel([]byte(tomlData))
		require.NoError(t, err)
		assert.Equal(t, ThinkingParserNone, model.ThinkingParser.Type)
	})

	t.Run("returns error for invalid TOML", func(t *testing.T) {
		tomlData := `
name = "invalid
provider = "test"
`
		_, err := LoadModel([]byte(tomlData))
		assert.Error(t, err)
	})
}

func TestModelValidate(t *testing.T) {
	t.Run("validates minimal valid model", func(t *testing.T) {
		model := &Model{
			Name:     "test-model",
			Provider: "test-provider",
			ThinkingParser: &ThinkingParserConfig{
				Type: ThinkingParserNone,
			},
		}
		err := model.Validate()
		assert.NoError(t, err)
	})

	t.Run("requires model name", func(t *testing.T) {
		model := &Model{
			Provider: "test-provider",
		}
		err := model.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")
	})

	t.Run("requires provider", func(t *testing.T) {
		model := &Model{
			Name: "test-model",
		}
		err := model.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider is required")
	})

	t.Run("validates thinking parser config", func(t *testing.T) {
		model := &Model{
			Name:     "test-model",
			Provider: "test-provider",
			ThinkingParser: &ThinkingParserConfig{
				Type: ThinkingParserInBand,
				// Missing delimiters
			},
		}
		err := model.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid thinking parser config")
	})

	t.Run("allows nil thinking parser", func(t *testing.T) {
		model := &Model{
			Name:     "test-model",
			Provider: "test-provider",
		}
		err := model.Validate()
		assert.NoError(t, err)
	})
}

func TestThinkingParserConfigValidate(t *testing.T) {
	t.Run("validates none type", func(t *testing.T) {
		config := &ThinkingParserConfig{
			Type: ThinkingParserNone,
		}
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("validates in_band type with delimiters", func(t *testing.T) {
		config := &ThinkingParserConfig{
			Type:           ThinkingParserInBand,
			StartDelimiter: "<think>",
			EndDelimiter:   "</think>",
		}
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("requires start_delimiter for in_band", func(t *testing.T) {
		config := &ThinkingParserConfig{
			Type:         ThinkingParserInBand,
			EndDelimiter: "</think>",
		}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "start_delimiter")
	})

	t.Run("requires end_delimiter for in_band", func(t *testing.T) {
		config := &ThinkingParserConfig{
			Type:           ThinkingParserInBand,
			StartDelimiter: "<think>",
		}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "end_delimiter")
	})

	t.Run("validates out_of_band type with field_path", func(t *testing.T) {
		config := &ThinkingParserConfig{
			Type:      ThinkingParserOutOfBand,
			FieldPath: "thinking",
		}
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("requires field_path for out_of_band", func(t *testing.T) {
		config := &ThinkingParserConfig{
			Type: ThinkingParserOutOfBand,
		}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "field_path")
	})

	t.Run("rejects unknown parser type", func(t *testing.T) {
		config := &ThinkingParserConfig{
			Type: "unknown",
		}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown parser type")
	})
}

func TestAutoDetectThinkingParser(t *testing.T) {
	tests := []struct {
		name          string
		modelName     string
		expectedType  ThinkingParserType
		expectedField string
		expectedStart string
		expectedEnd   string
	}{
		{
			name:          "Claude model",
			modelName:     "claude-3-5-sonnet-20241022",
			expectedType:  ThinkingParserOutOfBand,
			expectedField: "thinking",
		},
		{
			name:          "Claude model uppercase",
			modelName:     "CLAUDE-3-OPUS",
			expectedType:  ThinkingParserOutOfBand,
			expectedField: "thinking",
		},
		{
			name:          "o1 model",
			modelName:     "o1-preview",
			expectedType:  ThinkingParserOutOfBand,
			expectedField: "reasoning.summary",
		},
		{
			name:          "o1-mini model",
			modelName:     "o1-mini",
			expectedType:  ThinkingParserOutOfBand,
			expectedField: "reasoning.summary",
		},
		{
			name:          "o3 model",
			modelName:     "o3-mini",
			expectedType:  ThinkingParserOutOfBand,
			expectedField: "reasoning.summary",
		},
		{
			name:          "QwQ model",
			modelName:     "qwq-32b-preview",
			expectedType:  ThinkingParserInBand,
			expectedStart: "<think>",
			expectedEnd:   "</think>",
		},
		{
			name:          "Qwen model",
			modelName:     "qwen-7b",
			expectedType:  ThinkingParserInBand,
			expectedStart: "<think>",
			expectedEnd:   "</think>",
		},
		{
			name:          "DeepSeek R1 model",
			modelName:     "deepseek-r1-distill-qwen-32b",
			expectedType:  ThinkingParserInBand,
			expectedStart: "<think>",
			expectedEnd:   "</think>",
		},
		{
			name:         "GPT-4 (no thinking)",
			modelName:    "gpt-4-turbo",
			expectedType: ThinkingParserNone,
		},
		{
			name:         "Llama (no thinking)",
			modelName:    "llama-3-70b",
			expectedType: ThinkingParserNone,
		},
		{
			name:         "Gemma (no thinking)",
			modelName:    "gemma-7b",
			expectedType: ThinkingParserNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := autoDetectThinkingParser(tt.modelName)
			assert.Equal(t, tt.expectedType, config.Type)
			if tt.expectedField != "" {
				assert.Equal(t, tt.expectedField, config.FieldPath)
			}
			if tt.expectedStart != "" {
				assert.Equal(t, tt.expectedStart, config.StartDelimiter)
			}
			if tt.expectedEnd != "" {
				assert.Equal(t, tt.expectedEnd, config.EndDelimiter)
			}
		})
	}
}
