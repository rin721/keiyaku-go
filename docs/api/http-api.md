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
version: 1.3
related_rules: [GOV-P0-001, GOV-P0-002, GOV-P1-001]
source_of_truth: [docs/adr/20260515-adopt-gin-gorm-clean-architecture.md, docs/adr/20260519-adopt-remote-service-plugin-system.md]
derived_from: [docs/architecture/system-design.md, docs/adr/20260519-adopt-remote-service-plugin-system.md]
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

## Auth

| Method | Path | Auth | Description |
| --- | --- | --- | --- |
| POST | `/api/v1/auth/register` | No | 注册用户并返回 Token |
| POST | `/api/v1/auth/login` | No | 登录并返回 Token |

## User

| Method | Path | Auth | Description |
| --- | --- | --- | --- |
| GET | `/api/v1/users/me` | JWT | 获取当前用户资料 |

## Article

| Method | Path | Auth | Description |
| --- | --- | --- | --- |
| POST | `/api/v1/articles` | JWT | 创建文章，可选择立即发布 |
| GET | `/api/v1/articles/{id}` | No | 获取已发布文章详情 |
| GET | `/api/v1/articles` | No | 分页获取已发布文章 |

## Plugin

插件注册接口使用插件注册 token，不使用用户 JWT。插件管理接口默认由现有 JWT + Casbin 控制。插件网关根据已注册 route 的 `auth_policy` 决定是否需要用户身份。

| Method | Path | Auth | Description |
| --- | --- | --- | --- |
| POST | `/api/v1/plugins/registrations` | Plugin token | 注册远端插件实例、manifest 与路由 |
| POST | `/api/v1/plugins/{plugin_key}/instances/{instance_id}/heartbeat` | Plugin token | 刷新插件实例租约 |
| DELETE | `/api/v1/plugins/{plugin_key}/instances/{instance_id}` | Plugin token | 注销插件实例 |
| GET | `/api/v1/plugins` | JWT + Casbin | 查询插件服务列表 |
| GET | `/api/v1/plugins/{plugin_key}` | JWT + Casbin | 查询插件服务、实例和路由详情 |
| GET | `/api/v1/plugins/{plugin_key}/instances` | JWT + Casbin | 查询插件实例健康、租约和管理状态 |
| POST | `/api/v1/plugins/{plugin_key}/disable` | JWT + Casbin | 禁用插件服务 |
| POST | `/api/v1/plugins/{plugin_key}/enable` | JWT + Casbin | 启用插件服务 |
| POST | `/api/v1/plugins/{plugin_key}/instances/{instance_id}/disable` | JWT + Casbin | 禁用插件实例 |
| GET | `/api/v1/plugins/{plugin_key}/audit-events` | JWT + Casbin | 查询插件审计事件，支持 `limit` |
| ANY | `/api/v1/extensions/{plugin_key}/*path` | Route policy | 按插件注册路由代理 HTTP 请求 |

插件实例响应包含 `health_status`、`last_checked_at`、`consecutive_failures` 和 `last_error_at`。`health_status=unknown` 或 `healthy` 可路由，`unhealthy` 不可路由。

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

## Reserved CMS/RBAC

分类、标签、评论和后台 RBAC 在首版保留数据结构与路由扩展点，后续按模块补齐 Handler 与 Usecase。
