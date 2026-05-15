---
state_id: ADR-20260515-BACKEND-002
doc_role: adr
memory_level: L0
state_scope: global
scope: repo
authority_level: ssot_decision
owners: [tech-lead]
status: accepted
effective_date: 2026-05-15
version: 1.0
related_rules: [GOV-P0-001, GOV-P0-002, GOV-P1-001, GOV-P1-002, GOV-P1-003, GOV-P1-005, GOV-P1-006]
source_of_truth: [docs/adr/20260515-adopt-gin-gorm-clean-architecture.md]
derived_from: [docs/governance/rules.md, docs/adr/20260515-default-backend-direction.md, docs/conventions/layering.md, docs/conventions/dependency-injection.md]
read_when: [governance_change, boundary_sensitive, migration_sensitive, security_sensitive]
update_when: [default_rule_changed, default_behavior_changed, convention_changed, adr_accepted]
conflict_policy: accepted_adr_overrides_default_backend_direction
rollback_target: [docs/adr/20260515-default-backend-direction.md, docs/governance/rules.md]
verification_target: [scripts/check-governance-sync.ps1, scripts/check-governance-map.ps1, scripts/check-layering.ps1]
change_reason: align implementation stack with product architecture request
---

# ADR 20260515：采用 Gin 与 GORM 的 Clean Architecture 后端方向

- 状态：accepted
- 日期：2026-05-15
- 负责人：tech-lead

## 背景与上下文

仓库已有默认方向 ADR，原定未来后端采用 Echo 与 sqlc，并保持手动依赖注入。当前产品实施计划要求以博客/内容管理系统为例，从零开始落地企业级 Go 后端，并明确指定 Gin、GORM v2、MySQL 8.0、Redis、Zap、Viper、JWT、Casbin 与 Snowflake。

该要求改变了默认 HTTP 框架和 Repository 实现辅助工具，属于默认行为变化。为了避免实现阶段形成未授权偏离，需要用新的 ADR 对项目实现方向做出裁决。

## 决策内容

本项目应用代码的默认后端方向调整为：

- HTTP 传输适配器采用 Gin。
- MySQL 访问与持久化映射采用 GORM v2。
- Redis 客户端采用 go-redis/v9。
- 日志采用 Uber Zap，输出结构化 JSON，并预留动态日志级别与脱敏边界。
- 配置采用 Viper，支持 YAML 与环境变量覆盖。
- 认证采用 JWT，授权采用 Casbin v3。
- 业务 ID 采用 Twitter Snowflake 算法，不使用自增 ID 作为外部业务标识。
- 依赖注入继续采用显式手动构造函数，不引入运行时反射型 DI 容器。

该决策不降低 P0/P1 分层要求：Gin DTO、GORM Model、Redis、Casbin、JWT 等外部框架类型不得穿透到 Domain；Repository Port 归属于 Application/Domain 边界侧，具体实现放在 Infrastructure。

## 后果评估

正面收益是实现方向与产品目标一致，Gin 与 GORM 的生态可加速博客/CMS 首版落地，并保持手动装配带来的依赖可读性。代价是需要同步治理规则与自动化预期，并在评审中持续关注 DTO/PO/Domain Model 隔离，防止 GORM Model 被误用为领域模型。

## 备选方案

- 继续沿用 Echo + sqlc。
  优点是与既有默认 ADR 完全一致；缺点是违背当前明确技术约束。
- 引入运行时 DI 容器。
  缺点是隐藏依赖图，与仓库依赖注入约定冲突，因此不采纳。

## 后续事项

- [x] 更新仓库级规则中的默认后端方向。
- [x] 新增架构设计、API 契约和核心表结构文档。
- [x] 实现首版工程骨架与手动依赖装配入口。
- [x] 运行治理、分层与 Go 测试验证。
