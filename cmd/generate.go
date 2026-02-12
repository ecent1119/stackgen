package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/stackgen-cli/stackgen/internal/generator"
	"github.com/stackgen-cli/stackgen/internal/models"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate configuration from stackgen.yaml",
	Long: `Generate Docker Compose configuration from an existing stackgen.yaml file.

This command reads the configuration from stackgen.yaml and generates
the Docker Compose files, Dockerfiles, and environment files.

Examples:
  stackgen generate                           # Generate from ./stackgen.yaml
  stackgen generate --config my.yaml          # Use custom config file
  stackgen generate --dry-run                 # Preview without writing files
  stackgen generate --force                   # Overwrite existing files
  stackgen generate --compose-out custom.yml  # Custom compose output path`,
	RunE: runGenerate,
}

func init() {
	rootCmd.AddCommand(generateCmd)
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// Find config file
	configPath := cfgFile
	if configPath == "" {
		configPath = "stackgen.yaml"
	}

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s\nRun 'stackgen init' to create a new configuration", configPath)
	}

	// Read config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var project models.Project
	if err := yaml.Unmarshal(data, &project); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	color.Cyan("🔧 Generating from %s...\n", configPath)

	// Generate
	gen := generator.New(&project)
	output, err := gen.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate configuration: %w", err)
	}

	// Output
	if dryRun {
		color.Yellow("\n📋 Dry run - previewing generated files:\n")
		output.Print()
		return nil
	}

	// Determine output directory
	outputDir := project.OutputDir
	if composeOut != "" {
		outputDir = filepath.Dir(composeOut)
	}
	if outputDir == "" {
		outputDir = "."
	}
	absOutput, _ := filepath.Abs(outputDir)

	// Check for existing files and prompt if --force not set
	if !forceWrite {
		composeFileName := "docker-compose.yml"
		if composeOut != "" {
			composeFileName = filepath.Base(composeOut)
		}
		composePath := filepath.Join(absOutput, composeFileName)
		if _, err := os.Stat(composePath); err == nil {
			prompt := promptui.Prompt{
				Label:     fmt.Sprintf("File %s exists. Overwrite", composePath),
				IsConfirm: true,
			}
			_, err := prompt.Run()
			if err != nil {
				color.Yellow("Cancelled.")
				return nil
			}
		}
	}
	
	if err := output.WriteToDir(absOutput); err != nil {
		return fmt.Errorf("failed to write files: %w", err)
	}

	color.Green("\n✅ Configuration regenerated successfully!\n")
	
	return nil
}
