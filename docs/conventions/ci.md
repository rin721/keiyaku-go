---
state_id: CONV-CI-001
doc_role: convention
memory_level: L1
state_scope: module
scope: ci
authority_level: binding
owners: [tech-lead]
status: active
effective_date: 2026-05-15
version: 2.0
related_rules: []
source_of_truth: [docs/governance/automation-matrix.md, docs/governance/change-management.md]
derived_from: [docs/governance/automation-matrix.md, docs/governance/change-management.md]
read_when: [test_or_ci, governance_change]
update_when: [ci_changed, automation_changed, default_rule_changed]
conflict_policy: binding_must_yield_to_ssot
rollback_target: [docs/governance/automation-matrix.md, docs/governance/change-management.md]
verification_target: [scripts/check-governance.ps1, scripts/check-go-package-state.ps1, scripts/check-governance-map.ps1]
---

# CI 约定

CI 负责执行可机器化的治理规则，不承载需要人工判断的设计评审。

## 必备检查

- 治理文档存在性检查。
- 元数据完整性检查。
- rule_id 链接检查。
- 治理债务和 break-glass 到期检查。
- 分层依赖方向检查。
- 测试约定检查。
- Secret 扫描。
- Go package 存在时执行 `go test ./...` 和 `golangci-lint`。

## Go package 状态

是否执行 Go lint/test 不能只看 `go.mod`，必须检测是否存在可分析 Go package。当前仓库允许存在 Go module 基线但暂时没有 Go package。

## 本地与 CI 一致性

Makefile、本地 pre-commit 和 GitHub Actions 应调用同一批治理脚本，避免本地通过但 CI 失败。
