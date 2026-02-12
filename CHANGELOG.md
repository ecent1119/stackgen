# Changelog

All notable changes to stackgen will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-02-09

### Added

- Initial release of stackgen
- **CLI/TUI interface** with interactive selection
- **Datastores**: PostgreSQL, MySQL, SQL Server (Developer Edition), Neo4j (Community Edition), Redis, Redis Stack
- **Runtimes**: Go, Node.js, Python, Java, Rust, C# / .NET
- **Profiles**: web-app, api, ml, fullstack, java-enterprise, dotnet, rust-api
- **Test generation**: `stackgen test` command with TUI for generating:
  - Test containers (Dockerfile.test)
  - Test compose files (docker-compose.test.yml)
  - Test function scaffolding (unit/integration/e2e)
- **Commands**: init, list, add, generate, test
- **Features**:
  - Interactive TUI selection wizard
  - Profile-based quick setup
  - Auto-generated environment variables
  - Connection string generation
  - Multi-stage Dockerfile templates
  - Health checks for all datastores
  - Non-root container defaults
  - Volume persistence
  - Network isolation
  - `--dry-run` preview mode
- **Platforms**:
  - Linux x64 (amd64)
  - Windows x64 (amd64)
  - macOS Silicon (arm64)
  - macOS Intel (amd64)

### Notes

- For local development and testing only
- SQL Server uses Developer Edition (development use only)
- Neo4j uses Community Edition
- All datastores use official Docker images
