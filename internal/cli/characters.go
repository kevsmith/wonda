package cli

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/poiesic/wonda/internal/scenarios"
	"github.com/spf13/cobra"
)

var charactersCommand = &cobra.Command{
	Use:     "characters",
	Short:   "Work with wonda character definitions",
	Aliases: []string{"c"},
}

var showCharacterCommand = &cobra.Command{
	Use:   "show <character-name>",
	Short: "View a specific character definition",
	Args:  cobra.ExactArgs(1),
	Run:   showCharacter,
}

var editCharacterCommand = &cobra.Command{
	Use:   "edit <character-name>",
	Short: "Open a character definition in $EDITOR",
	Args:  cobra.ExactArgs(1),
	Run:   editCharacter,
}

var newCharacterCommand = &cobra.Command{
	Use:   "new <character-name>",
	Short: "Create a new character definition",
	Args:  cobra.ExactArgs(1),
	Run:   newCharacter,
}

var listCharactersCommand = &cobra.Command{
	Use:   "list",
	Short: "List all available character definitions",
	Run:   listCharacters,
}

const characterTemplate = `version = "1.0.0"

[basics]
# Required: Character archetype or name
archetype = ""

# Required: Core definition of who/what the character is (10-1000 characters)
description = ""

# Optional: Detailed history and context (max 2000 characters)
background = ""

# Required: How the character speaks and interacts (10-500 characters)
communication_style = ""

# Required: How the character makes choices (10-500 characters)
decision_style = ""

# Required: Behavioral characteristics (minimum 1)
traits = []

# Optional: Areas of expertise
skills = []

# Optional: Core beliefs and principles
values = []
`

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

	// Create the file with template
	if err := os.WriteFile(tomlFile, []byte(characterTemplate), 0644); err != nil {
		reportErrorAndDieP(tomlFile, err)
	}

	reportSuccess(fmt.Sprintf("Created character definition: %s", tomlFile))

	// Validate the template (will fail validation due to empty fields, but that's expected)
	_, err := scenarios.LoadCharacter([]byte(characterTemplate))
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
