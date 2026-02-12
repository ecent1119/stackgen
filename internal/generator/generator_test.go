package generator

import (
	"strings"
	"testing"

	"github.com/stackgen-cli/stackgen/internal/models"
)

func TestGeneratePostgres(t *testing.T) {
	project := &models.Project{
		Name:      "testproject",
		OutputDir: ".",
		Datastores: []models.Datastore{
			{
				Type:         models.DatastorePostgres,
				Name:         "postgres",
				Port:         5432,
				InternalPort: 5432,
				Tag:          "16-alpine",
			},
		},
	}

	gen := New(project)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check compose file contains postgres
	if !strings.Contains(output.ComposeYAML, "postgres:16-alpine") {
		t.Error("ComposeYAML should contain postgres:16-alpine")
	}

	// Check env file contains DATABASE_URL
	if !strings.Contains(output.EnvFile, "DATABASE_URL") {
		t.Error("EnvFile should contain DATABASE_URL")
	}

	// Check health check is present
	if !strings.Contains(output.ComposeYAML, "pg_isready") {
		t.Error("ComposeYAML should contain postgres health check")
	}
}

func TestGenerateMultipleDatastores(t *testing.T) {
	project := &models.Project{
		Name:      "multitest",
		OutputDir: ".",
		Datastores: []models.Datastore{
			{
				Type:         models.DatastorePostgres,
				Name:         "postgres",
				Port:         5432,
				InternalPort: 5432,
				Tag:          "16-alpine",
			},
			{
				Type:         models.DatastoreRedis,
				Name:         "redis",
				Port:         6379,
				InternalPort: 6379,
				Tag:          "7-alpine",
			},
		},
	}

	gen := New(project)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(output.ComposeYAML, "postgres:") {
		t.Error("ComposeYAML should contain postgres service")
	}

	if !strings.Contains(output.ComposeYAML, "redis:") {
		t.Error("ComposeYAML should contain redis service")
	}

	if !strings.Contains(output.EnvFile, "REDIS_URL") {
		t.Error("EnvFile should contain REDIS_URL")
	}
}

func TestGenerateRuntime(t *testing.T) {
	project := &models.Project{
		Name:      "runtimetest",
		OutputDir: ".",
		Runtimes: []models.Runtime{
			{
				Type:         models.RuntimeGo,
				Name:         "go-app",
				Framework:    "stdlib",
				Port:         8080,
				InternalPort: 8080,
				BuildContext: "go-app",
				Dockerfile:   "Dockerfile",
			},
		},
	}

	gen := New(project)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check Dockerfile was generated
	if _, ok := output.Dockerfiles["go-app"]; !ok {
		t.Error("Dockerfile should be generated for go-app")
	}

	// Check compose has build context
	if !strings.Contains(output.ComposeYAML, "build:") {
		t.Error("ComposeYAML should contain build configuration")
	}
}

func TestGenerateMSSQL(t *testing.T) {
	project := &models.Project{
		Name:      "mssqltest",
		OutputDir: ".",
		Datastores: []models.Datastore{
			{
				Type:         models.DatastoreMSSQL,
				Name:         "mssql",
				Port:         1433,
				InternalPort: 1433,
				Tag:          "2022-latest",
			},
		},
	}

	gen := New(project)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check MSSQL uses Developer edition
	if !strings.Contains(output.ComposeYAML, "MSSQL_PID: Developer") {
		t.Error("ComposeYAML should specify Developer edition for MSSQL")
	}

	// Check ACCEPT_EULA is set
	if !strings.Contains(output.ComposeYAML, "ACCEPT_EULA: \"Y\"") {
		t.Error("ComposeYAML should have ACCEPT_EULA set")
	}
}

func TestGenerateNeo4j(t *testing.T) {
	project := &models.Project{
		Name:      "neo4jtest",
		OutputDir: ".",
		Datastores: []models.Datastore{
			{
				Type:         models.DatastoreNeo4j,
				Name:         "neo4j",
				Port:         7474,
				InternalPort: 7474,
				Tag:          "5",
			},
		},
	}

	gen := New(project)
	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check Neo4j uses community edition
	if !strings.Contains(output.ComposeYAML, "neo4j:5-community") {
		t.Error("ComposeYAML should use Neo4j community edition")
	}

	// Check bolt port is configured
	if !strings.Contains(output.EnvFile, "NEO4J_URI") {
		t.Error("EnvFile should contain NEO4J_URI")
	}
}

func TestPasswordGeneration(t *testing.T) {
	pw1 := generatePassword(16)
	pw2 := generatePassword(16)

	if pw1 == pw2 {
		t.Error("Passwords should be unique")
	}

	if len(pw1) != 16 {
		t.Errorf("Password length should be 16, got %d", len(pw1))
	}
}

func TestStrongPasswordGeneration(t *testing.T) {
	pw := generateStrongPassword(16)

	if len(pw) != 16 {
		t.Errorf("Password length should be 16, got %d", len(pw))
	}

	// Should contain required character types
	hasUpper := strings.ContainsAny(pw, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	hasLower := strings.ContainsAny(pw, "abcdefghijklmnopqrstuvwxyz")
	hasDigit := strings.ContainsAny(pw, "0123456789")
	hasSpecial := strings.ContainsAny(pw, "!@#$%")

	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		t.Error("Strong password should contain uppercase, lowercase, digit, and special char")
	}
}
