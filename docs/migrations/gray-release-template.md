# Migration Gray-Release Template

Use this template for high-risk P1 DDL changes, such as modifying the type, meaning, nullability, or uniqueness of a core column.

## Principles

- DDL must be repeatable, rollbackable, or compensatable.
- Migrations must use a migration version table or equivalent duplicate-execution protection.
- Every migration must include preflight checks before changing schema or data.
- Do not assume MySQL-style DDL can be fully rolled back by a transaction.
- Rollback plans should prefer forward-compatible compensation over destructive reversal.

## Required Metadata

- Change owner:
- ADR link:
- Target table:
- Target columns:
- Risk level:
- Compatibility window:
- Backfill strategy:
- Read switch strategy:
- Rollback or compensation strategy:
- Observability signals:

## Step 1: V1 Add Compatible Schema

- Add the target column, table, index, or constraint in a backward-compatible way.
- Keep new columns nullable or provide safe defaults when needed.
- Keep old read and write paths unchanged.
- Add preflight checks that skip already-applied schema changes.
- Verify old application versions can still run against the new schema.

## Step 2: V2 Dual Write and Backfill

- Update code to write both the old and new fields.
- Start a resumable background backfill for historical data.
- Use idempotent backfill batches with progress checkpoints.
- Add metrics for backfill progress, lag, failure count, and data mismatch count.
- Keep reads on the old field until backfill is complete and verified.

## Step 3: V3 Switch Read Path

- Switch reads to the new field behind a feature flag or controlled rollout.
- Continue dual writes during the observation period.
- Compare old and new values through sampling or consistency checks.
- Roll back by switching reads back to the old field if mismatches exceed the agreed threshold.

## Step 4: V4 Remove Old Path

- After the observation period, stop dual writes.
- Remove old code paths and stale compatibility logic.
- Drop the old field only after data retention, backup, and rollback windows are satisfied.
- Keep the migration version and audit trail.

## Acceptance Checklist

- [ ] ADR is linked for high-risk or non-default behavior.
- [ ] Preflight checks protect every non-idempotent operation.
- [ ] Backfill can resume after interruption.
- [ ] Duplicate execution is harmless.
- [ ] Rollback or compensation is documented.
- [ ] Metrics and logs include TraceID or CorrelationID for background work.
- [ ] Old and new application versions can coexist during rollout.
