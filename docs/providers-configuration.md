# Wonda Providers Configuration

**Version**: 1.0.0
**Status**: Draft
**Last Updated**: 2025-10-13

## Overview

The `providers.toml` file configures LLM provider endpoints and authentication for Wonda. This allows scenarios to reference providers by name while keeping sensitive configuration (API keys, base URLs) separate from scenario definitions.

## Design Principles

1. **Separation of Concerns**: Provider configuration separate from scenario definitions
2. **Security**: API keys stored in configuration, not in scenario files
3. **Flexibility**: Support both cloud and self-hosted providers
4. **Simplicity**: Minimal configuration required per provider

## File Location

**Default path**: `~/.config/wonda/providers.toml` (Linux/macOS) or `%APPDATA%\wonda\providers.toml` (Windows)

Alternative paths can be specified via:
- Command line flag: `--providers-config /path/to/providers.toml`
- Environment variable: `WONDA_PROVIDERS_CONFIG=/path/to/providers.toml`

## TOML Format

```toml
# Anthropic Claude API
[providers.anthropic]
base_url = "https://api.anthropic.com/v1"
api_key = "sk-ant-..."  # Optional: can also use ANTHROPIC_API_KEY env var

# OpenAI API
[providers.openai]
base_url = "https://api.openai.com/v1"
api_key = "sk-..."  # Optional: can also use OPENAI_API_KEY env var

# Google Gemini API
[providers.google]
base_url = "https://generativelanguage.googleapis.com/v1"
api_key = "..."  # Optional: can also use GOOGLE_API_KEY env var

# Ollama (self-hosted, no API key needed)
[providers.ollama]
base_url = "http://localhost:11434"

# Custom provider
[providers.custom]
base_url = "https://my-llm-server.example.com/v1"
api_key = "custom-key-123"
```

## Field Descriptions

### Provider Name (from section header)

- Unique identifier for the provider
- Format: `[providers.provider_name]`
- Referenced in scenario files via `provider = "provider_name"`
- Examples: `anthropic`, `openai`, `ollama`, `google`, `custom`
- **Naming requirements**:
  - Must start with an alphabetic character (a-z, A-Z)
  - Can contain alphanumeric characters, dashes (`-`), and underscores (`_`)
  - Cannot start with numbers (e.g., `99problems` is invalid)
- **Environment variable transformation**:
  - Converted to uppercase
  - Dashes (`-`) and spaces are replaced with underscores (`_`)
  - Example: `ollama-local` → `OLLAMA_LOCAL_API_KEY`
  - Example: `my custom provider` → `MY_CUSTOM_PROVIDER_API_KEY`

### base_url (required)

**Type**: string (URL)
**Description**: Base URL for the provider's API endpoint
**Examples**:
- Cloud providers: `"https://api.anthropic.com/v1"`, `"https://api.openai.com/v1"`
- Self-hosted: `"http://localhost:11434"`, `"http://192.168.1.100:8080"`

### api_key (optional)

**Type**: string
**Description**: API key for authentication
**Environment variable fallback**: If not specified, Wonda will check `<PROVIDER_NAME>_API_KEY` where `<PROVIDER_NAME>` is the uppercase version of the provider name
**Security notes**:
- Store securely, do not commit to version control
- Can be omitted if using environment variables (recommended)
- Not needed for self-hosted providers without authentication

## Environment Variable Fallback

If `api_key` is not specified in the configuration file, Wonda will check for environment variables using the pattern `<PROVIDER_NAME>_API_KEY` where `<PROVIDER_NAME>` is derived from the provider name in the TOML section header.

**Transformation Rules**:
1. Convert to uppercase
2. Replace dashes (`-`) with underscores (`_`)
3. Replace spaces with underscores (`_`)
4. Append `_API_KEY` suffix

**Naming Pattern**: `[providers.{name}]` → `{NAME}_API_KEY`

**Examples**:
- `[providers.anthropic]` → `ANTHROPIC_API_KEY`
- `[providers.openai]` → `OPENAI_API_KEY`
- `[providers.google]` → `GOOGLE_API_KEY`
- `[providers.ollama-local]` → `OLLAMA_LOCAL_API_KEY` (dash → underscore)
- `[providers.my-custom-llm]` → `MY_CUSTOM_LLM_API_KEY` (dashes → underscores)
- `[providers."my provider"]` → `MY_PROVIDER_API_KEY` (space → underscore)

This pattern allows using the same underlying provider with different API keys by creating multiple provider entries:

```toml
# Use different Anthropic keys for different purposes
[providers.anthropic-dev]
base_url = "https://api.anthropic.com/v1"
# Uses ANTHROPIC_DEV_API_KEY

[providers.anthropic-prod]
base_url = "https://api.anthropic.com/v1"
# Uses ANTHROPIC_PROD_API_KEY
```

**Example without api_key in config**:
```toml
[providers.anthropic]
base_url = "https://api.anthropic.com/v1"
# No api_key specified - will use ANTHROPIC_API_KEY env var
```

## Usage in Scenarios

Providers configured in `providers.toml` are referenced by name in scenario files:

```toml
# In scenario file
[scenario.defaults]
provider = "anthropic"  # References [providers.anthropic] from providers.toml
model = "claude-3-5-sonnet-20241022"

[agents.Alice]
character = "pragmatist"
provider = "ollama"  # References [providers.ollama] from providers.toml
model = "llama3.1:8b"
```

## Minimal Configuration

Self-hosted setup with Ollama (no API keys required):

```toml
[providers.ollama]
base_url = "http://localhost:11434"
```

## Security Best Practices

1. **File Permissions**: Set restrictive permissions on `providers.toml`
   ```bash
   chmod 600 ~/.config/wonda/providers.toml
   ```

2. **Version Control**: Add to `.gitignore`
   ```
   providers.toml
   .config/wonda/providers.toml
   ```

3. **Environment Variables**: Prefer environment variables for API keys in production
   ```bash
   export ANTHROPIC_API_KEY="sk-ant-..."
   export OPENAI_API_KEY="sk-..."
   ```

4. **Separate Configurations**: Use different `providers.toml` files for development/production

## Example Configurations

### Cloud-Only Setup

```toml
[providers.anthropic]
base_url = "https://api.anthropic.com/v1"
api_key = "sk-ant-..."

[providers.openai]
base_url = "https://api.openai.com/v1"
api_key = "sk-..."
```

### Self-Hosted Only

```toml
[providers.ollama]
base_url = "http://localhost:11434"
```

### Hybrid Setup

```toml
# Cloud provider for powerful models
[providers.anthropic]
base_url = "https://api.anthropic.com/v1"
# API key from ANTHROPIC_API_KEY env var

# Local provider for development/testing
[providers.ollama]
base_url = "http://localhost:11434"
```

### Multiple Ollama Instances

```toml
[providers.ollama-local]
base_url = "http://localhost:11434"

[providers.ollama-server]
base_url = "http://192.168.1.100:11434"
```

### Multiple API Keys for Same Provider

Use different API keys for the same provider (e.g., separate dev/prod accounts, different rate limits, cost tracking):

```toml
[providers.anthropic-dev]
base_url = "https://api.anthropic.com/v1"
# Uses ANTHROPIC_DEV_API_KEY environment variable

[providers.anthropic-prod]
base_url = "https://api.anthropic.com/v1"
# Uses ANTHROPIC_PROD_API_KEY environment variable

[providers.openai-team-a]
base_url = "https://api.openai.com/v1"
api_key = "sk-proj-teamA..."

[providers.openai-team-b]
base_url = "https://api.openai.com/v1"
api_key = "sk-proj-teamB..."
```

Then in scenarios, reference by the specific provider name:
```toml
[scenario.defaults]
provider = "anthropic-dev"  # Uses dev account
model = "claude-3-5-sonnet-20241022"
```

## Validation

Wonda validates provider configuration on startup:

- **Invalid provider name**: Error - provider names must start with an alphabetic character (a-z, A-Z)
- **Missing base_url**: Error - provider cannot be used
- **Invalid URL format**: Error - must be valid HTTP/HTTPS URL
- **Missing api_key**: Warning if no environment variable found (for cloud providers)
- **Unreferenced providers**: No warning - unused providers are OK

## Related Documentation

- [Scenario Definition](./scenario-definition.md) - Using providers in scenarios
- [Agent Architecture](./agent-architecture.md) - How agents interact with providers
