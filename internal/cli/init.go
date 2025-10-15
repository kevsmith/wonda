package cli

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
)

const (
	emptyProviderConfig = `# Wonda Providers Configuration
# API Key Environment Variables:
# If api_key is not specified, Wonda looks for <PROVIDER_NAME>_API_KEY
# where <PROVIDER_NAME> is transformed from the provider name using:
#   1. Convert to uppercase
#   2. Replace dashes (-) with underscores (_)
#   3. Replace spaces with underscores (_)
#
# Examples:
#   [providers.anthropic] → ANTHROPIC_API_KEY
#   [providers.ollama-local] → OLLAMA_LOCAL_API_KEY
#   [providers."my provider"] → MY_PROVIDER_API_KEY
#
# Provider Naming Requirements:
#   - Must start with an alphabetic character (a-z, A-Z)
#   - Can contain alphanumeric characters, dashes, and underscores`
	claudeSonnetConfig = `# Anthropic Claude 3.5 Sonnet
# Thinking extraction is auto-detected based on model name
name = "claude-3-5-sonnet-20241022"
provider = "anthropic"`
	chatGPT4TurboConfg = `# OpenAI GPT-4 Turbo
# No thinking parser needed - auto-detected as "none"
name = "gpt-4-turbo"
provider = "openai"`
)

var modelConfigs = []string{"claude-sonnet.toml", "gpt4-turbo.toml"}
var subdirs = []string{"models", "characters", "scenarios"}

var initCommand = &cobra.Command{
	Use:   "init",
	Short: "Initialize new Wonda configuration",
	Run:   initConfig,
}

func createStructure() {
	info, err := os.Stat(configDir)
	if err != nil {
		if !os.IsNotExist(err) {
			reportErrorAndDie(err)
		}
	}
	if info == nil {
		if err := os.MkdirAll(configDir, 0744); err != nil {
			reportErrorAndDie(err)
		}
	}
	for _, subdir := range subdirs {
		fullSubdir := path.Join(configDir, subdir)
		info, err = os.Stat(fullSubdir)
		if err != nil {
			if !os.IsNotExist(err) {
				reportErrorAndDieP(fullSubdir, err)
			}
		}
		if info == nil {
			if err := os.Mkdir(fullSubdir, 0744); err != nil {
				reportErrorAndDieP(fullSubdir, err)
			}
		}
	}
}

func createPlaceholders() {
	// providers.toml
	tomlFile := path.Join(configDir, "providers.toml")
	if _, err := os.Stat(tomlFile); err != nil {
		if os.IsNotExist(err) {
			if err := os.WriteFile(tomlFile, []byte(emptyProviderConfig), 0644); err != nil {
				reportErrorAndDieP(tomlFile, err)
			}
		} else {
			reportErrorAndDieP(tomlFile, err)
		}
	} else {
		reportWarning(fmt.Sprintf("skipped existing file %s", tomlFile))
	}
	modelsDir := path.Join(configDir, "models")
	for _, modelFile := range modelConfigs {
		fileContents := ""
		switch modelFile {
		case "claude-sonnet.toml":
			modelFile = path.Join(modelsDir, modelFile)
			fileContents = claudeSonnetConfig
		case "gpt4-turbo.toml":
			modelFile = path.Join(modelsDir, modelFile)
			fileContents = chatGPT4TurboConfg
		}
		if fileContents != "" {
			if _, err := os.Stat(modelFile); err != nil {
				if os.IsNotExist(err) {
					if err := os.WriteFile(modelFile, []byte(fileContents), 0644); err != nil {
						reportErrorAndDieP(modelFile, err)
					}
				} else {
					reportErrorAndDieP(modelFile, err)
				}
			} else {
				reportWarning(fmt.Sprintf("skipped existing file %s", modelFile))
			}
		}
	}
}

func initConfig(cmd *cobra.Command, args []string) {
	createStructure()
	createPlaceholders()
	reportSuccess("Done")
}
