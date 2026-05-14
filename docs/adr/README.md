---
doc_role: adr_index
scope: repo
authority_level: binding
owners: [tech-lead]
status: active
effective_date: 2026-05-15
version: 1.0
related_rules: []
read_when: [governance_change, boundary_sensitive, migration_sensitive, async_sensitive]
update_when: [adr_policy_changed, governance_structure_changed, default_rule_changed]
---

# 架构决策记录

ADR 是 Keiyaku 的设计意图来源。`docs/adr/*.md` 中状态为 `accepted` 且 `authority_level: ssot_decision` 的 ADR 才是偏离默认规则与重大设计取舍的裁决真相。任何偏离 P1 治理规范或改变默认行为的变更，都必须先记录 ADR，并在合入前完成审批。

## 命名规范

使用：

```text
docs/adr/YYYYMMDD-short-kebab-name.md
```

示例：

```text
docs/adr/20260509-adopt-echo-sqlc.md
docs/adr/20260509-allow-wire-for-admin-module.md
```

## 状态

元数据 `status` 使用以下状态之一：

- `draft`：提案
- `accepted`：通过
- `deprecated`：已弃用或已被替代
- `historical`：仅保留历史背景，不再作为现行裁决

## 审批要求

改变或偏离 P1 的 ADR，必须由技术负责人 Review 后，相关代码或配置才可合入。

P0 不允许偏离。如果某个方案看起来必须突破 P0，说明设计需要调整。

## 模板

复制 `0000-template.md`，替换日期、标题、状态、背景、决策与后果评估。
