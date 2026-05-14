---
doc_role: adr
scope: repo
authority_level: ssot_decision
owners: [tech-lead]
status: accepted
effective_date: 2026-05-15
version: 1.0
related_rules: []
read_when: [governance_change]
update_when: [governance_structure_changed, metadata_standard_changed, automation_changed]
---

# ADR 20260515：治理 SSOT 结构整固

- 状态：accepted
- 日期：2026-05-15
- 负责人：tech-lead
- 关联 Issue 或 PR：

## 背景与上下文
<!-- ADR-001 -->

仓库已经具备导航、规则、AI 执行协议、专题约定、评审清单与治理脚本的基础骨架，但旧的 `docs/architecture/governance.md` 仍保留大量重复规则，元数据标签词表也尚未收敛，导致规则来源、路由方式和自动化边界存在双真相与漂移风险。

## 决策内容
<!-- ADR-002 -->

采用分层治理结构，并明确各层职责：

- `docs/governance/README.md` 作为导航真相。
- `docs/governance/rules.md` 作为仓库规则真相。
- `docs/governance/ai-execution.md` 作为 AI 执行协议真相。
- `docs/governance/metadata-schema.md` 作为元数据字段和标签词表真相。
- `docs/conventions/*.md` 只承载局部默认约定。
- `docs/adr/*.md` 中状态为 `accepted` 的文件作为重大取舍与偏离治理的裁决真相。
- `docs/review/*.md` 只保留难自动化的人工检查。
- `docs/architecture/governance.md` 降级为历史背景，不再承载现行规则正文。

同时建立统一任务标签词表、异常模板、治理同步校验和自动化归位矩阵，确保 Prompt 只保留 first-hop、路由、自检和升级条件。

## 后果评估
<!-- ADR-003 -->

正面收益是降低上下文加载成本、减少双真相和误读，并让治理变更更容易闭环到 ADR、评审与 CI。代价是治理资产数量略有增加，需要维护元数据 schema 与治理脚本的一致性。

## 备选方案

- 继续沿用单一治理文档。
  缺点是无法同时满足路由、规则、AI 执行、评审与自动化归位，且 Prompt 容易重新膨胀成规则堆栈。
- 只补一份 AI Prompt，不调整治理结构。
  缺点是把稳定工程政策错误地放回 Prompt，无法形成可持续自动化闭环。

## 后续事项

- [x] 文档已更新。
- [x] `metadata-schema.md`、`automation-matrix.md` 与评审清单已按需同步。
- [x] 如果决策改变了可执行规则，CI 或静态检查已更新。
- [x] 如涉及迁移或高风险变更，已记录回滚或补偿方案。
