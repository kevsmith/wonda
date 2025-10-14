# Wonda Model Configuration Examples

This directory contains example model configuration files for Wonda.

## Usage

1. Copy the example files you need to `~/.config/wonda/models/`
2. Rename them if desired (filename becomes the display name)
3. Adjust the configuration as needed for your setup

```bash
# Example: Set up Claude
mkdir -p ~/.config/wonda/models
cp models.toml.example/claude-opus.toml ~/.config/wonda/models/claude-opus.toml
```

## Configuration Structure

Each model configuration file contains:

- **name**: The API model identifier (e.g., "claude-3-5-sonnet-20241022")
- **provider**: Reference to a provider name defined in `providers.toml`
- **thinking_parser** (optional): Configuration for extracting thinking/reasoning from responses

## Thinking Parser Auto-Detection

Wonda automatically detects the appropriate thinking parser based on model name patterns:

| Model Pattern | Parser Type | Configuration |
|--------------|-------------|---------------|
| `claude-*` | out_of_band | `thinking` field |
| `o1-*`, `o3-*` | out_of_band | `reasoning.summary` field |
| `qwq*`, `qwen*` | in_band | `<think>...</think>` delimiters |
| `deepseek-r1*` | in_band | `<think>...</think>` delimiters |
| Others | none | No thinking extraction |

## Manual Configuration

You can override auto-detection by explicitly configuring the thinking parser:

### Out-of-Band Parser

For models that return thinking in a separate API response field:

```toml
[thinking_parser]
type = "out_of_band"
field_path = "reasoning.summary"  # JSONPath-like field accessor
```

### In-Band Parser

For models that embed thinking in response text with delimiters:

```toml
[thinking_parser]
type = "in_band"
start_delimiter = "<think>"
end_delimiter = "</think>"
```

### No Thinking

To disable thinking extraction:

```toml
[thinking_parser]
type = "none"
```

## Examples

- **claude-opus.toml**: Anthropic Claude with auto-detected thinking
- **o1-preview.toml**: OpenAI o1 reasoning model with auto-detected thinking
- **gpt-4-turbo.toml**: Standard GPT-4 model without thinking
- **qwq-local.toml**: Local Qwen reasoning model via Ollama
- **custom-reasoning.toml**: Custom model with explicit thinking configuration
