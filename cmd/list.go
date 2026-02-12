package cmd

import (
	"fmt"

	"github.com/stackgen-cli/stackgen/internal/models"
	"github.com/stackgen-cli/stackgen/internal/profiles"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list [datastores|runtimes|profiles]",
	Short: "List available datastores, runtimes, or profiles",
	Long: `List available components for stackgen configuration.

Examples:
  stackgen list datastores  # Show all available datastores
  stackgen list runtimes    # Show all available runtimes
  stackgen list profiles    # Show all preset profiles
  stackgen list             # Show everything`,
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		listDatastores()
		fmt.Println()
		listRuntimes()
		fmt.Println()
		listProfiles()
		return nil
	}

	switch args[0] {
	case "datastores", "datastore", "ds":
		listDatastores()
	case "runtimes", "runtime", "rt":
		listRuntimes()
	case "profiles", "profile", "p":
		listProfiles()
	default:
		return fmt.Errorf("unknown category: %s. Use: datastores, runtimes, or profiles", args[0])
	}

	return nil
}

func listDatastores() {
	color.Cyan("📦 Available Datastores:\n\n")
	
	fmt.Printf("  %-15s %-35s %-10s %s\n", 
		color.HiWhiteString("TYPE"), 
		color.HiWhiteString("DESCRIPTION"),
		color.HiWhiteString("PORT"),
		color.HiWhiteString("EDITION"))
	fmt.Println("  " + color.HiBlackString("─────────────────────────────────────────────────────────────────────────"))

	for _, dsType := range models.AvailableDatastores() {
		info := models.GetDatastoreInfo(dsType)
		fmt.Printf("  %-15s %-35s %-10d %s\n",
			color.YellowString(string(dsType)),
			info.Description,
			info.DefaultPort,
			color.HiBlackString(info.Edition))
	}
}

func listRuntimes() {
	color.Cyan("⚡ Available Runtimes:\n\n")
	
	fmt.Printf("  %-12s %-30s %-10s %s\n",
		color.HiWhiteString("TYPE"),
		color.HiWhiteString("DESCRIPTION"),
		color.HiWhiteString("PORT"),
		color.HiWhiteString("FRAMEWORKS"))
	fmt.Println("  " + color.HiBlackString("─────────────────────────────────────────────────────────────────────────"))

	for _, rtType := range models.AvailableRuntimes() {
		info := models.GetRuntimeInfo(rtType)
		frameworks := ""
		for i, fw := range info.Frameworks {
			if i > 0 {
				frameworks += ", "
			}
			frameworks += fw
		}
		fmt.Printf("  %-12s %-30s %-10d %s\n",
			color.YellowString(string(rtType)),
			info.Description,
			info.DefaultPort,
			color.HiBlackString(frameworks))
	}
}

func listProfiles() {
	color.Cyan("🎯 Available Profiles:\n\n")
	
	fmt.Printf("  %-18s %s\n",
		color.HiWhiteString("PROFILE"),
		color.HiWhiteString("DESCRIPTION"))
	fmt.Println("  " + color.HiBlackString("─────────────────────────────────────────────────────────────────────────"))

	for _, profile := range profiles.AvailableProfiles() {
		fmt.Printf("  %-18s %s\n",
			color.YellowString(profile.Name),
			profile.Description)
		
		// Show components
		var components []string
		for _, ds := range profile.Datastores {
			components = append(components, string(ds))
		}
		for _, rt := range profile.Runtimes {
			components = append(components, string(rt.Type))
		}
		fmt.Printf("  %-18s %s\n", "", color.HiBlackString("→ "+joinComponents(components)))
	}
	
	fmt.Println()
	color.HiBlackString("  Use: stackgen init --profile <name>")
}

func joinComponents(components []string) string {
	result := ""
	for i, c := range components {
		if i > 0 {
			result += " + "
		}
		result += c
	}
	return result
}
