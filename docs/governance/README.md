---
doc_role: navigation
scope: repo
authority_level: ssot_navigation
owners: [tech-lead]
status: active
effective_date: 2026-05-15
version: 1.0
related_rules: []
read_when: [all_tasks, governance_change, context_routing]
update_when: [document_tree_changed, ssot_changed, routing_changed, metadata_standard_changed]
---

# 治理导航

本文档是仓库治理体系的导航真相。它负责说明先读什么、不同任务应加载哪些上下文、不同文档冲突时如何裁决，以及治理变更应从哪里进入。

## SSOT 注册表

- 导航真相：[docs/governance/README.md](/D:/coder/go/keiyaku-go/docs/governance/README.md)
- 仓库级治理规则真相：[docs/governance/rules.md](/D:/coder/go/keiyaku-go/docs/governance/rules.md)
- AI 执行协议真相：[docs/governance/ai-execution.md](/D:/coder/go/keiyaku-go/docs/governance/ai-execution.md)
- 治理演进与冲突处理真相：[docs/governance/change-management.md](/D:/coder/go/keiyaku-go/docs/governance/change-management.md)
- 自动化归位矩阵：[docs/governance/automation-matrix.md](/D:/coder/go/keiyaku-go/docs/governance/automation-matrix.md)
- 治理债务与受控破例登记：[docs/governance/exceptions.yaml](/D:/coder/go/keiyaku-go/docs/governance/exceptions.yaml)
- 设计偏离与重大取舍裁决真相：[docs/adr](/D:/coder/go/keiyaku-go/docs/adr) 下状态为 Accepted 的 ADR
- 人工评审清单：[docs/review/checklist.md](/D:/coder/go/keiyaku-go/docs/review/checklist.md) 与 [docs/review/governance-change-checklist.md](/D:/coder/go/keiyaku-go/docs/review/governance-change-checklist.md)
- 专题设计约定：[docs/conventions](/D:/coder/go/keiyaku-go/docs/conventions) 下的具体文档
- 自动检查与 CI：`scripts/`、`.golangci.yml`、`.gitleaks.toml`、`Makefile` 和 `.github/workflows/`

## 冲突优先级

当文档或自动化表达不一致时，按以下顺序裁决：

1. P0 规则和安全/数据不可逆约束。
2. Accepted ADR。
3. [docs/governance/rules.md](/D:/coder/go/keiyaku-go/docs/governance/rules.md)。
4. 相关专题约定文档。
5. [docs/governance/ai-execution.md](/D:/coder/go/keiyaku-go/docs/governance/ai-execution.md)。
6. 评审清单。
7. `AGENTS.md`、`CLAUDE.md` 或其他工具入口。

如果自动化规则与更高优先级文档冲突，先停止扩散变更，更新规则来源或登记受控破例，再调整脚本/CI。

## First-Hop 路径

所有 AI 开发代理进入仓库后都应使用同一条 first-hop 路径：

1. 读取 [AGENTS.md](/D:/coder/go/keiyaku-go/AGENTS.md) 或工具专属薄入口。
2. 读取本文档，确定 SSOT、冲突优先级和任务路由。
3. 读取 [docs/governance/ai-execution.md](/D:/coder/go/keiyaku-go/docs/governance/ai-execution.md)，确认执行协议。
4. 按任务类型读取最小充分上下文。
5. 在编辑前确认是否需要 ADR、治理债务登记、break-glass 或人工升级。

不要为了保险而默认读完整个 `docs/` 目录。上下文加载应由 `read_when`、任务路径和变更类型共同决定。

## 任务路由矩阵

- `pkg/` 工具包或共享包变更：
  读取 [docs/governance/rules.md](/D:/coder/go/keiyaku-go/docs/governance/rules.md)、[docs/conventions/pkg.md](/D:/coder/go/keiyaku-go/docs/conventions/pkg.md)、[docs/conventions/layering.md](/D:/coder/go/keiyaku-go/docs/conventions/layering.md) 和 [docs/governance/automation-matrix.md](/D:/coder/go/keiyaku-go/docs/governance/automation-matrix.md)。
- 架构边界、分层、依赖方向、DTO/PO/领域模型变更：
  读取 [docs/governance/rules.md](/D:/coder/go/keiyaku-go/docs/governance/rules.md)、[docs/conventions/layering.md](/D:/coder/go/keiyaku-go/docs/conventions/layering.md)、[docs/adr/README.md](/D:/coder/go/keiyaku-go/docs/adr/README.md) 和相关 Accepted ADR。
- 治理规则、导航、AI 执行协议、元数据标准变更：
  读取本文档、[docs/governance/rules.md](/D:/coder/go/keiyaku-go/docs/governance/rules.md)、[docs/governance/ai-execution.md](/D:/coder/go/keiyaku-go/docs/governance/ai-execution.md)、[docs/governance/change-management.md](/D:/coder/go/keiyaku-go/docs/governance/change-management.md)、[docs/governance/automation-matrix.md](/D:/coder/go/keiyaku-go/docs/governance/automation-matrix.md)、[docs/adr/0000-template.md](/D:/coder/go/keiyaku-go/docs/adr/0000-template.md) 和 [docs/review/governance-change-checklist.md](/D:/coder/go/keiyaku-go/docs/review/governance-change-checklist.md)。
- 迁移、回填、灰度、回滚、数据兼容变更：
  读取 [docs/governance/rules.md](/D:/coder/go/keiyaku-go/docs/governance/rules.md)、[docs/conventions/migrations.md](/D:/coder/go/keiyaku-go/docs/conventions/migrations.md)、[docs/templates/migration-plan.md](/D:/coder/go/keiyaku-go/docs/templates/migration-plan.md) 和相关 ADR。
- 异步任务、重试、幂等、定时任务变更：
  读取 [docs/governance/rules.md](/D:/coder/go/keiyaku-go/docs/governance/rules.md)、[docs/conventions/async-jobs.md](/D:/coder/go/keiyaku-go/docs/conventions/async-jobs.md)、[docs/conventions/testing.md](/D:/coder/go/keiyaku-go/docs/conventions/testing.md) 和相关 ADR。
- 测试约定变更：
  读取 [docs/conventions/testing.md](/D:/coder/go/keiyaku-go/docs/conventions/testing.md)、[docs/governance/automation-matrix.md](/D:/coder/go/keiyaku-go/docs/governance/automation-matrix.md) 和 [scripts/check-test-conventions.ps1](/D:/coder/go/keiyaku-go/scripts/check-test-conventions.ps1)。
- CI、lint、治理脚本、安全扫描变更：
  读取 [docs/conventions/ci.md](/D:/coder/go/keiyaku-go/docs/conventions/ci.md)、[docs/conventions/security-logging.md](/D:/coder/go/keiyaku-go/docs/conventions/security-logging.md)、[docs/governance/automation-matrix.md](/D:/coder/go/keiyaku-go/docs/governance/automation-matrix.md)、`Makefile`、`.github/workflows/` 和相关脚本。

## 元数据标准

面向治理的 Markdown 文档必须使用 YAML front matter。除非文档类型确实不适合，否则至少包含：

- `doc_role`：文档角色，例如 `navigation`、`governance_rules`、`ai_execution`、`convention`、`adr`、`review_checklist`、`template`。
- `scope`：适用范围，例如 `repo`、`pkg`、`internal/domain`、`ci`、`security`。
- `authority_level`：裁决权重，例如 `ssot_navigation`、`ssot_rules`、`binding`、`advisory`、`checklist`。
- `owners`：维护责任人或角色。
- `status`：生命周期，例如 `draft`、`active`、`accepted`、`deprecated`。
- `effective_date`：生效日期，格式为 `YYYY-MM-DD`。
- `version`：文档版本。
- `related_rules`：关联的 `rule_id`、ADR 或自动化项。
- `read_when`：触发读取的任务或上下文标签。
- `update_when`：触发同步更新的条件。

机器扫描应优先读取元数据，而不是靠文件名猜测文档职责。

## 维护规则

- 新增或修改治理规则时，必须判断它应归入规则文档、专题约定、ADR、评审清单还是自动化。
- 可由脚本、lint、测试或 CI 稳定检查的规则，不应长期停留在 Prompt 正文里。
- Prompt 和工具入口只保留路由逻辑、执行约束和升级条件。
- 改变默认工程风格、架构边界或治理政策时，必须评估是否需要 ADR。
- 允许历史代码暂时不完全符合新治理，但必须通过 [docs/governance/exceptions.yaml](/D:/coder/go/keiyaku-go/docs/governance/exceptions.yaml) 可追踪。
