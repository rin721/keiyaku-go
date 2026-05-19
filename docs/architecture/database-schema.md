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
version: 2.0
related_rules: [GOV-P0-001, GOV-P1-003]
source_of_truth: [docs/adr/20260515-adopt-gin-gorm-clean-architecture.md, docs/adr/20260519-adopt-remote-service-plugin-system.md, docs/adr/20260519-adopt-plugin-v3-contract.md, docs/adr/20260519-split-blog-plugin-and-iam-service.md]
derived_from: [docs/conventions/migrations.md, docs/architecture/system-design.md, docs/adr/20260519-adopt-plugin-v3-contract.md, docs/adr/20260519-split-blog-plugin-and-iam-service.md]
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
- `articles`、`categories`、`tags`、`article_tags`、`comments`：历史主服务内容表，当前作为 legacy unused 保留；新 Article 业务由 Blog 插件独立持有。
- `casbin_rule`：Casbin v3 策略持久化表。

所有业务表使用 `BIGINT` 业务 ID，不依赖自增 ID 暴露业务量。GORM Model 只存在于 Infrastructure，不能作为 Domain Model 或 HTTP DTO 返回。

## IAM 会话表

IAM refresh session 迁移位于 `migrations/000004_iam_refresh_sessions.up.sql`，覆盖：

- `iam_refresh_sessions`：记录 refresh token JTI、用户、状态、过期时间、轮换关系和吊销时间。

登录和注册会创建 active refresh session；refresh 必须命中 active 且未过期的 session，并在事务内创建新 session、将旧 session 标记为 `rotated`；logout 将当前用户仍 active 的 sessions 标记为 `revoked`。

## 插件注册表

插件注册表 v3 破坏式迁移位于 `migrations/000006_plugin_v3_contract.up.sql`。该迁移会重建旧 `plugin_*` 表，旧插件必须重新注册。当前表结构覆盖：

- `plugin_services`：插件服务身份、协议、当前 manifest hash、`openapi_url`、管理状态和元数据。
- `plugin_instances`：插件运行实例、base URL、版本、心跳时间、租约过期时间和实例状态。
- `plugin_routes`：插件 manifest v3 下的 `route_id`、`gateway_path`、上游路径、鉴权策略和转发选项。
- `plugin_route_claims`：跨插件 route ownership 锁表，注册事务内拒绝 exact/prefix/ANY 重叠。
- `plugin_audit_events`：插件注册、心跳、注销、健康变化、管理操作和网关失败摘要事件。
- `plugin_signature_nonces`：控制面 HMAC nonce 去重表，防止签名重放。

插件注册表只保存主服务路由和实例状态，不保存插件业务数据。插件业务表由插件服务自行维护。

## Blog 插件业务表

Blog 插件迁移位于 `plugins/blog/migrations/000001_blog_schema.up.sql`，覆盖：

- `blog_articles`：Blog 文章主体，使用 Snowflake ID，支持草稿、发布、归档。
- `blog_article_revisions`：文章版本快照，创建文章时写入 v1 revision。
- `blog_categories`：分类预留表，首版允许 `category_id=0`。
- `blog_tags`、`blog_article_tags`：标签标准化与文章标签关系。

Blog 插件使用独立 MySQL DSN，不跨库 join 主服务 `users` 表，只保存网关注入用户上下文中的 `author_id`。

## 插件审计事件

`plugin_audit_events` 是运维审计摘要表，不作为安全不可抵赖日志。`metadata_json` 只允许保存摘要字段，不保存 token、Authorization、Cookie、请求体或完整响应体。主服务插件维护任务会按 `plugins.audit_retention_days` 清理过期审计事件，并清理过期 `plugin_signature_nonces`。

主要索引：

- `idx_plugin_audit_plugin_created(plugin_key, created_at)`：按插件查询最近事件。
- `idx_plugin_audit_action(action)`：按事件类型排查。

## 回滚说明

- 插件 v3 migration 是破坏式迁移，down 只恢复空 v2 表结构，不恢复旧注册数据。
- 关闭健康检查可通过 `plugins.health_check_interval: 0s` 完成。
- 关闭路由缓存可通过 `plugins.route_cache_ttl: 0s` 完成。
- 若需要回滚插件 v3 schema，可执行 `migrations/000006_plugin_v3_contract.down.sql`，随后插件必须按目标版本重新注册。
- 若需要回滚 IAM refresh session schema，可执行 `migrations/000004_iam_refresh_sessions.down.sql`。
