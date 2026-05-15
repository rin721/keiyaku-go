---
state_id: GOV-RULES-001
doc_role: governance_rules
memory_level: L0
state_scope: global
scope: repo
authority_level: ssot_rules
owners: [tech-lead]
status: active
effective_date: 2026-05-15
version: 2.0
related_rules: [GOV-P0-001, GOV-P0-002, GOV-P0-003, GOV-P0-004, GOV-P1-001, GOV-P1-002, GOV-P1-003, GOV-P1-004, GOV-P1-005, GOV-P1-006]
source_of_truth: [docs/governance/rules.md]
derived_from: [docs/adr/20260515-default-backend-direction.md, docs/adr/20260515-adopt-gin-gorm-clean-architecture.md]
read_when: [all_tasks, governance_change, pkg_change, boundary_sensitive, migration_sensitive, async_sensitive, test_or_ci, security_sensitive]
update_when: [default_rule_changed, default_behavior_changed, adr_accepted, automation_changed, review_policy_changed]
conflict_policy: ssot_rules_override_lower_layers
rollback_target: [docs/adr/20260515-default-backend-direction.md]
verification_target: [scripts/check-governance.ps1, scripts/check-rule-links.ps1, scripts/check-governance-map.ps1]
---

# 工程治理规则

本文档是 Keiyaku-Go 的仓库级治理规则真相。专题约定、评审清单、脚本、lint 和 CI 都必须引用或落实本文档中的规则，不得另行定义冲突规则。

## 核心原则

- 意图由治理文档和 ADR 表达，事实由代码、契约、配置、迁移脚本、测试和 CI 验证。
- 规则分为 P0 绝对红线、P1 默认要求、P2 推荐实践。
- 可自动化的规则必须进入脚本、lint、测试或 CI；暂不可自动化的规则进入评审清单，并保留迁移到自动化的计划。
- 能力抽象优先于组件绑定。Gin 是传输适配器选择，GORM 是 Repository 实现辅助工具，二者不得穿透到领域模型或传输边界之外。

## 当前技术阶段

当前仓库已建立 Go module 基线，并进入 Gin + GORM 的应用骨架实现阶段。后端默认方向为 Gin + GORM v2，默认依赖注入方式为手动装配；当前应用实现方向由 [20260515-adopt-gin-gorm-clean-architecture.md](../adr/20260515-adopt-gin-gorm-clean-architecture.md) 裁决。任何再次改变默认技术方向的决策必须通过 ADR。

## P0 绝对红线

### GOV-P0-001 禁止模型与契约穿透

- 持久层模型，包括 PO 与 ORM Model，不得直接返回给 Handler 或序列化为 JSON 响应。
- Repository 与 Infrastructure 方法签名不得依赖传输 DTO 类型。
- Handler 必须先把 DTO 转换为 Application 层的 Command、Query、Criteria、Domain Object 或 Value Object，再向内传递。

### GOV-P0-002 禁止逆向依赖

- `pkg/` 只放业务无关的可复用组件，不得 import 任何 `internal/` 包。
- Domain、Repository、Infrastructure 不得依赖 `internal/api`、`internal/handler` 或 `handler/types` 中的传输契约。
- 底层必须对 HTTP、RPC、OpenAPI DTO 等外部通信契约保持失明。

### GOV-P0-003 禁止高危哈希用于密码存储

- 密码存储不得使用 MD5、SHA-1、SHA-256 或其他快速哈希算法。
- 密码存储必须使用 Argon2id 或 bcrypt，并统一定义成本参数。
- 密码哈希参数必须支持版本化，并支持登录时渐进式 rehash。
- SHA-256 或 SHA-512 仅可用于文件完整性、ETag、HMAC 等非密码场景。

### GOV-P0-004 禁止敏感信息明文落日志

- 日志系统必须具备脱敏 hook 或等效脱敏边界。
- 代码不得通过 `zap.Any`、`fmt %+v`、JSON dump 或类似方式打印完整请求体、配置对象、用户对象或第三方凭据对象。
- 敏感信息只能通过明确的脱敏白名单结构输出。

## P1 默认要求

### GOV-P1-001 全链路追踪

- 所有入口请求必须提取或生成 TraceID。
- 内部 HTTP/RPC 调用必须透传 TraceID。
- 所有结构化日志必须包含 TraceID。
- 异步任务、MQ 消息与定时任务必须携带 TraceID 或 CorrelationID。

### GOV-P1-002 务实 DDD 分层

- 分层职责、依赖方向、DTO 隔离与 Repository Port 归属以 `docs/conventions/layering.md` 为准。
- 可自动化的依赖方向检查必须进入 `scripts/check-layering.ps1`、`.golangci.yml` 或 CI。
- 模型泄露、抽象归属与边界语义等难以可靠自动化的判断进入评审清单。

### GOV-P1-003 数据库迁移治理

- DDL 变更必须可重复部署、可回滚或可补偿。
- 无法天然幂等的 migration 必须依赖版本表、防重复执行机制与前置检查。
- 重大表结构变更必须采用灰度策略，例如冗余列、双写、回填、读切换、旧字段下线。
- MySQL 风格 DDL 不得依赖“全量事务回滚”的技术假设。

### GOV-P1-004 异步任务模型

- 预计超过接口延迟预算、不可因进程退出而丢失、需要重试、需要削峰或需要跨实例消费的任务必须进入异步系统。
- 底层任务系统必须具备持久化、重试、死信队列、优雅关闭与可观测能力。
- 消费端必须通过幂等键、状态机、唯一约束或去重表处理重复投递。

### GOV-P1-005 测试执行策略

- Domain 与 Application 逻辑默认使用快速单元测试，并使用 Mock 或 Fake 屏蔽底层依赖。
- Repository 与 Infrastructure 行为默认通过 `testcontainers-go` 或等效方式验证真实中间件。
- 集成测试必须纳入 CI，但要通过 build tag 或 Make 目标与本地快速测试隔离。

### GOV-P1-006 依赖注入策略

- 默认采用手动依赖注入。
- 依赖图谱庞大的模块可使用 Wire 等编译期代码生成工具，生成代码必须提交并参与 review。
- 禁止引入运行时反射型依赖注入容器。

## P2 推荐实践

- 推荐接入 OpenTelemetry，实现 Trace、Metrics、Logs 关联。
- 推荐通过 Prometheus 兼容方式暴露运行时与核心业务指标。
- 简单结构映射推荐手写；高频或复杂映射推荐使用编译期生成工具。
- 局部高并发处理推荐使用有界并发。
- 超大文件导入导出推荐使用流式 API。
- 文件监听推荐结合事件去抖与定时全量扫描。
- 邮件发送应隐藏在 `pkg/mailer` 能力抽象之后。
- 推荐通过 Makefile 与 pre-commit 收口格式化、lint、静态检查与安全扫描。
