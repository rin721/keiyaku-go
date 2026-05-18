---
state_id: PLUGIN-DEV-001
doc_role: convention
memory_level: L1
state_scope: module
scope: repo
authority_level: binding
owners: [tech-lead]
status: active
effective_date: 2026-05-19
version: 1.0
related_rules: [GOV-P0-001, GOV-P0-002, GOV-P0-004, GOV-P1-001, GOV-P1-002, GOV-P1-006]
source_of_truth: [docs/adr/20260519-adopt-remote-service-plugin-system.md, docs/architecture/plugin-system.md]
derived_from: [pkg/plugin/README.md, docs/api/http-api.md]
read_when: [boundary_sensitive, security_sensitive, pkg_change]
update_when: [default_behavior_changed, convention_changed, adr_accepted, security_policy_changed]
conflict_policy: binding_must_yield_to_plugin_architecture
rollback_target: [docs/architecture/plugin-system.md, docs/adr/20260519-adopt-remote-service-plugin-system.md]
verification_target: [scripts/check-layering.ps1, scripts/check-governance-map.ps1, scripts/check-governance-sync.ps1]
---

# 插件开发文档

本文档面向插件开发者，说明如何创建、注册、调试和发布一个远端 HTTP 插件。插件系统设计见 [远端插件系统设计](../architecture/plugin-system.md)。

## 开发前准备

主服务必须完成：

- 执行 `migrations/000002_plugin_registry.up.sql`。
- 配置 `plugins.registration_tokens`。
- 配置 `plugins.allowed_plugin_keys`。
- 配置 `plugins.allowed_hosts` 或 `plugins.allowed_cidrs`。
- 本地调试如需 `127.0.0.1`，显式设置 `plugins.allow_loopback: true`。

示例本地配置：

```yaml
plugins:
  enabled: true
  registration_tokens:
    - "change-me-32-bytes-minimum-token"
  allowed_plugin_keys:
    - demo-plugin
  allowed_cidrs:
    - "127.0.0.1/32"
  allow_loopback: true
  allow_public_routes: false
```

## 创建插件样板

使用 `pluginctl init` 生成独立插件服务：

```powershell
go run ./cmd/pluginctl init `
  --dir ./plugin-demo `
  --module example.com/plugin-demo `
  --plugin-key demo-plugin `
  --name DemoPlugin
```

生成内容：

- `go.mod`：独立 Go module，带本地 `replace` 指向当前仓库。
- `main.go`：最小 HTTP 插件服务、注册逻辑、心跳 runner。
- `manifest.json`：插件声明。

## Manifest 格式

```json
{
  "schema_version": "v1",
  "plugin_key": "demo-plugin",
  "name": "DemoPlugin",
  "version": "0.1.0",
  "instance_id": "demo-plugin-local",
  "protocol": "http",
  "base_url": "http://127.0.0.1:9090",
  "health_path": "/healthz",
  "routes": [
    {
      "method": "GET",
      "match_type": "exact",
      "path": "/hello",
      "upstream_path": "/hello",
      "auth_policy": "authenticated",
      "timeout_ms": 5000,
      "forward_auth_header": false
    }
  ]
}
```

字段说明：

| 字段 | 要求 |
| --- | --- |
| `schema_version` | 首版固定为 `v1` |
| `plugin_key` | 小写字母开头，只允许小写字母、数字、短横线 |
| `instance_id` | 同一插件下唯一，建议包含部署实例标识 |
| `protocol` | 首版固定为 `http` |
| `base_url` | 插件服务可被主服务访问的根地址 |
| `health_path` | 插件健康检查路径 |
| `routes` | 至少声明一个路由 |

## 路由声明

| 字段 | 说明 |
| --- | --- |
| `method` | `GET`、`POST`、`PUT`、`PATCH`、`DELETE` 或 `ANY` |
| `match_type` | `exact` 或 `prefix` |
| `path` | 主服务网关下的路径，例如 `/hello` |
| `upstream_path` | 插件服务实际路径，例如 `/hello` |
| `auth_policy` | `inherit`、`authenticated`、`rbac`、`admin`、`public` |
| `timeout_ms` | 单次上游请求超时 |
| `forward_auth_header` | 是否透传用户原始 Authorization |

访问路径由主服务拼接：

```text
/api/v1/extensions/{plugin_key}{path}
```

示例 manifest 中 `/hello` 的访问地址是：

```text
/api/v1/extensions/demo-plugin/hello
```

## 使用 SDK 注册

插件服务可直接使用 `pkg/plugin`：

```go
manifest := plugin.Manifest{
	SchemaVersion: plugin.DefaultSchemaVersion,
	PluginKey:     "demo-plugin",
	Name:          "DemoPlugin",
	Version:       "0.1.0",
	InstanceID:    "demo-plugin-local",
	Protocol:      plugin.ProtocolHTTP,
	BaseURL:       "http://127.0.0.1:9090",
	HealthPath:    "/healthz",
	Routes: []plugin.Route{
		{
			Method:       plugin.MethodGet,
			MatchType:    plugin.MatchTypeExact,
			Path:         "/hello",
			UpstreamPath: "/hello",
			AuthPolicy:   plugin.AuthPolicyAuthenticated,
			TimeoutMS:    5000,
		},
	},
}

client := plugin.NewClient("http://127.0.0.1:8080", os.Getenv("KEIYAKU_PLUGIN_TOKEN"))
result, err := client.Register(context.Background(), manifest)
```

心跳：

```go
runner := plugin.HeartbeatRunner{
	Client:     client,
	PluginKey:  manifest.PluginKey,
	InstanceID: manifest.InstanceID,
	Interval:   10 * time.Second,
}
_ = runner.Run(ctx)
```

注销：

```go
_ = client.Unregister(ctx, manifest.PluginKey, manifest.InstanceID)
```

## 使用命令注册

校验 manifest：

```powershell
go run ./cmd/pluginctl manifest validate --manifest ./plugin-demo/manifest.json
```

注册：

```powershell
$env:KEIYAKU_PLUGIN_TOKEN="change-me-32-bytes-minimum-token"
go run ./cmd/pluginctl register --manifest ./plugin-demo/manifest.json --host http://127.0.0.1:8080
```

发送一次心跳：

```powershell
go run ./cmd/pluginctl heartbeat --plugin-key demo-plugin --instance-id demo-plugin-local --host http://127.0.0.1:8080
```

注销实例：

```powershell
go run ./cmd/pluginctl unregister --plugin-key demo-plugin --instance-id demo-plugin-local --host http://127.0.0.1:8080
```

## 插件服务实现要求

- 必须提供 `health_path`。
- 必须按 manifest 声明实现每个 upstream route。
- 必须把业务响应作为普通 HTTP 响应返回，不需要包装成主服务 `{code,msg,data}`。
- 不要依赖主服务 `internal` 包。
- 不要在日志里打印 token、Authorization、Cookie、完整请求体或用户敏感信息。
- 需要用户身份时读取 `X-Keiyaku-User-ID`、`X-Keiyaku-Username`、`X-Keiyaku-User-Roles`。
- 需要链路追踪时读取 `X-Trace-ID`。

## 网关上下文 Header

主服务默认传递：

| Header | 含义 |
| --- | --- |
| `X-Trace-ID` | 当前请求 TraceID |
| `X-Keiyaku-Plugin-Key` | 当前插件 key |
| `X-Keiyaku-User-ID` | 认证用户 ID |
| `X-Keiyaku-Username` | 认证用户名 |
| `X-Keiyaku-User-Roles` | 认证用户角色，逗号分隔 |
| `X-Forwarded-Host` | 原始 Host |
| `X-Forwarded-Proto` | 原始协议 |
| `X-Forwarded-Method` | 原始 HTTP method |

默认不会传递 `Authorization` 和 `Cookie`。

## 本地调试流程

1. 修改主服务配置，允许本地插件 key、token 和 loopback。
2. 启动 MySQL、Redis，并执行 migration。
3. 启动主服务。
4. 在另一个终端启动插件服务。
5. 插件启动时自动注册，或使用 `pluginctl register` 手动注册。
6. 访问 `/api/v1/extensions/{plugin_key}/...` 验证转发。

示例：

```powershell
Invoke-RestMethod `
  -Headers @{ Authorization = "Bearer <user-access-token>" } `
  http://127.0.0.1:8080/api/v1/extensions/demo-plugin/hello
```

## 发布检查清单

- [ ] `plugin_key` 已加入主服务白名单。
- [ ] 插件 base URL 已加入 `allowed_hosts` 或 `allowed_cidrs`。
- [ ] 生产 token 不写入代码和仓库文档。
- [ ] 插件服务有健康检查。
- [ ] 插件服务能持续心跳。
- [ ] route 的 `auth_policy` 与业务风险匹配。
- [ ] 不需要透传原始 Authorization 时保持 `forward_auth_header=false`。
- [ ] 插件日志不打印敏感 header、token、请求体或完整用户对象。
- [ ] 插件服务能处理主服务超时和重试缺失的情况。

## 常见问题

### 注册返回 401

检查 `Authorization: Bearer <token>` 是否传递，token 是否存在于主服务 `plugins.registration_tokens`。

### 注册返回 403

检查 `plugin_key` 是否存在于 `plugins.allowed_plugin_keys`。

### 注册返回 400

检查 manifest 格式、`base_url`、路由 path、method、match type 和 auth policy。

### 网关返回 503

插件实例不存在、已禁用、租约过期或 manifest hash 不匹配。确认插件仍在发送心跳。

### 插件收不到用户 token

这是默认行为。需要透传时在 route 中设置 `forward_auth_header=true`，并评估安全风险。
