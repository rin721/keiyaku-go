---
doc_role: metadata_schema
scope: repo
authority_level: binding
owners: [tech-lead]
status: active
effective_date: 2026-05-15
version: 1.0
related_rules: []
read_when: [governance_change]
update_when: [metadata_standard_changed, governance_structure_changed, automation_changed]
---

# 治理元数据标准

本文档定义治理文档 front matter 的必填字段、受控枚举和任务标签词表。机器扫描必须优先读取这里的定义，而不是靠文件名猜测文档职责。

## 必填字段

所有面向治理的 `*.md`、`*.yaml`、`*.yml` 文档必须包含：

- `doc_role`
- `scope`
- `authority_level`
- `owners`
- `status`
- `version` 或 `effective_date`
- `related_rules`
- `read_when`
- `update_when`

## `doc_role`

<!-- META-DOC_ROLE-START -->
- `ai_entry`
- `navigation`
- `governance_rules`
- `ai_execution`
- `governance_process`
- `automation_spec`
- `metadata_schema`
- `convention`
- `adr_index`
- `adr`
- `review_checklist`
- `exception_registry`
- `historical_reference`
- `template`
<!-- META-DOC_ROLE-END -->

## `scope`

<!-- META-SCOPE-START -->
- `repo`
- `pkg`
- `layering`
- `migrations`
- `async`
- `ci`
- `security`
- `testing`
- `dependency_injection`
- `review`
- `architecture_history`
<!-- META-SCOPE-END -->

## `authority_level`

<!-- META-AUTHORITY_LEVEL-START -->
- `entry`
- `ssot_navigation`
- `ssot_rules`
- `ssot_decision`
- `binding`
- `derived`
- `record`
- `template`
<!-- META-AUTHORITY_LEVEL-END -->

## `status`

<!-- META-STATUS-START -->
- `draft`
- `active`
- `accepted`
- `deprecated`
- `historical`
<!-- META-STATUS-END -->

## `read_when`

`read_when` 只使用任务标签，不再混用不同粒度的上下文词。

<!-- META-READ_WHEN-START -->
- `all_tasks`
- `governance_change`
- `pkg_change`
- `boundary_sensitive`
- `migration_sensitive`
- `async_sensitive`
- `test_or_ci`
- `security_sensitive`
- `review_change`
- `exception_review`
<!-- META-READ_WHEN-END -->

## `update_when`

`update_when` 只描述治理同步触发条件，不描述单次临时任务。

<!-- META-UPDATE_WHEN-START -->
- `routing_changed`
- `execution_protocol_changed`
- `metadata_standard_changed`
- `governance_structure_changed`
- `default_rule_changed`
- `default_behavior_changed`
- `adr_policy_changed`
- `adr_accepted`
- `convention_changed`
- `automation_changed`
- `review_policy_changed`
- `exception_policy_changed`
- `ci_changed`
- `security_policy_changed`
- `migration_policy_changed`
- `async_policy_changed`
- `test_policy_changed`
- `dependency_injection_policy_changed`
- `historical_status_changed`
- `exception_registry_changed`
- `template_changed`
<!-- META-UPDATE_WHEN-END -->

## 约束说明

- `AGENTS.md` 和 `CLAUDE.md` 使用 `doc_role: ai_entry` 与 `authority_level: entry`。
- 仓库导航真相必须使用 `doc_role: navigation` 与 `authority_level: ssot_navigation`。
- 仓库规则真相必须使用 `doc_role: governance_rules` 与 `authority_level: ssot_rules`。
- Accepted ADR 必须使用 `doc_role: adr`、`authority_level: ssot_decision` 和 `status: accepted`。
- 历史参考文档必须使用 `doc_role: historical_reference`，且 `status` 只能是 `historical` 或 `deprecated`。
- 任何新增枚举值都必须同步更新本文档、相关文档 front matter、`scripts/check-governance-taxonomy.ps1` 的解析行为以及治理检查。
