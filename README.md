# Keiyaku-Go

Keiyaku-Go 是一个基于 Go 构建的企业级博客与内容管理系统后端。当前实现采用 Gin 与 GORM，并按 Clean Architecture 组织代码，明确隔离 HTTP 交付层、应用用例层、领域层和基础设施适配器。

## 当前技术栈

- Web 框架：Gin
- 数据库：MySQL 8.0 + GORM v2
- 缓存：Redis + go-redis/v9
- 日志：Uber Zap JSON 日志，支持级别控制与日志切割
- 配置：Viper，支持 YAML 与环境变量覆盖
- 认证与授权：JWT + Casbin v3
- ID：Twitter Snowflake，用于生成对外业务 ID
- 依赖注入：只使用显式手动构造函数

## 架构概览

依赖方向如下：

```text
internal/domain <- internal/application <- internal/api
internal/domain <- internal/application <- internal/infrastructure
cmd/bootstrap -> api + application + infrastructure
```

HTTP DTO、响应结构和 HTTP 状态映射只存在于 Gin 适配层附近。GORM Model 只存在于基础设施层。Application 层拥有用例、Port、应用错误码和错误包装；Domain 实体不 import Gin、GORM、Redis、JWT、Casbin 或其他适配器包。

主链路：

1. 请求进入 Gin 路由和中间件链。
2. Handler 校验 DTO，并映射为应用层 Command 或 Query。
3. Application Service 编排领域逻辑和 Port。
4. Infrastructure Adapter 使用 GORM、Redis、JWT、Casbin 和 Snowflake 实现 Port。
5. Handler 返回统一响应结构：`{code:int,msg:string,data:any}`。

更多细节见：[系统结构设计](docs/architecture/system-design.md)。

## 目录说明

```text
cmd/            进程入口，包括 API Server 和迁移命令。
internal/       私有应用代码，按 Clean Architecture 分层。
api/            OpenAPI 风格的公开 API 契约。
configs/        YAML 配置模板。
migrations/     MySQL schema 迁移脚本。
deployments/    本地与部署资产，例如 Docker Compose。
docs/           架构、API、数据库、ADR 和治理文档。
pkg/            与业务无关的可复用包。
scripts/        治理、分层、测试和 CI 辅助脚本。
test/           跨模块或集成测试资产。
```

## 本地开发

启动 MySQL 和 Redis：

```powershell
docker compose -f deployments/docker/docker-compose.yml up -d
```

执行数据库迁移：

```powershell
go run ./cmd/migrate -dsn "keiyaku:keiyaku@tcp(127.0.0.1:3306)/keiyaku?charset=utf8mb4&parseTime=True&loc=UTC"
```

启动 API Server：

```powershell
go run ./cmd/api -config configs/config.yaml
```

健康检查：

```powershell
Invoke-RestMethod http://127.0.0.1:8080/healthz
```

## 验证

运行 Go 检查：

```powershell
go test ./...
go vet ./...
```

运行架构与治理检查：

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-layering.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/check-governance.ps1
```

## 关键文档

- [系统结构设计](docs/architecture/system-design.md)
- [HTTP API 契约](docs/api/http-api.md)
- [数据库结构](docs/architecture/database-schema.md)
- [Gin 与 GORM 架构决策](docs/adr/20260515-adopt-gin-gorm-clean-architecture.md)
- [治理入口](docs/governance/README.md)
- [AI 执行协议](docs/governance/ai-execution.md)

## 当前实现状态

首批后端切片已经落地：

- Auth 注册与登录用例
- 当前用户资料接口
- Article 创建、列表与详情用例
- Gin 中间件链：TraceID、Recovery、访问日志、CORS、限流、熔断、JWT、Casbin
- MySQL 初始迁移：users、roles、permissions、articles、categories、tags、comments、casbin_rule
