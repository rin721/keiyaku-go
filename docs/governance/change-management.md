---
state_id: GOV-CHANGE-001
doc_role: governance_process
memory_level: L0
state_scope: global
scope: repo
authority_level: binding
owners: [tech-lead]
status: active
effective_date: 2026-05-15
version: 2.0
related_rules: []
source_of_truth: [docs/governance/change-management.md]
derived_from: [docs/governance/rules.md, docs/adr/20260515-governance-state-model.md]
read_when: [governance_change, boundary_sensitive, migration_sensitive, async_sensitive, exception_review]
update_when: [default_behavior_changed, default_rule_changed, adr_policy_changed, exception_policy_changed, governance_structure_changed]
conflict_policy: change_management_orchestrates_sync_under_ssot
rollback_target: [docs/governance/rules.md, docs/adr/20260515-governance-state-model.md]
verification_target: [scripts/check-governance-sync.ps1, scripts/check-exception-expiry.ps1, scripts/check-governance-map.ps1]
---

# 治理变更管理

本文档定义默认行为冲突、治理演进、历史同步、治理债务和 break-glass 的处理方式。默认行为变更未完成闭环同步时，不得视为任务完成。

## 状态边界

- `Global State`：导航、规则、执行协议、Accepted ADR、异常登记和派生索引等仓库级长期状态。
- `Module State`：专题约定、评审清单、模板等局部长期状态。
- `Local State`：任务计划、变更范围、同步清单和验证范围。
- `Ephemeral State`：扫描结果、临时推断和对话草稿。

`Local State` 和 `Ephemeral State` 不得直接污染 `Global State` 或 `Module State`。只有补齐 ADR、同步面、回滚目标和验证目标后，才能提升为长期治理状态。

## 通用动作闭环

任何会产生仓库事实、任务结论、代码或文档变更、自动化行为、验证结果、同步判断或交付说明的操作，都是需要治理的动作。只读探索可以停留在 `Ephemeral State`，但一旦它支撑决策、计划、同步判断或交付说明，就必须按动作闭环处理。

每个动作必须经过同一套生命周期：

1. `Classify`：识别动作类型、任务标签、作用域和是否涉及代码、文档、自动化、配置、数据、依赖、项目说明或治理资产。
2. `Scope`：判断动作结果属于 `Ephemeral State`、`Local State`、`Module State` 还是 `Global State`，并识别是否触发 ADR、治理债务、break-glass 或人工审批。
3. `Sync`：先执行同步触发判定；触发同步时列出需要评估的同步面，需要更新时更新目标载体，不需要更新时记录依据。
4. `Execute`：按最小充分范围执行动作，不扩大到未授权目录或无关历史实现。
5. `Verify`：运行与动作影响面匹配的测试、脚本、lint、治理校验或人工检查；不能运行时记录原因。
6. `Record`：在交付说明、治理产物或评审记录中说明动作结果、验证结果、未更新同步面的依据和剩余风险。

动作闭环适用于“执行”和“不执行”两类结果。决定不更新文档、不改代码、不跑某项检查或不触发 ADR，也必须作为动作记录其依据。

## 同步触发判定

每个动作必须先判定是否触发同步更新。判定结论使用两类结果：

- `sync_required`：动作会改变仓库长期事实、默认行为、公共契约、项目说明、治理资产、自动化检查面、验证策略、目录结构、配置入口、依赖工具链、数据模型、安全模型或交付方式。
- `sync_not_required`：动作只产生临时观察、局部探索、无状态查询、未落地草案，或结论不会改变任何仓库事实和长期载体。

出现以下任一情况，必须判定为 `sync_required`，并进入同步面评估：

- 修改代码行为、包边界、依赖方向、公共接口、错误模型、配置结构、启动流程或目录结构。
- 修改 README、架构、API、数据库、运行说明、ADR、convention、治理规则、AI 执行协议、评审清单、异常登记或 Prompt。
- 修改治理脚本、lint、测试检查、Makefile、workflow、CI、预提交、派生索引生成逻辑或任何验证表面。
- 新增、删除或移动仓库跟踪文件，尤其是会影响读者理解、工具扫描、构建、测试或部署的文件。
- 运行会生成或刷新仓库跟踪产物的命令，例如刷新 [governance-map.json](governance-map.json)。
- 做出“无需更新某文档、无需 ADR、无需测试、无需同步索引”等交付相关判断。

只有同时满足以下条件，才能判定为 `sync_not_required`：

- 动作不修改仓库跟踪文件。
- 动作不改变任何默认规则、默认边界、默认工作流、公共契约或项目说明。
- 动作结果不作为长期事实写入代码、文档、脚本、CI、治理索引或交付承诺。
- 不存在按任务路由必须评估的同步面。

`sync_not_required` 不是省略动作；必须在交付说明、评审记录或治理产物中记录不触发同步的依据。

## 冲突级别

- L0 局部实现差异：只影响单个目录或能力，不改变默认规则、默认边界或默认工作流。可直接更新局部约定或实现。
- L1 默认行为变化：会成为后续默认写法或默认流程。必须先补 ADR，再同步治理规则、AI 执行协议、专题约定、评审清单、自动化矩阵和脚本/CI。
- L2 P1 偏离：偏离默认治理要求。必须补 ADR，并由 owner 审批。
- L3 P0、安全、权限模型、不可逆数据风险：停止实施，要求人工决策；P0 不允许通过 break-glass 豁免。
- L4 历史实现大量不一致：先登记治理债务，再制定迁移范围、验证方式、收益、风险和停止条件。

## 决策路径

1. 先判断是否触及 P0、安全、权限或不可逆数据风险；若是，立即停止并升级。
2. 若不是，再判断是否改变默认行为、默认边界、默认风格、默认依赖方式、默认工作流或默认执行路径。
3. 若不改变默认项，按局部实现处理。
4. 若改变默认项，按 L1/L2/L4 分级，并完成文档、ADR、评审、自动化和历史同步决策。

## 状态晋升规则

- 只影响单次任务的计划、验证和范围结论停留在 `Local State`。
- 只服务于当前分析过程的推断停留在 `Ephemeral State`。
- 只有当结论会成为默认行为、默认边界、默认风格或默认工作流时，才允许提升为 `Module State` 或 `Global State`。
- 提升前必须补齐：`source_of_truth`、`derived_from`、`rollback_target`、`verification_target`、所需 ADR 和同步面。

## 默认风格演进触发条件

- 某种实现从单点实践变成多个模块复用。
- 评审中反复批准同类偏离。
- 自动化发现同类问题高频出现。
- 新约定会改变目录边界、依赖方向、测试策略、迁移策略或 CI 门禁。
- 用户明确要求把某种风格作为长期默认。

触发后必须评估是否同步更新：`rules.md`、`README.md`、`ai-execution.md`、`metadata-schema.md`、相关 convention、ADR、review checklist、automation matrix、脚本、lint、测试和 CI。

## 治理维护循环

当出现以下任一信号时，应把局部修补升级为治理更新：

- 同类 metadata 漏填、索引漂移或路由误读反复出现。
- 评审中反复批准相同类型的受控偏离。
- 自动化反复发现同类问题，但规则仍只存在于人工评审中。
- 某种局部实践已经演化成新的默认行为。

## 自动化变更闭环

凡是修改治理脚本、lint、测试约定检查、CI workflow、派生索引生成逻辑或任何会改变治理检查面的自动化实现，都必须在同一变更中完成同步判断：

- 若自动化新增、删除或改变检查事项，必须同步更新 [automation-matrix.md](automation-matrix.md) 或对应专题约定，说明规则来源和执行位置。
- 若变更影响 AI 代理加载、判断、执行或交付自检流程，必须同步更新 [ai-execution.md](ai-execution.md)。
- 若变更影响治理文档、元数据、索引内容或索引生成逻辑，必须刷新 [governance-map.json](governance-map.json) 并运行对应校验。
- 不得只修改脚本、CI 或派生索引而跳过规则来源、同步面和验证目标；无法同步时必须说明阻塞原因，并按停止条件升级。

## 治理债务

治理债务用于暂时允许历史代码或历史文档不完全符合新治理，但必须可追踪、可复审、可关闭。债务登记在 `docs/governance/exceptions.yaml`。

必填字段：`rule_id`、`reason`、`owner`、`created_at`、`review_at` 或 `expiry`、`required_followup`、`rollback_target`、`verification_target`、`status`。

## Boy Scout Rule

当治理债务位于当前正在修改的文件、目录或紧邻调用链中，并同时满足以下条件时，AI 应主动建议顺带清理：

- 不改变默认行为、默认边界或公共接口。
- 不引入新的架构决策，不需要新增 ADR。
- 额外修改和验证成本不超过本次任务的大约 10% 到 20%。
- 清理后的验证路径与本次任务基本重合。

如果不满足这些条件，AI 必须显式说明为什么本次不适合顺带清理，并保留或补充治理债务标记。

## Break-glass

Break-glass 只用于紧急恢复或阻止重大风险扩大。

- 不得覆盖 P0。
- 必须设置 `expiry`。
- 必须写明 `required_followup`。
- 如果偏离默认规则超过一个发布周期，或改变默认行为，必须补临时 ADR。
- 到期必须关闭、续期或转成治理债务。

## 历史代码同步

默认采用 touched-code first。只有安全、日志、密码、依赖方向、迁移兼容、CI 触发等高风险规则，才发起专项全仓扫描。

停止条件：自动化误报高、影响范围不清、缺少 ADR、迁移收益低于成本、或需要超过当前任务范围的大规模重构。

## 默认行为变更闭环

只要变更会影响默认行为、默认约束、默认边界、默认风格、默认工作流或默认执行路径，就必须判断并同步：

- 是否需要新增或更新 ADR。
- 是否需要更新导航、执行协议、元数据 schema、专题约定、评审清单和自动化矩阵。
- 是否需要回填 `governance-map.json`、脚本、lint、测试和 CI。
- 是否需要 touched-code first 同步历史实现。
- 是否需要兼容层、迁移路径、回滚方案和验证方案。

## 同步判断记录

同步面判断本身就是治理动作。对每个按路由需要评估的同步面，结论必须显式记录：

- 如果需要更新，必须列出目标文件或自动化表面，并完成对应验证。
- 如果不需要更新，必须说明不更新的依据，例如未改变默认规则、未改变执行协议、未改变元数据枚举或未改变评审政策。
- 不得用沉默跳过代替“不需要更新”的判断；缺少判断记录时，变更不得视为治理闭环完成。

## 项目说明同步

项目说明文档同步判断也是治理动作。凡是变更会影响读者理解项目结构、目录职责、启动方式、配置方式、公开契约、架构边界、验证命令或当前实现状态，必须同步评估项目说明文档：

- 首要评估 [../../README.md](../../README.md)、[../architecture/system-design.md](../architecture/system-design.md)、[../api/http-api.md](../api/http-api.md) 以及与变更直接相关的架构、API、数据库或运行文档。
- 如果实现删除、移动或新增目录，必须检查项目目录说明是否仍准确。
- 如果公开 API、响应契约、启动命令或配置入口未变化，必须显式记录对应项目说明文档不需要更新的依据。
- 带治理元数据的项目说明文档发生变化后，必须刷新 [governance-map.json](governance-map.json) 并运行治理校验。
