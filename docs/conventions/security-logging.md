---
doc_role: convention
scope: security
authority_level: binding
owners: [tech-lead]
status: active
effective_date: 2026-05-15
version: 1.0
related_rules: [GOV-P0-003, GOV-P0-004, GOV-P1-001]
read_when: [security_sensitive, governance_change, test_or_ci]
update_when: [security_policy_changed, automation_changed, default_rule_changed]
---

# 安全与日志约定

## 密码与哈希

- 密码存储必须使用 Argon2id 或 bcrypt。
- 密码哈希参数必须版本化。
- 快速哈希只允许用于非密码安全模型，例如 HMAC、ETag 或文件完整性。

## 敏感日志

- 默认禁止打印完整请求体、配置对象、用户对象或第三方凭据对象。
- 禁止使用 `zap.Any` 或 `%+v` 输出未脱敏结构。
- 敏感信息只能通过明确白名单结构输出。

## 可观测性

- 结构化日志默认携带 TraceID。
- 异步任务、消息和定时任务默认携带 TraceID 或 CorrelationID。
