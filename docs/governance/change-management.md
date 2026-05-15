---
state_id: GOV-CHANGE-001
doc_role: governance_process
memory_level: L0
state_scope: global
scope: repo
authority_level: binding
owners: [tech-lead]
status: active
effective_date: 2026-05-15
version: 2.0
related_rules: []
source_of_truth: [docs/governance/change-management.md]
derived_from: [docs/governance/rules.md, docs/adr/20260515-governance-state-model.md]
read_when: [governance_change, boundary_sensitive, migration_sensitive, async_sensitive, exception_review]
update_when: [default_behavior_changed, default_rule_changed, adr_policy_changed, exception_policy_changed, governance_structure_changed]
conflict_policy: change_management_orchestrates_sync_under_ssot
rollback_target: [docs/governance/rules.md, docs/adr/20260515-governance-state-model.md]
verification_target: [scripts/check-governance-sync.ps1, scripts/check-exception-expiry.ps1, scripts/check-governance-map.ps1]
---

# 治理变更管理

本文档定义默认行为冲突、治理演进、历史同步、治理债务和 break-glass 的处理方式。默认行为变更未完成闭环同步时，不得视为任务完成。

## 状态边界

- `Global State`：导航、规则、执行协议、Accepted ADR、异常登记和派生索引等仓库级长期状态。
- `Module State`：专题约定、评审清单、模板等局部长期状态。
- `Local State`：任务计划、变更范围、同步清单和验证范围。
- `Ephemeral State`：扫描结果、临时推断和对话草稿。

`Local State` 和 `Ephemeral State` 不得直接污染 `Global State` 或 `Module State`。只有补齐 ADR、同步面、回滚目标和验证目标后，才能提升为长期治理状态。

## 冲突级别

- L0 局部实现差异：只影响单个目录或能力，不改变默认规则、默认边界或默认工作流。可直接更新局部约定或实现。
- L1 默认行为变化：会成为后续默认写法或默认流程。必须先补 ADR，再同步治理规则、AI 执行协议、专题约定、评审清单、自动化矩阵和脚本/CI。
- L2 P1 偏离：偏离默认治理要求。必须补 ADR，并由 owner 审批。
- L3 P0、安全、权限模型、不可逆数据风险：停止实施，要求人工决策；P0 不允许通过 break-glass 豁免。
- L4 历史实现大量不一致：先登记治理债务，再制定迁移范围、验证方式、收益、风险和停止条件。

## 决策路径

1. 先判断是否触及 P0、安全、权限或不可逆数据风险；若是，立即停止并升级。
2. 若不是，再判断是否改变默认行为、默认边界、默认风格、默认依赖方式、默认工作流或默认执行路径。
3. 若不改变默认项，按局部实现处理。
4. 若改变默认项，按 L1/L2/L4 分级，并完成文档、ADR、评审、自动化和历史同步决策。

## 状态晋升规则

- 只影响单次任务的计划、验证和范围结论停留在 `Local State`。
- 只服务于当前分析过程的推断停留在 `Ephemeral State`。
- 只有当结论会成为默认行为、默认边界、默认风格或默认工作流时，才允许提升为 `Module State` 或 `Global State`。
- 提升前必须补齐：`source_of_truth`、`derived_from`、`rollback_target`、`verification_target`、所需 ADR 和同步面。

## 默认风格演进触发条件

- 某种实现从单点实践变成多个模块复用。
- 评审中反复批准同类偏离。
- 自动化发现同类问题高频出现。
- 新约定会改变目录边界、依赖方向、测试策略、迁移策略或 CI 门禁。
- 用户明确要求把某种风格作为长期默认。

触发后必须评估是否同步更新：`rules.md`、`README.md`、`ai-execution.md`、`metadata-schema.md`、相关 convention、ADR、review checklist、automation matrix、脚本、lint、测试和 CI。

## 治理维护循环

当出现以下任一信号时，应把局部修补升级为治理更新：

- 同类 metadata 漏填、索引漂移或路由误读反复出现。
- 评审中反复批准相同类型的受控偏离。
- 自动化反复发现同类问题，但规则仍只存在于人工评审中。
- 某种局部实践已经演化成新的默认行为。

## 治理债务

治理债务用于暂时允许历史代码或历史文档不完全符合新治理，但必须可追踪、可复审、可关闭。债务登记在 `docs/governance/exceptions.yaml`。

必填字段：`rule_id`、`reason`、`owner`、`created_at`、`review_at` 或 `expiry`、`required_followup`、`rollback_target`、`verification_target`、`status`。

## Boy Scout Rule

当治理债务位于当前正在修改的文件、目录或紧邻调用链中，并同时满足以下条件时，AI 应主动建议顺带清理：

- 不改变默认行为、默认边界或公共接口。
- 不引入新的架构决策，不需要新增 ADR。
- 额外修改和验证成本不超过本次任务的大约 10% 到 20%。
- 清理后的验证路径与本次任务基本重合。

如果不满足这些条件，AI 必须显式说明为什么本次不适合顺带清理，并保留或补充治理债务标记。

## Break-glass

Break-glass 只用于紧急恢复或阻止重大风险扩大。

- 不得覆盖 P0。
- 必须设置 `expiry`。
- 必须写明 `required_followup`。
- 如果偏离默认规则超过一个发布周期，或改变默认行为，必须补临时 ADR。
- 到期必须关闭、续期或转成治理债务。

## 历史代码同步

默认采用 touched-code first。只有安全、日志、密码、依赖方向、迁移兼容、CI 触发等高风险规则，才发起专项全仓扫描。

停止条件：自动化误报高、影响范围不清、缺少 ADR、迁移收益低于成本、或需要超过当前任务范围的大规模重构。

## 默认行为变更闭环

只要变更会影响默认行为、默认约束、默认边界、默认风格、默认工作流或默认执行路径，就必须判断并同步：

- 是否需要新增或更新 ADR。
- 是否需要更新导航、执行协议、元数据 schema、专题约定、评审清单和自动化矩阵。
- 是否需要回填 `governance-map.json`、脚本、lint、测试和 CI。
- 是否需要 touched-code first 同步历史实现。
- 是否需要兼容层、迁移路径、回滚方案和验证方案。
