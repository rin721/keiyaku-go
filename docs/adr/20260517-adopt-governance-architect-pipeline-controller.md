---
state_id: ADR-20260517-PIPELINE-001
doc_role: adr
memory_level: L0
state_scope: global
scope: ai
authority_level: ssot_decision
owners: [tech-lead]
status: draft
effective_date: 2026-05-17
version: 1.0
related_rules: []
source_of_truth: [docs/adr/20260517-adopt-governance-architect-pipeline-controller.md]
derived_from: [docs/governance/ai-execution.md, docs/governance/change-management.md, docs/adr/20260515-governance-state-model.md]
read_when: [governance_change]
update_when: [execution_protocol_changed, governance_structure_changed, metadata_standard_changed, review_policy_changed, automation_changed]
conflict_policy: draft_adr_does_not_override_active_execution_protocol
rollback_target: [docs/governance/ai-execution.md, docs/governance/README.md, docs/governance/metadata-schema.md]
verification_target: [scripts/check-governance.ps1, scripts/check-governance-sync.ps1, scripts/check-governance-taxonomy.ps1, scripts/check-governance-map.ps1]
change_reason: introduce a stateful governance pipeline for AI governance tasks
---

# ADR 20260517：采用治理任务 Pipeline Controller

- 状态：draft
- 日期：2026-05-17
- 负责人：tech-lead
- 关联 Issue 或 PR：

## 背景与上下文
<!-- ADR-001 -->

仓库已经具备导航、规则、AI 执行协议、metadata v2、ADR、评审清单、治理脚本、派生索引与 CI。但治理或 Prompt 体系变更仍可能在单次对话中生成长期结论，而没有统一的阶段门禁、状态锁、Artifact Manifest 和显式审批绑定。

如果继续只依赖普通 first-hop 入口，复杂治理变更容易出现三类风险：

- 把长期治理结论留在对话上下文中，没有绑定目标载体。
- 把 Prompt 写成规则 SSOT，重新制造规则漂移。
- 在没有 owner approval 的情况下，把草案、pending 或 partial context 写成已生效默认流程。

## 决策内容
<!-- ADR-002 -->

引入治理任务 Pipeline Controller，作为 `governance_change` 以及 Prompt、ADR、治理脚本、CI、metadata schema、派生索引变更的专用执行流程。

控制器采用四阶段模型：

- `Evaluator`：扫描、分类、判断上下文质量和下一步。
- `Architect`：在门禁满足时输出 Artifact Manifest、目标载体、验证和回滚计划。
- `Maintain`：处理不改变默认行为或长期治理规则的轻量维护任务。
- `Blocked`：在上下文缺失、SSOT 冲突、审批缺失或高风险 unknown 时停止。

默认范围采用“治理任务强制、普通实现轻量”：

- 治理或 Prompt 体系变更必须使用 Controller 判断状态、作用域、SSOT、Artifact Manifest、验证和回滚。
- 普通代码实现任务不强制每轮输出 `decision_audit` 或 `PIPELINE_STATE_LOCK`。
- Prompt 文件只承载执行状态机、上下文路由、门禁、产物协议和状态封存；长期治理规则仍归入规则文档、ADR、评审清单、脚本、lint、测试、CI 或派生索引。

本 ADR 处于 `draft` 时，新 Prompt 和强制执行范围只是提案；只有本 ADR 被 owner 接受后，相关流程才成为默认治理执行路径。

## 后果评估
<!-- ADR-003 -->

正面收益：

- 治理变更必须显式声明 Artifact Manifest，长期结论有明确载体。
- 多 Agent 协作时可以通过状态锁和显式门禁降低状态污染风险。
- 普通开发任务不被强制输出重型审计块，保持日常实现路径轻量。

代价与约束：

- 治理任务输出格式更严格，需要维护 Prompt、AI 执行协议、评审清单、脚本和派生索引的一致性。
- 状态锁只适合治理任务流；若误用于普通实现任务，会增加噪音。
- 在 owner 接受本 ADR 前，任何强制语气都不得被解释为已生效默认规则。

## 备选方案

- 对所有 AI 任务强制输出状态锁。
  缺点是普通实现任务噪音大，容易降低执行效率。
- 只把用户提供的控制器原文复制到 `AGENTS.md`。
  缺点是入口会膨胀成巨型 Prompt，并把稳定规则重新推回 Prompt。
- 完全不新增 Prompt 文件，只更新治理规则。
  缺点是缺少可直接调用的治理任务执行协议，无法稳定约束多 Agent 输出。

## 后续事项

- [ ] owner 接受本 ADR 后，将相关 Prompt 从 `draft` 转为 `active`。
- [ ] 文档已更新。
- [ ] `metadata-schema.md`、`automation-matrix.md` 与评审清单已按需同步。
- [ ] `governance-map.json` 与相关导出/校验脚本已按需同步。
- [ ] 如果决策改变了可执行规则，CI 或静态检查已更新。
