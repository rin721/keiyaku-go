---
doc_role: review_checklist
scope: review
authority_level: derived
owners: [tech-lead]
status: active
effective_date: 2026-05-15
version: 1.0
related_rules: [GOV-P0-001, GOV-P0-002, GOV-P0-003, GOV-P0-004, GOV-P1-001, GOV-P1-002, GOV-P1-003, GOV-P1-004, GOV-P1-005, GOV-P1-006]
read_when: [review_change, governance_change]
update_when: [review_policy_changed, default_rule_changed, automation_changed]
---

# 代码评审清单

本清单只保留暂时无法可靠自动化的判断项。可自动化规则应进入脚本、lint、测试或 CI。

## P0 人工确认

- [ ] GOV-P0-001：没有把 PO、ORM Model 或持久层模型伪装成中性类型返回给 Handler 或 JSON 响应。
- [ ] GOV-P0-001：Repository、Domain、Infrastructure 方法签名没有语义上依赖 transport DTO。
- [ ] GOV-P0-003：密码存储方案具备版本化参数和渐进式 rehash 设计。
- [ ] GOV-P0-004：敏感日志输出经过明确白名单结构，而不是只依赖调用点自觉。

## P1 人工确认

- [ ] GOV-P1-001：入口、下游调用、异步任务和日志的 TraceID/CorrelationID 传播路径完整。
- [ ] GOV-P1-002：Repository Port 归属符合业务语义，而不是按实现便利放置。
- [ ] GOV-P1-003：迁移方案具备前置检查、兼容窗口、回滚或补偿方案。
- [ ] GOV-P1-004：异步消费者具备幂等设计，重复投递不会破坏业务状态。
- [ ] GOV-P1-005：测试覆盖关键业务分支，Mock/Fake 没有掩盖真实边界风险。
- [ ] GOV-P1-006：依赖注入方式没有引入运行时反射容器或隐藏生成代码。

## P2 建议项

- [ ] 观测、指标、文件处理、邮件能力、并发控制等推荐实践已按当前变更风险合理取舍。
- [ ] 如果某个局部实践正在成为默认风格，已按 `change-management.md` 判断是否需要 ADR 与闭环同步。
