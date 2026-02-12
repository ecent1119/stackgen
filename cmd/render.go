package cmd

import (
	"github.com/spf13/cobra"
)

var renderCmd = &cobra.Command{
	Use:   "render",
	Short: "Generate configuration from stackgen.yaml (alias for generate)",
	Long: `Render Docker Compose configuration from an existing stackgen.yaml file.
This is an alias for the 'generate' command.

Examples:
  stackgen render                    # Render from ./stackgen.yaml
  stackgen render --config my.yaml   # Use custom config file
  stackgen render --dry-run          # Preview without writing files`,
	RunE: runGenerate, // Reuse the same function as generate
}

func init() {
	rootCmd.AddCommand(renderCmd)
}
