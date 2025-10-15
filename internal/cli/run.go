package cli

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/poiesic/wonda/internal/scenarios"
	"github.com/poiesic/wonda/internal/simulations"
	"github.com/spf13/cobra"
)

var runCommand = &cobra.Command{
	Use:     "run <scenario-name>",
	Aliases: []string{"r"},
	Short:   "Run a simulation from a scenario definition",
	Args:    cobra.ExactArgs(1),
	Run:     runSimulation,
}

func init() {
	rootCommand.AddCommand(runCommand)
}

func runSimulation(cmd *cobra.Command, args []string) {
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
	fmt.Printf("Initializing simulation...\n\n")
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
