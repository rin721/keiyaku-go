---
doc_role: convention
scope: internal
authority_level: binding
owners: [tech-lead]
status: active
effective_date: 2026-05-15
version: 1.0
related_rules: [GOV-P0-001, GOV-P0-002, GOV-P1-002]
read_when: [pkg_change, boundary_sensitive, governance_change]
update_when: [layering_rule_changed, default_rule_changed, adr_accepted, automation_changed]
---

# Layering Convention

This document is the local convention for package boundaries, dependency direction, and DTO isolation.

Repository-wide rules still live in [../architecture/governance.md](/D:/coder/go/keiyaku-go/docs/architecture/governance.md) until `docs/governance/rules.md` is introduced. This document narrows those rules into day-to-day layering guidance.

## Layer Responsibilities

- `internal/api` or `internal/handler`: transport parsing, request validation, response assembly, and DTO adaptation.
- `internal/application`: use-case orchestration, transaction boundaries, commands, queries, criteria, and application-facing ports.
- `internal/domain`: business invariants, entities, value objects, domain services, and domain-facing ports.
- `internal/repository`: persistence adapters that implement ports owned by domain or application.
- `internal/infrastructure`: external system adapters and technical infrastructure.
- `pkg`: reusable, business-agnostic packages that must not depend on repository internals.

## Dependency Direction

- `pkg` must not import any `internal` package.
- `internal/domain` must not import `internal/application`, `internal/repository`, `internal/infrastructure`, `internal/api`, or `internal/handler`.
- `internal/application` must not import transport packages such as `internal/api` or `internal/handler`.
- `internal/repository` and `internal/infrastructure` must not import transport packages such as `internal/api` or `internal/handler`.
- Lower layers must not depend on transport DTOs, OpenAPI contracts, HTTP request models, or handler-local types.

## DTO and Model Boundaries

- Handler code converts transport DTOs into application commands, queries, criteria, domain objects, or value objects before crossing inward.
- Repository and infrastructure method signatures must not depend on transport DTO types.
- Persistence models, including PO and ORM models, must not be returned to handlers or serialized directly as JSON responses.
- Mapping code should live at the boundary where the source and target models meet; avoid letting transport or persistence models leak across multiple layers.

## Repository Port Ownership

- Domain object persistence ports belong in `internal/domain`.
- Use-case-specific complex query or reporting ports belong in `internal/application`.
- Adapter packages implement ports; they do not define upstream business interfaces.

## Automation

Machine-checkable parts of this convention are enforced by [../../scripts/check-layering.ps1](/D:/coder/go/keiyaku-go/scripts/check-layering.ps1) and the `depguard` rules in [../../.golangci.yml](/D:/coder/go/keiyaku-go/.golangci.yml).

Human review is still required for model leakage that cannot be reliably detected by imports alone, such as whether a returned type is effectively a persistence model under a neutral name.
