# AGENTS.md

请使用中文回答。

---

## 项目概述

**Zero-Admin** 是一套基于 go-zero 框架实现的企业级电商系统，采用微服务架构。

### 技术栈
- **Go 1.25** + go-zero 1.9.3
- **GORM** + MySQL (ORM持久层)
- **gRPC** (服务间通信)
- **MongoDB** / **Redis** / **Elasticsearch**

### 微服务模块
| 模块 | 描述 |
|------|------|
| sys | 系统管理（用户、角色、菜单、部门、租户） |
| ums | 会员管理 |
| pms | 商品管理 |
| oms | 订单管理 |
| sms | 营销管理 |
| cms | 内容管理 |
| search | 搜索服务 |

---

## 构建与测试命令

### Makefile 常用命令

```bash
# 安装依赖
make deps

# 构建所有服务
make build

# 清理构建产物
make clean

# 格式化代码 (goctl)
make format

# 生成代码 (API + RPC)
make gen

# 生成 Model 代码 (GORM gen)
make model

# 构建 Docker 镜像
make image

# 运行测试
make test
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
go test ./...

# 查看测试输出
go test ./... -v -count=1
```

### 服务启动

```bash
# 启动所有服务
make start

# 停止所有服务
make stop

# 重启所有服务
make restart
```

---

## 代码风格指南

### 1. 命名规范

```go
// 结构体：PascalCase
type UpdateUserRoleListLogic struct {
    ctx    context.Context
    svcCtx *svc.ServiceContext
    Logger logx.Logger
}

// 函数/方法：PascalCase
func NewUpdateUserRoleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserRoleListLogic

// 变量/参数：camelCase
userId := in.UserId
roleIds := in.RoleIds

// 常量：PascalCase 或 全大写+下划线
const UserActivationDisabled int32 = 0
const MAX_RETRY_COUNT = 3
```

### 2. 导入组织

```go
import (
    "context"
    "errors"
    "fmt"
    "strings"
    "time"

    // 第三方库
    "github.com/zeromicro/go-zero/core/logc"
    "gorm.io/gorm"
    "gorm.io/gorm/clause"

    // 本地包
    "github.com/feihua/zero-admin/rpc/sys/gen/model"
    "github.com/feihua/zero-admin/rpc/sys/gen/query"
    logiccommon "github.com/feihua/zero-admin/rpc/sys/internal/logic/common"
    "github.com/feihua/zero-admin/rpc/sys/internal/svc"
    "github.com/feihua/zero-admin/rpc/sys/sysclient"
)
```

**顺序**：标准库 → 第三方库 → 本地包（按字母排序）

### 3. 错误处理

```go
// 使用 errors.New() 创建错误
return nil, errors.New("不允许操作超级管理员用户")

// 包装错误
return nil, fmt.Errorf("角色[%s]主体范围配置非法: %w", row.RoleName, scopeErr)

// 检查错误类型
if errors.Is(err, gorm.ErrRecordNotFound) {
    return nil, errors.New("指定的租户不存在")
}

// 事务回滚
err = query.Q.Transaction(func(tx *query.Query) error {
    // 操作
    if err != nil {
        return err  // 自动回滚
    }
    return nil
})
```

### 4. 日志记录

```go
// 使用 logc (带上下文)
logc.Errorf(l.ctx, "删除用户与角色的关联失败,参数:%+v,异常:%s", in, err.Error())

// 使用 logx.WithContext
Logger: logx.WithContext(ctx),

// Info 日志
logx.Infof("mysql已连接")
```

### 5. 数据库模型 (GORM)

```go
// 使用 gorm tag 映射列名
type UserScopeBinding struct {
    UserID           int64  `gorm:"column:user_id"`
    ScopeType        string `gorm:"column:scope_type"`
    PlatformID       int64  `gorm:"column:platform_id"`
    TenantID         int64  `gorm:"column:tenant_id"`
}

// 实现 TableName 方法
func (*UserScopeBinding) TableName() string {
    return "sys_user_scope"
}
```

### 6. gRPC Service 结构

```go
// Logic 结构体
type UpdateUserRoleListLogic struct {
    ctx    context.Context
    svcCtx *svc.ServiceContext
    logx.Logger
}

// 构造函数
func NewUpdateUserRoleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserRoleListLogic {
    return &UpdateUserRoleListLogic{
        ctx:    ctx,
        svcCtx: svcCtx,
        Logger: logx.WithContext(ctx),
    }
}

// 业务方法
func (l *UpdateUserRoleListLogic) UpdateUserRoleList(in *sysclient.UpdateUserRoleListReq) (*sysclient.UpdateUserRoleListResp, error) {
    // 业务逻辑
}
```

### 7. 注释规范

```go
// UpdateUserRoleList 分配用户角色
/*
Author: LiuFeiHua
Date: 2024/5/23 17:38
*/
func (l *UpdateUserRoleListLogic) UpdateUserRoleList(in *sysclient.UpdateUserRoleListReq) (*sysclient.UpdateUserRoleListResp, error) {
    // 1.判断是否为超级管理员
    // 2.删除用户与角色的关联
    // 3.添加用户与角色的关联
}
```

### 8. 测试模式

```go
// Mock gRPC 客户端
type mockProductSpuService struct {
    productspuservice.ProductSpuService
    addFn func(context.Context, *pmsclient.ProductSpuReq, ...grpc.CallOption) (*pmsclient.ProductSpuResp, error)
}

func (m *mockProductSpuService) AddProductSpu(ctx context.Context, in *pmsclient.ProductSpuReq, opts ...grpc.CallOption) (*pmsclient.ProductSpuResp, error) {
    return m.addFn(ctx, in, opts...)
}

// 测试函数命名: Test{Method}{Scenario}
func TestAddProductSpuPassesScopeAndNestedDetailIDsToRPC(t *testing.T) {
    // given
    ctx := newAdminProductSpuContext("merchant", 1, 88, 3001)
    
    // when
    logic := NewAddProductSpuLogic(ctx, &svc.ServiceContext{
        ProductSpuService: &mockProductSpuService{...},
    })
    _, err := logic.AddProductSpu(&types.AddProductSpuReq{...})
    
    // then
    if err != nil {
        t.Fatalf("AddProductSpu returned error: %v", err)
    }
}
```

### 9. Context 传递

```go
// 在请求中传递用户信息
ctx := context.WithValue(ctx, "userId", json.Number("1001"))
ctx = context.WithValue(ctx, "userName", "tester")
ctx = context.WithValue(ctx, "scopeType", scopeType)

// 使用 WithContext 执行数据库操作
count, err := query.SysUser.WithContext(l.ctx).Where(query.SysUser.ID.Eq(in.UserId)).Count()
```

### 10. 目录结构

```
rpc/
  {service}/
    {service}.go              # main 入口
    etc/{service}.yaml       # 配置文件
    internal/
      config/config.go       # 配置结构体
      logic/                 # 业务逻辑
        {service}service/
          *logic.go
      svc/servicecontext.go # 依赖注入
    gen/
      model/                 # GORM 模型
      query/                 # GORM Gen 查询
    proto/                   # proto 文件
api/
  {api}/
    {api}.go                 # main 入口
    internal/
      handler/               # HTTP handler
      logic/                 # 业务逻辑
      types/                 # 请求/响应类型
      middleware/            # 中间件
pkg/
  {shared_lib}/             # 共享包
```

---

## 代码生成 (goctl)

```bash
# 生成 API 代码
goctl api go -api ./api/admin/doc/api/admin.api -dir ./api/admin/

# 生成 RPC 代码
goctl rpc protoc rpc/sys/sys.proto \
  --go_out=./rpc/sys/ \
  --go-grpc_out=./rpc/sys/ \
  --zrpc_out=./rpc/sys/ -m

# 格式化 API 目录
goctl api format --dir api/admin/doc/api
```

---

## 静态分析与格式化

项目已配置 **golangci-lint**（见 `.golangci.yml`）。使用以下命令：

```bash
# 安装依赖后运行
make deps

# 运行 golangci-lint 检查
make lint

# 自动修复可修复的问题
make lint-fix

# 基础检查（无 golangci-lint 时）
go fmt ./...
goimports -l -w .
go vet ./...
```

### golangci-lint 配置

启用以下 linter：
- `errcheck` - 检查未处理的错误
- `govet` / `staticcheck` - 可疑代码检查
- `unused` / `ineffassign` - 未使用代码检测
- `stylecheck` - 代码风格检查
- `goimports` - 导入整理
- `misspell` - 拼写错误检查
- `bodyclose` - HTTP 响应体未关闭检测
- `dupl` - 代码重复检测

> 注：如需跳过某些文件，可添加 `//nolint:lintername` 注释

---

## 常见模式

### ServiceContext 依赖注入

```go
type ServiceContext struct {
    Config   config.Config
    DB       *gorm.DB
    Redis    *redis.Redis
    RedisKey string
}

func NewServiceContext(c config.Config) *ServiceContext {
    db, err := gorm.Open(mysql.Open(c.Mysql.Datasource), &gorm.Config{...})
    if err != nil {
        panic(err)
    }
    return &ServiceContext{Config: c, DB: db}
}
```

### 事务处理

```go
err = query.Q.Transaction(func(tx *query.Query) error {
    // 执行多个数据库操作
    if _, err = q.WithContext(l.ctx).Where(...).Delete(); err != nil {
        return err
    }
    if err = q.WithContext(l.ctx).CreateInBatches(...); err != nil {
        return err
    }
    return nil
})
```

### Scope 治理模式

项目使用统一的 `scope.GovernanceScope` 进行多租户、多商户治理：

```go
// 验证角色在当前 scope 内
if err = logiccommon.ValidateRoleIDsInScope(l.ctx, l.svcCtx.DB, currentScope, in.RoleIds); err != nil {
    return nil, err
}
```

---

## 后续 BMAD 相关规则 (保持不变)

## Agent 任务感知与 codex-autopilot 绑定

- 当前 agent 只负责本仓库：`/Users/helay/Documents/GitHub/zero-admin`
- 对应 codex-autopilot tmux 窗口：`zero-admin`
- 当用户问“我有什么任务没完成 / 现在在做什么 / 还有什么待办 / 当前任务是什么”时，**不要只根据聊天记忆回答**，也不要优先说“记忆不可用”。
- 必须优先检查与本仓库相关的真实状态来源：
  1. `git status --short`
  2. `git log --oneline -10`
  3. `~/.autopilot/state/zero-admin.json`
  4. `~/.autopilot/logs/watchdog.log` 里与 `zero-admin` 相关的最近记录
  5. 必要时查看 tmux 窗口 `autopilot:zero-admin` 的 pane 输出
- 回答任务类问题时，优先输出：
  - 已完成
  - 进行中
  - 未完成/待处理
  - 卡点/风险
  - 下一步建议
- 如果 autopilot 状态读不到，要明确说明是“autopilot 状态不可读”或“tmux 状态不可读”，不要笼统说“没有任务”或“只有当前群聊上下文可见”。

## Codex 会话质量守则

- 用户说“codex 清理上下文”时：在当前 Codex 会话里执行 `/compact`。
- 用户说“codex 继续 / 恢复会话”时：优先执行 `codex resume --last`。
- 用户说“codex 新会话”时：不要 `resume --last`，直接启动一个全新的 Codex 会话：`codex --dangerously-bypass-approvals-and-sandbox`。
- 对本仓库来说，**避免 Codex 上下文污染优先于少开新会话**。
- 只要出现以下迹象，就优先考虑 `/compact` 或直接新开会话：
  - 对当前需求、分支、任务目标开始混淆
  - 反复引用过时计划或旧实现
  - 把别的项目/别的模块上下文带进来
  - 连续多轮修改后推理质量明显下降
- 一旦怀疑上下文被污染，宁可保守地新开会话，也不要硬接着做。
- 在上下文不干净时：
  - 不做架构决策
  - 不做大规模重构
  - 不删改核心逻辑
- Compact 或新会话之后，先重新阅读仓库内与当前任务直接相关的约束文件，例如：`AGENTS.md`、`CONVENTIONS.md`、`prd-todo.md`、`task_plan.md`、`findings.md`、`progress.md`（若存在），再继续编码。

## Agent 内置 BMAD 调度与飞书进展回传规则

- `zero-admin` 是一个**已有的项目 agent**，不是临时脚本包装器；后续 BMAD 自动推进能力必须**集成到这个 agent 的既有规则体系中**，而不是另起一套平行 agent。
- 该 agent 的职责分两层：
  1. **对外层（Feishu 可见）**：读取真实状态、判断当前最优下一步、汇报进展、汇报阻塞、请求人工确认。
  2. **执行层（codex-autopilot）**：通过 tmux / watchdog / Codex CLI 实际执行 create-story、dev-story、code-review 等 BMAD 节点。
- 该 agent 需要把以下内容视为默认工作模式：
  - 项目推进按 **BMAD 流程**执行；
  - **每完成一个 BMAD 节点，就切换到一个新会话语义**；
  - 每次调度都必须先读取真实状态，再判断“当前最优解”；
  - codex-autopilot 是执行器，不是流程定义者。
- 该 agent 在被周期性触发、主动轮询、或用户要求“继续推进 / 自动化运行 / 看现在做到哪了”时，必须优先读取以下真实状态源：
  1. `_opcos/implementation-artifacts/sprint-status.yaml`
  2. 当前 story 文件
  3. `git status --short`
  4. `git log --oneline -10`
  5. `~/.autopilot/state/zero-admin.json`
  6. `~/.autopilot/logs/watchdog.log`
  7. 必要时 tmux 窗口 `autopilot:zero-admin` pane 输出
- 该 agent 每次只能输出并派发**一个明确的 BMAD 节点动作**。
- **节点规则可以固定，但 story key 严禁硬编码。**
- 所有带 story key 的动作都必须由 agent 在执行当下根据真实状态**动态推导**，而不是把 `2-2`、`2-3` 一类编号写死在规则、脚本或提示词里。
- 允许输出的动作形态例如：
  - `create-story`
  - `refine-story:<derived_story_key>`
  - `dev-story:<derived_story_key>`
  - `analyze-current-state:<derived_story_key>`
  - `code-review:<derived_story_key>`
  - `fix-review-findings:<derived_story_key>`
  - `blocked:<reason>`
- `<derived_story_key>` 的推导必须至少结合：
  1. `sprint-status.yaml` 中按顺序出现的 `in-progress` / `ready-for-dev` story
  2. 对应 story 文件状态
  3. `git status --short` 中是否已存在该 story 的实现痕迹
  4. autopilot / watchdog 当前是否存在阻塞、空闲、已派发但未收口状态
- 如果代码状态与 story 工件状态不一致，优先输出 `analyze-current-state:<derived_story_key>`，不得直接盲派发 `dev-story` 或 `code-review`。
- 禁止把模糊任务直接交给执行层，例如：
  - “继续做”
  - “你看着办”
  - “顺着往下推进”
- 该 agent 在 Feishu 中必须输出**阶段性可见进展**，但避免刷屏。默认只在以下时机发送：
  1. 调度判断出新的最优动作
  2. 已派发执行任务
  3. 关键里程碑完成
  4. 出现明确阻塞
  5. 一个 BMAD 节点完成
- Feishu 进展汇报必须尽量短、结构固定、便于群内快速扫读。默认结构：
  - 当前节点：<node>
  - 当前状态：<1~3 行总结>
  - 最优下一步：<action>
  - 是否已派发：已派发 / 未派发
  - 阻塞/风险：<没有则写“无硬阻塞”>
- 若只是定时检查且无明显变化，必须使用极简模板：
  - 当前节点：<node>
  - 当前状态：无状态变化
  - 最优下一步：维持当前节点
  - 是否已派发：否
  - 阻塞/风险：无硬阻塞
- 对群内消息做节流：
  - 不要每次轮询都发长文
  - 无状态变化时必须极简
  - 只有在“新动作已派发 / 阻塞变化 / 节点完成”时才允许发较完整说明
- **节点完成回传是必需的**。create-story / refine-story / dev-story / code-review / fix-review-findings / analyze-current-state 任一节点完成后，必须再次在 Feishu 中回传一次“完成结果”，默认模板：
  - 节点完成：<node>
  - 结果：<完成了什么>
  - 结论：<可继续 / 需人工确认 / blocked>
  - 下一步建议：<next-action>
  - 关键风险：<没有则写“无新的硬阻塞”>
- 默认允许自动执行的 BMAD 节点：
  - create-story
  - story refine
  - dev-story
  - analyze-current-state
  - code-review
  - review 修复
  - 测试 / 校验 / 文档状态收口
- 默认必须先请求人工确认的动作：
  - 跨 epic 改范围
  - 大规模 schema 重构
  - 生产发布
  - 删除大量代码 / 数据
  - 改 BMAD 主计划 / sprint 顺序
- 如果代码状态与 BMAD 工件状态不一致（例如：代码已改很多，但 story 仍是 `ready-for-dev`），该 agent **不得盲目继续开发**，必须优先触发 `analyze-current-state:<story>`，先做阶段性验收分析，再决定下一步。
- 如果执行层卡在登录页、权限页、shell recovery、测试红线或关键工件缺失，该 agent 必须将其判定为 **blocked**，并在 Feishu 中直接汇报阻塞点，而不是继续派发下一个节点。
- **“仍在执行中”不能只靠 story=in-progress 或工作区有未提交改动来判断。** 只有同时满足下列信号中的至少一项，才允许在群里使用“继续推进 / 仍在执行中 / 正在补齐”这类表述：
  1. tmux pane 最近有新增输出，且内容显示正在继续处理当前 story
  2. 最近出现新的测试执行、文件变更或提交推进信号
  3. Codex 没有停在普通 prompt / 等待输入态，而是处于明确的工作态
- 如果 story 未完成，但当前只看到“tmux 停在 prompt、watchdog 连续 idle、没有新的输出/测试/提交”，则必须降级判断为：
  - `当前节点未完成，但执行层处于等待/半静止状态`
  - 此时不得继续宣称“自动推进中”，也不得重复派发同一动作，除非确认需要续派发。
- 若 `watchdog/state` 与 tmux pane 观测不一致，优先在群里明确写出“状态源不一致”，避免把不确定状态包装成确定结论。
- **自动化不仅要会派发节点，也要会治理会话。** 当检测到状态源不一致、上下文污染、半静止等待或交互阻断时，必须优先执行会话治理动作，而不是继续盲派发下一节点。
- 默认会话治理动作分四类：
  1. `analyze-current-state:<derived_story_key>`：用于状态不一致、代码与工件不同步、上下文不可信时，先做阶段性验收分析。
  2. `compact-current-session`：仅当当前 story 仍正确、只是上下文过长或轻度污染时使用。
  3. `start-clean-session-for:<derived_story_key>`：当旧会话已明显污染、或准备切换到下一个 BMAD 节点时，必须启动全新 Codex 会话，不得继续沿用旧会话。
  4. `hold-and-report`：出现硬阻塞、关键工件缺失、登录/权限卡死时，不派发，只汇报。
- 自动化判定顺序：
  1. 先判定是否 `blocked`
  2. 再判定是否“状态源不一致 / 上下文污染”
  3. 若需要，先 `analyze-current-state`
  4. 分析后若当前 story 仍可继续：轻污染则 `compact-current-session`，重污染则 `start-clean-session-for:<same_story>`
  5. 只有在当前节点已有明确完成证据时，才允许切到下一节点，并且必须使用 `start-clean-session-for:<next_story_or_next_node>`
- **严禁在旧会话里直接跨 story 推进。** 例如：不能在 2.2 的旧会话里直接续推 2.3；若要推进 2.3，必须先判断 2.2 是否完成或应中止，然后以 `start-clean-session-for:2-3` 形式开启全新会话。
- 若判定为“当前节点未完成，但执行层处于等待/半静止状态”，默认最优动作不是继续硬派发，而是先 `analyze-current-state`，确认是否该续派发、compact、还是新会话。

## 绑定项目群回复规则

- 在已绑定的 Feishu 项目群里，**不需要 @ 才回复**。
- 只要消息来自当前绑定项目群，且内容属于以下类型，就应直接回复，不要过度克制：
  - 明确提问
  - 任务推进 / 继续执行
  - 状态询问 / 进度追问
  - 要求总结、汇报、给方案、做判断
  - 与当前仓库、当前任务、当前 autopilot/codex 状态直接相关的指令
- 不要因为“群聊礼貌”或“避免打扰”而默认沉默；在这些项目群里，及时反馈比克制更重要。
- 如果信息不足，也要直接回一个简短但有用的澄清或当前状态，而不是不回。
- 只有在明显与本项目无关、别人彼此闲聊、或你确实无法提供价值时，才保持安静。
