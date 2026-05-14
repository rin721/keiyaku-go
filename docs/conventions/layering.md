---
doc_role: convention
scope: layering
authority_level: binding
owners: [tech-lead]
status: active
effective_date: 2026-05-15
version: 1.0
related_rules: [GOV-P0-001, GOV-P0-002, GOV-P1-001, GOV-P1-002, GOV-P1-003]
read_when: [boundary_sensitive, pkg_change, governance_change]
update_when: [default_rule_changed, default_behavior_changed, adr_accepted, automation_changed]
---

# 分层约定

本文档是包边界、依赖方向和 DTO 隔离的局部约定。仓库级规则真相是 [docs/governance/rules.md](../governance/rules.md)，本文档将其中的分层规则细化为日常执行约定。

## 分层职责

- `internal/api`：HTTP、RPC、CLI 或其他传输入口；只负责协议适配、请求解析、响应映射和调用应用服务。
- `internal/application`：用例编排、事务边界、命令/查询处理、端口接口定义。
- `internal/domain`：领域实体、值对象、领域服务和领域不变量。
- `internal/infrastructure`：数据库、队列、缓存、外部服务、文件系统、时钟等具体适配器。
- `internal/repository`：只有在仓库约定明确采用独立 repository 层时使用；否则 repository port 应归属于 application/domain，具体实现放在 infrastructure。
- `pkg`：可复用、无业务私有依赖的工具包。

## 依赖方向

- 传输层可以依赖 application，不应直接依赖 infrastructure 实现。
- application 可以依赖 domain 和它自己定义的 port，不应依赖 transport DTO 或具体数据库实现。
- domain 不应依赖 application、infrastructure、transport、ORM 或框架类型。
- infrastructure 可以实现 application/domain 暴露的 port。
- `pkg` 不应导入 `internal`。

## DTO 与模型边界

- Transport DTO 只能存在于入口层附近，不应泄露到 domain。
- Persistence model 不应作为 domain model 在业务逻辑中流动。
- 跨边界传递时应显式映射，避免隐式复用结构体造成语义污染。
- 如果为了性能或兼容性需要复用结构，必须有注释说明边界原因，并在必要时登记 ADR 或治理债务。

## Repository Port 归属

- 默认将 repository port 放在调用它的 application/domain 边界一侧。
- 具体数据库实现放在 infrastructure。
- 如果引入独立 `internal/repository` 层，需要说明它解决的边界问题，并避免变成贫血转发层。
- 变更默认 repository 归属模型时，需要 ADR。

## 自动化

分层规则优先由 [scripts/check-layering.ps1](../../scripts/check-layering.ps1) 检查。人工评审只处理脚本难以判断的语义问题，例如 DTO/PO 语义泄露、port 归属是否合理、例外是否需要 ADR。
