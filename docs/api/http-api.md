---
state_id: API-HTTP-001
doc_role: convention
memory_level: L1
state_scope: module
scope: repo
authority_level: binding
owners: [tech-lead]
status: active
effective_date: 2026-05-15
version: 1.0
related_rules: [GOV-P0-001, GOV-P0-002, GOV-P1-001]
source_of_truth: [docs/adr/20260515-adopt-gin-gorm-clean-architecture.md]
derived_from: [docs/architecture/system-design.md]
read_when: [boundary_sensitive, security_sensitive]
update_when: [default_behavior_changed, convention_changed, adr_accepted]
conflict_policy: binding_must_yield_to_ssot
rollback_target: [docs/architecture/system-design.md]
verification_target: [scripts/check-layering.ps1, scripts/check-governance-map.ps1]
---

# HTTP API 契约

所有接口使用 `/api/v1` 前缀，响应体统一为：

```json
{"code":0,"msg":"ok","data":{}}
```

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
| GET | `/api/v1/articles/:id` | No | 获取已发布文章详情 |
| GET | `/api/v1/articles` | No | 分页获取已发布文章 |

## Reserved CMS/RBAC

分类、标签、评论和后台 RBAC 在首版保留数据结构与路由扩展点，后续按模块补齐 Handler 与 Usecase。
