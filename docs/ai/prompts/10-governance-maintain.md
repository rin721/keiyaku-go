---
state_id: AI-PROMPT-MAINTAIN-001
doc_role: ai_prompt
memory_level: L0
state_scope: global
scope: ai
authority_level: binding
owners: [tech-lead]
status: draft
effective_date: 2026-05-17
version: 1.0
related_rules: []
source_of_truth: [docs/governance/ai-execution.md, docs/governance/change-management.md]
derived_from: [docs/adr/20260517-adopt-governance-architect-pipeline-controller.md, docs/governance/ai-execution.md, docs/governance/change-management.md]
read_when: [governance_change]
update_when: [execution_protocol_changed, governance_structure_changed, review_policy_changed]
conflict_policy: prompt_must_yield_to_governance_ssot
rollback_target: [docs/governance/ai-execution.md, docs/adr/20260517-adopt-governance-architect-pipeline-controller.md]
verification_target: [scripts/check-governance.ps1, scripts/check-governance-sync.ps1, scripts/check-governance-map.ps1]
---

# Governance Maintain Prompt

本 Prompt 是治理任务中的轻量 Maintain 路径草案。它用于判断某个任务无需进入重型 Architect 输出时，代理如何保持最小上下文、最小产物和可验证结果。

## 使用条件

可以使用 Maintain 路径的任务必须同时满足：

- 不改变默认行为、默认边界、默认风格、默认工作流或长期治理政策。
- 不新增 SSOT，不改变 Accepted ADR，不调整 metadata schema 或治理索引结构。
- 不涉及安全、权限、支付、不可逆数据风险或跨模块协作默认方式。
- 验证路径清楚，且不需要 owner approval。

如果任一条件不满足，回到 [00-governance-architect-controller.md](00-governance-architect-controller.md) 的 Evaluator 或 Blocked 阶段。

## 输出要求

Maintain 输出应说明：

- 当前分类和为何不需要治理闭环。
- 本次实际读写范围。
- 执行的验证命令或无法验证的原因。
- 是否触达治理债务；若触达，说明是否符合 Boy Scout Rule。

Maintain 不应输出完整 Artifact Manifest，也不应把 Local 或 Ephemeral 结论提升为长期治理状态。
