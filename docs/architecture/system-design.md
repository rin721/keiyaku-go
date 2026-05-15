---
state_id: ARCH-SYSTEM-001
doc_role: convention
memory_level: L1
state_scope: module
scope: repo
authority_level: binding
owners: [tech-lead]
status: active
effective_date: 2026-05-15
version: 1.0
related_rules: [GOV-P0-001, GOV-P0-002, GOV-P1-001, GOV-P1-002, GOV-P1-006]
source_of_truth: [docs/adr/20260515-adopt-gin-gorm-clean-architecture.md]
derived_from: [docs/governance/rules.md, docs/conventions/layering.md, docs/conventions/dependency-injection.md]
read_when: [boundary_sensitive, governance_change]
update_when: [default_behavior_changed, convention_changed, adr_accepted]
conflict_policy: binding_must_yield_to_ssot
rollback_target: [docs/adr/20260515-adopt-gin-gorm-clean-architecture.md]
verification_target: [scripts/check-layering.ps1, scripts/check-governance-map.ps1]
---

# 系统结构设计

## 分层结构

```mermaid
flowchart LR
  Client["Client / Admin UI"] --> Gin["Gin Router + Middleware"]
  Gin --> Handler["HTTP Handler + DTO"]
  Handler --> Usecase["Application Usecase"]
  Usecase --> Domain["Domain Entity / Value Object"]
  Usecase --> Port["Application Port"]
  Port --> Infra["Infrastructure Adapter"]
  Infra --> MySQL["MySQL 8.0 / GORM Model"]
  Infra --> Redis["Redis / go-redis"]
  Infra --> Security["JWT / Casbin / Snowflake"]
```

依赖方向固定为 `api -> application -> domain`。`infrastructure` 只能实现 `application` 或 `domain` 暴露的 Port，不允许把 Gin DTO、GORM Model、Redis、JWT 或 Casbin 类型传入 Domain。

## 请求链路

1. 请求进入 Gin HTTP Server。
2. 中间件处理 TraceID、Recovery、结构化访问日志、CORS、限流、熔断、JWT 与 Casbin。
3. Handler 绑定并校验 DTO，把 DTO 显式转换为 Application Command 或 Query。
4. Application 编排业务用例、事务边界、Repository Port、Cache Port、Token Port 与 IDGenerator Port。
5. Domain 执行业务不变量校验，不依赖任何外部框架。
6. Infrastructure 通过 GORM/Redis/Casbin/JWT/Snowflake 完成具体适配。
7. Handler 把应用结果映射为统一 `{code,msg,data}` JSON。

## 启动链路

```mermaid
flowchart TD
  Main["cmd/api/main.go"] --> Config["Load Viper Config"]
  Config --> Logger["Build Zap Logger"]
  Logger --> Store["Open MySQL + Redis"]
  Store --> Security["Build JWT + Casbin + Snowflake"]
  Security --> App["Construct Repositories + Usecases + Handlers"]
  App --> Router["Register Gin Routes"]
  Router --> Server["Start HTTP Server"]
  Server --> Shutdown["SIGINT/SIGTERM Graceful Shutdown"]
```

`cmd/api/main.go` 只负责进程生命周期。依赖装配集中在 `internal/bootstrap`，并通过显式构造函数自下而上创建。
