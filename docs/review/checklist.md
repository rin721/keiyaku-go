# 代码评审清单

每个 Pull Request 都必须使用此清单。P0 项一旦违规，必须阻断合入。

## P0：绝对红线

- [ ] 模型安全：Repository 或 Infrastructure 没有把 PO、GORM Model 或持久层模型直接返回给 Handler 或 JSON 响应。
<!-- REV-P0-001 -->
- [ ] 契约安全：Repository、Domain、Infrastructure 的方法签名不依赖 `internal/api`、`internal/handler` 或 `handler/types` 中的 DTO。
<!-- REV-P0-002 -->
- [ ] 依赖方向：`pkg/` 不 import `internal/`；底层不 import 上层传输契约。
- [ ] 密码存储：密码代码使用 Argon2id 或 bcrypt，并具备统一、版本化成本参数与渐进式 Rehash 能力。
- [ ] 密码存储：密码场景没有使用 MD5、SHA-1、SHA-256 或其他快速哈希算法。
- [ ] 敏感日志：没有通过 `zap.Any`、`%+v`、JSON dump 或类似整体输出方式打印完整请求体、配置对象、用户对象或三方凭据对象。
<!-- REV-P0-003 -->
- [ ] 敏感日志：任何敏感日志输出都经过明确的脱敏白名单结构。

## P1：默认要求

- [ ] 链路追踪：入口请求会提取或生成 TraceID。
<!-- REV-P1-001 -->
- [ ] 链路追踪：下游 HTTP/RPC 调用会透传 TraceID。
- [ ] 链路追踪：结构化日志始终包含 TraceID。
- [ ] 异步追踪：后台任务、消息与定时任务携带 TraceID 或 CorrelationID。
- [ ] 分层约束：Handler 负责入参校验，并把 DTO 转换为 Application 层 Command、Query、Criteria、DO 或 Value Object。
- [ ] 分层约束：Application 用例负责业务编排与事务边界。
- [ ] 分层约束：Domain 负责业务不变量，并保持与传输协议无关。
- [ ] Repository Port：领域持久化 Port 放在 Domain；复杂用例查询 Port 放在 Application。
- [ ] Migration 安全：DDL 变更可重复、可回滚或可补偿，并具备版本表与前置检查保护。
- [ ] Migration 安全：重大表结构变更使用增量灰度步骤，而不是依赖 DDL 事务回滚。
- [ ] 异步可靠性：需要重试、持久化、削峰、跨实例处理或长时间执行的任务进入异步系统。
- [ ] 幂等设计：消息与任务消费者通过幂等键、状态机、唯一约束或去重表处理重复投递。
<!-- REV-P1-002 -->
- [ ] 测试策略：Domain 与 Application 逻辑有快速单元测试，并使用 Mock 或 Fake。
- [ ] 测试策略：Repository 与 Infrastructure 通过 `testcontainers-go` 使用真实中间件做集成测试。
- [ ] 测试策略：集成测试通过 Build Tag 或 Make 目标与快速测试隔离。
- [ ] 依赖注入：默认使用手动依赖注入；使用 Wire 等编译期生成时，生成代码必须参与 Review。
- [ ] 依赖注入：没有引入运行时反射型 DI 容器。

## P2：推荐实践

- [ ] 可观测性：需要跨服务关联时，使用 OpenTelemetry 关联 Trace、Metrics 与 Logs。
- [ ] 指标：核心运行时与业务信号通过 Prometheus 兼容方式暴露。
- [ ] 数据映射：简单 DTO/DO/PO 映射手写；高频或复杂映射使用编译期生成。
- [ ] 资源控制：本地高并发处理使用有界并发。
- [ ] 文件处理：大文件使用流式 API。
- [ ] 文件监听：文件监听同时具备事件去抖与定时全量扫描。
- [ ] 邮件能力：发信能力隐藏在 `pkg/mailer` 抽象之后。
- [ ] Git 自动化：Makefile 与 pre-commit 在本地执行格式化、lint、治理检查与安全扫描。
