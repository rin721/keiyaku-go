---
state_id: REVIEW-GOV-001
doc_role: review_checklist
memory_level: L1
state_scope: module
scope: review
authority_level: derived
owners: [tech-lead]
status: active
effective_date: 2026-05-15
version: 2.0
related_rules: []
source_of_truth: [docs/governance/rules.md, docs/governance/change-management.md, docs/governance/metadata-schema.md]
derived_from: [docs/governance/rules.md, docs/governance/change-management.md, docs/governance/metadata-schema.md]
read_when: [governance_change, review_change]
update_when: [review_policy_changed, governance_structure_changed, metadata_standard_changed]
conflict_policy: derived_must_yield_to_ssot
rollback_target: [docs/governance/change-management.md, docs/governance/metadata-schema.md]
verification_target: [scripts/check-governance-sync.ps1, scripts/check-governance-map.ps1]
---

# 治理变更评审清单

治理变更 PR 必须确认以下事项：

- [ ] 是否改变默认规则或默认设计风格。
- [ ] 如果改变默认规则，是否已补 ADR。
- [ ] 是否同步更新 `rules.md`、导航文档、`ai-execution.md`、`metadata-schema.md` 和相关 convention。
- [ ] 是否同步更新 `automation-matrix.md`、脚本、lint、测试或 CI。
- [ ] 是否同步更新 `governance-map.json` 及其生成/校验脚本。
- [ ] metadata v2 字段是否完整，`source_of_truth`、`derived_from`、`rollback_target`、`verification_target` 是否可追踪。
- [ ] 可机器检查的规则是否进入脚本、lint、测试或 CI。
- [ ] 难自动化的判断项是否进入评审清单。
- [ ] 是否需要登记治理债务、`exceptions.yaml` 或 break-glass。
- [ ] 是否把 Local/Ephemeral 结论误写成 Global/Module 长期状态。
- [ ] 是否需要历史代码同步策略、范围、`stop-condition` 和回滚思路。
