# Keiyaku-Go

Keiyaku-Go is an enterprise-grade blog and content management backend built with Go. The current implementation uses Gin and GORM under a Clean Architecture layout, with clear boundaries between HTTP delivery, application use cases, domain logic, and infrastructure adapters.

## Current Stack

- Web framework: Gin
- Database: MySQL 8.0 with GORM v2
- Cache: Redis with go-redis/v9
- Logging: Uber Zap JSON logs with level control and rotation
- Config: Viper with YAML and environment variable overrides
- AuthN/AuthZ: JWT and Casbin v3
- IDs: Twitter Snowflake IDs for public business identifiers
- Dependency injection: explicit manual constructors only

## Architecture

The dependency direction is:

```text
internal/api -> internal/application -> internal/domain
internal/infrastructure -> application/domain ports
```

HTTP DTOs live near Gin handlers. GORM models live only in infrastructure. Domain entities do not import Gin, GORM, Redis, JWT, Casbin, or any other adapter package.

Main flow:

1. Request enters Gin routes and middleware.
2. Handler validates DTOs and maps them to application commands or queries.
3. Application services orchestrate domain logic and ports.
4. Infrastructure adapters implement ports with GORM, Redis, JWT, Casbin, and Snowflake.
5. Handler returns the unified response shape: `{code:int,msg:string,data:any}`.

More detail: [system design](docs/architecture/system-design.md).

## Directory Guide

```text
cmd/            Process entrypoints, including API server and migration command.
internal/       Private application code following Clean Architecture.
api/            OpenAPI-style public API contract.
configs/        YAML configuration templates.
migrations/     MySQL schema migration scripts.
deployments/    Local and deployment assets such as Docker Compose.
docs/           Architecture, API, database, ADR, and governance documents.
pkg/            Business-agnostic reusable packages.
scripts/        Governance, layering, test, and CI helper scripts.
test/           Cross-module or integration test assets.
```

## Local Development

Start MySQL and Redis:

```powershell
docker compose -f deployments/docker/docker-compose.yml up -d
```

Run database migrations:

```powershell
go run ./cmd/migrate -dsn "keiyaku:keiyaku@tcp(127.0.0.1:3306)/keiyaku?charset=utf8mb4&parseTime=True&loc=UTC"
```

Start the API server:

```powershell
go run ./cmd/api -config configs/config.yaml
```

Health check:

```powershell
Invoke-RestMethod http://127.0.0.1:8080/healthz
```

## Verification

Run Go checks:

```powershell
go test ./...
go vet ./...
```

Run architecture and governance checks:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-layering.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-governance.ps1
```

## Key Documents

- [System design](docs/architecture/system-design.md)
- [HTTP API contract](docs/api/http-api.md)
- [Database schema](docs/architecture/database-schema.md)
- [Gin and GORM ADR](docs/adr/20260515-adopt-gin-gorm-clean-architecture.md)
- [Governance entry](docs/governance/README.md)
- [AI execution protocol](docs/governance/ai-execution.md)

## Current Implementation Status

The first backend slice is in place:

- Auth registration and login use cases
- Current user profile endpoint
- Article create, list, and detail use cases
- Gin middleware chain for trace ID, recovery, logging, CORS, rate limiting, circuit breaking, JWT, and Casbin
- MySQL schema migration for users, roles, permissions, articles, categories, tags, comments, and Casbin rules

