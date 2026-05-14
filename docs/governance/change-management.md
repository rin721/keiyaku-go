---
doc_role: governance
scope: repo
authority_level: binding
owners: [tech-lead]
status: active
effective_date: 2026-05-15
version: 1.0
related_rules: []
read_when: [governance_change, default_style_change, exception_review, boundary_sensitive]
update_when: [default_rule_changed, adr_accepted, exception_policy_changed, migration_policy_changed]
---

# 治理变更管理

本文档定义设计风格冲突、治理演进、历史同步、治理债务和 break-glass 的处理方式。

## 冲突级别

- L0 局部实现差异：只影响单个目录或能力，不改变默认规则。直接更新对应专题约定或实现。
- L1 默认风格变化：会成为后续默认写法。必须补 ADR，并同步治理规则、专题约定、评审清单和自动化矩阵。
- L2 P1 偏离：偏离默认治理要求。必须补 ADR，并由 owner 审批。
- L3 P0、安全、数据兼容或不可逆风险：停止实施，要求人工决策；P0 不允许通过 break-glass 豁免。
- L4 历史实现大量不一致：先登记治理债务，再制定迁移范围、验证方式、收益、风险和停止条件。

## 默认风格演进触发条件

- 某种实现从单点实践变成多个模块复用。
- 评审中反复批准同类偏离。
- 自动化发现同类问题高频出现。
- 新约定会改变目录边界、依赖方向、测试策略、迁移策略或 CI 门禁。
- 用户明确要求把某种风格作为长期默认。

触发后必须评估是否同步更新：`rules.md`、`ai-execution.md`、相关 convention、ADR、review checklist、automation matrix、脚本、lint、测试和 CI。

## 治理债务

治理债务用于暂时允许历史代码或历史文档不完全符合新治理，但必须可追踪、可复审、可关闭。债务登记在 `docs/governance/exceptions.yaml`。

必填字段：`rule_id`、`reason`、`owner`、`created_at`、`review_at` 或 `expiry`、`required_followup`、`status`。

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
