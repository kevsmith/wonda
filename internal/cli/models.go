package cli

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/poiesic/wonda/internal/config"
	"github.com/spf13/cobra"
)

var modelsCommand = &cobra.Command{
	Use:     "models",
	Short:   "Manage model configurations",
	Aliases: []string{"m"},
}

var showModelCommand = &cobra.Command{
	Use:     "show <model-name>",
	Short:   "Display model configuration",
	Aliases: []string{"s"},
	Args:    cobra.ExactArgs(1),
	Run:     showModel,
}

var editModelCommand = &cobra.Command{
	Use:     "edit <model-name>",
	Short:   "Edit model configuration in $EDITOR",
	Aliases: []string{"e"},
	Args:    cobra.ExactArgs(1),
	Run:     editModel,
}

var newModelCommand = &cobra.Command{
	Use:     "new <model-name>",
	Short:   "Create new model configuration",
	Aliases: []string{"n"},
	Args:    cobra.ExactArgs(1),
	Run:     newModel,
}

var listModelsCommand = &cobra.Command{
	Use:     "list",
	Short:   "List all model configurations",
	Aliases: []string{"l"},
	Run:     listModels,
}


func init() {
	modelsCommand.AddCommand(showModelCommand, editModelCommand, newModelCommand, listModelsCommand)
}

func showModel(cmd *cobra.Command, args []string) {
	modelName := args[0]
	if !strings.HasSuffix(modelName, ".toml") {
		modelName = modelName + ".toml"
	}
	tomlFile := path.Join(configDir, "models", modelName)
	contents, err := os.ReadFile(tomlFile)
	if err != nil {
		reportErrorAndDieP(tomlFile, err)
	}
	fmt.Printf("PATH: %s\n", tomlFile)
	fmt.Println(string(contents))
}

func editModel(cmd *cobra.Command, args []string) {
	modelName := args[0]
	if !strings.HasSuffix(modelName, ".toml") {
		modelName = modelName + ".toml"
	}
	tomlFile := path.Join(configDir, "models", modelName)
	if _, err := os.Stat(tomlFile); err != nil {
		reportErrorAndDieP(tomlFile, err)
	}
	editFile(tomlFile)
}

func newModel(cmd *cobra.Command, args []string) {
	modelName := args[0]
	if !strings.HasSuffix(modelName, ".toml") {
		modelName = modelName + ".toml"
	}
	tomlFile := path.Join(configDir, "models", modelName)

	// Check if file already exists
	if _, err := os.Stat(tomlFile); err == nil {
		reportErrorAndDieS(fmt.Sprintf("model configuration already exists: %s", tomlFile))
	}

	// Ensure models directory exists
	modelsDir := path.Join(configDir, "models")
	if err := os.MkdirAll(modelsDir, 0755); err != nil {
		reportErrorAndDieP(modelsDir, err)
	}

	// Get template content
	templateContent, err := config.GetTemplate("model")
	if err != nil {
		reportErrorAndDieS(fmt.Sprintf("Failed to load model template: %s", err.Error()))
	}

	// Create the file with template
	if err := os.WriteFile(tomlFile, []byte(templateContent), 0644); err != nil {
		reportErrorAndDieP(tomlFile, err)
	}

	reportSuccess(fmt.Sprintf("Created model configuration: %s", tomlFile))

	// Validate the template
	model, err := config.LoadModel([]byte(templateContent))
	if err != nil {
		reportWarning(fmt.Sprintf("Template validation warning: %s", err.Error()))
	} else if err := model.Validate(); err != nil {
		reportWarning(fmt.Sprintf("Template validation warning: %s", err.Error()))
	}

	// Open in editor
	editFile(tomlFile)
}

func listModels(cmd *cobra.Command, args []string) {
	modelsDir := path.Join(configDir, "models")

	entries, err := os.ReadDir(modelsDir)
	if err != nil {
		if os.IsNotExist(err) {
			reportWarning("No models directory found. Run 'wonda init' first.")
			return
		}
		reportErrorAndDieP(modelsDir, err)
	}

	if len(entries) == 0 {
		fmt.Println("No model configurations found.")
		return
	}

	fmt.Printf("Models in %s:\n\n", modelsDir)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".toml") {
			continue
		}

		modelFile := path.Join(modelsDir, entry.Name())
		contents, err := os.ReadFile(modelFile)
		if err != nil {
			fmt.Printf("  ❌ %s (error reading file)\n", entry.Name())
			continue
		}

		model, err := config.LoadModel(contents)
		if err != nil {
			fmt.Printf("  ❌ %s (invalid TOML)\n", entry.Name())
			continue
		}

		nameDisplay := strings.TrimSuffix(entry.Name(), ".toml")
		if model.Name != "" {
			fmt.Printf("  • %s\n", nameDisplay)
			fmt.Printf("    Model: %s\n", model.Name)
			if model.Provider != "" {
				fmt.Printf("    Provider: %s\n", model.Provider)
			}
			if model.ThinkingParser != nil && model.ThinkingParser.Type != config.ThinkingParserNone {
				fmt.Printf("    Thinking: %s\n", model.ThinkingParser.Type)
			}
		} else {
			fmt.Printf("  • %s (incomplete)\n", nameDisplay)
		}
	}
}
