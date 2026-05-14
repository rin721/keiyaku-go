---
doc_role: historical_reference
scope: architecture_history
authority_level: derived
owners: [tech-lead]
status: historical
effective_date: 2026-05-15
version: 1.0
related_rules: []
read_when: [governance_change]
update_when: [historical_status_changed, governance_structure_changed]
---

# 历史治理背景

本文档只保留治理体系演进前的历史背景，不再承载当前规则真相。现行治理应以 [docs/governance/rules.md](../governance/rules.md)、[docs/governance/README.md](../governance/README.md) 和状态为 `accepted` 的 ADR 为准。

## 保留原因

- 需要保留旧治理表达的迁移背景。
- 需要为历史讨论、旧链接和审计记录提供定位点。
- 需要明确旧文档已退出 SSOT，避免误读为现行规则来源。

## 迁移说明

- 现行规则已经迁移到 `docs/governance/`。
- 默认技术方向的裁决已迁移到 Accepted ADR，例如 [20260515-default-backend-direction.md](../adr/20260515-default-backend-direction.md)。
- 本文档不再复制 P0/P1/P2 规则正文，避免形成双真相。

如果需要理解为什么仓库从单一治理文档拆分为导航、规则、执行协议、专题约定、ADR、评审清单和自动化矩阵，请优先阅读 [20260515-governance-ssot-structure.md](../adr/20260515-governance-ssot-structure.md)。
