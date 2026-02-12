package models

// Project represents the entire generated configuration
type Project struct {
	Name       string      `yaml:"name"`
	OutputDir  string      `yaml:"output_dir"`
	Datastores []Datastore `yaml:"datastores"`
	Runtimes   []Runtime   `yaml:"runtimes"`
	Networks   []Network   `yaml:"networks"`
	Profile    string      `yaml:"profile,omitempty"`
}

// Datastore represents a database or cache service
type Datastore struct {
	Type        DatastoreType     `yaml:"type"`
	Name        string            `yaml:"name"`
	Image       string            `yaml:"image"`
	Tag         string            `yaml:"tag"`
	Port        int               `yaml:"port"`
	InternalPort int              `yaml:"internal_port"`
	Volumes     []Volume          `yaml:"volumes"`
	Environment map[string]string `yaml:"environment"`
	HealthCheck *HealthCheck      `yaml:"health_check,omitempty"`
	Networks    []string          `yaml:"networks"`
}

// DatastoreType enumerates supported datastores
type DatastoreType string

const (
	DatastorePostgres   DatastoreType = "postgres"
	DatastoreMySQL      DatastoreType = "mysql"
	DatastoreMSSQL      DatastoreType = "mssql"
	DatastoreNeo4j      DatastoreType = "neo4j"
	DatastoreRedis      DatastoreType = "redis"
	DatastoreRedisStack DatastoreType = "redis-stack"
)

// Runtime represents a language/framework container
type Runtime struct {
	Type         RuntimeType       `yaml:"type"`
	Name         string            `yaml:"name"`
	Framework    string            `yaml:"framework,omitempty"`
	Port         int               `yaml:"port"`
	InternalPort int               `yaml:"internal_port"`
	BuildContext string            `yaml:"build_context"`
	Dockerfile   string            `yaml:"dockerfile"`
	Volumes      []Volume          `yaml:"volumes"`
	Environment  map[string]string `yaml:"environment"`
	Command      string            `yaml:"command,omitempty"`
	DependsOn    []string          `yaml:"depends_on"`
	Networks     []string          `yaml:"networks"`
}

// RuntimeType enumerates supported runtimes
type RuntimeType string

const (
	RuntimeGo     RuntimeType = "go"
	RuntimeNode   RuntimeType = "node"
	RuntimePython RuntimeType = "python"
	RuntimeJava   RuntimeType = "java"
	RuntimeRust   RuntimeType = "rust"
	RuntimeCSharp RuntimeType = "csharp"
)

// Volume represents a Docker volume mount
type Volume struct {
	Source   string `yaml:"source"`
	Target   string `yaml:"target"`
	Type     string `yaml:"type"` // bind, volume, tmpfs
	ReadOnly bool   `yaml:"read_only,omitempty"`
}

// Network represents a Docker network
type Network struct {
	Name   string `yaml:"name"`
	Driver string `yaml:"driver"`
}

// HealthCheck represents a Docker health check
type HealthCheck struct {
	Test        []string `yaml:"test"`
	Interval    string   `yaml:"interval"`
	Timeout     string   `yaml:"timeout"`
	Retries     int      `yaml:"retries"`
	StartPeriod string   `yaml:"start_period"`
}

// EnvVar represents an environment variable with metadata
type EnvVar struct {
	Key         string `yaml:"key"`
	Value       string `yaml:"value"`
	Description string `yaml:"description,omitempty"`
	Secret      bool   `yaml:"secret,omitempty"`
}

// ComposeService represents a service in docker-compose.yml
type ComposeService struct {
	Image         string            `yaml:"image,omitempty"`
	Build         *ComposeBuild     `yaml:"build,omitempty"`
	ContainerName string            `yaml:"container_name,omitempty"`
	Ports         []string          `yaml:"ports,omitempty"`
	Volumes       []string          `yaml:"volumes,omitempty"`
	Environment   map[string]string `yaml:"environment,omitempty"`
	EnvFile       []string          `yaml:"env_file,omitempty"`
	DependsOn     []string          `yaml:"depends_on,omitempty"`
	Networks      []string          `yaml:"networks,omitempty"`
	HealthCheck   *ComposeHealth    `yaml:"healthcheck,omitempty"`
	Restart       string            `yaml:"restart,omitempty"`
	Command       string            `yaml:"command,omitempty"`
	User          string            `yaml:"user,omitempty"`
}

// ComposeBuild represents build configuration
type ComposeBuild struct {
	Context    string `yaml:"context"`
	Dockerfile string `yaml:"dockerfile"`
}

// ComposeHealth represents healthcheck in compose format
type ComposeHealth struct {
	Test        []string `yaml:"test"`
	Interval    string   `yaml:"interval"`
	Timeout     string   `yaml:"timeout"`
	Retries     int      `yaml:"retries"`
	StartPeriod string   `yaml:"start_period"`
}

// ComposeFile represents the full docker-compose.yml structure
type ComposeFile struct {
	Version  string                    `yaml:"version,omitempty"`
	Services map[string]ComposeService `yaml:"services"`
	Volumes  map[string]interface{}    `yaml:"volumes,omitempty"`
	Networks map[string]interface{}    `yaml:"networks,omitempty"`
}

// AvailableDatastores returns all supported datastore types
func AvailableDatastores() []DatastoreType {
	return []DatastoreType{
		DatastorePostgres,
		DatastoreMySQL,
		DatastoreMSSQL,
		DatastoreNeo4j,
		DatastoreRedis,
		DatastoreRedisStack,
	}
}

// AvailableRuntimes returns all supported runtime types
func AvailableRuntimes() []RuntimeType {
	return []RuntimeType{
		RuntimeGo,
		RuntimeNode,
		RuntimePython,
		RuntimeJava,
		RuntimeRust,
		RuntimeCSharp,
	}
}

// DatastoreInfo provides metadata about a datastore
type DatastoreInfo struct {
	Type        DatastoreType
	DisplayName string
	Description string
	DefaultPort int
	Edition     string
}

// GetDatastoreInfo returns metadata for a datastore type
func GetDatastoreInfo(t DatastoreType) DatastoreInfo {
	info := map[DatastoreType]DatastoreInfo{
		DatastorePostgres: {
			Type:        DatastorePostgres,
			DisplayName: "PostgreSQL",
			Description: "Powerful open-source relational database",
			DefaultPort: 5432,
			Edition:     "Official Image",
		},
		DatastoreMySQL: {
			Type:        DatastoreMySQL,
			DisplayName: "MySQL",
			Description: "Popular open-source relational database",
			DefaultPort: 3306,
			Edition:     "Official Image",
		},
		DatastoreMSSQL: {
			Type:        DatastoreMSSQL,
			DisplayName: "SQL Server",
			Description: "Microsoft SQL Server (Developer Edition)",
			DefaultPort: 1433,
			Edition:     "Developer Edition - for development use only",
		},
		DatastoreNeo4j: {
			Type:        DatastoreNeo4j,
			DisplayName: "Neo4j",
			Description: "Graph database for connected data",
			DefaultPort: 7474,
			Edition:     "Community Edition",
		},
		DatastoreRedis: {
			Type:        DatastoreRedis,
			DisplayName: "Redis",
			Description: "In-memory data store and cache",
			DefaultPort: 6379,
			Edition:     "Community",
		},
		DatastoreRedisStack: {
			Type:        DatastoreRedisStack,
			DisplayName: "Redis Stack",
			Description: "Redis with JSON, Search, TimeSeries modules",
			DefaultPort: 6379,
			Edition:     "Community",
		},
	}
	return info[t]
}

// RuntimeInfo provides metadata about a runtime
type RuntimeInfo struct {
	Type        RuntimeType
	DisplayName string
	Description string
	DefaultPort int
	Frameworks  []string
}

// GetRuntimeInfo returns metadata for a runtime type
func GetRuntimeInfo(t RuntimeType) RuntimeInfo {
	info := map[RuntimeType]RuntimeInfo{
		RuntimeGo: {
			Type:        RuntimeGo,
			DisplayName: "Go",
			Description: "Fast, statically typed language",
			DefaultPort: 8080,
			Frameworks:  []string{"stdlib", "gin", "fiber", "echo"},
		},
		RuntimeNode: {
			Type:        RuntimeNode,
			DisplayName: "Node.js",
			Description: "JavaScript runtime for server-side",
			DefaultPort: 3000,
			Frameworks:  []string{"express", "fastify", "nextjs", "nestjs"},
		},
		RuntimePython: {
			Type:        RuntimePython,
			DisplayName: "Python",
			Description: "Versatile scripting language",
			DefaultPort: 8000,
			Frameworks:  []string{"fastapi", "flask", "django"},
		},
		RuntimeJava: {
			Type:        RuntimeJava,
			DisplayName: "Java",
			Description: "Enterprise-grade JVM language",
			DefaultPort: 8080,
			Frameworks:  []string{"spring-boot", "quarkus", "micronaut"},
		},
		RuntimeRust: {
			Type:        RuntimeRust,
			DisplayName: "Rust",
			Description: "Memory-safe systems language",
			DefaultPort: 8080,
			Frameworks:  []string{"actix-web", "axum", "rocket"},
		},
		RuntimeCSharp: {
			Type:        RuntimeCSharp,
			DisplayName: "C# / .NET",
			Description: "Microsoft .NET platform",
			DefaultPort: 5000,
			Frameworks:  []string{"aspnetcore", "minimal-api"},
		},
	}
	return info[t]
}
