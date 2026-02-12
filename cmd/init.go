package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/stackgen-cli/stackgen/internal/generator"
	"github.com/stackgen-cli/stackgen/internal/models"
	"github.com/stackgen-cli/stackgen/internal/profiles"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var (
	projectName string
	outputDir   string
	profileName string
	skipPrompts bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new stackgen configuration",
	Long: `Initialize a new Docker Compose configuration interactively.

Select datastores (Postgres, MySQL, MSSQL, Neo4j, Redis, Redis Stack)
and runtimes (Go, Node, Python, Java, Rust, C#) to generate a complete
local development environment.

Examples:
  stackgen init                    # Interactive mode
  stackgen init --name myproject   # Specify project name
  stackgen init --profile web-app  # Use a preset profile
  stackgen init --dry-run          # Preview without writing files`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVarP(&projectName, "name", "n", "", "project name (defaults to current directory name)")
	initCmd.Flags().StringVar(&projectName, "stack-name", "", "project name (alias for --name)")
	initCmd.Flags().StringVarP(&outputDir, "output", "o", ".", "output directory")
	initCmd.Flags().StringVarP(&profileName, "profile", "p", "", "use a preset profile (web-app, api, ml, fullstack, etc.)")
	initCmd.Flags().BoolVarP(&skipPrompts, "yes", "y", false, "skip confirmation prompts")
}

func runInit(cmd *cobra.Command, args []string) error {
	// Get project name
	if projectName == "" {
		cwd, _ := os.Getwd()
		projectName = filepath.Base(cwd)

		if !skipPrompts {
			prompt := promptui.Prompt{
				Label:   "Project name",
				Default: projectName,
			}
			result, err := prompt.Run()
			if err != nil {
				return err
			}
			projectName = sanitizeName(result)
		}
	}

	color.Cyan("\n🚀 stackgen - Local Development Environment Generator\n")
	fmt.Printf("   Project: %s\n\n", color.YellowString(projectName))

	var project *models.Project

	// Check if using a profile
	if profileName != "" {
		profile := profiles.GetProfile(profileName)
		if profile == nil {
			return fmt.Errorf("unknown profile: %s. Run 'stackgen list profiles' to see available profiles", profileName)
		}
		project = profiles.BuildProjectFromProfile(profile, projectName, outputDir)
		color.Green("✓ Using profile: %s\n", profile.Name)
		fmt.Printf("  %s\n\n", profile.Description)
	} else {
		// Interactive selection
		var err error
		project, err = interactiveInit(projectName, outputDir)
		if err != nil {
			return err
		}
	}

	// Generate configuration
	gen := generator.New(project)
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

	// Write files
	absOutput, _ := filepath.Abs(outputDir)
	if err := output.WriteToDir(absOutput); err != nil {
		return fmt.Errorf("failed to write files: %w", err)
	}

	// Success message
	color.Green("\n✅ stackgen configuration generated successfully!\n\n")
	fmt.Println("Generated files:")
	fmt.Printf("  • %s\n", color.CyanString("docker-compose.yml"))
	fmt.Printf("  • %s\n", color.CyanString(".env"))
	fmt.Printf("  • %s\n", color.CyanString(".env.example"))
	fmt.Printf("  • %s\n", color.CyanString(".gitignore"))
	for name := range output.Dockerfiles {
		fmt.Printf("  • %s\n", color.CyanString(name+"/Dockerfile"))
	}

	fmt.Println("\nNext steps:")
	color.Yellow("  1. Review the generated .env file and adjust values as needed")
	color.Yellow("  2. Run: docker compose up -d")
	color.Yellow("  3. Check status: docker compose ps")
	fmt.Println()

	color.New(color.FgHiBlack).Println("⚠️  For local development and testing only.")
	color.New(color.FgHiBlack).Println("   Review configurations before any production use.")
	fmt.Println()

	return nil
}

func interactiveInit(name, outDir string) (*models.Project, error) {
	project := &models.Project{
		Name:      name,
		OutputDir: outDir,
	}

	// Select datastores
	color.Cyan("Select datastores:\n")
	selectedDatastores, err := selectDatastores()
	if err != nil {
		return nil, err
	}

	// Configure selected datastores
	portOffset := 0
	for _, dsType := range selectedDatastores {
		info := models.GetDatastoreInfo(dsType)
		ds := models.Datastore{
			Type:         dsType,
			Name:         string(dsType),
			Port:         info.DefaultPort + portOffset,
			InternalPort: info.DefaultPort,
			Tag:          getDefaultTag(dsType),
		}
		project.Datastores = append(project.Datastores, ds)
		fmt.Printf("  ✓ %s (port %d)\n", info.DisplayName, ds.Port)
	}

	// Select runtimes
	fmt.Println()
	color.Cyan("Select runtimes:\n")
	selectedRuntimes, err := selectRuntimes()
	if err != nil {
		return nil, err
	}

	// Configure selected runtimes
	var dependsOn []string
	for _, ds := range project.Datastores {
		dependsOn = append(dependsOn, ds.Name)
	}

	runtimePortOffset := 0
	for _, rtType := range selectedRuntimes {
		info := models.GetRuntimeInfo(rtType)
		
		// Select framework
		framework := selectFramework(rtType, info.Frameworks)
		
		rt := models.Runtime{
			Type:         rtType,
			Name:         string(rtType) + "-app",
			Framework:    framework,
			Port:         info.DefaultPort + runtimePortOffset,
			InternalPort: info.DefaultPort,
			BuildContext: string(rtType) + "-app",
			Dockerfile:   "Dockerfile",
			DependsOn:    dependsOn,
		}
		project.Runtimes = append(project.Runtimes, rt)
		fmt.Printf("  ✓ %s [%s] (port %d)\n", info.DisplayName, framework, rt.Port)
		runtimePortOffset += 1000
	}

	return project, nil
}

func selectDatastores() ([]models.DatastoreType, error) {
	items := []struct {
		Name        string
		Type        models.DatastoreType
		Description string
		Edition     string
	}{
		{"PostgreSQL", models.DatastorePostgres, "Relational database", "Official Image"},
		{"MySQL", models.DatastoreMySQL, "Relational database", "Official Image"},
		{"SQL Server", models.DatastoreMSSQL, "Microsoft SQL Server", "Developer Edition"},
		{"Neo4j", models.DatastoreNeo4j, "Graph database", "Community Edition"},
		{"Redis", models.DatastoreRedis, "In-memory cache", "Community"},
		{"Redis Stack", models.DatastoreRedisStack, "Redis + modules", "Community"},
		{"[Done]", "", "Finish selection", ""},
	}

	var selected []models.DatastoreType
	selectedMap := make(map[models.DatastoreType]bool)

	for {
		var displayItems []string
		for _, item := range items {
			if item.Type == "" {
				displayItems = append(displayItems, color.GreenString(item.Name))
			} else {
				check := "○"
				if selectedMap[item.Type] {
					check = color.GreenString("●")
				}
				edition := ""
				if item.Edition != "" {
					edition = color.HiBlackString(" [%s]", item.Edition)
				}
				displayItems = append(displayItems, fmt.Sprintf("%s %s - %s%s", check, item.Name, item.Description, edition))
			}
		}

		prompt := promptui.Select{
			Label: "Toggle datastores (select [Done] when finished)",
			Items: displayItems,
			Size:  7,
		}

		idx, _, err := prompt.Run()
		if err != nil {
			return nil, err
		}

		if items[idx].Type == "" {
			// Done selected
			break
		}

		// Toggle selection
		dsType := items[idx].Type
		if selectedMap[dsType] {
			delete(selectedMap, dsType)
		} else {
			selectedMap[dsType] = true
			selected = append(selected, dsType)
		}
	}

	// Return in selection order
	var result []models.DatastoreType
	for _, ds := range selected {
		if selectedMap[ds] {
			result = append(result, ds)
		}
	}
	return result, nil
}

func selectRuntimes() ([]models.RuntimeType, error) {
	items := []struct {
		Name        string
		Type        models.RuntimeType
		Description string
	}{
		{"Go", models.RuntimeGo, "Fast, statically typed"},
		{"Node.js", models.RuntimeNode, "JavaScript runtime"},
		{"Python", models.RuntimePython, "Versatile scripting"},
		{"Java", models.RuntimeJava, "Enterprise JVM"},
		{"Rust", models.RuntimeRust, "Memory-safe systems"},
		{"C# / .NET", models.RuntimeCSharp, "Microsoft .NET"},
		{"[Done]", "", "Finish selection"},
	}

	var selected []models.RuntimeType
	selectedMap := make(map[models.RuntimeType]bool)

	for {
		var displayItems []string
		for _, item := range items {
			if item.Type == "" {
				displayItems = append(displayItems, color.GreenString(item.Name))
			} else {
				check := "○"
				if selectedMap[item.Type] {
					check = color.GreenString("●")
				}
				displayItems = append(displayItems, fmt.Sprintf("%s %s - %s", check, item.Name, item.Description))
			}
		}

		prompt := promptui.Select{
			Label: "Toggle runtimes (select [Done] when finished)",
			Items: displayItems,
			Size:  7,
		}

		idx, _, err := prompt.Run()
		if err != nil {
			return nil, err
		}

		if items[idx].Type == "" {
			break
		}

		rtType := items[idx].Type
		if selectedMap[rtType] {
			delete(selectedMap, rtType)
		} else {
			selectedMap[rtType] = true
			selected = append(selected, rtType)
		}
	}

	var result []models.RuntimeType
	for _, rt := range selected {
		if selectedMap[rt] {
			result = append(result, rt)
		}
	}
	return result, nil
}

func selectFramework(rtType models.RuntimeType, frameworks []string) string {
	if len(frameworks) <= 1 {
		if len(frameworks) == 1 {
			return frameworks[0]
		}
		return ""
	}

	info := models.GetRuntimeInfo(rtType)
	prompt := promptui.Select{
		Label: fmt.Sprintf("Select %s framework", info.DisplayName),
		Items: frameworks,
	}

	_, result, err := prompt.Run()
	if err != nil {
		return frameworks[0]
	}
	return result
}

func sanitizeName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")
	return name
}

func getDefaultTag(dsType models.DatastoreType) string {
	tags := map[models.DatastoreType]string{
		models.DatastorePostgres:   "16-alpine",
		models.DatastoreMySQL:      "8.0",
		models.DatastoreMSSQL:      "2022-latest",
		models.DatastoreNeo4j:      "5",
		models.DatastoreRedis:      "7-alpine",
		models.DatastoreRedisStack: "latest",
	}
	return tags[dsType]
}
