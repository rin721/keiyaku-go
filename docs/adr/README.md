---
doc_role: adr_index
scope: repo
authority_level: ssot
owners: [tech-lead]
status: active
effective_date: 2026-05-15
version: 1.0
related_rules: []
read_when: [governance_change, boundary_sensitive, default_style_change, operational_sensitive]
update_when: [adr_policy_changed, default_rule_changed, governance_process_changed]
---

# 架构决策记录

ADR 是 Keiyaku 的设计意图来源。任何偏离 P1 治理规范的变更，都必须先记录 ADR，并在合入前完成审批。

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

使用以下状态之一：

- Proposed：提案
- Accepted：通过
- Deprecated：已弃用
- Superseded：已被替代

## 审批要求

改变或偏离 P1 的 ADR，必须由技术负责人 Review 后，相关代码或配置才可合入。

P0 不允许偏离。如果某个方案看起来必须突破 P0，说明设计需要调整。

## 模板

复制 `0000-template.md`，替换日期、标题、状态、背景、决策与后果评估。
