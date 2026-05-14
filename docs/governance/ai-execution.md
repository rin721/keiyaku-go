---
doc_role: ai_execution
scope: repo
authority_level: ssot
owners: [tech-lead]
status: active
effective_date: 2026-05-15
version: 1.0
related_rules: []
read_when: [all_tasks]
update_when: [execution_protocol_changed, routing_changed, self_check_changed, escalation_policy_changed]
---

# AI Execution Protocol

This document is the execution truth for AI development agents in this repository. It defines how agents should load context, decide whether a change is local or governance-impacting, and determine when to stop for human review.

## Entry Protocol

For every task:

1. Read [AGENTS.md](/D:/coder/go/keiyaku-go/AGENTS.md).
2. Read [README.md](/D:/coder/go/keiyaku-go/docs/governance/README.md).
3. Classify the task before loading more documents.
4. Load only the documents routed by the navigation truth.
5. Identify whether the task touches:
   - repository-wide defaults
   - architecture boundaries
   - migrations or async execution semantics
   - governance assets themselves
6. Only then inspect or edit code.

## Context Loading Rules

- Do not read the whole governance tree by default.
- Prefer the smallest sufficient set of governance documents for the task.
- Prefer documents with `authority_level: ssot` or `authority_level: binding` before reading `derived` or `template` documents.
- If a document has a narrow `scope` that clearly matches the task, load it before unrelated repo-wide supporting material.
- If the task changes governance itself, load both the navigation truth and the execution truth before touching any governance artifact.

## Task Classification

Classify each task into one of these buckets before editing:

- `implementation_local`: a code change that follows existing defaults
- `boundary_sensitive`: changes layering, contracts, imports, persistence boundaries, or transport/domain separation
- `operational_sensitive`: touches migrations, rollout, async jobs, idempotency, retries, observability, or security
- `governance_change`: changes prompts, governance docs, review checklists, scripts, lint, tests, or CI gates
- `default_style_change`: introduces a new default pattern that others are expected to follow

Classification outcomes:

- `implementation_local`: proceed after loading the routed docs
- `boundary_sensitive`: proceed only after checking for applicable ADRs
- `operational_sensitive`: proceed only after checking rollout, observability, and failure-mode expectations
- `governance_change`: update docs, review surfaces, and automation together where applicable
- `default_style_change`: require ADR unless the navigation truth explicitly says otherwise

## ADR Decision Rules

An agent must stop and require ADR creation or update before proceeding when the change:

- modifies a repository-wide default
- changes architecture boundaries or dependency direction
- changes the default migration or async execution model
- changes what reviewers or CI should treat as a required gate
- turns a one-off pattern into a default team pattern

An agent may proceed without ADR when the change:

- only applies a current default to a local implementation
- tightens automation without changing the underlying policy
- updates wording, examples, or non-authoritative explanation without changing behavior

## Self-Check Before Editing

Before editing, the agent must be able to answer:

1. What task type is this
2. Which document is the navigation truth
3. Which document currently holds the applicable rule truth
4. Whether an Accepted ADR already governs this scope
5. Whether this change should update docs, review checklist, or automation in the same change

If any answer is unclear, pause and resolve that ambiguity first.

## Self-Check After Editing

After editing, the agent should verify:

1. The chosen documents still reflect the actual workflow
2. Routing instructions still point to existing files
3. Metadata is present and consistent
4. Any new default or binding behavior has a matching governance home
5. Any newly automatable rule is not left only in prompt text

## Stop and Escalate Conditions

Stop and ask for human decision when:

- the change appears to violate a P0 rule
- two active authoritative documents conflict without a clear priority winner
- a change needs a temporary exception but no exception mechanism is yet defined in-repo
- the repository state no longer matches a claimed governance phase and the next intended phase is unclear
- a proposed default style would require broad historical backfill with uncertain scope or risk

## Update Propagation Rules

When governance-impacting changes are made, evaluate whether the same change also requires updates to:

- repository-wide governance rules
- ADRs
- review checklists
- governance scripts
- lint/test/CI configuration
- agent entry or routing documents

Do not assume prompt text alone is a sufficient implementation of governance.

## Metadata Handling

Agents must treat metadata as executable routing hints.

Required fields expected on governance-facing documents:

- `doc_role`
- `scope`
- `authority_level`
- `owners`
- `status`
- `version` or `effective_date`
- `related_rules`
- `read_when`
- `update_when`

Agent behavior:

- Use `read_when` to decide whether to load the document.
- Use `scope` to discard irrelevant governance context.
- Use `authority_level` to resolve which document is truth and which is only guidance or record.
- Use `update_when` to decide whether a change must fan out into other governance assets.

## Current Interim Rule Source

Until a dedicated `docs/governance/rules.md` is introduced, repository-wide rule truth remains [../architecture/governance.md](/D:/coder/go/keiyaku-go/docs/architecture/governance.md). Agents should not invent a replacement truth source during implementation work.
