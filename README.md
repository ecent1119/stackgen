# stackgen — Local Dev Stack Generator (CLI/TUI)

**stackgen** is a free, open source CLI/TUI tool for generating **local development and test Docker Compose environments** using **production-aligned reference configurations**.

It helps developers spin up realistic multi-service stacks quickly, without hand-writing compose files, Dockerfiles, or environment wiring.

No SaaS. No accounts. No background services.

---

## What stackgen does

stackgen interactively generates a ready-to-use local dev stack by:

* Selecting services (databases, caches, runtimes)
* Generating Docker Compose files
* Creating language-specific build/runtime containers
* Wiring services together via a generated `.env`
* Targeting an existing project folder or creating a new one
* **Generating test containers and test function scaffolding**

The result is a **deterministic, editable baseline** you fully control.

---

## Installation

### Download

Download the appropriate binary for your platform:

| Platform | File |
|----------|------|
| **Linux x64** | `stackgen-VERSION-linux-amd64.zip` |
| **Windows x64** | `stackgen-VERSION-windows-amd64.zip` |
| **macOS Silicon (M1/M2/M3)** | `stackgen-VERSION-darwin-arm64.zip` |
| **macOS Intel** | `stackgen-VERSION-darwin-amd64.zip` |

### Install

**macOS / Linux:**
```bash
unzip stackgen-*.zip
chmod +x stackgen
sudo mv stackgen /usr/local/bin/
```

**Windows:**
Extract the zip and add the directory to your PATH, or run directly.

---

## Quick Start

```bash
# Interactive mode - launches TUI
stackgen init

# Use a preset profile
stackgen init --profile web-app

# Generate test containers
stackgen test

# Preview without writing files
stackgen init --dry-run
```

---

## Supported Services (optional, selectable)

### Datastores

* Neo4j (Community Edition)
* PostgreSQL
* MySQL
* Microsoft SQL Server (Developer Edition)
* Redis
* Redis Stack (Community)

### Application Runtimes

* Go (build + server containers)
* Go test container
* Node.js (framework presets)
* Java
* Rust
* C#
* Python

All services use **official or widely adopted base images**, with clear version pinning and readable defaults.

---

## Generated Output

stackgen generates:

* `docker-compose.yml`
* Optional service-specific Dockerfiles
* `.env` and `.env.example`
* Named networks and volumes
* Predictable service naming
* Sensible local defaults
* **Test containers and test scaffolding** (via `stackgen test`)

Everything is **plain text and editable**.
Nothing is hidden or locked in.

---

## Commands

### `stackgen init`

Initialize a new stack configuration.

```bash
stackgen init                     # Interactive TUI
stackgen init --name myproject    # Specify project name
stackgen init --profile api       # Use preset profile
stackgen init --dry-run           # Preview output
```

### `stackgen test`

Generate test containers and test function scaffolding.

```bash
stackgen test                     # Interactive TUI
stackgen test --runtime go        # Generate Go test container
stackgen test --runtime node      # Generate Node.js test container
stackgen test --runtime python    # Generate Python test container
```

### `stackgen list`

List available components.

```bash
stackgen list                     # Show everything
stackgen list datastores          # Show datastores
stackgen list runtimes            # Show runtimes
stackgen list profiles            # Show preset profiles
```

### `stackgen add`

Add components to existing configuration.

```bash
stackgen add datastore postgres   # Add PostgreSQL
stackgen add runtime node         # Add Node.js
```

---

## Preset Profiles

| Profile | Components |
|---------|------------|
| `web-app` | Node.js + Postgres + Redis |
| `api` | Go + Postgres |
| `ml` | Python + Postgres + Redis |
| `fullstack` | Node + Go + Postgres + Redis + Neo4j |
| `java-enterprise` | Spring Boot + Postgres + Redis |
| `dotnet` | C# + SQL Server |
| `rust-api` | Rust + Postgres + Redis |

---

## Designed For

* Local development
* Integration testing
* Prototyping
* Reference architectures
* Onboarding new projects
* Standardizing dev environments

---

## Explicit Scope

stackgen is intentionally scoped for **local development and testing**.

It provides **production-aligned reference configurations**, but it is **not** a production deployment tool.

Generated configurations **must be reviewed, adapted, and secured** before any production use.

---

## What This Is Not

* Not a hosting platform
* Not a deployment service
* Not a compliance or security guarantee
* Not a managed environment

You own and control all generated output.

---

## Why stackgen Exists

Writing and maintaining compose files, env wiring, and build containers is repetitive, error-prone, and time-consuming.

stackgen compresses that work into minutes while keeping everything transparent and customizable.

You save time.
You keep control.

---

## Installation & Usage

* Free download from GitHub Releases
* No accounts or telemetry
* Works entirely locally
* See installation instructions above

---

## Support This Project

**stackgen is free and open source.**

If this tool saved you time, consider sponsoring:

[![Sponsor on GitHub](https://img.shields.io/badge/Sponsor-❤️-red?logo=github)](https://github.com/sponsors/ecent1119)

Your support helps maintain and improve this tool.

---

## License

MIT License - See [LICENSE](LICENSE) file.

---

## Disclaimer

This product is provided **as-is**, without warranty of any kind.

It is intended as a **reference and productivity tool**.
Users are responsible for reviewing generated configurations and ensuring suitability for their use case.

Upstream software licenses apply to generated configurations and referenced images.

See [DISCLAIMER.md](DISCLAIMER.md) for full details.
