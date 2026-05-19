---
state_id: ADR-20260519-PLUGIN-V2-001
doc_role: adr
memory_level: L0
state_scope: global
scope: repo
authority_level: ssot_decision
owners: [tech-lead]
status: accepted
effective_date: 2026-05-19
version: 1.0
related_rules: [GOV-P0-001, GOV-P0-002, GOV-P0-004, GOV-P1-001, GOV-P1-002, GOV-P1-003, GOV-P1-006]
source_of_truth: [docs/adr/20260519-adopt-plugin-v2-breaking-contract.md]
derived_from: [docs/adr/20260519-adopt-remote-service-plugin-system.md, docs/architecture/plugin-system.md]
read_when: [boundary_sensitive, migration_sensitive, security_sensitive]
update_when: [default_behavior_changed, convention_changed, adr_accepted]
conflict_policy: accepted_adr_defines_plugin_v2_contract
rollback_target: [docs/adr/20260519-adopt-remote-service-plugin-system.md, docs/architecture/plugin-system.md]
verification_target: [scripts/check-layering.ps1, scripts/check-governance-sync.ps1, scripts/check-governance-map.ps1]
change_reason: replace v1 plugin manifest and global token with explicit v2 gateway paths and per-plugin HMAC
---

# ADR 20260519：插件系统 v2 破坏式契约

- 状态：accepted
- 日期：2026-05-19
- 负责人：tech-lead

## 背景

远端 HTTP 插件系统 v1 已具备注册、心跳、健康检查、审计和网关转发能力，但生产收口时暴露出三个默认边界不够明确的问题：外部网关路径由 `plugin_key` 派生，注册身份依赖全局静态 token，网关到插件的签名也使用全局配置。继续用兼容补丁会让插件契约长期背负旧字段和旧路径模型。

## 决策

插件系统直接升级为 v2 破坏式契约：

- manifest 固定 `schema_version: "v2"`。
- route 使用 `route_id`、`gateway_path`、`upstream_path` 和字符串 `timeout`；删除 v1 的 `path` 与 `timeout_ms`。
- 插件显式声明完整 `gateway_path`，主服务只接受位于 `plugins.public_prefix` 下的路径。
- 废弃 `plugins.registration_tokens`、`plugins.allowed_plugin_keys` 和全局 `plugins.gateway_signing_secret`。
- 主服务使用 `plugins.trusted_plugins.<plugin_key>` 配置每个插件的 `registration_secret`、`gateway_secret`、host/CIDR 白名单和 loopback 许可。
- 控制面注册、心跳、注销使用 per-plugin registration secret HMAC；网关转发使用 per-plugin gateway secret HMAC。
- HMAC canonical string 固定为 `method + "\n" + path + "\n" + timestamp + "\n" + nonce + "\n" + sha256(body)`。
- 使用 `plugin_signature_nonces` 表阻止控制面签名重放。
- 通过破坏式 migration 重建 `plugin_*` 注册表；旧插件必须重新注册。

## 后果

正面收益：

- 插件外部路径成为显式契约，便于审计、OpenAPI 描述和跨插件冲突排查。
- 注册身份与网关身份从全局 secret 收敛到 per-plugin secret。
- SDK、CLI、Blog 示例和主服务共享同一个签名 canonical。

取舍：

- v1 插件 manifest 不再可用，部署时必须重新生成 v2 manifest。
- 旧插件注册表数据会被清空，插件进程需要重新注册。
- 本轮仍不引入 mTLS、gRPC、WebSocket、事件订阅、版本灰度或插件市场能力。
