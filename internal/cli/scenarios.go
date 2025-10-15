package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"
	"time"

	"github.com/poiesic/wonda/internal/config"
	"github.com/poiesic/wonda/internal/memory"
	"github.com/poiesic/wonda/internal/scenarios"
	"github.com/poiesic/wonda/internal/simulations"
	"github.com/spf13/cobra"
)

var scenariosCommand = &cobra.Command{
	Use:     "scenarios",
	Short:   "Manage scenario definitions",
	Aliases: []string{"s"},
}

var showScenarioCommand = &cobra.Command{
	Use:     "show <scenario-name>",
	Short:   "Display scenario definition",
	Aliases: []string{"s"},
	Args:    cobra.ExactArgs(1),
	Run:     showScenario,
}

var editScenarioCommand = &cobra.Command{
	Use:     "edit <scenario-name>",
	Short:   "Edit scenario definition in $EDITOR",
	Aliases: []string{"e"},
	Args:    cobra.ExactArgs(1),
	Run:     editScenario,
}

var newScenarioCommand = &cobra.Command{
	Use:     "new <scenario-name>",
	Short:   "Create new scenario definition",
	Aliases: []string{"n"},
	Args:    cobra.ExactArgs(1),
	Run:     newScenario,
}

var listScenariosCommand = &cobra.Command{
	Use:     "list",
	Short:   "List all scenario definitions",
	Aliases: []string{"l"},
	Run:     listScenarios,
}

var runScenarioCommand = &cobra.Command{
	Use:     "run <scenario-name>",
	Aliases: []string{"r"},
	Short:   "Run a simulation from a scenario definition",
	Args:    cobra.ExactArgs(1),
	Run:     runScenario,
}

func init() {
	scenariosCommand.AddCommand(showScenarioCommand, editScenarioCommand, newScenarioCommand, listScenariosCommand, runScenarioCommand)
}

func showScenario(cmd *cobra.Command, args []string) {
	scenarioName := args[0]
	if !strings.HasSuffix(scenarioName, ".toml") {
		scenarioName = scenarioName + ".toml"
	}
	tomlFile := path.Join(configDir, "scenarios", scenarioName)
	contents, err := os.ReadFile(tomlFile)
	if err != nil {
		reportErrorAndDieP(tomlFile, err)
	}
	fmt.Printf("PATH: %s\n", tomlFile)
	fmt.Println(string(contents))
}

func editScenario(cmd *cobra.Command, args []string) {
	scenarioName := args[0]
	if !strings.HasSuffix(scenarioName, ".toml") {
		scenarioName = scenarioName + ".toml"
	}
	tomlFile := path.Join(configDir, "scenarios", scenarioName)
	if _, err := os.Stat(tomlFile); err != nil {
		reportErrorAndDieP(tomlFile, err)
	}
	editFile(tomlFile)
}

func newScenario(cmd *cobra.Command, args []string) {
	scenarioName := args[0]
	if !strings.HasSuffix(scenarioName, ".toml") {
		scenarioName = scenarioName + ".toml"
	}
	tomlFile := path.Join(configDir, "scenarios", scenarioName)

	// Check if file already exists
	if _, err := os.Stat(tomlFile); err == nil {
		reportErrorAndDieS(fmt.Sprintf("scenario definition already exists: %s", tomlFile))
	}

	// Ensure scenarios directory exists
	scenariosDir := path.Join(configDir, "scenarios")
	if err := os.MkdirAll(scenariosDir, 0755); err != nil {
		reportErrorAndDieP(scenariosDir, err)
	}

	// Get template content
	templateContent, err := config.GetTemplate("scenario")
	if err != nil {
		reportErrorAndDieS(fmt.Sprintf("Failed to load scenario template: %s", err.Error()))
	}

	// Create the file with template
	if err := os.WriteFile(tomlFile, []byte(templateContent), 0644); err != nil {
		reportErrorAndDieP(tomlFile, err)
	}

	reportSuccess(fmt.Sprintf("Created scenario definition: %s", tomlFile))

	// Validate the template (will fail validation due to empty fields, but that's expected)
	_, err = scenarios.LoadScenario([]byte(templateContent))
	if err != nil {
		reportWarning(fmt.Sprintf("Template needs completion: %s", err.Error()))
	}

	// Open in editor
	editFile(tomlFile)
}

func listScenarios(cmd *cobra.Command, args []string) {
	scenariosDir := path.Join(configDir, "scenarios")

	entries, err := os.ReadDir(scenariosDir)
	if err != nil {
		if os.IsNotExist(err) {
			reportWarning("No scenarios directory found. Run 'wonda init' first.")
			return
		}
		reportErrorAndDieP(scenariosDir, err)
	}

	if len(entries) == 0 {
		fmt.Println("No scenario definitions found.")
		return
	}

	fmt.Printf("Scenarios in %s:\n\n", scenariosDir)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".toml") {
			continue
		}

		scenarioFile := path.Join(scenariosDir, entry.Name())
		contents, err := os.ReadFile(scenarioFile)
		if err != nil {
			fmt.Printf("  ❌ %s (error reading file)\n", entry.Name())
			continue
		}

		scenario, err := scenarios.LoadScenario(contents)
		if err != nil {
			fmt.Printf("  ❌ %s (invalid TOML)\n", entry.Name())
			continue
		}

		nameDisplay := strings.TrimSuffix(entry.Name(), ".toml")
		if scenario.Basics != nil && scenario.Basics.Name != "" {
			fmt.Printf("  • %s\n", nameDisplay)
			fmt.Printf("    Name: %s\n", scenario.Basics.Name)
			if scenario.Basics.Description != "" {
				// Truncate description if too long
				desc := scenario.Basics.Description
				if len(desc) > 60 {
					desc = desc[:57] + "..."
				}
				fmt.Printf("    Description: %s\n", desc)
			}
			if len(scenario.Agents) > 0 {
				agentNames := make([]string, 0, len(scenario.Agents))
				for name := range scenario.Agents {
					agentNames = append(agentNames, name)
				}
				fmt.Printf("    Agents: %d (%s)\n", len(scenario.Agents), strings.Join(agentNames, ", "))
			}
			if len(scenario.Goals) > 0 {
				fmt.Printf("    Goals: %d\n", len(scenario.Goals))
			}
			if len(scenario.Basics.Tags) > 0 {
				fmt.Printf("    Tags: %s\n", strings.Join(scenario.Basics.Tags, ", "))
			}
		} else {
			fmt.Printf("  • %s (incomplete)\n", nameDisplay)
		}
	}
}

func runScenario(cmd *cobra.Command, args []string) {
	// Ensure ONNX environment is cleaned up when simulation ends
	defer memory.DestroyONNXEnvironment()

	scenarioName := args[0]
	if !strings.HasSuffix(scenarioName, ".toml") {
		scenarioName = scenarioName + ".toml"
	}

	// Load scenario
	scenarioPath := path.Join(configDir, "scenarios", scenarioName)
	scenario, err := scenarios.LoadScenarioFromFile(scenarioPath)
	if err != nil {
		reportErrorAndDieP(scenarioPath, err)
	}

	// Create simulation
	sim := simulations.NewSimulation(scenario, configDir)

	// Initialize simulation (load characters, create agents)
	slog.Info("initializing simulation", "id", sim.ID.String())
	ctx := context.Background()

	// Set timeout based on scenario max_runtime
	timeout := scenario.Basics.MaxRuntime.ToDuration()
	if timeout == 0 {
		timeout = 30 * time.Minute // default
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := sim.Initialize(ctx); err != nil {
		reportErrorAndDieS(fmt.Sprintf("Failed to initialize simulation: %v", err))
	}

	// Start simulation
	fmt.Println()
	if err := sim.Start(ctx); err != nil {
		reportErrorAndDieS(fmt.Sprintf("Simulation error: %v", err))
	}
}
