---
state_id: GOV-META-001
doc_role: metadata_schema
memory_level: L0
state_scope: global
scope: repo
authority_level: binding
owners: [tech-lead]
status: active
effective_date: 2026-05-15
version: 2.0
related_rules: []
source_of_truth: [docs/governance/metadata-schema.md]
derived_from: [docs/adr/20260515-governance-ssot-structure.md, docs/adr/20260515-governance-state-model.md]
read_when: [governance_change]
update_when: [metadata_standard_changed, governance_structure_changed, automation_changed]
conflict_policy: metadata_schema_defines_governance_taxonomy
rollback_target: [docs/adr/20260515-governance-ssot-structure.md, docs/adr/20260515-governance-state-model.md]
verification_target: [scripts/check-governance-metadata.ps1, scripts/check-governance-taxonomy.ps1, scripts/check-governance-map.ps1]
---

# 治理元数据标准

本文档定义治理文档 front matter 的必填字段、受控枚举和任务标签词表。机器扫描必须优先读取这里的定义，而不是靠文件名猜测文档职责。

## 必填字段

metadata v2 适用于所有持久化治理资产。`governance-map.json` 作为 JSON 派生索引，不使用 front matter，但必须暴露等价字段。

所有面向治理的 `*.md`、`*.yaml`、`*.yml` 文档必须包含：

- `state_id`
- `doc_role`
- `memory_level`
- `state_scope`
- `scope`
- `authority_level`
- `owners`
- `status`
- `version` 或 `effective_date`
- `related_rules`
- `source_of_truth`
- `derived_from`
- `read_when`
- `update_when`
- `conflict_policy`
- `rollback_target`
- `verification_target`

## 条件字段

以下字段按需使用，但一旦出现就必须符合 schema 语义：

- `task_entrypoint`
- `depends_on`
- `impacts`
- `supersedes`
- `superseded_by`
- `change_reason`

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
- `governance_map`
<!-- META-DOC_ROLE-END -->

## `memory_level`

<!-- META-MEMORY_LEVEL-START -->
- `L0`
- `L1`
- `L2`
<!-- META-MEMORY_LEVEL-END -->

## `state_scope`

<!-- META-STATE_SCOPE-START -->
- `global`
- `module`
- `local`
- `ephemeral`
<!-- META-STATE_SCOPE-END -->

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
- `governance-map.json` 是派生索引，其等价元数据必须满足 `doc_role: governance_map`、`authority_level: derived` 和 `state_scope: global`。
- Accepted ADR 必须使用 `doc_role: adr`、`authority_level: ssot_decision` 和 `status: accepted`。
- 历史参考文档必须使用 `doc_role: historical_reference`，且 `status` 只能是 `historical` 或 `deprecated`。
- `source_of_truth` 指向当前状态真正依赖的 SSOT；派生产物不得把自己伪装成唯一真相源。
- `derived_from` 只记录明确来源，不做正文推断。
- `rollback_target` 和 `verification_target` 至少各有一个条目。
- `task_entrypoint: true` 只标记仓库的规范 first-hop 入口。
- 任何新增枚举值都必须同步更新本文档、相关文档 front matter、`scripts/check-governance-taxonomy.ps1` 的解析行为以及治理检查。

## 字段语义

- `state_id`：治理状态对象的稳定标识，需全仓唯一。
- `memory_level`：治理记忆层级；`L0` 表示仓库级长期记忆，`L1` 表示专题级长期记忆，`L2` 表示需要持久化的任务级记忆。
- `state_scope`：作用域边界；持久化文档通常使用 `global` 或 `module`，`local`/`ephemeral` 主要用于显式阻止误提升。
- `conflict_policy`：当前对象在冲突时如何让位，例如 `derived_must_yield_to_ssot`。
- `source_of_truth`：当前对象依赖的最终裁决来源。
- `derived_from`：当前对象由哪些来源整理、提炼或路由而来。
- `rollback_target`：回滚时要恢复到的文件、ADR 或规则来源。
- `verification_target`：完成后必须通过的脚本、测试或 CI 表面。
