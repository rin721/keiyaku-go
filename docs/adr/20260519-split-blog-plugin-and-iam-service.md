---
state_id: ADR-20260519-BLOG-IAM-001
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
source_of_truth: [docs/adr/20260519-split-blog-plugin-and-iam-service.md]
derived_from: [docs/adr/20260519-adopt-remote-service-plugin-system.md, docs/architecture/system-design.md]
read_when: [boundary_sensitive, migration_sensitive, security_sensitive]
update_when: [default_behavior_changed, convention_changed, adr_accepted]
conflict_policy: accepted_adr_defines_blog_and_iam_service_boundary
rollback_target: [docs/adr/20260519-adopt-remote-service-plugin-system.md, docs/architecture/system-design.md]
verification_target: [scripts/check-layering.ps1, scripts/check-governance-sync.ps1, scripts/check-governance-map.ps1]
change_reason: move article business into blog plugin and split identity into IAM service
---

# ADR 20260519：拆分 Blog 插件与 IAM 服务

- 状态：accepted
- 日期：2026-05-19
- 负责人：tech-lead

## 背景

Article 创建、列表与详情属于内容业务能力，不应继续固定在主服务内。插件系统已经提供远端 HTTP 插件注册、心跳、健康检查、审计和网关转发能力，因此 Article 可以作为 Blog 业务插件的首批用例迁出。

用户、认证、JWT 和 RBAC 是系统根信任能力。它们需要被主服务网关、管理接口和后续业务插件共同依赖，但不应作为普通业务插件注册，否则会削弱网关鉴权与插件控制面的信任边界。

## 决策

- Article 创建、列表与详情迁移到 `plugins/blog` 独立 Blog 插件服务。
- Blog 插件使用独立数据库连接、独立迁移和独立配置，不读写主服务 `articles` 或 `users` 表。
- 主服务不再注册 `/api/v1/articles`，Blog 业务统一通过 `/api/v1/extensions/blog/articles...` 访问。
- 主服务继续承担插件控制面、网关、审计、健康检查、路由缓存和平台运行能力。
- 用户、认证、JWT、RBAC 拆为 `cmd/iam` 独立服务。
- 主服务通过 IAM internal API 完成 access token introspection 和 authorize，不再直接解析 JWT 或加载 Casbin 策略。
- IAM 服务不经过插件注册系统，不作为普通业务插件。

## 后果

正面收益：

- Blog 业务可以独立部署、独立迁移、独立演进。
- 主服务边界收敛为平台控制面，减少业务耦合。
- IAM 成为明确的根信任服务，可被主服务和后续服务复用。

取舍：

- 文章访问路径从 `/api/v1/articles` 改为 `/api/v1/extensions/blog/articles...`，本阶段不提供兼容代理。
- 主服务调用 IAM 会增加一次内部网络依赖，因此新增 `/readyz` 暴露 IAM 可用性。
- 旧主服务内容表本阶段保留为 legacy unused，不做历史数据回填。

## 后续事项

- [x] 新增 `plugins/blog` 独立插件服务与迁移。
- [x] 主服务移除内置 Article 路由。
- [x] 新增 `cmd/iam` 独立 IAM 服务与主服务 IAM client。
- [x] 新增 IAM refresh token 持久会话、轮换与登出吊销。
- [ ] 后续补强 IAM 审计事件和 access token 吊销列表。
- [ ] 后续扩展 Blog 分类、标签、草稿、发布流、搜索和评论能力。
