---
doc_role: ai_entry
scope: repo
authority_level: binding_entry
owners: [tech-lead]
status: active
effective_date: 2026-05-15
version: 1.0
related_rules: []
read_when: [all_tasks]
update_when: [routing_changed, metadata_standard_changed, execution_protocol_changed]
---

# AI Agent Entry

This file is the first-hop entry for AI development agents working in this repository.

## Required Read Order

1. Read [docs/governance/README.md](/D:/coder/go/keiyaku-go/docs/governance/README.md).
2. Read [docs/governance/ai-execution.md](/D:/coder/go/keiyaku-go/docs/governance/ai-execution.md).
3. Read only the task-specific governance documents routed by the navigation document.
4. Read code only after the governance route is clear.

Do not load all governance documents by default. Route first, then load the smallest sufficient set.

## Current SSOT Map

- Navigation truth: [docs/governance/README.md](/D:/coder/go/keiyaku-go/docs/governance/README.md)
- AI execution truth: [docs/governance/ai-execution.md](/D:/coder/go/keiyaku-go/docs/governance/ai-execution.md)
- Repository-wide governance rules truth, current interim source: [docs/architecture/governance.md](/D:/coder/go/keiyaku-go/docs/architecture/governance.md)
- Deviation and design decision truth: Accepted ADRs under [docs/adr](/D:/coder/go/keiyaku-go/docs/adr)
- Governance exception and debt registry: [docs/governance/exceptions.yaml](/D:/coder/go/keiyaku-go/docs/governance/exceptions.yaml)

## Fast Routing

- `pkg/` utility or shared package work:
  Read navigation, execution protocol, repository-wide governance rules, then the layering convention. If the change alters default package style or boundary expectations, also read ADR guidance.
- Layering, dependency direction, DTO/domain/repository boundary work:
  Read navigation, execution protocol, repository-wide governance rules, the layering convention, then ADR guidance.
- Migration, gray release, backfill, rollout, rollback work:
  Read navigation, execution protocol, repository-wide governance rules, migration template, then ADR guidance if the change is high-risk or non-default.
- Async jobs, retries, idempotency, scheduled task work:
  Read navigation, execution protocol, repository-wide governance rules, then ADR guidance if the change alters the default model.
- Test, CI, lint, security scan, governance script work:
  Read navigation, execution protocol, repository-wide governance rules, review checklist, convention scripts, and the relevant workflow files.
- Governance or prompt system changes:
  Read navigation, execution protocol, repository-wide governance rules, ADR guidance, review checklist, and the governance automation scripts.

## Metadata Standard

All governance-facing documents should carry YAML front matter. The canonical field definitions live in the navigation truth document, but agents must expect at least these fields:

- `doc_role`
- `scope`
- `authority_level`
- `owners`
- `status`
- `version` or `effective_date`
- `related_rules`
- `read_when`
- `update_when`

Agents should use metadata for discovery and routing, not just filenames.

## Do Not

- Do not collapse all rules into a single prompt file.
- Do not treat prompts as the source of truth for stable engineering policy.
- Do not bypass ADRs when a change alters default behavior, architecture boundaries, or governance policy.
