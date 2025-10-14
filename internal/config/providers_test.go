package config

import (
	"os"
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvidersMarshalTOML(t *testing.T) {
	t.Run("single provider with api key", func(t *testing.T) {
		providers := NewProviders()
		apiKey := "sk-test-key"
		providers.Providers["anthropic"] = &Provider{
			BaseURL: "https://api.anthropic.com/v1",
			APIKey:  &apiKey,
		}

		buf, err := toml.Marshal(providers)
		require.NoError(t, err)
		require.NotEmpty(t, buf)

		result := string(buf)
		assert.Contains(t, result, "[providers.anthropic]")
		assert.Contains(t, result, "base_url = 'https://api.anthropic.com/v1'")
		assert.Contains(t, result, "api_key = 'sk-test-key'")
	})

	t.Run("provider without api key", func(t *testing.T) {
		providers := NewProviders()
		providers.Providers["ollama"] = &Provider{
			BaseURL: "http://localhost:11434",
			APIKey:  nil,
		}

		buf, err := toml.Marshal(providers)
		require.NoError(t, err)
		require.NotEmpty(t, buf)

		result := string(buf)
		assert.Contains(t, result, "[providers.ollama]")
		assert.Contains(t, result, "base_url = 'http://localhost:11434'")
		assert.NotContains(t, result, "api_key")
	})

	t.Run("multiple providers", func(t *testing.T) {
		providers := NewProviders()
		anthropicKey := "sk-ant-123"
		openaiKey := "sk-openai-456"

		providers.Providers["anthropic"] = &Provider{
			BaseURL: "https://api.anthropic.com/v1",
			APIKey:  &anthropicKey,
		}
		providers.Providers["openai"] = &Provider{
			BaseURL: "https://api.openai.com/v1",
			APIKey:  &openaiKey,
		}
		providers.Providers["ollama"] = &Provider{
			BaseURL: "http://localhost:11434",
			APIKey:  nil,
		}

		buf, err := toml.Marshal(providers)
		require.NoError(t, err)
		require.NotEmpty(t, buf)

		result := string(buf)
		assert.Contains(t, result, "[providers.anthropic]")
		assert.Contains(t, result, "[providers.openai]")
		assert.Contains(t, result, "[providers.ollama]")
	})
}

func TestProvidersUnmarshalTOML(t *testing.T) {
	t.Run("single provider with api key", func(t *testing.T) {
		tomlData := `
[providers.anthropic]
base_url = "https://api.anthropic.com/v1"
api_key = "sk-test-key"
`

		var providers Providers
		err := toml.Unmarshal([]byte(tomlData), &providers)
		require.NoError(t, err)

		require.Contains(t, providers.Providers, "anthropic")
		provider := providers.Providers["anthropic"]
		assert.Equal(t, "https://api.anthropic.com/v1", provider.BaseURL)
		require.NotNil(t, provider.APIKey)
		assert.Equal(t, "sk-test-key", *provider.APIKey)
	})

	t.Run("provider without api key", func(t *testing.T) {
		tomlData := `
[providers.ollama]
base_url = "http://localhost:11434"
`

		var providers Providers
		err := toml.Unmarshal([]byte(tomlData), &providers)
		require.NoError(t, err)

		require.Contains(t, providers.Providers, "ollama")
		provider := providers.Providers["ollama"]
		assert.Equal(t, "http://localhost:11434", provider.BaseURL)
		assert.Nil(t, provider.APIKey)
	})

	t.Run("multiple providers", func(t *testing.T) {
		tomlData := `
[providers.anthropic]
base_url = "https://api.anthropic.com/v1"
api_key = "sk-ant-123"

[providers.openai]
base_url = "https://api.openai.com/v1"
api_key = "sk-openai-456"

[providers.ollama]
base_url = "http://localhost:11434"

[providers.google]
base_url = "https://generativelanguage.googleapis.com/v1"
api_key = "google-key-789"
`

		var providers Providers
		err := toml.Unmarshal([]byte(tomlData), &providers)
		require.NoError(t, err)

		assert.Len(t, providers.Providers, 4)
		assert.Contains(t, providers.Providers, "anthropic")
		assert.Contains(t, providers.Providers, "openai")
		assert.Contains(t, providers.Providers, "ollama")
		assert.Contains(t, providers.Providers, "google")

		// Verify anthropic
		assert.Equal(t, "https://api.anthropic.com/v1", providers.Providers["anthropic"].BaseURL)
		require.NotNil(t, providers.Providers["anthropic"].APIKey)
		assert.Equal(t, "sk-ant-123", *providers.Providers["anthropic"].APIKey)

		// Verify ollama (no api key)
		assert.Equal(t, "http://localhost:11434", providers.Providers["ollama"].BaseURL)
		assert.Nil(t, providers.Providers["ollama"].APIKey)
	})

	t.Run("provider with custom name", func(t *testing.T) {
		tomlData := `
[providers.my-custom-provider]
base_url = "https://custom.example.com/v1"
api_key = "custom-key"
`

		var providers Providers
		err := toml.Unmarshal([]byte(tomlData), &providers)
		require.NoError(t, err)

		require.Contains(t, providers.Providers, "my-custom-provider")
		provider := providers.Providers["my-custom-provider"]
		assert.Equal(t, "https://custom.example.com/v1", provider.BaseURL)
		require.NotNil(t, provider.APIKey)
		assert.Equal(t, "custom-key", *provider.APIKey)
	})

	t.Run("provider with localhost url", func(t *testing.T) {
		tomlData := `
[providers.ollama-local]
base_url = "http://localhost:11434"

[providers.ollama-server]
base_url = "http://192.168.1.100:11434"
`

		var providers Providers
		err := toml.Unmarshal([]byte(tomlData), &providers)
		require.NoError(t, err)

		assert.Len(t, providers.Providers, 2)
		assert.Equal(t, "http://localhost:11434", providers.Providers["ollama-local"].BaseURL)
		assert.Equal(t, "http://192.168.1.100:11434", providers.Providers["ollama-server"].BaseURL)
	})
}

func TestProvidersRoundTrip(t *testing.T) {
	t.Run("single provider round trip", func(t *testing.T) {
		original := NewProviders()
		apiKey := "sk-test-key"
		original.Providers["anthropic"] = &Provider{
			BaseURL: "https://api.anthropic.com/v1",
			APIKey:  &apiKey,
		}

		// Marshal to TOML
		buf, err := toml.Marshal(original)
		require.NoError(t, err)

		// Unmarshal back
		var decoded Providers
		err = toml.Unmarshal(buf, &decoded)
		require.NoError(t, err)

		// Verify equality
		assert.Len(t, decoded.Providers, len(original.Providers))
		require.Contains(t, decoded.Providers, "anthropic")
		assert.Equal(t, original.Providers["anthropic"].BaseURL, decoded.Providers["anthropic"].BaseURL)
		require.NotNil(t, decoded.Providers["anthropic"].APIKey)
		assert.Equal(t, *original.Providers["anthropic"].APIKey, *decoded.Providers["anthropic"].APIKey)
	})

	t.Run("multiple providers round trip", func(t *testing.T) {
		original := NewProviders()
		anthropicKey := "sk-ant-123"
		openaiKey := "sk-openai-456"

		original.Providers["anthropic"] = &Provider{
			BaseURL: "https://api.anthropic.com/v1",
			APIKey:  &anthropicKey,
		}
		original.Providers["openai"] = &Provider{
			BaseURL: "https://api.openai.com/v1",
			APIKey:  &openaiKey,
		}
		original.Providers["ollama"] = &Provider{
			BaseURL: "http://localhost:11434",
			APIKey:  nil,
		}

		// Marshal to TOML
		buf, err := toml.Marshal(original)
		require.NoError(t, err)

		// Unmarshal back
		var decoded Providers
		err = toml.Unmarshal(buf, &decoded)
		require.NoError(t, err)

		// Verify equality
		assert.Len(t, decoded.Providers, len(original.Providers))

		for name, originalProvider := range original.Providers {
			require.Contains(t, decoded.Providers, name)
			decodedProvider := decoded.Providers[name]
			assert.Equal(t, originalProvider.BaseURL, decodedProvider.BaseURL)

			if originalProvider.APIKey == nil {
				assert.Nil(t, decodedProvider.APIKey)
			} else {
				require.NotNil(t, decodedProvider.APIKey)
				assert.Equal(t, *originalProvider.APIKey, *decodedProvider.APIKey)
			}
		}
	})

	t.Run("multiple round trips preserve data", func(t *testing.T) {
		original := NewProviders()
		apiKey := "sk-test-key"
		original.Providers["test"] = &Provider{
			BaseURL: "https://test.example.com",
			APIKey:  &apiKey,
		}

		current := original
		for i := 0; i < 3; i++ {
			buf, err := toml.Marshal(current)
			require.NoError(t, err, "marshal iteration %d failed", i)

			var decoded Providers
			err = toml.Unmarshal(buf, &decoded)
			require.NoError(t, err, "unmarshal iteration %d failed", i)

			assert.Len(t, decoded.Providers, len(original.Providers), "providers changed after iteration %d", i)
			assert.Equal(t, original.Providers["test"].BaseURL, decoded.Providers["test"].BaseURL, "base_url changed after iteration %d", i)
			require.NotNil(t, decoded.Providers["test"].APIKey, "api_key became nil after iteration %d", i)
			assert.Equal(t, *original.Providers["test"].APIKey, *decoded.Providers["test"].APIKey, "api_key changed after iteration %d", i)

			current = &decoded
		}
	})
}

func TestNewProviders(t *testing.T) {
	t.Run("creates valid providers", func(t *testing.T) {
		providers := NewProviders()
		require.NotNil(t, providers)
		require.NotNil(t, providers.Providers)
		assert.Empty(t, providers.Providers)
	})

	t.Run("can marshal new providers", func(t *testing.T) {
		providers := NewProviders()
		buf, err := toml.Marshal(providers)
		require.NoError(t, err)
		require.NotEmpty(t, buf)
	})

	t.Run("can add providers", func(t *testing.T) {
		providers := NewProviders()
		providers.Providers["test"] = &Provider{
			BaseURL: "https://test.example.com",
		}

		assert.Len(t, providers.Providers, 1)
		assert.Contains(t, providers.Providers, "test")
	})
}

func TestProviderLoadFromEnvironment(t *testing.T) {
	t.Run("valid provider names", func(t *testing.T) {
		validNames := []string{
			"anthropic",
			"openai",
			"ollama-local",
			"my_provider",
			"provider123",
			"a",
			"A",
			"Test-Provider_123",
		}

		for _, name := range validNames {
			t.Run(name, func(t *testing.T) {
				provider := &Provider{
					Name:    name,
					BaseURL: "https://example.com",
				}
				err := provider.LoadFromEnvironment()
				assert.NoError(t, err, "expected %s to be valid", name)
			})
		}
	})

	t.Run("invalid provider names", func(t *testing.T) {
		invalidNames := []string{
			"99problems",    // starts with number
			"123",           // all numbers
			"provider name", // contains space (should use quotes in TOML)
			"provider@test", // contains @
			"provider.test", // contains .
			"provider/test", // contains /
			"-provider",     // starts with dash
			"_provider",     // starts with underscore
			"",              // empty
		}

		for _, name := range invalidNames {
			t.Run(name, func(t *testing.T) {
				provider := &Provider{
					Name:    name,
					BaseURL: "https://example.com",
				}
				err := provider.LoadFromEnvironment()
				assert.Error(t, err, "expected %s to be invalid", name)
				assert.Contains(t, err.Error(), "invalid provider name")
			})
		}
	})

	t.Run("loads from environment when APIKey is nil", func(t *testing.T) {
		// Set environment variable
		os.Setenv("TEST_PROVIDER_API_KEY", "test-key-from-env")
		defer os.Unsetenv("TEST_PROVIDER_API_KEY")

		provider := &Provider{
			Name:    "test-provider",
			BaseURL: "https://example.com",
			APIKey:  nil,
		}

		err := provider.LoadFromEnvironment()
		require.NoError(t, err)
		require.NotNil(t, provider.APIKey)
		assert.Equal(t, "test-key-from-env", *provider.APIKey)
	})

	t.Run("does not load from environment when APIKey is set", func(t *testing.T) {
		// Set environment variable
		os.Setenv("TEST_PROVIDER_API_KEY", "env-key")
		defer os.Unsetenv("TEST_PROVIDER_API_KEY")

		existingKey := "existing-key"
		provider := &Provider{
			Name:    "test-provider",
			BaseURL: "https://example.com",
			APIKey:  &existingKey,
		}

		err := provider.LoadFromEnvironment()
		require.NoError(t, err)
		require.NotNil(t, provider.APIKey)
		assert.Equal(t, "existing-key", *provider.APIKey, "should not override existing APIKey")
	})

	t.Run("does not set APIKey when environment variable is empty", func(t *testing.T) {
		// Ensure environment variable is not set
		os.Unsetenv("EMPTY_TEST_API_KEY")

		provider := &Provider{
			Name:    "empty-test",
			BaseURL: "https://example.com",
			APIKey:  nil,
		}

		err := provider.LoadFromEnvironment()
		require.NoError(t, err)
		assert.Nil(t, provider.APIKey, "should remain nil when env var not set")
	})

	t.Run("transformation rules - dashes to underscores", func(t *testing.T) {
		os.Setenv("OLLAMA_LOCAL_API_KEY", "dash-test-key")
		defer os.Unsetenv("OLLAMA_LOCAL_API_KEY")

		provider := &Provider{
			Name:    "ollama-local",
			BaseURL: "http://localhost:11434",
			APIKey:  nil,
		}

		err := provider.LoadFromEnvironment()
		require.NoError(t, err)
		require.NotNil(t, provider.APIKey)
		assert.Equal(t, "dash-test-key", *provider.APIKey)
	})

	t.Run("transformation rules - uppercase", func(t *testing.T) {
		os.Setenv("ANTHROPIC_API_KEY", "uppercase-test-key")
		defer os.Unsetenv("ANTHROPIC_API_KEY")

		provider := &Provider{
			Name:    "anthropic",
			BaseURL: "https://api.anthropic.com/v1",
			APIKey:  nil,
		}

		err := provider.LoadFromEnvironment()
		require.NoError(t, err)
		require.NotNil(t, provider.APIKey)
		assert.Equal(t, "uppercase-test-key", *provider.APIKey)
	})

	t.Run("transformation rules - mixed case to uppercase", func(t *testing.T) {
		os.Setenv("MY_CUSTOM_PROVIDER_API_KEY", "mixed-case-key")
		defer os.Unsetenv("MY_CUSTOM_PROVIDER_API_KEY")

		provider := &Provider{
			Name:    "My-Custom-Provider",
			BaseURL: "https://example.com",
			APIKey:  nil,
		}

		err := provider.LoadFromEnvironment()
		require.NoError(t, err)
		require.NotNil(t, provider.APIKey)
		assert.Equal(t, "mixed-case-key", *provider.APIKey)
	})

	t.Run("transformation rules - multiple dashes", func(t *testing.T) {
		os.Setenv("MY_VERY_LONG_PROVIDER_NAME_API_KEY", "multi-dash-key")
		defer os.Unsetenv("MY_VERY_LONG_PROVIDER_NAME_API_KEY")

		provider := &Provider{
			Name:    "my-very-long-provider-name",
			BaseURL: "https://example.com",
			APIKey:  nil,
		}

		err := provider.LoadFromEnvironment()
		require.NoError(t, err)
		require.NotNil(t, provider.APIKey)
		assert.Equal(t, "multi-dash-key", *provider.APIKey)
	})

	t.Run("transformation rules - underscores preserved", func(t *testing.T) {
		os.Setenv("MY_PROVIDER_API_KEY", "underscore-key")
		defer os.Unsetenv("MY_PROVIDER_API_KEY")

		provider := &Provider{
			Name:    "my_provider",
			BaseURL: "https://example.com",
			APIKey:  nil,
		}

		err := provider.LoadFromEnvironment()
		require.NoError(t, err)
		require.NotNil(t, provider.APIKey)
		assert.Equal(t, "underscore-key", *provider.APIKey)
	})
}

func TestLoadProviders(t *testing.T) {
	t.Run("loads providers with environment variables", func(t *testing.T) {
		os.Setenv("ANTHROPIC_API_KEY", "env-anthropic-key")
		os.Setenv("OLLAMA_LOCAL_API_KEY", "env-ollama-key")
		defer os.Unsetenv("ANTHROPIC_API_KEY")
		defer os.Unsetenv("OLLAMA_LOCAL_API_KEY")

		tomlData := `
[providers.anthropic]
base_url = "https://api.anthropic.com/v1"

[providers.ollama-local]
base_url = "http://localhost:11434"
`

		providers, err := LoadProviders([]byte(tomlData))
		require.NoError(t, err)

		// Check anthropic loaded from env
		require.Contains(t, providers.Providers, "anthropic")
		assert.Equal(t, "anthropic", providers.Providers["anthropic"].Name)
		require.NotNil(t, providers.Providers["anthropic"].APIKey)
		assert.Equal(t, "env-anthropic-key", *providers.Providers["anthropic"].APIKey)

		// Check ollama-local loaded from env
		require.Contains(t, providers.Providers, "ollama-local")
		assert.Equal(t, "ollama-local", providers.Providers["ollama-local"].Name)
		require.NotNil(t, providers.Providers["ollama-local"].APIKey)
		assert.Equal(t, "env-ollama-key", *providers.Providers["ollama-local"].APIKey)
	})

	t.Run("prefers config api_key over environment", func(t *testing.T) {
		os.Setenv("ANTHROPIC_API_KEY", "env-key")
		defer os.Unsetenv("ANTHROPIC_API_KEY")

		tomlData := `
[providers.anthropic]
base_url = "https://api.anthropic.com/v1"
api_key = "config-key"
`

		providers, err := LoadProviders([]byte(tomlData))
		require.NoError(t, err)

		require.Contains(t, providers.Providers, "anthropic")
		require.NotNil(t, providers.Providers["anthropic"].APIKey)
		assert.Equal(t, "config-key", *providers.Providers["anthropic"].APIKey, "should prefer config key over env")
	})

	t.Run("returns error for invalid provider name", func(t *testing.T) {
		tomlData := `
[providers.99problems]
base_url = "https://example.com"
`

		_, err := LoadProviders([]byte(tomlData))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid provider name")
	})

	t.Run("sets provider Name field", func(t *testing.T) {
		tomlData := `
[providers.test]
base_url = "https://example.com"
`

		providers, err := LoadProviders([]byte(tomlData))
		require.NoError(t, err)

		require.Contains(t, providers.Providers, "test")
		assert.Equal(t, "test", providers.Providers["test"].Name)
	})
}
