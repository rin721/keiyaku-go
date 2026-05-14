---
doc_role: automation_spec
scope: repo
authority_level: derived
owners: [tech-lead]
status: active
effective_date: 2026-05-15
version: 1.0
related_rules: [GOV-P0-001, GOV-P0-002, GOV-P1-002]
read_when: [governance_change, ci_change, review_change, boundary_sensitive]
update_when: [rule_placement_changed, automation_changed, review_checklist_changed, convention_changed]
---

# Automation Matrix

This matrix assigns governance content to the right enforcement surface. Stable policy belongs in governance or convention docs, repeatable checks belong in scripts/lint/CI, and judgment-heavy checks stay in review.

## Layering Rules

| Concern | Source | Enforcement |
| --- | --- | --- |
| Layer responsibilities | `docs/conventions/layering.md` | Review |
| `pkg` must not import `internal` | `docs/conventions/layering.md` | `scripts/check-layering.ps1`, `.golangci.yml` depguard |
| Domain must not import upper layers or adapters | `docs/conventions/layering.md` | `scripts/check-layering.ps1` |
| Application/repository/infrastructure must not import transport packages | `docs/conventions/layering.md` | `scripts/check-layering.ps1`, `.golangci.yml` depguard |
| DTO and persistence model leakage | `docs/conventions/layering.md` | Review first, future static checks when reliable |
| Repository port ownership | `docs/conventions/layering.md` | Review |

## Test Conventions

| Concern | Source | Enforcement |
| --- | --- | --- |
| Domain/application tests stay fast | `docs/architecture/governance.md` | `scripts/check-test-conventions.ps1` for testcontainers imports; review for coverage quality |
| Testcontainers tests must be isolated from default test runs | `docs/architecture/governance.md` | `scripts/check-test-conventions.ps1` checks `integration` build tag |
| Repository/infrastructure integration tests use real middleware | `docs/architecture/governance.md` | `scripts/check-test-conventions.ps1` for tagged files; review for scenario quality |
| Go package detection before lint/test | `docs/governance/exceptions.yaml` | Pending CI/Make update under `EXC-20260515-003` |

## Prompt Boundary

Agent prompts and entry documents should only route tasks to this matrix, the relevant convention document, and the relevant script. They should not restate the detailed layering or testing rules.
