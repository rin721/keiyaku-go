---
doc_role: review_checklist
scope: review
authority_level: derived
owners: [tech-lead]
status: active
effective_date: 2026-05-15
version: 1.0
related_rules: []
read_when: [governance_change, review_change]
update_when: [review_policy_changed, governance_structure_changed, metadata_standard_changed]
---

# 治理变更评审清单

治理变更 PR 必须确认以下事项：

- [ ] 是否改变默认规则或默认设计风格。
- [ ] 如果改变默认规则，是否已补 ADR。
- [ ] 是否同步更新 `rules.md`、导航文档、`ai-execution.md`、`metadata-schema.md` 和相关 convention。
- [ ] 是否同步更新 `automation-matrix.md`、脚本、lint、测试或 CI。
- [ ] 可机器检查的规则是否进入脚本、lint、测试或 CI。
- [ ] 难自动化的判断项是否进入评审清单。
- [ ] 是否需要登记治理债务、`exceptions.yaml` 或 break-glass。
- [ ] 是否需要历史代码同步策略、范围、`stop-condition` 和回滚思路。
