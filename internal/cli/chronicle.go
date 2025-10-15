package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/poiesic/wonda/internal/chronicle"
	"github.com/spf13/cobra"
)

var chronicleCommand = &cobra.Command{
	Use:     "chronicle",
	Aliases: []string{"ch"},
	Short:   "Work with chronicle files",
	Long:    "Commands for viewing and exporting simulation chronicles",
}

var chronicleExportCommand = &cobra.Command{
	Use:     "export <chronicle-file>",
	Aliases: []string{"e"},
	Short:   "Export a chronicle file to readable format",
	Long:    "Export a chronicle JSONL file to Markdown (default) or pretty JSON",
	Args:    cobra.ExactArgs(1),
	Run:     chronicleExport,
}

var chronicleTailCommand = &cobra.Command{
	Use:     "tail <chronicle-file>",
	Aliases: []string{"t"},
	Short:   "Stream chronicle entries as they're written",
	Long:    "Continuously monitor a chronicle file and output new entries in Markdown format",
	Args:    cobra.ExactArgs(1),
	Run:     chronicleTail,
}

var exportFormat string
var tailPollInterval time.Duration

func init() {
	rootCommand.AddCommand(chronicleCommand)
	chronicleCommand.AddCommand(chronicleExportCommand, chronicleTailCommand)

	chronicleExportCommand.Flags().StringVar(&exportFormat, "format", "markdown", "Output format: markdown or json")
	chronicleTailCommand.Flags().DurationVar(&tailPollInterval, "interval", 100*time.Millisecond, "Polling interval for checking file updates")
}

func chronicleExport(cmd *cobra.Command, args []string) {
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

func chronicleTail(cmd *cobra.Command, args []string) {
	chroniclePath := args[0]

	// Check if file exists
	fileInfo, err := os.Stat(chroniclePath)
	if err != nil {
		if os.IsNotExist(err) {
			reportErrorAndDieS(fmt.Sprintf("Chronicle file not found: %s", chroniclePath))
		}
		reportErrorAndDieS(fmt.Sprintf("Failed to access file: %v", err))
	}

	// Open the file
	file, err := os.Open(chroniclePath)
	if err != nil {
		reportErrorAndDieS(fmt.Sprintf("Failed to open file: %v", err))
	}
	defer file.Close()

	// Read and output existing contents
	var metadata *chronicle.Metadata
	lineCount := 0
	lastSize := fileInfo.Size()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		lineCount++
		if line == "" {
			continue
		}

		// Parse and output the entry
		if err := parseLine(line, &metadata); err != nil {
			reportErrorAndDieS(fmt.Sprintf("Failed to parse line %d: %v", lineCount, err))
		}
	}

	if err := scanner.Err(); err != nil {
		reportErrorAndDieS(fmt.Sprintf("Error reading file: %v", err))
	}

	// Start polling for new content
	for {
		time.Sleep(tailPollInterval)

		// Check current file size
		fileInfo, err := os.Stat(chroniclePath)
		if err != nil {
			if os.IsNotExist(err) {
				reportErrorAndDieS("Chronicle file was deleted")
			}
			reportErrorAndDieS(fmt.Sprintf("Failed to stat file: %v", err))
		}

		currentSize := fileInfo.Size()

		// Check for truncation
		if currentSize < lastSize {
			reportErrorAndDieS("Chronicle file was truncated")
		}

		// Check if there's new data
		if currentSize > lastSize {
			// Read new content
			newScanner := bufio.NewScanner(file)
			for newScanner.Scan() {
				line := newScanner.Text()
				lineCount++
				if line == "" {
					continue
				}

				// Parse and output the entry
				if err := parseLine(line, &metadata); err != nil {
					reportErrorAndDieS(fmt.Sprintf("Failed to parse line %d: %v", lineCount, err))
				}
			}

			if err := newScanner.Err(); err != nil {
				reportErrorAndDieS(fmt.Sprintf("Error reading new content: %v", err))
			}

			// Update size tracking
			lastSize = currentSize
		}
	}
}

// parseLine parses a single JSONL line and outputs it as Markdown.
// Updates metadata pointer if it encounters a metadata line.
func parseLine(line string, metadata **chronicle.Metadata) error {
	// Determine type
	var typeCheck struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal([]byte(line), &typeCheck); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	switch typeCheck.Type {
	case "metadata":
		var m chronicle.Metadata
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			return fmt.Errorf("failed to parse metadata: %w", err)
		}
		*metadata = &m
		outputMetadataMarkdown(&m)

	case "turn":
		var t chronicle.Turn
		if err := json.Unmarshal([]byte(line), &t); err != nil {
			return fmt.Errorf("failed to parse turn: %w", err)
		}
		outputTurnMarkdown(&t)

	default:
		return fmt.Errorf("unknown entry type: %s", typeCheck.Type)
	}

	return nil
}

// outputMetadataMarkdown outputs metadata as Markdown header.
func outputMetadataMarkdown(m *chronicle.Metadata) {
	fmt.Printf("# Simulation Chronicle: %s\n\n", m.Scenario)
	fmt.Printf("**Simulation ID:** `%s`  \n", m.SimulationID)
	fmt.Printf("**Location:** %s  \n", m.Location)
	fmt.Printf("**Time:** %s  \n", m.Time)
	if m.Atmosphere != "" {
		fmt.Printf("**Atmosphere:** %s  \n", m.Atmosphere)
	}
	fmt.Printf("**Started:** %s  \n", m.StartTime.Format("2006-01-02 15:04:05"))
	fmt.Println()
	fmt.Println("---")
	fmt.Println()
}

// outputTurnMarkdown outputs a turn as Markdown.
func outputTurnMarkdown(t *chronicle.Turn) {
	fmt.Printf("## Turn %d\n\n", t.Number)

	for _, event := range t.Events {
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
	outputMetadataMarkdown(metadata)

	// Duration (we know total turns when exporting)
	fmt.Printf("**Duration:** %d turns  \n", len(turns))
	fmt.Println()
	fmt.Println("---")
	fmt.Println()

	// Turns
	for _, turn := range turns {
		outputTurnMarkdown(&turn)
	}
}
