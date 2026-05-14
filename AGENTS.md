---
doc_role: ai_entry
scope: repo
authority_level: binding_entry
owners: [tech-lead]
status: active
effective_date: 2026-05-15
version: 1.0
related_rules: []
read_when: [all_tasks]
update_when: [routing_changed, metadata_standard_changed, execution_protocol_changed]
---

# AI 代理入口

本文档是 AI 开发代理进入本仓库的 first-hop 入口。

## 必读顺序

1. 读取 [docs/governance/README.md](/D:/coder/go/keiyaku-go/docs/governance/README.md)。
2. 读取 [docs/governance/ai-execution.md](/D:/coder/go/keiyaku-go/docs/governance/ai-execution.md)。
3. 只读取导航文档按任务路由出的治理上下文。
4. 在治理路由清晰之后，再阅读代码。

不要默认加载全部治理文档。先完成路由，再加载满足任务所需的最小上下文集合。

## 当前 SSOT 地图

- 导航真相：[docs/governance/README.md](/D:/coder/go/keiyaku-go/docs/governance/README.md)
- AI 执行真相：[docs/governance/ai-execution.md](/D:/coder/go/keiyaku-go/docs/governance/ai-execution.md)
- 仓库级治理规则真相：[docs/governance/rules.md](/D:/coder/go/keiyaku-go/docs/governance/rules.md)
- 偏离与设计决策真相：[docs/adr](/D:/coder/go/keiyaku-go/docs/adr) 下状态为 Accepted 的 ADR
- 治理破例与债务登记真相：[docs/governance/exceptions.yaml](/D:/coder/go/keiyaku-go/docs/governance/exceptions.yaml)

## 快速路由

- `pkg/` 工具包或共享包变更：
  读取导航、执行协议、仓库级治理规则、pkg 专题约定，再读取分层约定。如果变更会改变默认包风格或边界预期，还需要读取 ADR 指引。
- 分层、依赖方向、DTO、领域模型、Repository 边界变更：
  读取导航、执行协议、仓库级治理规则、分层约定，再读取 ADR 指引。
- 迁移、灰度发布、回填、发布、回滚变更：
  读取导航、执行协议、仓库级治理规则、迁移约定、迁移模板；如果变更高风险或偏离默认方案，还需要读取 ADR 指引。
- 异步任务、重试、幂等、定时任务变更：
  读取导航、执行协议、仓库级治理规则、异步任务约定、测试约定；如果变更改变默认模型，还需要读取 ADR 指引。
- 测试、CI、lint、安全扫描、治理脚本变更：
  读取导航、执行协议、仓库级治理规则、测试/CI/安全约定、相关约定脚本和 workflow 文件。
- 治理或 Prompt 体系变更：
  读取导航、执行协议、仓库级治理规则、ADR 指引、评审清单和治理自动化脚本。

## 元数据标准

所有面向治理的文档都应携带 YAML front matter。字段定义以导航真相文档为准，但代理至少必须识别以下字段：

- `doc_role`
- `scope`
- `authority_level`
- `owners`
- `status`
- `version` 或 `effective_date`
- `related_rules`
- `read_when`
- `update_when`

代理应使用元数据做发现和路由，而不是只依赖文件名。

## 禁止事项

- 不要把全部规则折叠进单个 Prompt 文件。
- 不要把 Prompt 当作稳定工程政策的真相来源。
- 当变更会改变默认行为、架构边界或治理政策时，不要绕过 ADR。
