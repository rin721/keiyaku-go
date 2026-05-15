---
state_id: CONV-PKG-001
doc_role: convention
memory_level: L1
state_scope: module
scope: pkg
authority_level: binding
owners: [tech-lead]
status: active
effective_date: 2026-05-15
version: 2.0
related_rules: [GOV-P0-002]
source_of_truth: [docs/governance/rules.md, docs/conventions/layering.md]
derived_from: [docs/governance/rules.md, docs/conventions/layering.md]
read_when: [pkg_change, governance_change]
update_when: [default_rule_changed, default_behavior_changed, adr_accepted, automation_changed]
conflict_policy: binding_must_yield_to_ssot
rollback_target: [docs/governance/rules.md, docs/conventions/layering.md]
verification_target: [scripts/check-layering.ps1, scripts/check-governance-map.ps1]
---

# pkg 设计约定

`pkg/` 用于业务无关、可复用、可独立测试的能力组件。

## 放入条件

- 能力可被多个业务模块复用。
- 不依赖 `internal/` 业务包。
- 不暴露业务实体、传输 DTO 或持久层模型。
- 可以通过单元测试验证核心行为。

## 默认边界

- `pkg` 可以依赖标准库和稳定第三方库。
- `pkg` 不得 import `internal`。
- `pkg` 不得包含业务流程编排。
- 业务适配代码放在 `internal`，不要反向塞进 `pkg`。

## 命名与演进

- 包名表达能力，不表达具体业务场景。
- 新增通用能力前，优先确认是否只是当前业务的局部抽象。
- 如果某个 `pkg` 设计风格将成为默认模式，必须更新本文档，并评估是否需要 ADR 与自动化联动。
