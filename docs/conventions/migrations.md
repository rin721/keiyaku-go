---
state_id: CONV-MIG-001
doc_role: convention
memory_level: L1
state_scope: module
scope: migrations
authority_level: binding
owners: [tech-lead]
status: active
effective_date: 2026-05-15
version: 2.0
related_rules: [GOV-P1-003]
source_of_truth: [docs/governance/rules.md, docs/governance/change-management.md]
derived_from: [docs/governance/rules.md, docs/governance/change-management.md]
read_when: [migration_sensitive, governance_change]
update_when: [migration_policy_changed, adr_accepted, automation_changed]
conflict_policy: binding_must_yield_to_ssot
rollback_target: [docs/governance/rules.md, docs/governance/change-management.md]
verification_target: [scripts/check-governance-sync.ps1, scripts/check-governance-map.ps1]
---

# Migration 约定

数据库迁移必须以可重复部署、可观测、可回滚或可补偿为默认目标。

## 默认要求

- 每个迁移在修改 schema 或数据前必须具备前置检查。
- 无法天然幂等的迁移必须依赖版本表或等效机制。
- 高风险变更使用灰度步骤：兼容 schema、双写、回填、读切换、旧路径移除。
- 不得假设 MySQL 风格 DDL 可通过事务完整回滚。

## ADR 触发

高风险表结构变更、非默认回滚策略、不可逆数据变更、跨版本兼容窗口变化，必须补 ADR。

## 模板

灰度迁移执行计划使用 `docs/migrations/gray-release-template.md`。
