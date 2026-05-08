# Code Review Checklist

Use this checklist for every pull request. P0 items must block merge when violated.

## P0: Absolute Bans

- [ ] Model safety: repository or infrastructure code does not return PO, GORM model, or persistence model directly to handlers or JSON responses.
- [ ] Contract safety: repository, domain, and infrastructure signatures do not depend on DTOs from `internal/api`, `internal/handler`, or `handler/types`.
- [ ] Dependency direction: `pkg/` does not import `internal/`; lower layers do not import upper-layer transport contracts.
- [ ] Password storage: password code uses Argon2id or bcrypt with centralized, versioned parameters and progressive rehash support.
- [ ] Password storage: MD5, SHA-1, SHA-256, or other fast hashes are not used for password storage.
- [ ] Sensitive logging: full request bodies, configuration objects, users, and third-party credential objects are not logged through `zap.Any`, `%+v`, JSON dumps, or equivalent whole-object logging.
- [ ] Sensitive logging: any sensitive log output goes through explicit redacted allowlist structures.

## P1: Default Requirements

- [ ] Traceability: entry requests extract or generate TraceID.
- [ ] Traceability: downstream HTTP/RPC calls propagate TraceID.
- [ ] Traceability: structured logs always include TraceID.
- [ ] Async traceability: background jobs, messages, and scheduled tasks carry TraceID or CorrelationID.
- [ ] Layering: handlers validate input and map DTOs to application commands, queries, criteria, domain objects, or value objects.
- [ ] Layering: application use cases own orchestration and transaction boundaries.
- [ ] Layering: domain code owns business invariants and remains transport-agnostic.
- [ ] Repository ports: domain persistence ports live in domain; complex use-case query ports live in application.
- [ ] Migration safety: DDL changes are repeatable, rollbackable, or compensatable and guarded by version/preflight checks.
- [ ] Migration safety: major schema changes follow additive gray-release steps instead of relying on DDL transaction rollback.
- [ ] Async reliability: tasks that require retry, persistence, traffic smoothing, cross-instance processing, or long execution enter the async system.
- [ ] Idempotency: message and job consumers handle duplicate delivery through idempotency keys, state machines, unique constraints, or deduplication tables.
- [ ] Testing: domain and application logic has fast unit tests with mocks or fakes.
- [ ] Testing: repository and infrastructure behavior is covered by integration tests using real middleware through `testcontainers-go`.
- [ ] Testing: integration tests are separated from fast tests through build tags or Make targets.
- [ ] Dependency injection: manual DI is used by default; Wire-style compile-time generation is reviewed when adopted.
- [ ] Dependency injection: runtime reflection DI containers are not introduced.

## P2: Recommended Practices

- [ ] Observability: OpenTelemetry is used where cross-service trace, metrics, and logs need correlation.
- [ ] Metrics: Prometheus-compatible metrics are exposed for core runtime and business signals.
- [ ] Mapping: simple DTO/DO/PO mapping is hand-written; high-volume or complex mapping uses compile-time generation.
- [ ] Resource control: high-concurrency local processing uses bounded concurrency.
- [ ] File processing: large files use streaming APIs.
- [ ] File watching: file watchers combine debounce with periodic full scans.
- [ ] Mail: mail sending is hidden behind a `pkg/mailer` capability abstraction.
- [ ] Git Ops: Makefile and pre-commit hooks run formatting, linting, governance checks, and security scans locally.
