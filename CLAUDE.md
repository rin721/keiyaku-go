---
state_id: GOV-ENTRY-002
doc_role: ai_entry
memory_level: L0
state_scope: global
scope: repo
authority_level: entry
owners: [tech-lead]
status: active
effective_date: 2026-05-15
version: 2.0
related_rules: []
source_of_truth: [AGENTS.md, docs/governance/README.md, docs/governance/ai-execution.md]
derived_from: [AGENTS.md]
read_when: [all_tasks]
update_when: [routing_changed, execution_protocol_changed, metadata_standard_changed, governance_structure_changed]
conflict_policy: entry_must_yield_to_repo_governance
rollback_target: [AGENTS.md, docs/governance/README.md]
verification_target: [scripts/check-governance.ps1, scripts/check-governance-sync.ps1]
---

# Claude Code Entry

Claude Code 在本仓库中的入口规则与 `AGENTS.md` 保持一致。

## 读取顺序

1. 先读 [AGENTS.md](AGENTS.md)。
2. 再读 [docs/governance/README.md](docs/governance/README.md)。
3. 再读 [docs/governance/ai-execution.md](docs/governance/ai-execution.md)。
4. 如果任务属于 `governance_change`，或跨多个治理作用域，读取 [docs/governance/governance-map.json](docs/governance/governance-map.json)。
5. 按导航文档中的任务路由加载最小上下文。

本文件只做 Claude Code 薄适配，不复制稳定工程规则，不承担规则真相职责。
