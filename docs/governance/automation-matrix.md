---
doc_role: automation_spec
scope: repo
authority_level: derived
owners: [tech-lead]
status: active
effective_date: 2026-05-15
version: 1.0
related_rules: [GOV-P0-001, GOV-P0-002, GOV-P0-003, GOV-P0-004, GOV-P1-001, GOV-P1-002, GOV-P1-003, GOV-P1-004, GOV-P1-005, GOV-P1-006]
read_when: [governance_change, ci_change, review_change, boundary_sensitive, operational_sensitive]
update_when: [rule_placement_changed, automation_changed, review_checklist_changed, convention_changed]
---

# 自动化归位矩阵

本矩阵决定治理规则进入哪个执行表面。稳定规则进入 `rules.md`，局部默认进入 convention，可重复检查进入脚本、lint、测试或 CI，难以机器判断的内容进入评审清单。

## 架构与分层

| 事项 | 规则来源 | 执行位置 |
| --- | --- | --- |
| `pkg` 不得 import `internal` | GOV-P0-002 / `conventions/pkg.md` | `scripts/check-layering.ps1`、`.golangci.yml` |
| Domain 不得 import 上层或 adapter | GOV-P1-002 / `conventions/layering.md` | `scripts/check-layering.ps1` |
| Application/Repository/Infrastructure 不得 import transport | GOV-P0-002 / `conventions/layering.md` | `scripts/check-layering.ps1`、`.golangci.yml` |
| DTO/PO/ORM 模型语义泄露 | GOV-P0-001 / `conventions/layering.md` | 评审清单，后续可扩展静态检查 |
| Repository Port 归属 | GOV-P1-002 / `conventions/layering.md` | 评审清单 |

## 测试

| 事项 | 规则来源 | 执行位置 |
| --- | --- | --- |
| Domain/Application 测试保持快速 | GOV-P1-005 / `conventions/testing.md` | `scripts/check-test-conventions.ps1` + 评审 |
| testcontainers 测试必须带 `integration` build tag | GOV-P1-005 / `conventions/testing.md` | `scripts/check-test-conventions.ps1` |
| Repository/Infrastructure 集成测试使用真实中间件 | GOV-P1-005 / `conventions/testing.md` | 评审清单，脚本做基础提示 |
| 有 Go module 但无 package 时安全跳过 lint/test | `conventions/ci.md` | `scripts/check-go-package-state.ps1`、Makefile、CI |

## 安全与日志

| 事项 | 规则来源 | 执行位置 |
| --- | --- | --- |
| 禁止密码使用快速哈希 | GOV-P0-003 / `conventions/security-logging.md` | `scripts/check-governance.ps1` 基础扫描 + gosec |
| 禁止明文敏感日志 | GOV-P0-004 / `conventions/security-logging.md` | `scripts/check-governance.ps1` 基础扫描 + 评审 |
| Secret 扫描 | GOV-P0-004 | gitleaks / pre-commit / CI |
| TraceID/CorrelationID 传播 | GOV-P1-001 | 评审清单，后续补集成测试 |

## 迁移与异步

| 事项 | 规则来源 | 执行位置 |
| --- | --- | --- |
| 高风险 migration 使用灰度模板 | GOV-P1-003 / `conventions/migrations.md` | 评审清单 + ADR |
| 迁移具备前置检查、补偿或回滚 | GOV-P1-003 | 评审清单 |
| 异步任务具备持久化、重试、死信、幂等 | GOV-P1-004 / `conventions/async-jobs.md` | 评审清单 + ADR |
| 回填任务具备 checkpoint | GOV-P1-004 | 评审清单 |

## 治理资产

| 事项 | 规则来源 | 执行位置 |
| --- | --- | --- |
| 治理文档存在 | `governance/README.md` | `scripts/check-governance.ps1` |
| 元数据完整性 | `governance/README.md` | `scripts/check-governance-metadata.ps1` |
| rule_id 链接完整性 | `governance/rules.md` | `scripts/check-rule-links.ps1` |
| debt / break-glass 到期 | `change-management.md` | `scripts/check-exception-expiry.ps1` |
| 治理变更同步面 | `change-management.md` | `review/governance-change-checklist.md` |

## Prompt 边界

Agent 入口只保留 first-hop、任务路由、自检、ADR/破例升级条件。稳定规则不得只存在于 Prompt 中。
