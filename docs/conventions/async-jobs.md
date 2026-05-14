---
doc_role: convention
scope: async
authority_level: binding
owners: [tech-lead]
status: active
effective_date: 2026-05-15
version: 1.0
related_rules: [GOV-P1-001, GOV-P1-004]
read_when: [async_sensitive, governance_change]
update_when: [async_policy_changed, adr_accepted, automation_changed]
---

# 异步任务约定

当任务不能可靠地在请求生命周期内完成，或需要重试、削峰、跨实例消费时，应进入异步系统。

## 默认能力

- 持久化任务状态。
- 支持重试和死信队列。
- 支持优雅关闭。
- 暴露可观测指标和结构化日志。
- 携带 TraceID 或 CorrelationID。

## 消费端要求

- 使用幂等键、状态机、唯一约束或去重表处理重复投递。
- 失败必须可重试或可补偿。
- 长耗时回填任务必须记录 checkpoint。

## ADR 触发

如果异步方案缺少持久化、重试、死信或幂等能力，必须补 ADR 并说明风险控制。
