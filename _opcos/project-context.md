---
project_name: 九克城
user_name: David
date: 2026-03-14
sections_completed:
  - technology_stack
  - language_rules
  - framework_rules
  - testing_rules
  - quality_rules
  - workflow_rules
  - anti_patterns
status: complete
rule_count: 55
optimized_for_llm: true
---

# Project Context for AI Agents

_This file contains critical rules and patterns that AI agents must follow when implementing code in this project. Focus on unobvious details that agents might otherwise miss._

---

## Technology Stack & Versions

### Backend (Go)
- **Go**: 1.25
- **go-zero**: 1.9.3
- **GORM**: 1.31.1 + GORM Gen 0.3.27
- **MongoDB Driver**: v1.17.6
- **Redis**: v9.16.0
- **gRPC**: v1.77.0
- **JWT**: golang-jwt/jwt v4.5.2
- **Elasticsearch**: v9.2.0
- **RabbitMQ**: v1.10.0
- **Prometheus**: v1.21.1

### Frontend (React)
- **React**: 17.0.0
- **Ant Design Pro**: 5.2.0
- **Umi**: 3.5.0
- **TypeScript**: 4.5.0

### Mobile (Flutter)
- **Flutter SDK**: ^3.10.7
- **Provider**: ^6.1.5+1 (状态管理)
- **Dio**: ^5.9.0 (HTTP 客户端)
- **cached_network_image**: ^3.4.1 (图片缓存)
- **shared_preferences**: ^2.5.4 (本地存储)
- **card_swiper**: ^3.0.1 (卡片轮播)
- **easy_refresh**: ^3.4.0 (下拉刷新)
- **flutter_spinkit**: ^5.2.2 (加载动画)
- **bottom_sheet**: ^4.0.4 (底部弹窗)

---

## 基础环境要求

### 服务器信息

| 属性 | 值 |
|------|------|
| **服务器地址** | 47.107.224.56 |
| **SSH 端口** | 22 |
| **用户名** | root |
| **密码** | Qianmai1# |
| **SSH 命令** | `ssh root@47.107.224.56` |

### 必须的基础设施服务

| 服务 | 版本 | 端口 | 用途 |
|------|------|------|------|
| **MySQL** | 8.0+ | 3306 | 主数据库 (gozero) |
| **Redis** | 6.0+ | 16379 | 缓存 (密码: 123456) |
| **MongoDB** | 7.0+ | 27017 | 会员数据存储 |
| **Elasticsearch** | 8.x | 9200 | 搜索服务 (cluster: opencoze) |
| **RabbitMQ** | 3.9+ | 5672 | 消息队列 (用户: test/test) |
| **Etcd** | 3.5+ | Docker 2379 | 服务注册与发现 (Milvus) |

### API 网关服务

| 服务 | 端口 | 说明 |
|------|------|------|
| **admin-api** | 8888 | 管理端 API 网关 |
| **front-api** | 9999 | 用户端 API 网关 (Flutter 移动端对接) |

### RPC 微服务

| 服务 | 端口 | 说明 |
|------|------|------|
| **sys-rpc** | 8070 | 系统管理 |
| **ums-rpc** | 8081 | 会员管理 |
| **pms-rpc** | 8082 | 商品管理 |
| **oms-rpc** | 8083 | 订单管理 |
| **sms-rpc** | 8084 | 营销管理 |
| **cms-rpc** | 8085 | 内容管理 |
| **search-rpc** | 8088 | 搜索服务 |

### 监控服务

| 服务 | 端口 | 说明 |
|------|------|------|
| **Prometheus** | 各服务独立 | 指标收集 |

### Flutter 移动端配置

- **API 地址**: `http://47.107.224.56:9999` (生产环境)
- **静态资源**: 阿里云 OSS (macro-oss.oss-cn-shenzhen.aliyuncs.com)

### 部署的访问地址

| 服务 | 地址 | 说明 |
|------|------|------|
| **管理后台** | http://47.107.224.56:8000 | 管理端前端 (Ant Design Pro) |
| **API 文档** | http://47.107.224.56:8888/swagger | Swagger API 文档 |
| **用户端 API** | http://47.107.224.56:9999 | 移动端 API |

### 账号信息

| 系统 | 账号 | 密码 |
|------|------|------|
| **管理后台** | admin | 123456 |
| **MySQL** | root | 12341qweqfsd2356 |
| **Redis** | - | 123456 |
| **MongoDB** | admin | admin123 |
| **RabbitMQ** | test | test |

---

## Critical Implementation Rules

### Language-Specific Rules (Go)

1. **Import 顺序**: 标准库 → 外部包 → 内部包
2. **错误处理**: 使用 `errorx.NewDefaultError()` + `logc.Errorf()`
3. **字符串处理**: 使用 `strings.TrimSpace()` 处理输入
4. **结构体注释**: 包含 Author 和 Date
5. **上下文传递**: Logic 层接收 `context.Context`

### Framework-Specific Rules (go-zero)

1. **三层架构**: Handler → Logic → Service
2. **Handler 模式**:
   ```go
   func XxxHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
       return func(w http.ResponseWriter, r *http.Request) {
           var req types.XxxReq
           if err := httpx.Parse(r, &req); err != nil { ... }
           l := xxx.NewXxxLogic(r.Context(), svcCtx)
           resp, err := l.Xxx(&req)
           // 统一返回
       }
   }
   ```
3. **Logic 结构体**:
   ```go
   type XxxLogic struct { logx.Logger; ctx context.Context; svcCtx *svc.ServiceContext }
   ```
4. **类型定义**: `form:"field,default=1"` / `json:"field,optional"`
5. **服务访问**: 通过 `svcCtx.XxxService` 调用 RPC

### Testing Rules

- 使用 Playwright 进行 E2E 测试 (web-admin)
- Go 层暂无单元测试，建议为 Logic 层添加
- API 测试使用 httptest
- Flutter: 使用 flutter_test 进行单元测试

### Framework-Specific Rules (Flutter)

1. **状态管理**: 使用 Provider ^6.1.5+1
2. **API 调用**: 使用 Dio ^5.9.0，配置拦截器处理认证和错误
3. **目录结构**:
   ```
   lib/
   ├── models/          # 数据模型
   ├── services/        # API 服务层
   ├── providers/       # Provider 状态管理
   ├── screens/         # 页面
   ├── widgets/         # 通用组件
   └── utils/           # 工具类
   ```
4. **Widget 构建**: 使用 const 构造函数优化性能
5. **图片加载**: 使用 cached_network_image 进行缓存
6. **列表刷新**: 使用 easy_refresh 实现下拉刷新
7. **状态更新**: 使用 ChangeNotifier + Consumer 或 context.watch

### Code Quality & Style Rules

- **Go**: gofmt/goimports 格式化
- **React**: ESLint (@umijs/fabric) + Prettier
- **命名**: snake_case (文件) / PascalCase (结构体)
- **目录结构**: 必须遵循 api/admin/internal/* 模式

### Development Workflow Rules

1. **代码生成**: `generate-code` + `goctl`
2. **构建**: `make build` / `make deps`
3. **服务模块**: admin-api, front-api, sys/ums/oms/pms/cms rpc

### Critical Don't-Miss Rules

**禁止事项:**
- ❌ 手动修改生成代码 (rpc/*/gen/, types/types.go)
- ❌ 在 API 层写业务逻辑
- ❌ 绕过 go-zero 框架约定
- ❌ 直接操作数据库 (必须用 GORM)

**边界情况:**
- ✅ 空值检查: `strings.TrimSpace()`
- ✅ 分页默认值: `default=1`, `default=20`
- ✅ 错误日志: 包含请求参数

**安全规则:**
- 密码加密: 使用项目约定方式
- JWT: 使用 golang-jwt 库
- SQL注入: GORM 参数化查询

---

## Usage Guidelines

**For AI Agents:**
- Read this file before implementing any code
- Follow ALL rules exactly as documented
- When in doubt, prefer the more restrictive option
- Update this file if new patterns emerge

**For Humans:**
- Keep this file lean and focused on agent needs
- Update when technology stack changes
- Review quarterly for outdated rules
- Remove rules that become obvious over time

---

_Last Updated: 2026-03-14_
