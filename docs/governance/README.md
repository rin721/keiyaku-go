---
doc_role: navigation
scope: repo
authority_level: ssot_navigation
owners: [tech-lead]
status: active
effective_date: 2026-05-15
version: 1.0
related_rules: []
read_when: [all_tasks, governance_change]
update_when: [routing_changed, metadata_standard_changed, governance_structure_changed, default_rule_changed, adr_accepted]
---

# 治理导航

本文档是仓库治理体系的导航真相。它只负责 first-hop、冲突优先级、任务路由和读取顺序，不复制稳定工程规则。

## SSOT 注册表

- 导航真相：[docs/governance/README.md](README.md)
- 仓库级治理规则真相：[docs/governance/rules.md](rules.md)
- AI 执行协议真相：[docs/governance/ai-execution.md](ai-execution.md)
- 元数据标准真相：[docs/governance/metadata-schema.md](metadata-schema.md)
- 治理演进与冲突处理真相：[docs/governance/change-management.md](change-management.md)
- 自动化归位矩阵：[docs/governance/automation-matrix.md](automation-matrix.md)
- 治理债务与受控破例登记：[docs/governance/exceptions.yaml](exceptions.yaml)
- 设计偏离与重大取舍裁决真相：[docs/adr](../adr) 下状态为 Accepted 的 ADR
- 人工评审清单：[docs/review/checklist.md](../review/checklist.md) 与 [docs/review/governance-change-checklist.md](../review/governance-change-checklist.md)
- 专题设计约定：[docs/conventions](../conventions) 下的具体文档
- 自动检查与 CI：`scripts/`、`.golangci.yml`、`.gitleaks.toml`、`Makefile` 和 `.github/workflows/`

## 冲突优先级

当文档或自动化表达不一致时，按以下顺序裁决：

1. P0 规则和安全/数据不可逆约束。
2. Accepted ADR。
3. [docs/governance/rules.md](rules.md)。
4. 相关专题约定文档。
 5. [docs/governance/metadata-schema.md](metadata-schema.md)。
 6. [docs/governance/ai-execution.md](ai-execution.md)。
 7. 评审清单。
 8. `AGENTS.md`、`CLAUDE.md` 或其他工具入口。

如果自动化规则与更高优先级文档冲突，先停止扩散变更，更新规则来源或登记受控破例，再调整脚本/CI。

## First-Hop 路径

所有 AI 开发代理进入仓库后都应使用同一条 first-hop 路径：

1. 读取 [AGENTS.md](../../AGENTS.md) 或工具专属薄入口。
2. 读取本文档，确定 SSOT、冲突优先级和任务路由。
3. 读取 [docs/governance/ai-execution.md](ai-execution.md)，确认执行协议。
4. 按任务类型读取最小充分上下文。
5. 在编辑前确认是否需要 ADR、治理债务登记、break-glass 或人工升级。

不要为了保险而默认读完整个 `docs/` 目录。上下文加载应由 `read_when`、任务标签和变更类型共同决定。

## 任务路由矩阵

- `pkg_change`：
  读取 [docs/governance/rules.md](rules.md)、[docs/conventions/pkg.md](../conventions/pkg.md)、[docs/conventions/layering.md](../conventions/layering.md) 和 [docs/governance/automation-matrix.md](automation-matrix.md)。
- `boundary_sensitive`：
  读取 [docs/governance/rules.md](rules.md)、[docs/conventions/layering.md](../conventions/layering.md)、[docs/conventions/dependency-injection.md](../conventions/dependency-injection.md)、[docs/governance/change-management.md](change-management.md)、[docs/adr/README.md](../adr/README.md) 和相关 Accepted ADR。
- `governance_change`：
  读取本文档、[docs/governance/rules.md](rules.md)、[docs/governance/ai-execution.md](ai-execution.md)、[docs/governance/metadata-schema.md](metadata-schema.md)、[docs/governance/change-management.md](change-management.md)、[docs/governance/automation-matrix.md](automation-matrix.md)、[docs/adr/0000-template.md](../adr/0000-template.md) 和 [docs/review/governance-change-checklist.md](../review/governance-change-checklist.md)。
- `migration_sensitive`：
  读取 [docs/governance/rules.md](rules.md)、[docs/conventions/migrations.md](../conventions/migrations.md)、[docs/migrations/gray-release-template.md](../migrations/gray-release-template.md) 和相关 ADR。
- `async_sensitive`：
  读取 [docs/governance/rules.md](rules.md)、[docs/conventions/async-jobs.md](../conventions/async-jobs.md)、[docs/conventions/testing.md](../conventions/testing.md) 和相关 ADR。
- `test_or_ci`：
  读取 [docs/conventions/ci.md](../conventions/ci.md)、[docs/conventions/security-logging.md](../conventions/security-logging.md)、[docs/governance/automation-matrix.md](automation-matrix.md)、`Makefile`、`.github/workflows/` 和相关脚本。
- `security_sensitive`：
  读取 [docs/governance/rules.md](rules.md)、[docs/conventions/security-logging.md](../conventions/security-logging.md)、[docs/governance/automation-matrix.md](automation-matrix.md)。
- `review_change`：
  读取 [docs/review/checklist.md](../review/checklist.md)、[docs/review/governance-change-checklist.md](../review/governance-change-checklist.md) 和 [docs/governance/automation-matrix.md](automation-matrix.md)。
- `exception_review`：
  读取 [docs/governance/change-management.md](change-management.md)、[docs/governance/exceptions.yaml](exceptions.yaml)、[docs/governance/exceptions.template.yaml](exceptions.template.yaml) 和相关 ADR。

## 元数据标准

面向治理的 Markdown 和 YAML 治理文档必须使用 YAML front matter。字段定义与受控枚举以 [docs/governance/metadata-schema.md](metadata-schema.md) 为准，至少包含：

- `doc_role`：文档角色，取值受 `metadata-schema.md` 控制。
- `scope`：适用范围，取值受 `metadata-schema.md` 控制。
- `authority_level`：裁决权重，取值受 `metadata-schema.md` 控制。
- `owners`：维护责任人或角色。
- `status`：生命周期，取值受 `metadata-schema.md` 控制。
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
- 改变默认工程风格、默认边界、默认依赖方式、默认工作流或治理政策时，必须评估是否需要 ADR 和闭环同步。
- 允许历史代码暂时不完全符合新治理，但必须通过 [docs/governance/exceptions.yaml](exceptions.yaml) 可追踪。
- Markdown 文档链接必须使用相对路径；不得写入本机绝对路径，例如 Windows 盘符路径或 `/Users/...`、`/home/...` 这类机器相关路径。
