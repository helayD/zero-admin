---
project_name: zero-admin
user_name: David
date: 2026-03-11
sections_completed: []
existing_patterns_found: 0
---

# Project Context for AI Agents

_This file contains critical rules and patterns that AI agents must follow when implementing code in this project. Focus on unobvious details that agents might otherwise miss._

---

## Technology Stack & Versions

_Discovered during context generation_

### Backend (Go)

- **Go**: 1.25
- **go-zero**: 1.9.3 (核心微服务框架)
- **MongoDB Driver**: go.mongodb.org/mongo-driver v1.17.6
- **MySQL ORM**: GORM v1.31.1 + GORM Gen v0.3.27
- **Redis**: github.com/redis/go-redis/v9 v9.16.0
- **gRPC**: google.golang.org/grpc v1.77.0
- **JWT**: github.com/golang-jwt/jwt/v4 v4.5.2
- **Elasticsearch**: github.com/elastic/go-elasticsearch/v9 v9.2.0
- **RabbitMQ**: github.com/rabbitmq/amqp091-go v1.10.0
- **Prometheus**: github.com/prometheus/client_golang v1.21.1
- **Jaeger (Tracing)**: OpenTelemetry exporters
- **JSON**: github.com/bytedance/sonic v1.15.0

### Frontend (React)

- **React**: ^17.0.0
- **Ant Design Pro**: 5.2.0
- **Umi**: ^3.5.0
- **TypeScript**: ^4.5.0
- **ECharts**: ^5.4.3
- **Playwright**: ^1.17.0 (E2E Testing)

### Mobile (Flutter)

- **Flutter Mall**: flutter-mall 电商模块

### Infrastructure

- **Docker**: 容器化部署
- **MySQL**: 主数据库
- **Redis**: 缓存
- **MongoDB**: 文档存储
- **Elasticsearch**: 搜索引擎
- **RabbitMQ**: 消息队列

## Critical Implementation Rules

### Go Backend Architecture (go-zero)

1. **三层架构模式**: Handler → Logic → Service
   - Handler: 解析HTTP请求参数，调用Logic
   - Logic: 业务逻辑处理，调用RPC服务
   - Service: RPC服务实现，直接操作数据库

2. **API-RPC通信**:
   - API层处理HTTP请求
   - 通过gRPC调用RPC服务
   - 使用proto定义接口

3. **代码生成**:
   - 使用 `goctl` 工具生成代码
   - 使用 `generate-code` 工具根据数据库表生成

4. **目录结构**:
   ```
   api/admin/          # HTTP API层
   rpc/[module]/       # RPC服务层 (cms, ums, pms, oms等)
   web-admin/          # React前端
   flutter-mall/      # Flutter移动端
   consumer/          # 消费者服务
   job/               # 定时任务
   ```

5. **命名规范**:
   - 文件名: 小写字母+下划线 (snake_case)
   - 结构体: PascalCase
   - Handler: [Name]Handler
   - Logic: [Name]Logic

6. **错误处理**:
   - 使用 errorx 包定义业务错误
   - 统一返回格式: {Code, Message, Data}

### Frontend Architecture (React + Ant Design Pro)

1. **目录结构**:
   - src/pages/ - 页面组件
   - src/components/ - 通用组件
   - src/services/ - API服务
   - src/models/ - 状态管理

2. **开发命令**:
   - `npm run dev` - 开发模式
   - `npm run build` - 构建生产版本
   - `npm run lint` - 代码检查
   - `npm run tsc` - TypeScript检查

3. **Ant Design Pro组件**:
   - 使用 ProTable, ProForm, ProLayout 等高级组件
   - 遵循 Pro 的代码规范

### Database

1. **ORM**: 使用 GORM + GORM Gen
2. **生成模型**: 通过数据库表自动生成
3. **事务处理**: 使用 GORM 事务

### API Response Format

统一响应格式:
```json
{
  "code": "000000",
  "message": "成功消息",
  "data": {},
  "current": 1,
  "pageSize": 10,
  "total": 100
}
```

### Development Workflow

1. **后端开发流程**:
   - 设计数据库表
   - 使用 generate-code 生成基础代码
   - 在 rpc 层实现业务逻辑
   - 在 api 层添加 HTTP 接口

2. **前端开发流程**:
   - 调用 API 服务
   - 使用 Ant Design 组件构建页面

### Critical Rules

1. **不要手动修改生成的代码** - 使用 goctl 重新生成
2. **API 层只做参数校验和转发** - 业务逻辑在 RPC 层
3. **统一错误处理** - 使用项目定义的 errorx
4. **数据库操作使用事务** - 确保数据一致性
5. **遵循 go-zero 框架规范** - 不要绕过框架约定

---

_Context generated on 2026-03-11_
