package generator

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/stackgen-cli/stackgen/internal/models"
	"github.com/stackgen-cli/stackgen/internal/templates"
	"gopkg.in/yaml.v3"
)

// Generator handles the generation of Docker Compose configurations
type Generator struct {
	project    *models.Project
	compose    *models.ComposeFile
	envVars    []models.EnvVar
	dockerfiles map[string]string
}

// New creates a new Generator
func New(project *models.Project) *Generator {
	return &Generator{
		project:     project,
		compose:     &models.ComposeFile{Services: make(map[string]models.ComposeService)},
		envVars:     []models.EnvVar{},
		dockerfiles: make(map[string]string),
	}
}

// Generate creates all configuration files
func (g *Generator) Generate() (*GeneratedOutput, error) {
	// Initialize networks
	g.compose.Networks = map[string]interface{}{
		g.project.Name + "-network": map[string]string{"driver": "bridge"},
	}
	g.compose.Volumes = make(map[string]interface{})

	networkName := g.project.Name + "-network"

	// Process datastores
	for _, ds := range g.project.Datastores {
		service, envs, err := g.generateDatastoreService(ds, networkName)
		if err != nil {
			return nil, fmt.Errorf("failed to generate datastore %s: %w", ds.Name, err)
		}
		g.compose.Services[ds.Name] = service
		g.envVars = append(g.envVars, envs...)
		
		// Add volume
		volumeName := ds.Name + "-data"
		g.compose.Volumes[volumeName] = map[string]interface{}{}
	}

	// Process runtimes
	for _, rt := range g.project.Runtimes {
		service, envs, dockerfile, err := g.generateRuntimeService(rt, networkName)
		if err != nil {
			return nil, fmt.Errorf("failed to generate runtime %s: %w", rt.Name, err)
		}
		g.compose.Services[rt.Name] = service
		g.envVars = append(g.envVars, envs...)
		if dockerfile != "" {
			g.dockerfiles[rt.Name] = dockerfile
		}
	}

	return g.buildOutput()
}

func (g *Generator) generateDatastoreService(ds models.Datastore, network string) (models.ComposeService, []models.EnvVar, error) {
	var service models.ComposeService
	var envs []models.EnvVar

	volumeName := ds.Name + "-data"
	password := generatePassword(16)

	switch ds.Type {
	case models.DatastorePostgres:
		service = models.ComposeService{
			Image:         "postgres:" + ds.Tag,
			ContainerName: g.project.Name + "-" + ds.Name,
			Ports:         []string{fmt.Sprintf("%d:5432", ds.Port)},
			Volumes:       []string{fmt.Sprintf("%s:/var/lib/postgresql/data", volumeName)},
			Environment: map[string]string{
				"POSTGRES_USER":     "${POSTGRES_USER:-postgres}",
				"POSTGRES_PASSWORD": "${POSTGRES_PASSWORD}",
				"POSTGRES_DB":       "${POSTGRES_DB:-" + g.project.Name + "}",
			},
			Networks: []string{network},
			Restart:  "unless-stopped",
			HealthCheck: &models.ComposeHealth{
				Test:        []string{"CMD-SHELL", "pg_isready -U postgres"},
				Interval:    "10s",
				Timeout:     "5s",
				Retries:     5,
				StartPeriod: "10s",
			},
		}
		envs = []models.EnvVar{
			{Key: "POSTGRES_USER", Value: "postgres", Description: "PostgreSQL username"},
			{Key: "POSTGRES_PASSWORD", Value: password, Description: "PostgreSQL password", Secret: true},
			{Key: "POSTGRES_DB", Value: g.project.Name, Description: "PostgreSQL database name"},
			{Key: "DATABASE_URL", Value: fmt.Sprintf("postgresql://postgres:%s@%s:5432/%s", password, ds.Name, g.project.Name), Description: "PostgreSQL connection string", Secret: true},
		}

	case models.DatastoreMySQL:
		service = models.ComposeService{
			Image:         "mysql:" + ds.Tag,
			ContainerName: g.project.Name + "-" + ds.Name,
			Ports:         []string{fmt.Sprintf("%d:3306", ds.Port)},
			Volumes:       []string{fmt.Sprintf("%s:/var/lib/mysql", volumeName)},
			Environment: map[string]string{
				"MYSQL_ROOT_PASSWORD": "${MYSQL_ROOT_PASSWORD}",
				"MYSQL_DATABASE":      "${MYSQL_DATABASE:-" + g.project.Name + "}",
				"MYSQL_USER":          "${MYSQL_USER:-app}",
				"MYSQL_PASSWORD":      "${MYSQL_PASSWORD}",
			},
			Networks: []string{network},
			Restart:  "unless-stopped",
			HealthCheck: &models.ComposeHealth{
				Test:        []string{"CMD", "mysqladmin", "ping", "-h", "localhost"},
				Interval:    "10s",
				Timeout:     "5s",
				Retries:     5,
				StartPeriod: "30s",
			},
		}
		rootPassword := generatePassword(16)
		envs = []models.EnvVar{
			{Key: "MYSQL_ROOT_PASSWORD", Value: rootPassword, Description: "MySQL root password", Secret: true},
			{Key: "MYSQL_DATABASE", Value: g.project.Name, Description: "MySQL database name"},
			{Key: "MYSQL_USER", Value: "app", Description: "MySQL application user"},
			{Key: "MYSQL_PASSWORD", Value: password, Description: "MySQL application password", Secret: true},
			{Key: "MYSQL_URL", Value: fmt.Sprintf("mysql://app:%s@%s:3306/%s", password, ds.Name, g.project.Name), Description: "MySQL connection string", Secret: true},
		}

	case models.DatastoreMSSQL:
		service = models.ComposeService{
			Image:         "mcr.microsoft.com/mssql/server:" + ds.Tag,
			ContainerName: g.project.Name + "-" + ds.Name,
			Ports:         []string{fmt.Sprintf("%d:1433", ds.Port)},
			Volumes:       []string{fmt.Sprintf("%s:/var/opt/mssql", volumeName)},
			Environment: map[string]string{
				"ACCEPT_EULA":       "Y",
				"MSSQL_SA_PASSWORD": "${MSSQL_SA_PASSWORD}",
				"MSSQL_PID":         "Developer",
			},
			Networks: []string{network},
			Restart:  "unless-stopped",
			HealthCheck: &models.ComposeHealth{
				Test:        []string{"CMD-SHELL", "/opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P \"$MSSQL_SA_PASSWORD\" -Q \"SELECT 1\" || exit 1"},
				Interval:    "10s",
				Timeout:     "5s",
				Retries:     5,
				StartPeriod: "30s",
			},
		}
		// MSSQL requires strong passwords
		strongPassword := generateStrongPassword(16)
		envs = []models.EnvVar{
			{Key: "MSSQL_SA_PASSWORD", Value: strongPassword, Description: "SQL Server SA password (Developer Edition)", Secret: true},
			{Key: "MSSQL_URL", Value: fmt.Sprintf("Server=%s,1433;Database=master;User Id=sa;Password=%s;TrustServerCertificate=True", ds.Name, strongPassword), Description: "SQL Server connection string", Secret: true},
		}

	case models.DatastoreNeo4j:
		service = models.ComposeService{
			Image:         "neo4j:" + ds.Tag + "-community",
			ContainerName: g.project.Name + "-" + ds.Name,
			Ports:         []string{fmt.Sprintf("%d:7474", ds.Port), fmt.Sprintf("%d:7687", ds.Port+213)},
			Volumes: []string{
				fmt.Sprintf("%s:/data", volumeName),
				fmt.Sprintf("%s-logs:/logs", ds.Name),
			},
			Environment: map[string]string{
				"NEO4J_AUTH": "${NEO4J_AUTH:-neo4j/password}",
			},
			Networks: []string{network},
			Restart:  "unless-stopped",
			HealthCheck: &models.ComposeHealth{
				Test:        []string{"CMD-SHELL", "wget --no-verbose --tries=1 --spider http://localhost:7474 || exit 1"},
				Interval:    "10s",
				Timeout:     "5s",
				Retries:     5,
				StartPeriod: "30s",
			},
		}
		g.compose.Volumes[ds.Name+"-logs"] = map[string]interface{}{}
		envs = []models.EnvVar{
			{Key: "NEO4J_AUTH", Value: "neo4j/" + password, Description: "Neo4j authentication (Community Edition)", Secret: true},
			{Key: "NEO4J_URI", Value: fmt.Sprintf("bolt://%s:7687", ds.Name), Description: "Neo4j Bolt connection URI"},
		}

	case models.DatastoreRedis:
		service = models.ComposeService{
			Image:         "redis:" + ds.Tag,
			ContainerName: g.project.Name + "-" + ds.Name,
			Ports:         []string{fmt.Sprintf("%d:6379", ds.Port)},
			Volumes:       []string{fmt.Sprintf("%s:/data", volumeName)},
			Command:       "redis-server --appendonly yes --requirepass ${REDIS_PASSWORD}",
			Networks:      []string{network},
			Restart:       "unless-stopped",
			HealthCheck: &models.ComposeHealth{
				Test:        []string{"CMD", "redis-cli", "--pass", "${REDIS_PASSWORD}", "ping"},
				Interval:    "10s",
				Timeout:     "5s",
				Retries:     5,
				StartPeriod: "5s",
			},
		}
		envs = []models.EnvVar{
			{Key: "REDIS_PASSWORD", Value: password, Description: "Redis password", Secret: true},
			{Key: "REDIS_URL", Value: fmt.Sprintf("redis://:%s@%s:6379", password, ds.Name), Description: "Redis connection string", Secret: true},
		}

	case models.DatastoreRedisStack:
		service = models.ComposeService{
			Image:         "redis/redis-stack:" + ds.Tag,
			ContainerName: g.project.Name + "-" + ds.Name,
			Ports:         []string{fmt.Sprintf("%d:6379", ds.Port), fmt.Sprintf("%d:8001", ds.Port+1622)},
			Volumes:       []string{fmt.Sprintf("%s:/data", volumeName)},
			Environment: map[string]string{
				"REDIS_ARGS": "--requirepass ${REDIS_STACK_PASSWORD}",
			},
			Networks: []string{network},
			Restart:  "unless-stopped",
			HealthCheck: &models.ComposeHealth{
				Test:        []string{"CMD", "redis-cli", "--pass", "${REDIS_STACK_PASSWORD}", "ping"},
				Interval:    "10s",
				Timeout:     "5s",
				Retries:     5,
				StartPeriod: "5s",
			},
		}
		envs = []models.EnvVar{
			{Key: "REDIS_STACK_PASSWORD", Value: password, Description: "Redis Stack password (Community)", Secret: true},
			{Key: "REDIS_STACK_URL", Value: fmt.Sprintf("redis://:%s@%s:6379", password, ds.Name), Description: "Redis Stack connection string", Secret: true},
		}
	}

	return service, envs, nil
}

func (g *Generator) generateRuntimeService(rt models.Runtime, network string) (models.ComposeService, []models.EnvVar, string, error) {
	service := models.ComposeService{
		Build: &models.ComposeBuild{
			Context:    rt.BuildContext,
			Dockerfile: rt.Dockerfile,
		},
		ContainerName: g.project.Name + "-" + rt.Name,
		Ports:         []string{fmt.Sprintf("%d:%d", rt.Port, rt.InternalPort)},
		Volumes:       []string{fmt.Sprintf("./%s:/app", rt.BuildContext)},
		EnvFile:       []string{".env"},
		Networks:      []string{network},
		Restart:       "unless-stopped",
		DependsOn:     rt.DependsOn,
	}

	var envs []models.EnvVar
	var dockerfile string

	switch rt.Type {
	case models.RuntimeGo:
		dockerfile = templates.GoDockerfile(rt.Framework)
		envs = []models.EnvVar{
			{Key: "GO_ENV", Value: "development", Description: "Go environment"},
			{Key: "PORT", Value: fmt.Sprintf("%d", rt.InternalPort), Description: "Application port"},
		}

	case models.RuntimeNode:
		dockerfile = templates.NodeDockerfile(rt.Framework)
		envs = []models.EnvVar{
			{Key: "NODE_ENV", Value: "development", Description: "Node environment"},
			{Key: "PORT", Value: fmt.Sprintf("%d", rt.InternalPort), Description: "Application port"},
		}

	case models.RuntimePython:
		dockerfile = templates.PythonDockerfile(rt.Framework)
		envs = []models.EnvVar{
			{Key: "PYTHON_ENV", Value: "development", Description: "Python environment"},
			{Key: "PORT", Value: fmt.Sprintf("%d", rt.InternalPort), Description: "Application port"},
		}

	case models.RuntimeJava:
		dockerfile = templates.JavaDockerfile(rt.Framework)
		envs = []models.EnvVar{
			{Key: "JAVA_ENV", Value: "development", Description: "Java environment"},
			{Key: "PORT", Value: fmt.Sprintf("%d", rt.InternalPort), Description: "Application port"},
		}

	case models.RuntimeRust:
		dockerfile = templates.RustDockerfile(rt.Framework)
		envs = []models.EnvVar{
			{Key: "RUST_ENV", Value: "development", Description: "Rust environment"},
			{Key: "PORT", Value: fmt.Sprintf("%d", rt.InternalPort), Description: "Application port"},
		}

	case models.RuntimeCSharp:
		dockerfile = templates.CSharpDockerfile(rt.Framework)
		envs = []models.EnvVar{
			{Key: "ASPNETCORE_ENVIRONMENT", Value: "Development", Description: ".NET environment"},
			{Key: "ASPNETCORE_URLS", Value: fmt.Sprintf("http://+:%d", rt.InternalPort), Description: "ASP.NET Core URLs"},
		}
	}

	return service, envs, dockerfile, nil
}

func (g *Generator) buildOutput() (*GeneratedOutput, error) {
	output := &GeneratedOutput{
		Dockerfiles: g.dockerfiles,
	}

	// Generate docker-compose.yml
	composeYAML, err := yaml.Marshal(g.compose)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal compose file: %w", err)
	}
	output.ComposeYAML = addComposeHeader(string(composeYAML))

	// Generate .env
	var envBuilder strings.Builder
	envBuilder.WriteString("# Generated by stackgen - For local development only\n")
	envBuilder.WriteString("# WARNING: Do not commit this file to version control\n\n")
	for _, env := range g.envVars {
		if env.Description != "" {
			envBuilder.WriteString(fmt.Sprintf("# %s\n", env.Description))
		}
		envBuilder.WriteString(fmt.Sprintf("%s=%s\n", env.Key, env.Value))
	}
	output.EnvFile = envBuilder.String()

	// Generate .env.example
	var envExampleBuilder strings.Builder
	envExampleBuilder.WriteString("# Environment variables for stackgen\n")
	envExampleBuilder.WriteString("# Copy this file to .env and fill in the values\n\n")
	for _, env := range g.envVars {
		if env.Description != "" {
			envExampleBuilder.WriteString(fmt.Sprintf("# %s\n", env.Description))
		}
		if env.Secret {
			envExampleBuilder.WriteString(fmt.Sprintf("%s=<your-%s>\n", env.Key, strings.ToLower(strings.ReplaceAll(env.Key, "_", "-"))))
		} else {
			envExampleBuilder.WriteString(fmt.Sprintf("%s=%s\n", env.Key, env.Value))
		}
	}
	output.EnvExampleFile = envExampleBuilder.String()

	// Generate .gitignore
	output.GitIgnore = templates.GitIgnore()

	return output, nil
}

// GeneratedOutput holds all generated files
type GeneratedOutput struct {
	ComposeYAML    string
	EnvFile        string
	EnvExampleFile string
	GitIgnore      string
	Dockerfiles    map[string]string
}

// WriteToDir writes all generated files to the specified directory
func (out *GeneratedOutput) WriteToDir(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	files := map[string]string{
		"docker-compose.yml": out.ComposeYAML,
		".env":               out.EnvFile,
		".env.example":       out.EnvExampleFile,
		".gitignore":         out.GitIgnore,
	}

	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", name, err)
		}
	}

	// Write Dockerfiles
	for name, content := range out.Dockerfiles {
		dockerDir := filepath.Join(dir, name)
		if err := os.MkdirAll(dockerDir, 0755); err != nil {
			return fmt.Errorf("failed to create dockerfile directory: %w", err)
		}
		path := filepath.Join(dockerDir, "Dockerfile")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write Dockerfile for %s: %w", name, err)
		}
	}

	return nil
}

// Print outputs all generated files to stdout (for --dry-run)
func (out *GeneratedOutput) Print() {
	fmt.Println("=== docker-compose.yml ===")
	fmt.Println(out.ComposeYAML)
	fmt.Println("\n=== .env ===")
	fmt.Println(out.EnvFile)
	fmt.Println("\n=== .env.example ===")
	fmt.Println(out.EnvExampleFile)
	for name, content := range out.Dockerfiles {
		fmt.Printf("\n=== %s/Dockerfile ===\n", name)
		fmt.Println(content)
	}
}

func generatePassword(length int) string {
	bytes := make([]byte, length/2)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func generateStrongPassword(length int) string {
	// MSSQL requires uppercase, lowercase, digit, and special char
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%"
	password := make([]byte, length)
	// Ensure at least one of each required type
	password[0] = 'A'
	password[1] = 'a'
	password[2] = '1'
	password[3] = '!'
	
	randBytes := make([]byte, length-4)
	rand.Read(randBytes)
	for i := 4; i < length; i++ {
		password[i] = chars[int(randBytes[i-4])%len(chars)]
	}
	return string(password)
}

func addComposeHeader(yaml string) string {
	header := `# Generated by stackgen - Local Development Environment Generator
# For local development and testing only.
# Review configurations before any production use.
#
# Usage:
#   docker compose up -d      # Start all services
#   docker compose down       # Stop all services
#   docker compose logs -f    # View logs
#

`
	return header + yaml
}

// RenderTemplate renders a template with the given data
func RenderTemplate(tmpl string, data interface{}) (string, error) {
	t, err := template.New("template").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
