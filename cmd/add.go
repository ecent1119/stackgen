package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/stackgen-cli/stackgen/internal/generator"
	"github.com/stackgen-cli/stackgen/internal/models"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var addCmd = &cobra.Command{
	Use:   "add [datastore|runtime] [type]",
	Short: "Add a datastore or runtime to existing configuration",
	Long: `Add a new datastore or runtime to an existing stackgen configuration.

Examples:
  stackgen add datastore postgres    # Add PostgreSQL
  stackgen add datastore redis       # Add Redis
  stackgen add runtime node          # Add Node.js runtime
  stackgen add runtime go            # Add Go runtime
  stackgen add                       # Interactive mode`,
	RunE: runAdd,
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	// Find config file
	configPath := cfgFile
	if configPath == "" {
		configPath = "stackgen.yaml"
	}

	// Check if stackgen.yaml exists, if not check for docker-compose.yml
	var project *models.Project
	
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Try to infer from docker-compose.yml
		if _, err := os.Stat("docker-compose.yml"); os.IsNotExist(err) {
			return fmt.Errorf("no configuration found. Run 'stackgen init' first")
		}
		// Create minimal project from directory name
		cwd, _ := os.Getwd()
		project = &models.Project{
			Name:      filepath.Base(cwd),
			OutputDir: ".",
		}
	} else {
		// Read existing config
		data, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		project = &models.Project{}
		if err := yaml.Unmarshal(data, project); err != nil {
			return fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Interactive or argument-based
	if len(args) < 2 {
		return interactiveAdd(project, configPath)
	}

	category := strings.ToLower(args[0])
	typeName := strings.ToLower(args[1])

	switch category {
	case "datastore", "ds", "d":
		return addDatastore(project, configPath, models.DatastoreType(typeName))
	case "runtime", "rt", "r":
		return addRuntime(project, configPath, models.RuntimeType(typeName))
	default:
		return fmt.Errorf("unknown category: %s. Use: datastore or runtime", category)
	}
}

func interactiveAdd(project *models.Project, configPath string) error {
	prompt := promptui.Select{
		Label: "What do you want to add?",
		Items: []string{"Datastore", "Runtime"},
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return err
	}

	if idx == 0 {
		// Add datastore
		items := make([]string, 0)
		for _, ds := range models.AvailableDatastores() {
			info := models.GetDatastoreInfo(ds)
			items = append(items, fmt.Sprintf("%s - %s [%s]", ds, info.Description, info.Edition))
		}

		dsPrompt := promptui.Select{
			Label: "Select datastore",
			Items: items,
		}

		dsIdx, _, err := dsPrompt.Run()
		if err != nil {
			return err
		}
		return addDatastore(project, configPath, models.AvailableDatastores()[dsIdx])
	} else {
		// Add runtime
		items := make([]string, 0)
		for _, rt := range models.AvailableRuntimes() {
			info := models.GetRuntimeInfo(rt)
			items = append(items, fmt.Sprintf("%s - %s", rt, info.Description))
		}

		rtPrompt := promptui.Select{
			Label: "Select runtime",
			Items: items,
		}

		rtIdx, _, err := rtPrompt.Run()
		if err != nil {
			return err
		}
		return addRuntime(project, configPath, models.AvailableRuntimes()[rtIdx])
	}
}

func addDatastore(project *models.Project, configPath string, dsType models.DatastoreType) error {
	// Check if already exists
	for _, ds := range project.Datastores {
		if ds.Type == dsType {
			return fmt.Errorf("%s is already in the configuration", dsType)
		}
	}

	info := models.GetDatastoreInfo(dsType)
	
	// Find available port
	port := info.DefaultPort
	usedPorts := make(map[int]bool)
	for _, ds := range project.Datastores {
		usedPorts[ds.Port] = true
	}
	for usedPorts[port] {
		port++
	}

	ds := models.Datastore{
		Type:         dsType,
		Name:         string(dsType),
		Port:         port,
		InternalPort: info.DefaultPort,
		Tag:          getDefaultTag(dsType),
	}
	project.Datastores = append(project.Datastores, ds)

	// Save and regenerate
	if err := saveAndRegenerate(project, configPath); err != nil {
		return err
	}

	color.Green("✅ Added %s (port %d)\n", info.DisplayName, port)
	return nil
}

func addRuntime(project *models.Project, configPath string, rtType models.RuntimeType) error {
	info := models.GetRuntimeInfo(rtType)
	
	// Select framework if multiple available
	framework := info.Frameworks[0]
	if len(info.Frameworks) > 1 {
		prompt := promptui.Select{
			Label: "Select framework",
			Items: info.Frameworks,
		}
		_, framework, _ = prompt.Run()
	}

	// Check for duplicate name
	baseName := string(rtType) + "-app"
	name := baseName
	counter := 1
	for {
		exists := false
		for _, rt := range project.Runtimes {
			if rt.Name == name {
				exists = true
				break
			}
		}
		if !exists {
			break
		}
		counter++
		name = fmt.Sprintf("%s-%d", baseName, counter)
	}

	// Find available port
	port := info.DefaultPort
	usedPorts := make(map[int]bool)
	for _, rt := range project.Runtimes {
		usedPorts[rt.Port] = true
	}
	for usedPorts[port] {
		port += 1000
	}

	// Build depends_on from datastores
	var dependsOn []string
	for _, ds := range project.Datastores {
		dependsOn = append(dependsOn, ds.Name)
	}

	rt := models.Runtime{
		Type:         rtType,
		Name:         name,
		Framework:    framework,
		Port:         port,
		InternalPort: info.DefaultPort,
		BuildContext: name,
		Dockerfile:   "Dockerfile",
		DependsOn:    dependsOn,
	}
	project.Runtimes = append(project.Runtimes, rt)

	// Save and regenerate
	if err := saveAndRegenerate(project, configPath); err != nil {
		return err
	}

	color.Green("✅ Added %s [%s] (port %d)\n", info.DisplayName, framework, port)
	return nil
}

func saveAndRegenerate(project *models.Project, configPath string) error {
	// Save stackgen.yaml
	data, err := yaml.Marshal(project)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Regenerate
	gen := generator.New(project)
	output, err := gen.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate configuration: %w", err)
	}

	outputDir := project.OutputDir
	if outputDir == "" {
		outputDir = "."
	}
	absOutput, _ := filepath.Abs(outputDir)
	
	return output.WriteToDir(absOutput)
}
