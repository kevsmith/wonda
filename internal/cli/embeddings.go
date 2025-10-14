package cli

import (
	"fmt"
	"os"
	"path"

	"github.com/poiesic/wonda/internal/config"
	"github.com/spf13/cobra"
)

var embeddingsCommand = &cobra.Command{
	Use:     "embeddings",
	Short:   "Work with wonda embedding configurations",
	Aliases: []string{"e"},
}

var showEmbeddingCommand = &cobra.Command{
	Use:   "show",
	Short: "View embeddings configuration",
	Run:   showEmbeddings,
}

var listEmbeddingsCommand = &cobra.Command{
	Use:   "list",
	Short: "List all configured embeddings",
	Run:   listEmbeddings,
}

var editEmbeddingsCommand = &cobra.Command{
	Use:   "edit",
	Short: "Open embeddings section in providers.toml in $EDITOR",
	Run:   editEmbeddings,
}

const embeddingsTemplate = `
# Add your embedding configurations below
# Example:
#
# [embeddings.local-nomic]
# provider = "ollama"
# model = "nomic-ai/nomic-embed-text-v1.5-GGUF"
# dimensions = 768
#
# [embeddings.openai-small]
# provider = "openai"
# model = "text-embedding-3-small"
# dimensions = 1536
`

func init() {
	embeddingsCommand.AddCommand(showEmbeddingCommand, listEmbeddingsCommand, editEmbeddingsCommand)
}

func showEmbeddings(cmd *cobra.Command, args []string) {
	tomlFile := path.Join(configDir, "providers.toml")
	contents, err := os.ReadFile(tomlFile)
	if err != nil {
		reportErrorAndDieP(tomlFile, err)
	}

	embeddings, err := config.LoadEmbeddings(contents)
	if err != nil {
		reportErrorAndDieS(fmt.Sprintf("Failed to parse embeddings from %s: %s", tomlFile, err.Error()))
	}

	if len(embeddings.Embeddings) == 0 {
		fmt.Println("No embeddings configured.")
		fmt.Println("\nTo add embeddings, run: wonda embeddings edit")
		fmt.Println("\nExample configuration:")
		fmt.Println(embeddingsTemplate)
		return
	}

	fmt.Printf("Embeddings in %s:\n\n", tomlFile)
	for name, emb := range embeddings.Embeddings {
		fmt.Printf("  [%s]\n", name)
		fmt.Printf("    Provider:   %s\n", emb.Provider)
		fmt.Printf("    Model:      %s\n", emb.Model)
		fmt.Printf("    Dimensions: %d\n", emb.Dimensions)
		fmt.Println()
	}
}

func listEmbeddings(cmd *cobra.Command, args []string) {
	tomlFile := path.Join(configDir, "providers.toml")
	contents, err := os.ReadFile(tomlFile)
	if err != nil {
		reportErrorAndDieP(tomlFile, err)
	}

	embeddings, err := config.LoadEmbeddings(contents)
	if err != nil {
		reportErrorAndDieS(fmt.Sprintf("Failed to parse embeddings: %s", err.Error()))
	}

	if len(embeddings.Embeddings) == 0 {
		fmt.Println("No embeddings configured.")
		fmt.Println("\nRun 'wonda embeddings edit' to add embeddings.")
		return
	}

	fmt.Printf("Configured embeddings:\n\n")
	for name, emb := range embeddings.Embeddings {
		fmt.Printf("  • %s\n", name)
		fmt.Printf("      %s via %s (%dd)\n", emb.Model, emb.Provider, emb.Dimensions)
	}
}

func editEmbeddings(cmd *cobra.Command, args []string) {
	tomlFile := path.Join(configDir, "providers.toml")
	if _, err := os.Stat(tomlFile); err != nil {
		reportErrorAndDieP(tomlFile, err)
	}
	editFile(tomlFile)
}
