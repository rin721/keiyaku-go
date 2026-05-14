---
doc_role: convention
scope: testing
authority_level: binding
owners: [tech-lead]
status: active
effective_date: 2026-05-15
version: 1.0
related_rules: [GOV-P1-005]
read_when: [test_or_ci, governance_change, async_sensitive]
update_when: [test_policy_changed, automation_changed, ci_changed, adr_accepted]
---

# 测试约定

## 默认分层

- Domain 与 Application 测试应快速、稳定、无外部中间件依赖。
- Repository 与 Infrastructure 的真实中间件行为应通过集成测试验证。
- 集成测试使用 `integration` build tag 与默认快速测试隔离。

## 自动化边界

- `scripts/check-test-conventions.ps1` 检查 testcontainers 测试是否带 `integration` build tag。
- Domain/Application 测试不得 import `testcontainers-go`。
- 覆盖充分性、Mock/Fake 质量、场景完整性仍由评审判断。

## CI 期望

- 默认测试执行快速测试。
- 集成测试应有独立 Make 目标或 CI 步骤。
- 当前仓库无 Go package 时，Go 测试和 lint 应安全跳过。
