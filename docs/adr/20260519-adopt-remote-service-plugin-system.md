---
state_id: ADR-20260519-PLUGIN-001
doc_role: adr
memory_level: L0
state_scope: global
scope: repo
authority_level: ssot_decision
owners: [tech-lead]
status: accepted
effective_date: 2026-05-19
version: 1.0
related_rules: [GOV-P0-001, GOV-P0-002, GOV-P1-001, GOV-P1-002, GOV-P1-003, GOV-P1-006]
source_of_truth: [docs/adr/20260519-adopt-remote-service-plugin-system.md]
derived_from: [docs/governance/rules.md, docs/conventions/layering.md, docs/conventions/dependency-injection.md, docs/conventions/migrations.md]
read_when: [boundary_sensitive, migration_sensitive, security_sensitive]
update_when: [default_behavior_changed, convention_changed, adr_accepted]
conflict_policy: accepted_adr_defines_remote_plugin_model
rollback_target: [docs/adr/20260515-adopt-gin-gorm-clean-architecture.md, docs/governance/rules.md]
verification_target: [scripts/check-layering.ps1, scripts/check-governance-sync.ps1, scripts/check-governance-map.ps1]
change_reason: introduce remote service plugin extension model
---

# ADR 20260519：采用远端服务插件系统

- 状态：accepted
- 日期：2026-05-19
- 负责人：tech-lead

## 背景与上下文

业务层后续需要可扩展的插件能力。插件需要在独立服务器上运行，并通过当前主服务的注册入口声明可响应的路由。该能力改变了项目默认业务扩展方式，涉及服务注册、路由网关、鉴权、心跳租约、数据库注册表和运维回滚，因此需要 ADR 固化边界。

当前仓库默认 Gin + GORM + Clean Architecture，依赖注入方式为显式手动构造函数。插件系统不得把外部服务模型、HTTP DTO、GORM Model 或运行时容器穿透到 Domain。

## 决策内容

项目采用远端服务插件模型：

- 插件是独立部署的 HTTP 服务，不作为 Go dynamic plugin 加载进主进程。
- 主服务提供插件注册中心、心跳续约、实例状态、路由解析和 HTTP 网关转发。
- 首版只落地 HTTP 插件协议；数据结构保留 `protocol` 字段，但非 HTTP 注册直接拒绝。
- 注册表持久化到 MySQL，区分插件服务、插件实例和插件路由三类状态。
- 插件注册、心跳和注销使用静态 Bearer token；生产环境 token 必须通过配置或环境变量注入。
- 插件 key 默认需要在主服务配置白名单中，插件 `base_url` 默认必须命中 host 或 CIDR 白名单。
- 网关默认不透传原始 `Authorization`，只透传 TraceID、插件 key 和脱敏用户上下文；需要透传时必须由路由声明显式开启。
- 插件侧 SDK 放在 `pkg/plugin`，只承载业务无关的 manifest、注册 client 和心跳 runner。
- 主服务运行时装配继续使用显式手动 DI，不引入运行时反射型 DI 容器。

该决策不建立第三方不可信插件市场，不提供代码沙箱，不支持首版 gRPC、WebSocket、SSE、事件订阅或 mTLS。

后续 [ADR 20260519：插件系统 v2 破坏式契约](20260519-adopt-plugin-v2-breaking-contract.md) 已将首版全局静态注册 token 与派生式 `/extensions/{plugin_key}` 路径升级为 per-plugin HMAC 与显式 `gateway_path`。本 ADR 仍裁决“远端 HTTP 服务插件”模型；v2 契约细节以后续 ADR 为准。

## 后果评估

正面收益：

- 业务扩展可以独立部署、独立迭代，不需要重新编译主服务。
- 主服务仍保留统一入口、鉴权、TraceID、网关错误映射和注册状态管理。
- 插件系统不突破现有分层和依赖注入治理要求。

取舍：

- HTTP 远端调用增加网络延迟和上游不可用风险，需要心跳、超时和网关错误映射兜底。
- 静态注册 token 是首版低复杂度方案，后续高安全环境可升级为 per-plugin secret、HMAC 或 mTLS。
- 同一 `plugin_key` 首版只允许一个 active manifest hash，多版本灰度需要使用不同 plugin key 或后续新增版本路由。

## 备选方案

- Go dynamic plugin：跨平台和部署复杂度高，且不适合 Windows 与独立服务器运行诉求，不采纳。
- 运行时 DI 容器或服务定位器：会隐藏依赖图，与仓库默认依赖注入约定冲突，不采纳。
- 首版直接支持 gRPC、事件订阅和 WebSocket：范围过大，先保留协议扩展点，HTTP 稳定后再补 ADR 或专题约定。
- mTLS 首版：安全性更高，但配置和证书运维成本更高，当前先用 token + 白名单 + 网关签名配置项。

## 后续事项

- [x] 新增插件注册表迁移。
- [x] 新增 `pkg/plugin` SDK、注册中心、HTTP 网关和 `pluginctl`。
- [x] 同步 README、系统设计、HTTP API 和数据库结构文档。
- [x] 运行 Go 测试、OpenAPI 检查、分层检查和治理校验。
