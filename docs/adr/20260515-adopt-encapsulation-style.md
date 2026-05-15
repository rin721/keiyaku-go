---
state_id: ADR-20260515-ENCAP-001
doc_role: adr
memory_level: L0
state_scope: global
scope: repo
authority_level: ssot_decision
owners: [tech-lead]
status: accepted
effective_date: 2026-05-15
version: 1.0
related_rules: [GOV-P0-002, GOV-P1-002, GOV-P1-005]
source_of_truth: [docs/adr/20260515-adopt-encapsulation-style.md]
derived_from: [docs/governance/change-management.md, docs/conventions/pkg.md]
read_when: [governance_change, pkg_change, boundary_sensitive, review_change]
update_when: [default_behavior_changed, convention_changed, review_policy_changed, automation_changed]
conflict_policy: accepted_adr_defines_default_encapsulation_style
rollback_target: [docs/conventions/pkg.md, docs/governance/change-management.md]
verification_target: [scripts/check-governance-sync.ps1, scripts/check-governance-map.ps1, scripts/check-layering.ps1]
---

# ADR 20260515：采用可复用封装风格

- 状态：accepted
- 日期：2026-05-15
- 负责人：tech-lead
- 关联 Issue 或 PR：无

## 背景与上下文
<!-- ADR-001 -->

`pkg/cli` 引入了轻量级 CLI 封装，用于统一 `urfave/cli`、Survey 和 pterm 的使用方式。该实现风格不只适用于 CLI，也适用于后续邮件、缓存、认证、观测、文件处理等可复用能力封装。

如果该风格只停留在代码示例或局部 README 中，后续模块容易继续出现魔法字符串散落、错误类型不统一、概述文档缺少使用说明、业务逻辑穿透共享包等问题。用户已明确要求这类封装风格作为长期默认，因此需要把 Local State 提升为仓库级设计决策，并同步专题约定与评审面。

## 决策内容
<!-- ADR-002 -->

仓库默认采用“可复用封装风格”编写跨模块能力包：

- 能力包必须先明确边界：共享基础能力放在 `pkg`，业务装配和业务流程留在 `cmd` 或 `internal` 的适配边界。
- 新增可复用封装时，应优先拆分概述文档、常量、数据类型、错误类型、核心工具代码和测试。
- 概述文档必须包含使用说明，说明典型入口、参数或配置方式、错误处理方式和最小示例。
- 常量和专用类型应承载稳定名称、枚举、flag、环境变量、状态、错误分类等，避免魔法字符串在业务代码中扩散。
- 错误类型应能表达错误分类、执行阶段和原始错误，并让入口层可以统一退出码、日志或响应映射。
- 封装包不得 import `internal`，不得承载业务流程编排，也不得暴露业务 DTO、PO 或 ORM Model。
- 当前无法稳定自动化判断的风格项进入评审清单；可自动化的依赖方向仍由分层脚本和 lint 约束。

本 ADR 不允许偏离 P0，也不改变 Gin、GORM、手动依赖注入等既有技术方向。

## 后果评估
<!-- ADR-003 -->

正面收益：

- 新能力包更容易阅读、测试和复用。
- 命名、错误、常量和文档结构更加一致。
- `pkg` 与业务层边界更清晰，降低共享包被业务实现污染的风险。

取舍：

- 小型封装会多出少量文件，但换来稳定的阅读路径和演进空间。
- 该风格不能机械套用到所有目录；单点业务逻辑不应为了形式拆出无意义文件。
- 文档使用说明需要维护，能力行为变化时必须同步更新。

## 备选方案

- 只在 `pkg/cli` README 中记录：不足以成为长期默认，也无法指导其他封装。
- 将所有风格写进 `AGENTS.md`：会把稳定工程政策塞回 Prompt，违反治理边界。
- 立即新增脚本强制检查文件结构：当前语义判断较多，自动化容易误报，先放入约定和评审清单。

## 后续事项

- [x] 文档已更新。
- [x] `automation-matrix.md` 与评审清单已按需同步。
- [x] `governance-map.json` 与相关导出/校验脚本已按需同步。
- [x] 如果决策改变了可执行规则，CI 或静态检查已更新：本次仅复用既有分层检查，不新增结构强制脚本。
- [x] 如涉及迁移或高风险变更，已记录回滚或补偿方案：不涉及数据迁移。
