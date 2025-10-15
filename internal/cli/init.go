package cli

import (
	"fmt"
	"os"
	"path"

	"github.com/poiesic/wonda/internal/config/templates"
	"github.com/spf13/cobra"
)


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
			providersTemplate, err := templates.FS.ReadFile("providers_template.toml")
			if err != nil {
				reportErrorAndDie(fmt.Errorf("failed to read providers template: %w", err))
			}
			if err := os.WriteFile(tomlFile, providersTemplate, 0644); err != nil {
				reportErrorAndDieP(tomlFile, err)
			}
		} else {
			reportErrorAndDieP(tomlFile, err)
		}
	} else {
		reportWarning(fmt.Sprintf("skipped existing file %s", tomlFile))
	}

	// Example model config
	modelsDir := path.Join(configDir, "models")
	exampleModelPath := path.Join(modelsDir, "example_model.toml")
	if _, err := os.Stat(exampleModelPath); err != nil {
		if os.IsNotExist(err) {
			content, err := templates.FS.ReadFile("model_template.toml")
			if err != nil {
				reportErrorAndDie(fmt.Errorf("failed to read model template: %w", err))
			}
			if err := os.WriteFile(exampleModelPath, content, 0644); err != nil {
				reportErrorAndDieP(exampleModelPath, err)
			}
		} else {
			reportErrorAndDieP(exampleModelPath, err)
		}
	} else {
		reportWarning(fmt.Sprintf("skipped existing file %s", exampleModelPath))
	}
}

func initConfig(cmd *cobra.Command, args []string) {
	createStructure()
	createPlaceholders()
	reportSuccess("Done")
}
