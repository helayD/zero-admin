# 禁止使用 goctl 自动生成代码

## 绝对禁止

**严禁**在任何情况下使用 `goctl` 工具（包括但不限于以下命令）生成代码：

- `goctl go zero-admin/rpc/...` 
- `goctl -m` / `goctl --module`
- `goctl new`
- `goctl api`
- `goctl rpc`
- `goctl kube`
- `goctl env`
- 任何 `goctl` 子命令

**原因**：`goctl` 生成的代码与本项目代码风格存在严重冲突，包括但不限于：

| 维度 | goctl 生成 | 本项目规范 |
|------|-----------|-----------|
| 文件命名 | `addlogic.go` | `add_order_logic.go`（下划线分隔） |
| 包名 | `logic` | `{service}logic`（带服务名前缀） |
| 结构体 | `AddLogic` | `AddOrderLogic`（PascalCase） |
| 方法注释 | `// Add 添加` / 无注释 | `// AddOrder 添加订单(app)`（方法名+中文+用途标注） |
| 错误处理 | `return nil, err` | `logc.Errorf` + `errors.New` / `fmt.Errorf` |
| 日志 | 无或 `logx` 少量 | `logc.Errorf/Infof/Debugf`（带上下文） |
| 返回值 | 空响应骨架 | 完整业务逻辑实现 |
| Imports | 仅基础包 | 完整（`logc`, `gorm`, `query`, `common` 等） |

## 替代方案

当需要创建新的 Logic 文件时，必须**手动编写**，参考以下规范：

### 文件命名

```
{operation}_{entity}_logic.go   // 下划线分隔全小写
```

示例：
- `add_order_logic.go`
- `query_order_detail_logic.go`
- `update_order_status_logic.go`

### 包名

```
{service}logic   // 服务名 + logic，后缀
```

示例：
- `orderservicelogic`
- `cartitemservicelogic`

### 结构体 & 构造函数

```go
type AddOrderLogic struct {
    ctx    context.Context
    svcCtx *svc.ServiceContext
    logx.Logger
}

func NewAddOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddOrderLogic {
    return &AddOrderLogic{
        ctx:    ctx,
        svcCtx: svcCtx,
        Logger: logx.WithContext(ctx),
    }
}
```

### 方法注释格式

```go
// {MethodName} {中文描述}({调用方})
// Author: {Author}
// Date: {Date} {Time}
func (l *AddOrderLogic) AddOrder(in *omsclient.AddOrderReq) (*omsclient.AddOrderResp, error) {
    // 业务逻辑...
}
```

### Imports 规范（按顺序）

```go
import (
    "context"
    "errors"
    "time"

    "github.com/feihua/zero-admin/rpc/oms/gen/query"
    "github.com/feihua/zero-admin/rpc/oms/internal/logic/common"
    "github.com/feihua/zero-admin/rpc/oms/internal/svc"
    "github.com/feihua/zero-admin/rpc/oms/omsclient"
    "github.com/zeromicro/go-zero/core/logc"
    "gorm.io/gorm"

    "github.com/zeromicro/go-zero/core/logx"
)
```

## 违规处理

- 如发现 `goctl` 生成的骨架文件（如包含 `// todo: add your logic here`），**立即删除**
- 禁止基于 `goctl` 骨架进行"填充式"开发——必须从零编写完整业务逻辑
