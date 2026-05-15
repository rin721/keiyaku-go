---
state_id: ADR-20260515-STATE-001
doc_role: adr
memory_level: L0
state_scope: global
scope: repo
authority_level: ssot_decision
owners: [tech-lead]
status: accepted
effective_date: 2026-05-15
version: 1.0
related_rules: []
source_of_truth: [docs/adr/20260515-governance-state-model.md]
derived_from: [docs/governance/README.md, docs/governance/change-management.md]
read_when: [governance_change]
update_when: [governance_structure_changed, metadata_standard_changed, automation_changed]
conflict_policy: accepted_adr_overrides_default_governance_shape
rollback_target: [docs/governance/README.md, docs/governance/metadata-schema.md, docs/governance/change-management.md]
verification_target: [scripts/check-governance.ps1, scripts/check-governance-map.ps1, scripts/check-governance-sync.ps1]
supersedes: []
change_reason: promote governance assets from document-oriented routing to state-oriented routing
---

# ADR 20260515：治理状态模型升级

- 状态：accepted
- 日期：2026-05-15
- 负责人：tech-lead
- 关联 Issue 或 PR：

## 背景与上下文
<!-- ADR-001 -->

仓库已经具备导航、规则、AI 执行协议、专题约定、评审清单、治理脚本与 CI 的第一代治理骨架，但当前系统仍主要围绕“文档职责”组织。AI 代理可以知道“这份文档是什么”，却不能稳定回答“这个状态从哪里来、能覆盖谁、怎么回滚、怎么验证”。

如果继续只扩写 Prompt 或新增说明文档，会放大规则漂移、上下文浪费和多 Agent 协作下的状态污染风险。

## 决策内容
<!-- ADR-002 -->

采用 metadata v2 与派生索引双层模型，把治理资产视为可寻址的状态对象：

- 所有持久化治理资产必须携带 `state_id`、`memory_level`、`state_scope`、`source_of_truth`、`derived_from`、`conflict_policy`、`rollback_target` 和 `verification_target`。
- `Global State` 与 `Module State` 继续由现有 SSOT、Accepted ADR、专题约定、评审清单和异常登记承载，不新增第二套规则正文。
- `Local State` 和 `Ephemeral State` 只作为执行与分析边界存在，不得直接提升为长期治理，除非完成 ADR、同步面、回滚目标和验证目标闭环。
- 新增 `docs/governance/governance-map.json` 作为机器可读派生索引，用于状态发现、读写路由和多 Agent 协作；它必须让位于 SSOT 与 Accepted ADR。
- 新增导出与校验脚本，把 metadata v2、一致性、索引新鲜度和 lineage 校验接入本地检查、pre-commit 与 CI。

## 后果评估
<!-- ADR-003 -->

正面收益：

- AI 代理可以先识别状态，再识别作用域，再识别真相源，最后决定持久化位置。
- 派生索引可降低首次扫描成本，同时不牺牲 SSOT 边界。
- 回滚路径、验证路径和同步面显式化后，治理变更更容易闭环。

代价与约束：

- 所有治理文档的 front matter 需要补齐 metadata v2。
- 新增 `governance-map.json` 后，必须维护导出脚本和校验脚本的一致性。
- 这次升级属于默认治理行为变化，后续任何同类调整都必须继续走 ADR。

## 备选方案

- 继续只维护 metadata v1。
  缺点是无法显式建模状态来源、索引 lineage 和回滚验证链路。
- 把治理索引做成新的 SSOT。
  缺点是会制造第三套真相源，违反“派生产物必须让位于 SSOT”的原则。
- 继续在 Agent 入口里堆叠说明。
  缺点是会把稳定政策重新推回 Prompt，无法形成自动化闭环。

## 后续事项

- [x] 更新 metadata schema、导航、执行协议、变更管理和评审清单。
- [x] 新增治理索引与导出/校验脚本。
- [x] 把 metadata v2 接入治理检查、本地入口和 CI。
- [x] 为现有治理资产补齐状态字段与追踪链路。
