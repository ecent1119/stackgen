package profiles

import "github.com/stackgen-cli/stackgen/internal/models"

// Profile represents a preset configuration
type Profile struct {
	Name        string
	Description string
	Datastores  []models.DatastoreType
	Runtimes    []RuntimeConfig
}

// RuntimeConfig holds runtime configuration for a profile
type RuntimeConfig struct {
	Type      models.RuntimeType
	Framework string
}

// AvailableProfiles returns all preset profiles
func AvailableProfiles() []Profile {
	return []Profile{
		{
			Name:        "web-app",
			Description: "Full-stack web application (Node.js + Postgres + Redis)",
			Datastores:  []models.DatastoreType{models.DatastorePostgres, models.DatastoreRedis},
			Runtimes:    []RuntimeConfig{{Type: models.RuntimeNode, Framework: "express"}},
		},
		{
			Name:        "api",
			Description: "REST API backend (Go + Postgres)",
			Datastores:  []models.DatastoreType{models.DatastorePostgres},
			Runtimes:    []RuntimeConfig{{Type: models.RuntimeGo, Framework: "stdlib"}},
		},
		{
			Name:        "ml",
			Description: "Machine learning / data science (Python + Postgres + Redis)",
			Datastores:  []models.DatastoreType{models.DatastorePostgres, models.DatastoreRedis},
			Runtimes:    []RuntimeConfig{{Type: models.RuntimePython, Framework: "fastapi"}},
		},
		{
			Name:        "fullstack",
			Description: "Complete microservices stack (Node + Go + Postgres + Redis + Neo4j)",
			Datastores:  []models.DatastoreType{models.DatastorePostgres, models.DatastoreRedis, models.DatastoreNeo4j},
			Runtimes: []RuntimeConfig{
				{Type: models.RuntimeNode, Framework: "express"},
				{Type: models.RuntimeGo, Framework: "stdlib"},
			},
		},
		{
			Name:        "java-enterprise",
			Description: "Enterprise Java stack (Spring Boot + Postgres + Redis)",
			Datastores:  []models.DatastoreType{models.DatastorePostgres, models.DatastoreRedis},
			Runtimes:    []RuntimeConfig{{Type: models.RuntimeJava, Framework: "spring-boot"}},
		},
		{
			Name:        "dotnet",
			Description: ".NET Core application (C# + SQL Server)",
			Datastores:  []models.DatastoreType{models.DatastoreMSSQL},
			Runtimes:    []RuntimeConfig{{Type: models.RuntimeCSharp, Framework: "aspnetcore"}},
		},
		{
			Name:        "rust-api",
			Description: "High-performance Rust API (Rust + Postgres + Redis)",
			Datastores:  []models.DatastoreType{models.DatastorePostgres, models.DatastoreRedis},
			Runtimes:    []RuntimeConfig{{Type: models.RuntimeRust, Framework: "actix-web"}},
		},
	}
}

// GetProfile returns a profile by name
func GetProfile(name string) *Profile {
	for _, p := range AvailableProfiles() {
		if p.Name == name {
			return &p
		}
	}
	return nil
}

// BuildProjectFromProfile creates a Project from a profile
func BuildProjectFromProfile(profile *Profile, projectName, outputDir string) *models.Project {
	project := &models.Project{
		Name:      projectName,
		OutputDir: outputDir,
		Profile:   profile.Name,
	}

	// Add datastores with default ports
	portOffset := 0
	for _, dsType := range profile.Datastores {
		info := models.GetDatastoreInfo(dsType)
		ds := models.Datastore{
			Type:         dsType,
			Name:         string(dsType),
			Port:         info.DefaultPort + portOffset,
			InternalPort: info.DefaultPort,
			Tag:          getDefaultTag(dsType),
		}
		project.Datastores = append(project.Datastores, ds)
	}

	// Add runtimes
	runtimePortOffset := 0
	var dependsOn []string
	for _, dsType := range profile.Datastores {
		dependsOn = append(dependsOn, string(dsType))
	}

	for _, rtConfig := range profile.Runtimes {
		info := models.GetRuntimeInfo(rtConfig.Type)
		rt := models.Runtime{
			Type:         rtConfig.Type,
			Name:         string(rtConfig.Type) + "-app",
			Framework:    rtConfig.Framework,
			Port:         info.DefaultPort + runtimePortOffset,
			InternalPort: info.DefaultPort,
			BuildContext: string(rtConfig.Type) + "-app",
			Dockerfile:   "Dockerfile",
			DependsOn:    dependsOn,
		}
		project.Runtimes = append(project.Runtimes, rt)
		runtimePortOffset += 1000
	}

	return project
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
