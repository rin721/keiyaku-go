---
doc_role: ai_entry
scope: repo
authority_level: entry
owners: [tech-lead]
status: active
effective_date: 2026-05-15
version: 1.0
related_rules: []
read_when: [all_tasks]
update_when: [routing_changed, execution_protocol_changed, metadata_standard_changed, governance_structure_changed]
---

# Claude Code Entry

Claude Code 在本仓库中的入口规则与 `AGENTS.md` 保持一致。

## 读取顺序

1. 先读 [AGENTS.md](AGENTS.md)。
2. 再读 [docs/governance/README.md](docs/governance/README.md)。
3. 再读 [docs/governance/ai-execution.md](docs/governance/ai-execution.md)。
4. 按导航文档中的任务路由加载最小上下文。

本文件只做 Claude Code 薄适配，不复制稳定工程规则，不承担规则真相职责。
