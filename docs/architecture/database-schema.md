---
state_id: ARCH-DB-001
doc_role: convention
memory_level: L1
state_scope: module
scope: migrations
authority_level: binding
owners: [tech-lead]
status: active
effective_date: 2026-05-19
version: 1.1
related_rules: [GOV-P0-001, GOV-P1-003]
source_of_truth: [docs/adr/20260515-adopt-gin-gorm-clean-architecture.md, docs/adr/20260519-adopt-remote-service-plugin-system.md]
derived_from: [docs/conventions/migrations.md, docs/architecture/system-design.md, docs/adr/20260519-adopt-remote-service-plugin-system.md]
read_when: [migration_sensitive, boundary_sensitive]
update_when: [migration_policy_changed, default_behavior_changed, adr_accepted]
conflict_policy: binding_must_yield_to_ssot
rollback_target: [docs/adr/20260515-adopt-gin-gorm-clean-architecture.md]
verification_target: [scripts/check-governance-map.ps1]
---

# 核心数据表结构

首版迁移位于 `migrations/000001_init_cms_schema.up.sql`，覆盖：

- `users`：用户账号、密码哈希、状态与展示信息。
- `roles`、`permissions`、`user_roles`、`role_permissions`：后台 RBAC 扩展表。
- `articles`：文章主体，使用 Snowflake ID，支持草稿、发布、归档。
- `categories`、`tags`、`article_tags`：内容分类与标签。
- `comments`：文章评论，预留审核状态。
- `casbin_rule`：Casbin v3 策略持久化表。

所有业务表使用 `BIGINT` 业务 ID，不依赖自增 ID 暴露业务量。GORM Model 只存在于 Infrastructure，不能作为 Domain Model 或 HTTP DTO 返回。

插件注册表迁移位于 `migrations/000002_plugin_registry.up.sql`，覆盖：

- `plugin_services`：插件服务身份、协议、当前 manifest hash、管理状态和元数据。
- `plugin_instances`：插件运行实例、base URL、版本、心跳时间、租约过期时间和实例状态。
- `plugin_routes`：插件 manifest 下的 HTTP 路由、匹配方式、上游路径、鉴权策略和转发选项。

插件注册表只保存主服务路由和实例状态，不保存插件业务数据。插件业务表由插件服务自行维护。
