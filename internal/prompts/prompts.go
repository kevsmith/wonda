package prompts

import (
	"embed"
	"fmt"
)

// FS contains all prompt template files embedded at build time.
// Only files ending in _prompt.md are embedded to avoid accidentally
// packaging documentation or other markdown files.
//
//go:embed *_prompt.md
var FS embed.FS

// GetPrompt retrieves a prompt template by name.
// The name should not include the "_prompt.md" suffix - it will be added automatically.
//
// Example:
//
//	content, err := prompts.GetPrompt("agent_turn")  // reads agent_turn_prompt.md
func GetPrompt(name string) (string, error) {
	filename := name + "_prompt.md"
	content, err := FS.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt '%s': %w", name, err)
	}
	return string(content), nil
}
