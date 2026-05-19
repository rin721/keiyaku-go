---
state_id: ADR-20260519-PLUGIN-V3-001
doc_role: adr
memory_level: L0
state_scope: global
scope: repo
authority_level: ssot_decision
owners: [tech-lead]
status: accepted
effective_date: 2026-05-19
version: 1.0
related_rules: [GOV-P0-001, GOV-P0-002, GOV-P0-004, GOV-P1-001, GOV-P1-002, GOV-P1-003, GOV-P1-006]
source_of_truth: [docs/adr/20260519-adopt-plugin-v3-contract.md]
derived_from: [docs/adr/20260519-adopt-remote-service-plugin-system.md, docs/adr/20260519-adopt-plugin-v2-breaking-contract.md, docs/architecture/plugin-system.md]
read_when: [boundary_sensitive, migration_sensitive, security_sensitive, pkg_change]
update_when: [default_behavior_changed, convention_changed, adr_accepted, security_policy_changed, migration_policy_changed]
conflict_policy: accepted_adr_defines_plugin_v3_contract
rollback_target: [docs/adr/20260519-adopt-plugin-v2-breaking-contract.md, docs/architecture/plugin-system.md]
verification_target: [scripts/check-layering.ps1, scripts/check-governance-sync.ps1, scripts/check-governance-map.ps1]
change_reason: tighten plugin route ownership, signature canonicalization, outbound policy, and lifecycle behavior
---

# ADR 20260519: Adopt Plugin v3 Contract

- Status: accepted
- Date: 2026-05-19
- Owner: tech-lead

## Context

The v2 remote HTTP plugin contract established per-plugin secrets, explicit `gateway_path`, health checks, audit events, and gateway forwarding. The next production hardening step is to make route ownership and network trust explicit enough that one plugin cannot accidentally or intentionally claim another plugin's path, replay signed gateway requests, or redirect the host into unsafe outbound targets.

Keeping v2 compatibility would leave old canonical strings, route ownership gaps, and weaker scaffold defaults in place. This change therefore adopts a destructive v3 contract.

## Decision

Plugin v3 is not backward compatible with v2.

- Manifest `schema_version` is fixed to `"v3"`; v2 manifests are rejected.
- HMAC canonical string is `method + "\n" + path + "\n" + raw_query + "\n" + timestamp + "\n" + nonce + "\n" + sha256(body)`.
- SDK gateway verification supports an expected plugin key and a nonce replay store. Official scaffolds and the Blog plugin use both.
- SDK provides `LifecycleRunner` for register, heartbeat, backoff with jitter, automatic re-register on 404 or manifest mismatch, and unregister on shutdown.
- `plugins.trusted_plugins.<plugin_key>` includes route prefix, auth policy, method, outbound host/CIDR, loopback, and insecure HTTP policy. `registration_secret` and `gateway_secret` are both required for enabled plugins.
- Default route ownership is `/api/v1/extensions/{plugin_key}`. Cross-plugin path claims require explicit trust config and still conflict with existing claims.
- The v3 migration rebuilds `plugin_*` tables and adds `plugin_route_claims`.
- Registration validates and writes route claims in the same transaction. Exact/prefix and method/ANY overlaps across plugins are rejected.
- Registration, health checks, and gateway forwarding re-check outbound URL policy. Production defaults require HTTPS and reject loopback, link-local, metadata, and private IPs unless explicitly allowed by host/CIDR policy.
- Gateway request and response headers use allowlists. `Authorization` is forwarded only when the route opts in; `Cookie`, `Forwarded`, `X-Real-IP`, and response `Set-Cookie` are not forwarded by default.
- Manifest `openapi_url` is persisted for management details and diagnostics.

## Non-Goals

- No gRPC plugin protocol.
- No WebSocket, SSE, async event plugin, or plugin marketplace.
- No runtime DI container.
- No multi-version canary routing under one plugin key.

## Consequences

Positive outcomes:

- Plugin route ownership becomes deterministic and auditable.
- SDK lifecycle behavior is shared by pluginctl scaffolds and first-party examples.
- Gateway signatures bind the query string and support replay protection.
- Outbound and header policy reduces SSRF and credential-leak risk.

Trade-offs:

- Existing v2 manifests and registry rows must be regenerated and re-registered.
- Operators must explicitly configure local-development exceptions such as loopback and insecure HTTP.
- MySQL migration is destructive for plugin registry tables.

## Verification

- `go test ./cmd/pluginctl ./pkg/plugin ./internal/application/plugin ./internal/api/http/handler ./internal/api/http/router ./internal/infrastructure/config`
- `go test ./...` inside `plugins/blog`
- `go test ./...`
- `./scripts/check-layering.ps1`
- `./scripts/check-governance-sync.ps1`
- `./scripts/check-governance-map.ps1`
