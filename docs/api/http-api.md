---
state_id: API-HTTP-001
doc_role: convention
memory_level: L1
state_scope: module
scope: repo
authority_level: binding
owners: [tech-lead]
status: active
effective_date: 2026-05-19
version: 3.0
related_rules: [GOV-P0-001, GOV-P0-002, GOV-P1-001]
source_of_truth: [docs/adr/20260515-adopt-gin-gorm-clean-architecture.md, docs/adr/20260519-adopt-remote-service-plugin-system.md, docs/adr/20260519-adopt-plugin-v3-contract.md, docs/adr/20260519-split-blog-plugin-and-iam-service.md]
derived_from: [docs/architecture/system-design.md, docs/adr/20260519-adopt-plugin-v3-contract.md, docs/adr/20260519-split-blog-plugin-and-iam-service.md]
read_when: [boundary_sensitive, security_sensitive]
update_when: [default_behavior_changed, convention_changed, adr_accepted]
conflict_policy: binding_must_yield_to_ssot
rollback_target: [docs/architecture/system-design.md]
verification_target: [scripts/check-layering.ps1, scripts/check-governance-map.ps1]
---

# HTTP API 契约

业务接口使用 `/api/v1` 前缀，主服务自有错误响应体统一为：

```json
{"code":0,"msg":"ok","data":{}}
```

插件网关代理到上游插件后的业务响应原样透传；只有主服务自身发现路由不存在、插件不可用、上游连接失败或超时时，才返回统一响应结构。

## IAM Service

| Method | Path | Auth | Description |
| --- | --- | --- | --- |
| POST | `/api/v1/auth/register` | No | IAM 服务注册用户并返回 Token |
| POST | `/api/v1/auth/login` | No | IAM 服务登录并返回 Token |
| POST | `/api/v1/auth/refresh` | Refresh token | IAM 服务刷新 Token |
| POST | `/api/v1/auth/logout` | JWT | IAM 服务登出 |
| GET | `/api/v1/users/me` | JWT | IAM 服务获取当前用户资料 |
| POST | `/internal/v1/tokens/introspect` | Service token | 主服务内部校验 access token |
| POST | `/internal/v1/authorize` | Service token | 主服务内部授权决策 |

IAM 是独立服务，不经过插件注册系统。主服务通过 IAM client 调用 internal API 完成 token 校验与授权决策。

`/api/v1/auth/login` 与 `/api/v1/auth/register` 会创建持久化 refresh session；`/api/v1/auth/refresh` 必须提交 refresh token，并在 MySQL 中原子轮换旧 session；`/api/v1/auth/logout` 会吊销当前用户仍处于 active 状态的 refresh sessions。access token 仍保持短期无状态 JWT，由 IAM introspection 检查签名、过期时间和用户启用状态。

## Blog Plugin

Article 创建、列表与详情已迁移到 `blog` 插件。主服务不再注册 `/api/v1/articles`。

| Method | Path | Auth | Description |
| --- | --- | --- | --- |
| POST | `/api/v1/extensions/blog/articles` | Route RBAC | 通过 Blog 插件创建文章 |
| GET | `/api/v1/extensions/blog/articles` | Route RBAC | 通过 Blog 插件分页获取已发布文章 |
| GET | `/api/v1/extensions/blog/articles/{id}` | Route RBAC | 通过 Blog 插件获取已发布文章详情 |

## Plugin

插件注册、心跳和注销接口使用 per-plugin HMAC，不使用用户 JWT。插件管理接口默认由 IAM token + Casbin 授权控制。插件网关根据已注册 route 的 `auth_policy` 决定是否需要用户身份。

| Method | Path | Auth | Description |
| --- | --- | --- | --- |
| POST | `/api/v1/plugins/registrations` | Plugin HMAC | 注册远端插件实例、manifest v3 与路由 |
| POST | `/api/v1/plugins/{plugin_key}/instances/{instance_id}/heartbeat` | Plugin HMAC | 刷新插件实例租约 |
| DELETE | `/api/v1/plugins/{plugin_key}/instances/{instance_id}` | Plugin HMAC | 注销插件实例 |
| GET | `/api/v1/plugins` | JWT + Casbin | 查询插件服务列表 |
| GET | `/api/v1/plugins/{plugin_key}` | JWT + Casbin | 查询插件服务、实例和路由详情 |
| GET | `/api/v1/plugins/{plugin_key}/instances` | JWT + Casbin | 查询插件实例健康、租约和管理状态 |
| POST | `/api/v1/plugins/{plugin_key}/disable` | JWT + Casbin | 禁用插件服务 |
| POST | `/api/v1/plugins/{plugin_key}/enable` | JWT + Casbin | 启用插件服务 |
| POST | `/api/v1/plugins/{plugin_key}/instances/{instance_id}/disable` | JWT + Casbin | 禁用插件实例 |
| POST | `/api/v1/plugins/{plugin_key}/instances/{instance_id}/enable` | JWT + Casbin | 启用插件实例 |
| GET | `/api/v1/plugins/{plugin_key}/diagnostics` | JWT + Casbin | 查询路由匹配和实例可路由诊断，支持 `method` 与 `path` 查询参数 |
| GET | `/api/v1/plugins/{plugin_key}/audit-events` | JWT + Casbin | 查询插件审计事件，支持 `limit` |
| ANY | `plugins.public_prefix/*path` | Route policy | 按 manifest v3 的完整 `gateway_path` 代理 HTTP 请求 |

插件实例响应包含 `health_status`、`last_checked_at`、`consecutive_failures` 和 `last_error_at`。`health_status=unknown` 或 `healthy` 可路由，`unhealthy` 不可路由。

Plugin HMAC 请求必须包含 `X-Keiyaku-Plugin-Key`、`X-Keiyaku-Timestamp`、`X-Keiyaku-Nonce` 和 `X-Keiyaku-Signature`。v3 签名 canonical 为 `method + "\n" + path + "\n" + raw_query + "\n" + timestamp + "\n" + nonce + "\n" + sha256(body)`。

注册请求体受 `plugins.max_registration_body_bytes` 限制，插件网关业务请求体受 `plugins.max_gateway_body_bytes` 限制。网关会丢弃客户端传入的 `X-Keiyaku-*`、`X-Forwarded-*`、`Forwarded`、`X-Real-IP` 和默认 `Authorization`/`Cookie`，再由主服务重建可信上下文；响应默认不回传 `Set-Cookie`。

## OpenAPI Generation

`api/openapi.yaml` 是生成物，不手动编辑。HTTP API 契约由 handler 上的 OpenAPI 注释和 `internal/api/http/dto` 中的 DTO 结构生成：

```powershell
go run ./cmd/openapi generate
go run ./cmd/openapi generate --check
```

## API Docs

Swagger API 文档由 HTTP 路由构造器自动注入：

| Method | Path | Auth | Description |
| --- | --- | --- | --- |
| GET | `/docs` | No | Swagger UI |
| GET | `/docs/openapi.yaml` | No | OpenAPI 3.0 契约 |

## Reserved Blog/RBAC

分类、标签、评论和后台 RBAC 在 Blog 插件中保留数据结构与路由扩展点，后续按模块补齐 Handler 与 Usecase。
