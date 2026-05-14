---
doc_role: map
scope: repo
authority_level: ssot
owners: [tech-lead]
status: active
effective_date: 2026-05-15
version: 1.0
related_rules: []
read_when: [all_tasks]
update_when: [document_tree_changed, routing_changed, ssot_changed, metadata_standard_changed]
---

# Governance Navigation

This document is the navigation truth for repository governance. It defines what to read first, what to read next, and how to resolve conflicts between governance artifacts.

## SSOT Registry

- Navigation truth: this file
- AI execution truth: [ai-execution.md](/D:/coder/go/keiyaku-go/docs/governance/ai-execution.md)
- Repository-wide governance rules truth, current interim source: [../architecture/governance.md](/D:/coder/go/keiyaku-go/docs/architecture/governance.md)
- Deviation and design-decision truth: Accepted ADRs under [../adr](/D:/coder/go/keiyaku-go/docs/adr)
- Governance exception and debt registry: [exceptions.yaml](/D:/coder/go/keiyaku-go/docs/governance/exceptions.yaml)
- Automation placement matrix: [automation-matrix.md](/D:/coder/go/keiyaku-go/docs/governance/automation-matrix.md)
- Layering convention: [../conventions/layering.md](/D:/coder/go/keiyaku-go/docs/conventions/layering.md)
- Human review truth: [../review/checklist.md](/D:/coder/go/keiyaku-go/docs/review/checklist.md)
- Migration rollout template: [../migrations/gray-release-template.md](/D:/coder/go/keiyaku-go/docs/migrations/gray-release-template.md)

Note: `docs/governance/rules.md` does not exist yet. Until it is introduced, `docs/architecture/governance.md` remains the active repository-wide governance rules source.

## Conflict Priority

Use this order when documents conflict:

1. P0 repository-wide governance rules
2. Accepted ADRs that explicitly apply to the current scope
3. P1 repository-wide governance rules
4. This navigation file for routing and read order
5. AI execution protocol for agent behavior
6. Review checklists and templates

Rules:

- P0 cannot be waived by prompt wording.
- If an Accepted ADR modifies a default P1 rule for a defined scope, the ADR wins inside that scope.
- If two active documents conflict and priority does not resolve it, stop and request human clarification.

## First-Hop Path

Every agent entering the repository should follow this path:

1. Read [AGENTS.md](/D:/coder/go/keiyaku-go/AGENTS.md).
2. Read this file.
3. Read [ai-execution.md](/D:/coder/go/keiyaku-go/docs/governance/ai-execution.md).
4. Classify the task.
5. Load only the routed governance files for that task.
6. Then inspect code, scripts, or workflows.

## Task Routing Matrix

### `pkg/` tools, shared utilities, reusable components

Read in order:

1. [ai-execution.md](/D:/coder/go/keiyaku-go/docs/governance/ai-execution.md)
2. [automation-matrix.md](/D:/coder/go/keiyaku-go/docs/governance/automation-matrix.md)
3. [../architecture/governance.md](/D:/coder/go/keiyaku-go/docs/architecture/governance.md)
4. [../conventions/layering.md](/D:/coder/go/keiyaku-go/docs/conventions/layering.md)
5. Relevant Accepted ADRs if the change alters default package boundaries or reuse rules

Focus:

- `pkg/` must not depend on `internal/`
- Shared code must remain business-agnostic
- If the change proposes a new default `pkg` style, route through ADR first

### Layering, architecture boundaries, DTO/domain/repository interaction

Read in order:

1. [ai-execution.md](/D:/coder/go/keiyaku-go/docs/governance/ai-execution.md)
2. [automation-matrix.md](/D:/coder/go/keiyaku-go/docs/governance/automation-matrix.md)
3. [../architecture/governance.md](/D:/coder/go/keiyaku-go/docs/architecture/governance.md)
4. [../conventions/layering.md](/D:/coder/go/keiyaku-go/docs/conventions/layering.md)
5. [../adr/README.md](/D:/coder/go/keiyaku-go/docs/adr/README.md)
6. Any Accepted ADRs touching boundaries, framework choices, or default abstraction models

Focus:

- Dependency direction
- DTO leakage
- handler/application/domain/repository responsibilities
- Whether the change modifies a default architecture rule or only an implementation

### Governance document, prompt, or policy change

Read in order:

1. [ai-execution.md](/D:/coder/go/keiyaku-go/docs/governance/ai-execution.md)
2. [automation-matrix.md](/D:/coder/go/keiyaku-go/docs/governance/automation-matrix.md)
3. [../architecture/governance.md](/D:/coder/go/keiyaku-go/docs/architecture/governance.md)
4. [../adr/README.md](/D:/coder/go/keiyaku-go/docs/adr/README.md)
5. [../review/checklist.md](/D:/coder/go/keiyaku-go/docs/review/checklist.md)
6. [exceptions.yaml](/D:/coder/go/keiyaku-go/docs/governance/exceptions.yaml)
7. [../../scripts/check-governance.ps1](/D:/coder/go/keiyaku-go/scripts/check-governance.ps1)
8. [../../.github/workflows/governance.yml](/D:/coder/go/keiyaku-go/.github/workflows/governance.yml)

Focus:

- Whether the change creates a new default
- Whether ADR is required
- Whether review checklist and automation must be updated together

### Migration, rollout, gray release, backfill, rollback

Read in order:

1. [ai-execution.md](/D:/coder/go/keiyaku-go/docs/governance/ai-execution.md)
2. [../architecture/governance.md](/D:/coder/go/keiyaku-go/docs/architecture/governance.md)
3. [../migrations/gray-release-template.md](/D:/coder/go/keiyaku-go/docs/migrations/gray-release-template.md)
4. Relevant Accepted ADRs for non-default rollout or schema strategy

Focus:

- forward-compatible migration steps
- pre-checks
- dual write/backfill/read switch
- rollback or compensation strategy

### Async jobs, MQ, retries, idempotency, scheduled tasks

Read in order:

1. [ai-execution.md](/D:/coder/go/keiyaku-go/docs/governance/ai-execution.md)
2. [../architecture/governance.md](/D:/coder/go/keiyaku-go/docs/architecture/governance.md)
3. Relevant Accepted ADRs if the task changes the default async model

Focus:

- persistence
- retries
- dead-letter handling
- idempotency and correlation tracing

### Tests, lint, CI, security scans, governance automation

Read in order:

1. [ai-execution.md](/D:/coder/go/keiyaku-go/docs/governance/ai-execution.md)
2. [automation-matrix.md](/D:/coder/go/keiyaku-go/docs/governance/automation-matrix.md)
3. [../architecture/governance.md](/D:/coder/go/keiyaku-go/docs/architecture/governance.md)
4. [../review/checklist.md](/D:/coder/go/keiyaku-go/docs/review/checklist.md)
5. [../../scripts/check-layering.ps1](/D:/coder/go/keiyaku-go/scripts/check-layering.ps1)
6. [../../scripts/check-test-conventions.ps1](/D:/coder/go/keiyaku-go/scripts/check-test-conventions.ps1)
7. Relevant workflow, lint, or script file

Focus:

- what is policy vs what is automated
- false positives/false negatives
- whether automation matches the current repository state
- whether a convention belongs in docs, review, lint, tests, or a script

## Metadata Standard

All governance-facing documents should include YAML front matter with the following fields.

Required fields:

- `doc_role`: one of `map`, `governance`, `ai_execution`, `convention`, `adr_index`, `adr_record`, `review`, `template`, `automation_spec`, `exception_registry`, `ai_entry`
- `scope`: one of `repo`, `pkg`, `internal`, `migrations`, `ci`, `security`, or a narrower path such as `internal/domain`
- `authority_level`: one of `ssot`, `binding`, `binding_entry`, `record`, `derived`, `template`
- `owners`: array of owner roles or teams
- `status`: one of `draft`, `active`, `deprecated`, `superseded`
- `related_rules`: array of rule ids such as `GOV-P0-001`
- `read_when`: array of task triggers such as `all_tasks`, `governance_change`, `migration_change`, `pkg_change`, `ci_change`
- `update_when`: array of change triggers such as `routing_changed`, `default_rule_changed`, `adr_accepted`, `automation_changed`

At least one of these must be present:

- `version`
- `effective_date`

Recommended fields:

- `supersedes`
- `superseded_by`
- `conflict_priority`

Interpretation rules:

- Agents should use `doc_role`, `scope`, and `read_when` for discovery.
- Agents should use `authority_level` to decide whether a document is truth, a binding derivative, or a record.
- `related_rules` should be used to connect governance docs, ADRs, review checklists, and automation.
- `update_when` should drive governance maintenance when defaults change.

## Maintenance Rules

- If routing changes, update this file first.
- If default agent behavior changes, update [ai-execution.md](/D:/coder/go/keiyaku-go/docs/governance/ai-execution.md).
- If repository-wide default rules change, update the rules truth document and decide whether an ADR is required.
- If a change becomes automatable, move it out of prompt-only guidance and into scripts, lint, tests, or CI.
