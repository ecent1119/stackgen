# Disclaimer

## Intended Use

stackgen is a **local development environment scaffolding tool**. It generates Docker Compose configurations, Dockerfiles, environment files, and test scaffolding to help developers quickly set up local development and testing environments.

## Explicit Scope

stackgen is intentionally scoped for **local development and testing**.

It provides **production-aligned reference configurations**, but it is **not** a production deployment tool.

## Not For Production

**stackgen-generated configurations are for local development and testing only.**

The generated files are:
- **Reference configurations** — starting points that should be reviewed and customized
- **Production-aligned defaults** — sensible choices that may not fit all use cases
- **Development-focused** — optimized for developer experience, not production workloads

Before using any generated configuration in production:
1. Review all files and understand what they do
2. Remove or change development-specific settings
3. Implement proper secrets management
4. Add appropriate security configurations
5. Test thoroughly in staging environments

## What This Is Not

* Not a hosting platform
* Not a deployment service
* Not a compliance or security guarantee
* Not a managed environment

You own and control all generated output.

## No Warranties

stackgen is provided "as is" without warranty of any kind. The authors and distributors:

- Make no claims about correctness, completeness, or fitness for purpose
- Are not responsible for any issues arising from use of generated configurations
- Do not guarantee compatibility with any specific Docker, platform, or infrastructure version

## Third-Party Software

stackgen generates configurations for third-party software including:

| Software | Edition | License |
|----------|---------|---------|
| PostgreSQL | Official Docker Image | PostgreSQL License |
| MySQL | Official Docker Image | GPL v2 |
| SQL Server | **Developer Edition** | Microsoft EULA (Development use only) |
| Neo4j | **Community Edition** | GPL v3 |
| Redis | Community | BSD 3-Clause |
| Redis Stack | Community | Multiple (see Redis docs) |

**You are responsible for:**
- Understanding and complying with all applicable licenses
- Ensuring appropriate editions are used for your use case
- Obtaining proper licenses for any production or commercial use

## SQL Server Developer Edition

The generated SQL Server configuration uses **Developer Edition**, which is:
- Free for development and testing
- NOT licensed for production use
- Subject to Microsoft's licensing terms

For production SQL Server, you must obtain appropriate licensing from Microsoft.

## Neo4j Community Edition

The generated Neo4j configuration uses **Community Edition**, which is:
- Open source under GPL v3
- Has feature limitations compared to Enterprise Edition
- Subject to Neo4j's licensing terms

For enterprise features, you must obtain appropriate licensing from Neo4j.

## Limitation of Liability

In no event shall the authors, copyright holders, or distributors be liable for any claim, damages, or other liability arising from the use of stackgen or any generated configurations.

---

By using stackgen, you acknowledge that you have read and understood this disclaimer.
