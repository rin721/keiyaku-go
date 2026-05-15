---
state_id: CONV-DI-001
doc_role: convention
memory_level: L1
state_scope: module
scope: dependency_injection
authority_level: binding
owners: [tech-lead]
status: active
effective_date: 2026-05-15
version: 2.0
related_rules: [GOV-P1-006]
source_of_truth: [docs/governance/rules.md, docs/adr/20260515-default-backend-direction.md]
derived_from: [docs/governance/rules.md, docs/adr/20260515-default-backend-direction.md]
read_when: [boundary_sensitive, governance_change]
update_when: [dependency_injection_policy_changed, adr_accepted, automation_changed]
conflict_policy: binding_must_yield_to_ssot
rollback_target: [docs/governance/rules.md, docs/adr/20260515-default-backend-direction.md]
verification_target: [scripts/check-governance-sync.ps1, scripts/check-governance-map.ps1]
---

# 依赖注入约定

本文档细化 `GOV-P1-006`。仓库默认采用手动依赖注入；依赖注入约定属于默认行为的一部分，改变时必须触发 ADR 和闭环同步。

## 默认方式

- 默认通过显式构造函数、装配函数或 `main`/bootstrap 代码手动注入依赖。
- 依赖关系必须在代码中可读、可追踪，不依赖运行时反射或隐式注册。
- 依赖边界应配合 [layering.md](layering.md) 执行，不得借 DI 容器绕过分层规则。

## 允许的例外

- 对依赖图谱明显复杂、手写装配维护成本过高的模块，可使用 Wire 等编译期生成工具。
- 生成代码必须提交入库，并参与代码评审。
- 引入编译期生成工具后，仍应保留清晰的装配入口，不把业务初始化逻辑埋入生成器配置。

## 禁止事项

- 禁止引入运行时反射型依赖注入容器。
- 禁止通过服务定位器、全局注册表或隐式单例隐藏核心依赖关系。
- 禁止把 DI 选择当作局部实现细节悄悄改成新默认行为。

## ADR 触发

- 改变默认依赖注入方式。
- 在多个模块推广新的装配模式。
- 允许新的代码生成工具或框架级容器成为默认选项。

以上情形必须补 ADR，并同步 `rules.md`、`ai-execution.md`、相关专题约定、评审清单与自动化矩阵。
