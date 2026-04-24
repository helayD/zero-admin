# AGENTS.md

请使用中文回答。

---

## 项目概述

**Zero-Admin** 是基于 go-zero 框架的企业级电商系统，采用微服务架构。

- **技术栈**: Go 1.25 + go-zero 1.9.3 + GORM + MySQL + gRPC
- **模块**: sys(系统管理), ums(会员), pms(商品), oms(订单), sms(营销), cms(内容), search(搜索)

---

## 构建与测试命令

```bash
# 安装依赖 / 构建所有服务 / 清理
make deps && make build && make clean

# 格式化代码
make format

# 生成 Model 代码 (GORM gen) - 注意：RPC 代码需手动编写
make model

# 运行测试
make test

# 启动/停止/重启所有服务
make start && make stop && make restart
```

### 单个测试命令

```bash
# 运行单个测试文件
go test ./rpc/sys/internal/logic/userservice/... -run TestUpdateUserRoleList -v

# 运行单个测试函数
go test ./api/admin/internal/logic/pms/product_spu/... -run TestAddProductSpuPassesScopeAndNestedDetailIDsToRPC -v

# 运行带覆盖率
go test ./rpc/sys/... -coverprofile=coverage.out -covermode=atomic

# 运行所有测试
go test ./... -v -count=1
```

### 静态分析

```bash
make lint      # golangci-lint 检查
make lint-fix  # 自动修复
```

### 强制性约束

```text
- 禁止使用任何自动生成命令改写项目代码，包括但不限于：goctl、protoc、make gen。
- 禁止通过生成链覆盖或批量改写 api/*/internal/types、rpc/*/*.proto、rpc/*/smsclient 等文件。
- 需要修复问题时，必须优先采用手工最小改动；不要把“修改 .api/.proto 后再生成”当作默认方案。
- 如确需生成类操作，必须先得到用户明确许可；未获许可前一律视为禁止。
```

### C 端监管约束（数字藏品）

```text
Flutter（C 端面向普通用户的 UI）严禁展示任何区块链底层信息。
禁止在 C 端 UI 上展示以下字段：
  - chainType（链类型：蚂蚁链 / FISCO BCOS 等）
  - chainStatus / chainStatusText（链上状态）
  - chainTxId（链上交易 ID）
  - tokenId / tokenIdMasked（token 编号）
  - lastReceiptJson / lastReceiptSummary（链上回执）
  - 任何包含"蚂蚁链""AntChain""FISCO""区块链""链上""上链""链路"等关键词的文案

C 端只允许展示：卡片编号、模板名、活动名、获取时间、发放状态（mintStatusText）、
展示状态、合规状态、合规提示摘要。
如服务端返回的 mintStatusText 等允许字段包含底层链路文案，C 端 UI 必须转成发放/到账/处理结果口径。

B 端（Web Admin）不受此约束，运维人员可以看到完整链信息。
原因：数字藏品业务监管要求，C 端不得暴露底层区块链实现细节。
```

---

## 代码风格指南

### 命名规范

```go
// 结构体/函数/方法: PascalCase
type UpdateUserRoleListLogic struct {
    ctx    context.Context
    svcCtx *svc.ServiceContext
    Logger logx.Logger
}
func NewUpdateUserRoleListLogic(...) *UpdateUserRoleListLogic

// 变量/参数: camelCase
userId := in.UserId

// 常量: PascalCase 或 MAX_RETRY_COUNT
const UserActivationDisabled int32 = 0
```

### 导入组织 (顺序: 标准库 → 第三方库 → 本地包, 按字母排序)

```go
import (
    "context"
    "errors"
    "fmt"

    "github.com/zeromicro/go-zero/core/logc"
    "gorm.io/gorm"

    "github.com/feihua/zero-admin/rpc/sys/gen/model"
    "github.com/feihua/zero-admin/rpc/sys/gen/query"
    logiccommon "github.com/feihua/zero-admin/rpc/sys/internal/logic/common"
)
```

### 错误处理

```go
// 使用 errors.New() 或 fmt.Errorf()
return nil, errors.New("不允许操作超级管理员用户")
return nil, fmt.Errorf("角色[%s]主体范围配置非法: %w", row.RoleName, scopeErr)

// 检查错误类型
if errors.Is(err, gorm.ErrRecordNotFound) {
    return nil, errors.New("指定的租户不存在")
}

// 事务回滚 (自动)
err = query.Q.Transaction(func(tx *query.Query) error {
    if _, err = q.WithContext(l.ctx).Where(...).Delete(); err != nil {
        return err  // 自动回滚
    }
    return nil
})
```

### 日志记录

```go
// 使用 logc (带上下文)
logc.Errorf(l.ctx, "删除用户与角色的关联失败,参数:%+v,异常:%s", in, err.Error())

// 使用 logx.WithContext
Logger: logx.WithContext(ctx),
```

### API Handler 模式

```go
// api/*/internal/handler/{module}/{operation}handler.go
package user

import (
    "net/http"

    "github.com/feihua/zero-admin/api/admin/internal/logic/sys/user"
    "github.com/feihua/zero-admin/api/admin/internal/svc"
    "github.com/feihua/zero-admin/api/admin/internal/types"
    "github.com/zeromicro/go-zero/rest/httpx"
)

func AddUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req types.AddUserReq
        if err := httpx.Parse(r, &req); err != nil {
            httpx.ErrorCtx(r.Context(), w, err)
            return
        }

        l := user.NewAddUserLogic(r.Context(), svcCtx)
        resp, err := l.AddUser(&req)
        if err != nil {
            httpx.ErrorCtx(r.Context(), w, err)
        } else {
            httpx.OkJsonCtx(r.Context(), w, resp)
        }
    }
}
```

### API Middleware 模式

```go
// api/*/internal/middleware/{name}middleware.go
package middleware

type AddLogMiddleware struct {
    Sys operatelogservice.OperateLogService
}

func NewAddLogMiddleware(Sys operatelogservice.OperateLogService) *AddLogMiddleware {
    return &AddLogMiddleware{Sys: Sys}
}

func (m *AddLogMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 前置处理 (记录日志、鉴权等)
        startTime := time.Now()
        
        // 调用下一个处理器
        next(w, r)
        
        // 后置处理 (记录响应耗时等)
        duration := time.Since(startTime)
    }
}
```

### gRPC Logic 结构

```go
type UpdateUserRoleListLogic struct {
    ctx    context.Context
    svcCtx *svc.ServiceContext
    logx.Logger
}

func NewUpdateUserRoleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserRoleListLogic {
    return &UpdateUserRoleListLogic{
        ctx:    ctx,
        svcCtx: svcCtx,
        Logger: logx.WithContext(ctx),
    }
}

func (l *UpdateUserRoleListLogic) UpdateUserRoleList(in *sysclient.UpdateUserRoleListReq) (*sysclient.UpdateUserRoleListResp, error) {
    // 业务逻辑
}
```

### Config 结构

```go
// RPC Config - rpc/*/internal/config/config.go
package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
    zrpc.RpcServerConf

    Mysql struct {
        Datasource string
    }

    JWT struct {
        AccessSecret string
        AccessExpire int64
    }
}

// API Config - api/*/internal/config/config.go
package config

import (
    "github.com/zeromicro/go-zero/rest"
    "github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
    rest.RestConf

    SysRpc zrpc.RpcClientConf
    UmsRpc zrpc.RpcClientConf
    PmsRpc zrpc.RpcClientConf

    Auth struct {
        AccessSecret string
        AccessExpire int64
        ExcludeUrl   string
    }

    Redis struct {
        Address string
        Pass    string
    }
}
```

### 数据库模型 (GORM)

```go
type UserScopeBinding struct {
    UserID     int64  `gorm:"column:user_id"`
    ScopeType  string `gorm:"column:scope_type"`
    PlatformID int64  `gorm:"column:platform_id"`
    TenantID   int64  `gorm:"column:tenant_id"`
}

func (*UserScopeBinding) TableName() string {
    return "sys_user_scope"
}
```

### 注释规范

```go
// UpdateUserRoleList 分配用户角色
/*
Author: LiuFeiHua
Date: 2024/5/23 17:38
*/
func (l *UpdateUserRoleListLogic) UpdateUserRoleList(...) {
    // 1.判断是否为超级管理员
    // 2.删除用户与角色的关联
    // 3.添加用户与角色的关联
}
```

### 测试模式

```go
// Mock gRPC 客户端
type mockProductSpuService struct {
    productspuservice.ProductSpuService
    addFn func(context.Context, *pmsclient.ProductSpuReq, ...grpc.CallOption) (*pmsclient.ProductSpuResp, error)
}

// 测试函数命名: Test{Method}{Scenario}
func TestAddProductSpuPassesScopeAndNestedDetailIDsToRPC(t *testing.T) {
    ctx := newAdminProductSpuContext("merchant", 1, 88, 3001)
    logic := NewAddProductSpuLogic(ctx, &svc.ServiceContext{
        ProductSpuService: &mockProductSpuService{...},
    })
    _, err := logic.AddProductSpu(&types.AddProductSpuReq{...})
    if err != nil {
        t.Fatalf("AddProductSpu returned error: %v", err)
    }
}
```

---

## 目录结构

```
rpc/
  {service}/
    {service}.go              # main 入口
    etc/{service}.yaml        # 配置文件
    internal/
      config/config.go        # 配置结构体
      logic/{service}service/ # 业务逻辑 (*logic.go)
      svc/servicecontext.go   # 依赖注入
    gen/model/                # GORM 模型
    gen/query/                # GORM Gen 查询
    proto/                    # proto 文件 (手动编写)

api/
  {api}/
    {api}.go                  # main 入口
    internal/
      config/config.go        # 配置结构体
      handler/                # HTTP Handler (手动编写)
        {module}/
          *handler.go
      logic/                  # 业务逻辑
      types/                  # 请求/响应类型
      middleware/             # 中间件
        *middleware.go
```

---

## golangci-lint 配置

启用的 linter: `errcheck`, `gosimple`, `govet`, `ineffassign`, `staticcheck`, `unused`, `stylecheck`, `gofmt`, `goimports`, `misspell`, `unconvert`, `unparam`, `whitespace`, `bodyclose`, `dupl`, `goconst`, `gocritic`, `gochecknoinits`, `prealloc`

禁用: `lll` (行长度), `gochecknoglobals`, `wsl` (空行规则)

---

## 常见模式

### Scope 治理 (多租户/多商户)

```go
if err = logiccommon.ValidateRoleIDsInScope(l.ctx, l.svcCtx.DB, currentScope, in.RoleIds); err != nil {
    return nil, err
}
```

### Context 传递

```go
ctx := context.WithValue(ctx, "userId", json.Number("1001"))
count, err := query.SysUser.WithContext(l.ctx).Where(query.SysUser.ID.Eq(in.UserId)).Count()
```
