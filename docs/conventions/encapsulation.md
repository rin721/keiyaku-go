---
state_id: CONV-ENCAP-001
doc_role: convention
memory_level: L1
state_scope: module
scope: repo
authority_level: binding
owners: [tech-lead]
status: active
effective_date: 2026-05-15
version: 1.0
related_rules: [GOV-P0-002, GOV-P1-002, GOV-P1-005]
source_of_truth: [docs/adr/20260515-adopt-encapsulation-style.md, docs/governance/rules.md]
derived_from: [docs/adr/20260515-adopt-encapsulation-style.md, docs/conventions/pkg.md]
read_when: [governance_change, pkg_change, boundary_sensitive, review_change]
update_when: [default_behavior_changed, convention_changed, review_policy_changed, automation_changed]
conflict_policy: binding_must_yield_to_accepted_adr_and_rules
rollback_target: [docs/adr/20260515-adopt-encapsulation-style.md, docs/conventions/pkg.md]
verification_target: [scripts/check-layering.ps1, scripts/check-governance-map.ps1, scripts/check-governance-sync.ps1]
---

# 可复用封装风格约定

本文档定义仓库内可复用能力封装的默认风格。仓库级规则真相仍是 [docs/governance/rules.md](../governance/rules.md)，本约定由 [ADR 20260515：采用可复用封装风格](../adr/20260515-adopt-encapsulation-style.md) 裁决。

## 适用范围

适用于跨模块复用、可独立测试、希望形成稳定调用面的能力封装，例如 CLI、配置辅助、邮件、缓存、观测、文件处理、加密辅助、任务调度辅助等。

不适用于只服务单个用例的业务流程、一次性脚本、领域实体、传输 DTO 或持久层模型。

## 默认文件结构

可复用封装推荐按职责拆分文件，而不是把所有内容塞进一个大文件：

- `doc.go`：包级概述，说明边界和职责。
- `README.md`：概述文档，必须包含使用说明和最小示例。
- `constants.go`：默认值、稳定名称、枚举值、状态值和避免魔法字符串的常量。
- `types.go`：公开数据类型、配置结构、接口、上下文类型。
- `errors.go`：错误分类、错误类型、包装函数和错误码/退出码/状态映射。
- `options.go` 或 `config.go`：可选配置、默认配置和校验逻辑。
- `*.go` 核心实现文件：按能力拆分，文件名表达行为而不是业务场景。
- `*_test.go`：覆盖公开行为、错误包装和边界条件。

文件结构可以按复杂度裁剪。简单封装不必为了形式创建空文件，但一旦出现常量、类型、错误和核心逻辑混杂，应优先拆分。

## 概述文档要求

`README.md` 至少说明：

- 这个包解决什么问题，不解决什么问题。
- 典型入口函数或构造函数。
- 常用参数、flag、配置或环境变量。
- 错误处理方式。
- 一个可以复制改造的最小示例。
- 与分层、业务装配或安全边界相关的注意事项。

文档链接必须使用相对路径，不得写入本机绝对路径。

## 常量与数据类型

- 稳定字符串应通过常量或专用类型承载，例如命令名、flag 名、环境变量名、状态、阶段、错误分类。
- 枚举型字符串应定义专用类型，避免在调用点直接比较裸字符串。
- 配置结构应表达默认值、必填项和边界含义，避免把校验逻辑分散到调用方。
- 对外接口应描述能力，而不是绑定具体业务场景。

## 错误类型

- 可复用封装应提供统一错误类型或错误包装函数。
- 错误应保留原始 `error`，支持 `errors.Is` / `errors.As`。
- 错误分类应服务调用方决策，例如用法错误、交互错误、运行时错误、依赖错误。
- 入口层负责把错误映射为退出码、HTTP 响应或日志字段，封装包不应直接决定业务响应结构。

## 边界规则

- `pkg` 下的封装不得 import `internal`。
- 可复用封装不得承载业务流程编排。
- 业务 DTO、PO、ORM Model 不得作为通用封装的公开契约。
- 具体技术库可以被封装在能力包内部，但公开 API 应优先表达仓库需要的能力语义。

## 测试要求

- 至少覆盖公开构造函数、核心行为和错误分支。
- 交互、终端、外部依赖等难以直接测试的能力应通过接口抽象或测试替身隔离。
- 共享封装的测试应保持快速，避免默认依赖真实外部服务。

## 评审重点

- 文件拆分是否服务可读性，而不是形式化堆文件。
- README 是否包含实际使用说明。
- 是否存在可替换为常量或专用类型的魔法字符串。
- 错误是否保留分类和原始原因。
- 公开 API 是否泄露业务语义或具体适配层模型。
