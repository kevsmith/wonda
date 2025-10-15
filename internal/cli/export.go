package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/poiesic/wonda/internal/chronicle"
	"github.com/spf13/cobra"
)

var exportFormat string

var exportCommand = &cobra.Command{
	Use:   "export <chronicle-file>",
	Short: "Export a chronicle file to readable format",
	Long:  "Export a chronicle JSONL file to Markdown (default) or pretty JSON",
	Args:  cobra.ExactArgs(1),
	Run:   exportChronicle,
}

func init() {
	rootCommand.AddCommand(exportCommand)
	exportCommand.Flags().StringVar(&exportFormat, "format", "markdown", "Output format: markdown or json")
}

func exportChronicle(cmd *cobra.Command, args []string) {
	chroniclePath := args[0]

	// Read and parse the JSONL file
	metadata, turns, err := readChronicleFile(chroniclePath)
	if err != nil {
		reportErrorAndDieS(fmt.Sprintf("Failed to read chronicle: %v", err))
	}

	// Export based on format
	switch exportFormat {
	case "markdown", "md":
		exportMarkdown(metadata, turns)
	case "json":
		exportJSON(metadata, turns)
	default:
		reportErrorAndDieS(fmt.Sprintf("Unknown format: %s (use 'markdown' or 'json')", exportFormat))
	}
}

// readChronicleFile reads and parses a JSONL chronicle file.
func readChronicleFile(path string) (*chronicle.Metadata, []chronicle.Turn, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	var metadata *chronicle.Metadata
	var turns []chronicle.Turn

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Parse JSON to determine type
		var typeCheck struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal([]byte(line), &typeCheck); err != nil {
			return nil, nil, fmt.Errorf("failed to parse line: %w", err)
		}

		switch typeCheck.Type {
		case "metadata":
			var m chronicle.Metadata
			if err := json.Unmarshal([]byte(line), &m); err != nil {
				return nil, nil, fmt.Errorf("failed to parse metadata: %w", err)
			}
			metadata = &m
		case "turn":
			var t chronicle.Turn
			if err := json.Unmarshal([]byte(line), &t); err != nil {
				return nil, nil, fmt.Errorf("failed to parse turn: %w", err)
			}
			turns = append(turns, t)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	if metadata == nil {
		return nil, nil, fmt.Errorf("no metadata found in chronicle")
	}

	return metadata, turns, nil
}

// exportJSON exports the chronicle as pretty-printed JSON.
func exportJSON(metadata *chronicle.Metadata, turns []chronicle.Turn) {
	output := map[string]interface{}{
		"metadata": metadata,
		"turns":    turns,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		reportErrorAndDieS(fmt.Sprintf("Failed to encode JSON: %v", err))
	}
}

// exportMarkdown exports the chronicle as Markdown.
func exportMarkdown(metadata *chronicle.Metadata, turns []chronicle.Turn) {
	// Header
	fmt.Printf("# Simulation Chronicle: %s\n\n", metadata.Scenario)
	fmt.Printf("**Simulation ID:** `%s`  \n", metadata.SimulationID)
	fmt.Printf("**Location:** %s  \n", metadata.Location)
	fmt.Printf("**Time:** %s  \n", metadata.Time)
	if metadata.Atmosphere != "" {
		fmt.Printf("**Atmosphere:** %s  \n", metadata.Atmosphere)
	}
	fmt.Printf("**Started:** %s  \n", metadata.StartTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("**Duration:** %d turns  \n", len(turns))
	fmt.Println()
	fmt.Println("---")
	fmt.Println()

	// Turns
	for _, turn := range turns {
		fmt.Printf("## Turn %d\n\n", turn.Number)

		for _, event := range turn.Events {
			fmt.Printf("### %s\n\n", event.AgentName)

			// Reasoning
			if event.Reasoning != "" {
				fmt.Printf("**🧠 Reasoning:**\n")
				fmt.Printf("> %s\n\n", event.Reasoning)
			}

			// Dialogue
			if event.Dialogue != "" {
				fmt.Printf("**💬 Says:**\n")
				fmt.Printf("> \"%s\"\n\n", event.Dialogue)
			}

			// Emotion
			if event.Emotion != nil {
				fmt.Printf("**😊 Emotion:** %s (%d/10) → %s (%d/10)\n\n",
					event.Emotion.Before.Emotion,
					event.Emotion.Before.Intensity,
					event.Emotion.After.Emotion,
					event.Emotion.After.Intensity)
			}

			// Proposals
			if len(event.Proposals) > 0 {
				fmt.Printf("**🎯 Proposals:**\n")
				for _, proposal := range event.Proposals {
					fmt.Printf("- %s\n", proposal)
				}
				fmt.Println()
			}

			// Votes
			if len(event.Votes) > 0 {
				fmt.Printf("**🗳️ Votes:**\n")
				for _, vote := range event.Votes {
					voteSymbol := "✗"
					if vote.Choice == "yes" {
						voteSymbol = "✓"
					}
					fmt.Printf("- %s %s\n", voteSymbol, vote.ProposalID)
				}
				fmt.Println()
			}

			fmt.Println("---")
			fmt.Println()
		}
	}
}
