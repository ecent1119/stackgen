package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	version    = "1.0.0"
	cfgFile    string
	dryRun     bool
	forceWrite bool
	composeOut string
)

var rootCmd = &cobra.Command{
	Use:   "stackgen",
	Short: "Generate Docker Compose configurations for local development",
	Long: color.New(color.FgCyan).Sprint(`
stackgen - Local Development Environment Generator

Generate Docker Compose configurations with datastores and language
runtimes in seconds. Production-aligned defaults, opinionated structure,
zero vendor lock-in.

`) + color.New(color.FgYellow).Sprint("For local development and testing only.") + `
Generated configurations must be reviewed before any production use.`,
	Version: version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./stackgen.yaml)")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "output to stdout without writing files")
	rootCmd.PersistentFlags().BoolVarP(&forceWrite, "force", "f", false, "overwrite existing files without prompting")
	rootCmd.PersistentFlags().StringVar(&composeOut, "compose-out", "", "output path for docker-compose.yml (default: current directory)")
}
