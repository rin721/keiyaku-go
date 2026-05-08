# Architecture Decision Records

ADRs are the source of design intent for rinblog. Any deviation from P1 governance must be recorded as an ADR and approved before merge.

## Naming

Use:

```text
docs/adr/YYYYMMDD-short-kebab-name.md
```

Examples:

```text
docs/adr/20260509-adopt-echo-sqlc.md
docs/adr/20260509-allow-wire-for-admin-module.md
```

## Status

Use one of:

- Proposed
- Accepted
- Deprecated
- Superseded

## Required Review

An ADR that changes or deviates from P1 must be reviewed by the technical owner before the related code or configuration merges.

P0 deviations are not allowed. If a proposal appears to require a P0 deviation, the design must change.

## Template

Copy `0000-template.md` and replace the date, title, status, context, decision, and consequences.
