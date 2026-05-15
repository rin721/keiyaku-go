---
state_id: ADR-20260515-BACKEND-001
doc_role: adr
memory_level: L0
state_scope: global
scope: repo
authority_level: ssot_decision
owners: [tech-lead]
status: accepted
effective_date: 2026-05-15
version: 2.0
related_rules: [GOV-P1-002, GOV-P1-006]
source_of_truth: [docs/adr/20260515-default-backend-direction.md]
derived_from: [docs/governance/rules.md, docs/conventions/dependency-injection.md]
read_when: [boundary_sensitive, governance_change]
update_when: [default_rule_changed, dependency_injection_policy_changed, convention_changed]
conflict_policy: accepted_adr_overrides_default_backend_direction
rollback_target: [docs/governance/rules.md, docs/conventions/dependency-injection.md]
verification_target: [scripts/check-governance-sync.ps1, scripts/check-governance-map.ps1]
change_reason: establish default backend direction before application code exists
---

# ADR 20260515：默认后端方向与依赖注入方式

- 状态：accepted
- 日期：2026-05-15
- 负责人：tech-lead
- 关联 Issue 或 PR：

## 背景与上下文
<!-- ADR-001 -->

仓库当前仍处于治理基线阶段，尚未引入 HTTP API、数据库 schema 或业务代码，但已经需要为未来默认实现方向提供稳定裁决，以避免后续实现阶段在框架、Repository 辅助工具和依赖注入方式上反复摇摆。

## 决策内容
<!-- ADR-002 -->

仓库默认后端方向确定为：

- 传输适配器默认采用 Echo。
- Repository 实现辅助工具默认采用 sqlc。
- 依赖注入默认采用手动装配。
- 对复杂依赖图，可按 [docs/conventions/dependency-injection.md](../conventions/dependency-injection.md) 使用 Wire 等编译期生成工具。

该决策不改变现有 P0/P1 规则，仍要求框架和工具选择不得穿透领域边界；传输 DTO、持久层模型和具体中间件实现不得跨层泄露。

## 后果评估
<!-- ADR-003 -->

正面收益是为未来实现建立清晰默认方向，减少目录结构和装配方式的随意性，也方便在治理脚本、评审清单和专题约定中形成稳定预期。代价是未来若要切换默认方向，必须补 ADR，并同步 `rules.md`、`ai-execution.md`、专题约定、评审清单和自动化矩阵。

## 备选方案

- 暂不裁决默认方向，等业务代码出现后再定。
  缺点是会导致早期实现阶段反复形成局部风格，增加后续治理收口成本。
- 默认采用运行时 DI 容器。
  缺点是隐藏依赖关系，与当前分层和可读性目标冲突。

## 后续事项

- [x] 文档已更新。
- [x] `metadata-schema.md`、`automation-matrix.md` 与评审清单已按需同步。
- [x] 如果决策改变了可执行规则，CI 或静态检查已更新。
- [x] 如涉及迁移或高风险变更，已记录回滚或补偿方案。
