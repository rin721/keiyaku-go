# Keiyaku 工程治理规范

本文档是 Keiyaku 工程治理的设计意图来源。代码、契约、配置、迁移脚本与 CI 检查是可验证事实，必须与本文档保持一致。

## 核心原则

1. **意图即文档，事实即代码**：ADR 记录设计意图。代码、API 契约、配置、数据库迁移与 CI 检查共同构成最终可验证事实。CI 必须发现意图与事实之间的漂移。
2. **规范分级管控**：规范分为 P0 绝对红线、P1 默认要求、P2 推荐实践。
3. **能力抽象优于组件绑定**：架构优先定义系统需要的能力模型，再把具体库或中间件作为可替换实现接入。

## 默认技术方向

未来后端默认方向为 **Echo + sqlc**，但本仓库第一阶段只落地工程治理基线。本次不引入 HTTP API、数据库 schema、Go module 或业务代码。

后续实现必须把框架与数据库选择约束在能力接口之后。Echo 是传输适配器选择，sqlc 是 Repository 实现辅助工具；两者都不得导致传输 DTO 或持久层模型跨层泄露。

## P0：绝对红线

任何 P0 违规都必须阻断合入。可自动检测的规则必须进入 CI；暂不可自动检测的规则必须进入代码评审清单，并逐步沉淀为静态检查。

### 禁止模型与契约穿透

<!-- GOV-P0-001 -->

- 持久层模型，包括 PO 与 GORM Model，绝不能直接返回给 Handler 或序列化为 JSON 响应。
- Repository 与 Infrastructure 方法签名不得依赖 DTO 类型。
- Handler 必须先把 DTO 转换为 Application 层的 Command、Query、Criteria、DO 或 Value Object，再继续向下传递。

### 禁止逆向依赖

<!-- GOV-P0-002 -->

- `pkg/` 只放业务无关的可复用组件，不得 import 任何 `internal/` 包。
- Domain、Repository、Infrastructure 代码不得依赖 `internal/api`、`internal/handler` 或 `handler/types` 中的传输契约。
- 底层必须对 HTTP、RPC、OpenAPI DTO 等外部通信契约保持失明。

### 禁止高危哈希用于密码存储

<!-- GOV-P0-003 -->

- 密码存储不得使用 MD5、SHA-1、SHA-256 或其他快速哈希算法。
- 密码存储必须使用 Argon2id 或 bcrypt，并统一定义成本参数。
- 密码哈希参数必须支持版本化，并支持登录时渐进式 Rehash。
- SHA-256 或 SHA-512 仍可用于文件完整性、ETag、HMAC 等非密码场景，但必须符合对应安全模型。

### 禁止敏感信息明文落日志

<!-- GOV-P0-004 -->

- 日志系统必须具备脱敏 Hook 或等效脱敏边界。
- 代码不得通过 `zap.Any`、`fmt %+v`、JSON dump 或类似整体结构化输出方式打印完整请求体、配置对象、用户对象或三方凭据对象。
- 敏感信息只允许通过明确的脱敏白名单结构输出。

## P1：默认要求

P1 是工程默认标准。任何偏离都必须提交 ADR，并由技术负责人审批。

### 全链路追踪

<!-- GOV-P1-001 -->

- 所有入口请求必须提取或生成 TraceID。
- 内部 HTTP/RPC 调用必须透传 TraceID。
- 所有结构化日志必须包含 TraceID。
- 异步任务、MQ 消息与定时任务必须携带 TraceID 或 CorrelationID。

### 务实 DDD 分层

<!-- GOV-P1-002 -->

- `internal/api` 或 `internal/handler`：负责传输协议解析、入参校验与 DTO 装配。
- `internal/application`：负责用例编排、事务边界、Command、Query 与 Criteria。
- `internal/domain`：负责核心业务逻辑、DO、实体、值对象与业务不变量。
- Repository 接口按场景归属：
  - 领域对象持久化 Port 放在 `internal/domain`；
  - 面向用例的复杂查询或报表 Port 放在 `internal/application`。
- `internal/repository` 或其他基础设施 Adapter 只负责实现 Port，不得定义上层业务接口。
- `pkg`：业务无关共享库，不得向内依赖 `internal`。

### 数据库迁移治理

- DDL 变更必须可重复部署、可回滚或可补偿。
- 对无法天然幂等的 Migration，必须依赖 Migration 版本表、防重复执行机制与前置检查。
- 重大表结构变更必须采用灰度策略，例如冗余列、双写、回填、读切换、旧字段下线。
- MySQL 风格 DDL 变更不得依赖“全量事务回滚”的技术假设。

### 异步任务模型

当任务预计超过接口时延预算、不可因进程退出而丢失、需要重试、需要削峰，或需要跨实例消费时，必须进入异步系统。

底层任务系统必须具备持久化、重试、死信队列、优雅关闭与可观测能力。业务消费端必须通过幂等键、状态机、唯一约束或去重表保证重复投递下的幂等处理。

### 测试执行策略

- Domain 与 Application 逻辑必须通过快速单元测试覆盖，并使用 Mock 或 Fake 屏蔽底层依赖。
- Repository 与 Infrastructure 行为必须通过 `testcontainers-go` 拉起真实中间件验证。
- 集成测试必须纳入 CI，但要通过 Build Tag 或 Make 目标与本地快速测试隔离，例如 `go test -tags=integration ./...`。

### 依赖注入策略

- 默认采用手动依赖注入。
- 对依赖图谱庞大的模块，可使用 Wire 等编译期代码生成工具，生成代码必须提交并参与 Review。
- 禁止引入运行时反射型依赖注入容器。

## P2：推荐实践

- 推荐接入 OpenTelemetry，实现 Trace、Metrics、Logs 关联。
- 推荐通过 Prometheus 暴露运行时与核心业务指标。
- 简单结构推荐手写 Mapper；高频或复杂映射推荐使用 goverter、protoc、sqlc 等编译期生成工具。
- 局部高频计算推荐使用有界并发。
- 超大文件导入导出推荐使用流式 API。
- 文件监听推荐结合事件去抖与定时全量扫描。
- 邮件发送应隐藏在 `pkg/mailer` 能力抽象之后；生产环境优先使用信誉良好的 SMTP Relay。
- 推荐通过 Makefile 与 pre-commit 在本地收口格式化、lint、静态检查与安全扫描。

## CI 治理基线

CI 必须包含：

- 治理文档存在性检查。
- 架构规则检查，阻断 `pkg` import `internal`，并阻断底层 import 传输契约。
- `golangci-lint` 配置，至少包含 `errcheck`、`gosec`、`gocritic`、`wrapcheck`。
- 通过 gitleaks 或同等工具进行 Secret 扫描。
- API 契约出现后，补充 OpenAPI Drift Check。
