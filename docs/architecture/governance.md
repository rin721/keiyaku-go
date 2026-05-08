# rinblog Engineering Governance

This document is the source of intent for rinblog's engineering governance. Code, contracts, configuration, migrations, and CI checks are the verifiable facts that must stay aligned with this intent.

## Core Principles

1. **Intent is documentation, facts are code**: ADRs record design intent. Code, API contracts, configuration, database migrations, and CI checks form the final verifiable facts. CI must detect drift between intent and facts.
2. **Tiered governance**: Standards are grouped as P0 absolute bans, P1 default requirements, and P2 recommended practices.
3. **Capabilities over concrete components**: Architecture defines required capability models first, then selects libraries or middleware as replaceable implementations.

## Default Technology Direction

The preferred future backend direction is **Echo + sqlc**, but this repository intentionally starts with governance only. No HTTP API, database schema, Go module, or business code is introduced by this baseline.

Any later implementation must keep framework and database choices behind capability-oriented interfaces. Echo is a transport adapter choice; sqlc is a repository implementation helper. Neither may leak transport DTOs or persistence models across layer boundaries.

## P0: Absolute Bans

Any P0 violation must block merge. Automatically detectable rules must run in CI. Rules that are not yet automatically detectable must stay in the code review checklist and be promoted to static checks over time.

### No Model Penetration

- Persistence models, including PO and GORM models, must never be returned directly to handlers or serialized as JSON responses.
- Repository and infrastructure method signatures must not depend on DTO types.
- Handlers must map DTOs into application commands, queries, criteria, domain objects, or value objects before passing data downward.

### No Reverse Dependency

- `pkg/` contains business-agnostic reusable components and must not import any `internal/` package.
- Domain, repository, and infrastructure code must not depend on transport contracts from `internal/api`, `internal/handler`, or `handler/types`.
- Lower layers must be blind to HTTP, RPC, OpenAPI DTOs, and other external communication contracts.

### No Unsafe Password Hashing

- Password storage must not use MD5, SHA-1, SHA-256, or other fast hash algorithms.
- Password storage must use Argon2id or bcrypt with centralized cost parameters.
- Password hash parameters must be versioned and must support progressive rehash during login.
- SHA-256 or SHA-512 may still be used for non-password security models such as file integrity, ETags, and HMAC.

### No Plaintext Sensitive Data Logging

- The logging system must include a redaction hook or equivalent redaction boundary.
- Code must not dump full request bodies, configuration objects, user objects, or third-party credential objects through `zap.Any`, `fmt %+v`, JSON dumps, or equivalent structured whole-object logging.
- Sensitive data may only be logged through explicit redacted allowlist structures.

## P1: Default Requirements

P1 is the default engineering standard. Any deviation requires an ADR approved by the technical owner.

### Traceability

- Every entry request must extract or generate a TraceID.
- Internal HTTP or RPC calls must propagate TraceID.
- Structured logs must include TraceID.
- Async tasks, messages, and scheduled jobs must carry a TraceID or CorrelationID.

### Pragmatic DDD Layering

- `internal/api` or `internal/handler`: transport protocol parsing, request validation, and DTO assembly.
- `internal/application`: use case orchestration, transaction boundaries, commands, queries, and criteria.
- `internal/domain`: core business logic, domain objects, entities, value objects, and invariants.
- Repository interfaces belong by scenario:
  - domain object persistence ports live in `internal/domain`;
  - complex query or report ports for use cases live in `internal/application`.
- `internal/repository` or infrastructure adapters implement ports and must not define upper-layer business interfaces.
- `pkg`: business-agnostic shared libraries with no inward dependency on `internal`.

### Database Migrations

- DDL changes must be repeatable, rollbackable, or compensatable.
- Migrations must be guarded by a migration version table, duplicate-execution protection, and preflight checks where native idempotency is unavailable.
- Major schema changes must use gray-release patterns such as additive columns, dual writes, backfills, read switching, and later removal of old columns.
- Migrations must never rely on full transactional rollback for MySQL-style DDL.

### Async Task Model

Tasks must enter the async system when they are expected to exceed the API latency budget, must survive process exit, need retry, need traffic smoothing, or require cross-instance consumption.

The underlying task system must provide persistence, retry, dead-letter handling, graceful shutdown, and observability. Business consumers must be idempotent through idempotency keys, state machines, unique constraints, or deduplication tables.

### Test Execution Strategy

- Domain and application logic must be covered by fast unit tests with mocks or fakes for lower layers.
- Repository and infrastructure behavior must be verified against real middleware through `testcontainers-go`.
- Integration tests must be included in CI but separated from fast local tests by build tags or dedicated Make targets such as `go test -tags=integration ./...`.

### Dependency Injection

- Manual dependency injection is the default.
- Compile-time code generation such as Wire may be used for large dependency graphs, with generated code committed and reviewed.
- Runtime reflection dependency injection containers are forbidden.

## P2: Recommended Practices

- Prefer OpenTelemetry for trace, metrics, and log correlation.
- Prefer Prometheus for runtime and business metrics.
- Use hand-written mappers for simple structures; use compile-time generation such as goverter, protoc, or sqlc for frequent or complex mapping.
- Use bounded concurrency for local high-frequency computation.
- Use streaming APIs for very large file import/export.
- Combine debounced filesystem events with periodic full scans for file watching.
- Expose mail through a `pkg/mailer` capability abstraction; production should prefer reputable SMTP relays.
- Use Makefile and pre-commit hooks to keep formatting, linting, static checks, and security scanning close to the developer workflow.

## CI Governance Baseline

CI must include:

- Governance document presence checks.
- Architecture checks that prevent `pkg` from importing `internal` and prevent lower layers from importing transport contracts.
- `golangci-lint` configuration with `errcheck`, `gosec`, `gocritic`, and `wrapcheck`.
- Secret scanning through gitleaks or an equivalent scanner.
- OpenAPI drift checks once API contracts exist.
