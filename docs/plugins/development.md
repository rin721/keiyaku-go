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
version: 3.0
related_rules: [GOV-P0-001, GOV-P0-002, GOV-P0-004, GOV-P1-001, GOV-P1-002, GOV-P1-006]
source_of_truth: [docs/adr/20260519-adopt-plugin-v3-contract.md, docs/architecture/plugin-system.md]
derived_from: [pkg/plugin/README.md, docs/api/http-api.md, docs/adr/20260519-adopt-plugin-v3-contract.md]
read_when: [boundary_sensitive, security_sensitive, pkg_change]
update_when: [default_behavior_changed, convention_changed, adr_accepted, security_policy_changed]
conflict_policy: binding_must_yield_to_plugin_architecture
rollback_target: [docs/architecture/plugin-system.md, docs/adr/20260519-adopt-plugin-v3-contract.md]
verification_target: [scripts/check-layering.ps1, scripts/check-governance-map.ps1, scripts/check-governance-sync.ps1]
---

# 插件开发文档

Keiyaku-Go 插件系统 v3 只支持远端 HTTP 插件。插件是独立部署的服务，通过 per-plugin HMAC 注册，显式声明完整 `gateway_path`，并受 `trusted_plugins.<plugin_key>` 的路径、方法、鉴权和出站网络策略约束。v2 manifest、旧签名 canonical 和旧注册表数据不再兼容。

## 主服务配置

```yaml
plugins:
  enabled: true
  public_prefix: "/api/v1/extensions"
  max_registration_body_bytes: 1048576
  max_gateway_body_bytes: 10485760
  max_route_timeout: 5s
  trusted_plugins:
    demo-plugin:
      registration_secret: "change-me-demo-registration-secret-32b"
      gateway_secret: "change-me-demo-gateway-secret-32bytes"
      allowed_hosts: []
      allowed_cidrs:
        - "127.0.0.1/32"
      allowed_gateway_prefixes:
        - "/api/v1/extensions/demo-plugin"
      allowed_auth_policies:
        - "authenticated"
        - "rbac"
      allowed_methods:
        - "GET"
        - "POST"
      allow_loopback: true
      allow_insecure_http: true
```

生产环境每个 secret 至少 32 字节。`gateway_path` 必须位于 `public_prefix` 下，不能覆盖主服务内置路径。

## Manifest v3

```json
{
  "schema_version": "v3",
  "plugin_key": "demo-plugin",
  "name": "DemoPlugin",
  "version": "0.1.0",
  "instance_id": "demo-plugin-local",
  "protocol": "http",
  "base_url": "http://127.0.0.1:9090",
  "health_path": "/healthz",
  "routes": [
    {
      "route_id": "hello",
      "method": "GET",
      "match_type": "exact",
      "gateway_path": "/api/v1/extensions/demo-plugin/hello",
      "upstream_path": "/hello",
      "auth_policy": "authenticated",
      "timeout": "5s",
      "forward_auth_header": false
    }
  ]
}
```

Route 必填字段：`route_id`、`method`、`match_type`、`gateway_path`、`upstream_path`、`auth_policy`、`timeout`。v3 会拒绝 v1/v2 manifest；跨插件 exact/prefix/ANY 重叠路由会在注册事务内通过 `plugin_route_claims` 拒绝。

## HMAC 签名

注册、心跳和注销请求由 SDK 使用 registration secret 签名。请求必须包含：

- `X-Keiyaku-Plugin-Key`
- `X-Keiyaku-Timestamp`
- `X-Keiyaku-Nonce`
- `X-Keiyaku-Signature`

canonical string 为：

```text
METHOD + "\n" + PATH + "\n" + RAW_QUERY + "\n" + TIMESTAMP + "\n" + NONCE + "\n" + SHA256(body)
```

`RAW_QUERY` 是不包含 `?` 的原始查询串。主服务使用 `plugin_signature_nonces` 阻止控制面签名重放。网关转发到插件时使用 gateway secret 生成同类签名，插件服务应使用 `pkg/plugin.Verify` 校验。

推荐插件服务使用 `pkg/plugin.VerifySignedRequest` 读取并校验网关请求；它会在校验后恢复 `req.Body`，便于业务 Handler 继续读取。v3 插件应设置 `ExpectedPluginKey`，并传入 nonce store，例如本地进程内 `pkg/plugin.NewMemoryNonceStore()` 或可共享的持久化实现。对外业务接口应设置请求体上限，和主服务的 `plugins.max_gateway_body_bytes` 保持一致或更小。

## CLI

生成样板：

```powershell
go run ./cmd/pluginctl init `
  --dir ./plugin-demo `
  --module example.com/plugin-demo `
  --plugin-key demo-plugin `
  --name DemoPlugin
```

生成的样板服务启动时必须配置 `KEIYAKU_PLUGIN_REGISTRATION_SECRET` 与 `KEIYAKU_PLUGIN_GATEWAY_SECRET`；业务路由会校验 gateway HMAC，并使用 `LifecycleRunner` 自动注册、心跳、失败退避、必要时重注册和退出注销。

注册：

```powershell
$env:KEIYAKU_PLUGIN_REGISTRATION_SECRET="change-me-demo-registration-secret-32b"
go run ./cmd/pluginctl register --manifest ./plugin-demo/manifest.json --host http://127.0.0.1:8080
```

心跳与注销：

```powershell
go run ./cmd/pluginctl heartbeat --plugin-key demo-plugin --instance-id demo-plugin-local --host http://127.0.0.1:8080
go run ./cmd/pluginctl unregister --plugin-key demo-plugin --instance-id demo-plugin-local --host http://127.0.0.1:8080
```

## Blog 插件

内置 Blog 插件使用 v3 manifest，完整路径为：

```text
POST /api/v1/extensions/blog/articles
GET  /api/v1/extensions/blog/articles
GET  /api/v1/extensions/blog/articles/{id}
```

启动示例：

```powershell
$env:BLOG_MYSQL_DSN="blog:blog@tcp(127.0.0.1:3306)/blog?charset=utf8mb4&parseTime=True&loc=UTC"
$env:KEIYAKU_HOST="http://127.0.0.1:8080"
$env:BLOG_REGISTRATION_SECRET="change-me-blog-registration-secret-32b"
$env:BLOG_GATEWAY_SECRET="change-me-blog-gateway-secret-32bytes"
go run ./plugins/blog/cmd/blog
```

Blog 插件不读写主服务 `users` 或 `articles` 表。业务请求必须通过主服务网关进入，插件会校验 gateway HMAC，并读取 `X-Keiyaku-User-ID`、`X-Keiyaku-Username`、`X-Keiyaku-User-Roles`。

## 发布检查清单

- [ ] 插件已配置到 `plugins.trusted_plugins`。
- [ ] `registration_secret` 与 `gateway_secret` 已通过环境或密钥系统注入。
- [ ] `gateway_path` 全部位于 `plugins.public_prefix` 下。
- [ ] 插件 `base_url` 命中该插件的 host 或 CIDR 白名单。
- [ ] 插件实现 `health_path`，且健康检查不泄露敏感数据。
- [ ] 插件业务请求强制校验 gateway HMAC。
- [ ] 插件业务请求使用受限 body 读取，避免无界 `ReadAll`。
- [ ] 插件日志不打印 token、secret、Authorization、Cookie、请求体或完整用户对象。
