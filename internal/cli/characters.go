package cli

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/poiesic/wonda/internal/config"
	"github.com/poiesic/wonda/internal/scenarios"
	"github.com/spf13/cobra"
)

var charactersCommand = &cobra.Command{
	Use:     "characters",
	Short:   "Manage character definitions",
	Aliases: []string{"c"},
}

var showCharacterCommand = &cobra.Command{
	Use:     "show <character-name>",
	Short:   "Display character definition",
	Aliases: []string{"s"},
	Args:    cobra.ExactArgs(1),
	Run:     showCharacter,
}

var editCharacterCommand = &cobra.Command{
	Use:     "edit <character-name>",
	Short:   "Edit character definition in $EDITOR",
	Aliases: []string{"e"},
	Args:    cobra.ExactArgs(1),
	Run:     editCharacter,
}

var newCharacterCommand = &cobra.Command{
	Use:     "new <character-name>",
	Short:   "Create new character definition",
	Aliases: []string{"n"},
	Args:    cobra.ExactArgs(1),
	Run:     newCharacter,
}

var listCharactersCommand = &cobra.Command{
	Use:     "list",
	Short:   "List all character definitions",
	Aliases: []string{"l"},
	Run:     listCharacters,
}

func init() {
	charactersCommand.AddCommand(showCharacterCommand, editCharacterCommand, newCharacterCommand, listCharactersCommand)
}

func showCharacter(cmd *cobra.Command, args []string) {
	characterName := args[0]
	if !strings.HasSuffix(characterName, ".toml") {
		characterName = characterName + ".toml"
	}
	tomlFile := path.Join(configDir, "characters", characterName)
	contents, err := os.ReadFile(tomlFile)
	if err != nil {
		reportErrorAndDieP(tomlFile, err)
	}
	fmt.Printf("PATH: %s\n", tomlFile)
	fmt.Println(string(contents))
}

func editCharacter(cmd *cobra.Command, args []string) {
	characterName := args[0]
	if !strings.HasSuffix(characterName, ".toml") {
		characterName = characterName + ".toml"
	}
	tomlFile := path.Join(configDir, "characters", characterName)
	if _, err := os.Stat(tomlFile); err != nil {
		reportErrorAndDieP(tomlFile, err)
	}
	editFile(tomlFile)
}

func newCharacter(cmd *cobra.Command, args []string) {
	characterName := args[0]
	if !strings.HasSuffix(characterName, ".toml") {
		characterName = characterName + ".toml"
	}
	tomlFile := path.Join(configDir, "characters", characterName)

	// Check if file already exists
	if _, err := os.Stat(tomlFile); err == nil {
		reportErrorAndDieS(fmt.Sprintf("character definition already exists: %s", tomlFile))
	}

	// Ensure characters directory exists
	charactersDir := path.Join(configDir, "characters")
	if err := os.MkdirAll(charactersDir, 0755); err != nil {
		reportErrorAndDieP(charactersDir, err)
	}

	// Get template content
	templateContent, err := config.GetTemplate("character")
	if err != nil {
		reportErrorAndDieS(fmt.Sprintf("Failed to load character template: %s", err.Error()))
	}

	// Create the file with template
	if err := os.WriteFile(tomlFile, []byte(templateContent), 0644); err != nil {
		reportErrorAndDieP(tomlFile, err)
	}

	reportSuccess(fmt.Sprintf("Created character definition: %s", tomlFile))

	// Validate the template (will fail validation due to empty fields, but that's expected)
	_, err = scenarios.LoadCharacter([]byte(templateContent))
	if err != nil {
		reportWarning(fmt.Sprintf("Template needs completion: %s", err.Error()))
	}

	// Open in editor
	editFile(tomlFile)
}

func listCharacters(cmd *cobra.Command, args []string) {
	charactersDir := path.Join(configDir, "characters")

	entries, err := os.ReadDir(charactersDir)
	if err != nil {
		if os.IsNotExist(err) {
			reportWarning("No characters directory found. Run 'wonda init' first.")
			return
		}
		reportErrorAndDieP(charactersDir, err)
	}

	if len(entries) == 0 {
		fmt.Println("No character definitions found.")
		return
	}

	fmt.Printf("Characters in %s:\n\n", charactersDir)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".toml") {
			continue
		}

		characterFile := path.Join(charactersDir, entry.Name())
		contents, err := os.ReadFile(characterFile)
		if err != nil {
			fmt.Printf("  ❌ %s (error reading file)\n", entry.Name())
			continue
		}

		character, err := scenarios.LoadCharacter(contents)
		if err != nil {
			fmt.Printf("  ❌ %s (invalid TOML)\n", entry.Name())
			continue
		}

		nameDisplay := strings.TrimSuffix(entry.Name(), ".toml")
		if character.Basics != nil && character.Basics.Archetype != "" {
			fmt.Printf("  • %s\n", nameDisplay)
			fmt.Printf("    Archetype: %s\n", character.Basics.Archetype)
			if character.Basics.Description != "" {
				// Truncate description if too long
				desc := character.Basics.Description
				if len(desc) > 60 {
					desc = desc[:57] + "..."
				}
				fmt.Printf("    Description: %s\n", desc)
			}
			if len(character.Basics.Traits) > 0 {
				fmt.Printf("    Traits: %s\n", strings.Join(character.Basics.Traits, ", "))
			}
		} else {
			fmt.Printf("  • %s (incomplete)\n", nameDisplay)
		}
	}
}
