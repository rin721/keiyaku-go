---
doc_role: review
scope: repo
authority_level: derived
owners: [tech-lead]
status: active
effective_date: 2026-05-15
version: 1.0
related_rules: []
read_when: [governance_change, default_style_change]
update_when: [review_policy_changed, governance_process_changed]
---

# 治理变更评审清单

治理变更 PR 必须确认以下事项：

- [ ] 是否改变默认规则或默认设计风格。
- [ ] 如果改变默认规则，是否已补 ADR。
- [ ] 是否同步更新 `rules.md`、相关 convention、AI 执行协议和导航文档。
- [ ] 是否同步更新自动化矩阵。
- [ ] 可机器检查的规则是否进入脚本、lint、测试或 CI。
- [ ] 难自动化的判断项是否进入评审清单。
- [ ] 是否需要登记治理债务或 break-glass。
- [ ] 是否需要历史代码同步策略和停止条件。
