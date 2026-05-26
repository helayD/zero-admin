# Story 10-12: 抽卡消耗货币统一 —— 按 consume_type 支持积分或抽卡次数

## 背景与问题

当前系统存在两套独立货币体系：

| 字段 | 名称 | 来源 | 现状 |
|------|------|------|------|
| `ums_member_info.lottery_times` | 抽卡次数 | 每日登录 +3 | 抽卡活动硬编码消耗 |
| `ums_member_info.points` | 积分 | 每日登录 +10 | 商城/兑换使用，抽卡无法消耗 |

`sms_draw_activity.consume_type` 字段虽然已存在（值为 `lottery_times`），但 `participate_draw_logic.go` 和 `buildEligibility` 函数**完全忽略该字段**，硬编码操作 `lottery_times`。

业务需求：活动配置 `consume_type=points` 时应扣减积分参与转盘，统一为单一货币体系。

---

## 验收条件

- **AC1**：`consume_type=lottery_times` 行为与现在一致（无回归）
- **AC2**：`consume_type=points` 时，参与抽卡扣减 `ums_member_info.points`，资格检查也校验 `points` 余额
- **AC3**：资格不足时错误文案与货币类型对应（"剩余积分不足" / "剩余抽奖次数不足"）
- **AC4**：`sms_draw_activity` 管理后台编辑页可选择 `consume_type`（现有 web-admin 表单）
- **AC5**：`ums_member_lottery_grant_log` 的 `grant_type=daily_login_points` 正确写入 `points` 字段（现有逻辑验证，不新增）

---

## 技术方案

### 改动文件（共 3 处）

#### 1. `rpc/sms/internal/logic/drawparticipationservice/draw_participation_helper.go`

**a. 新增常量**
```go
drawConsumeTypePoints = "points"
```

**b. `drawMemberInfoSnapshot` 加 `Points` 字段**
```go
type drawMemberInfoSnapshot struct {
    MemberID     int64 `gorm:"column:member_id"`
    LotteryTimes int32 `gorm:"column:lottery_times"`
    Points       int32 `gorm:"column:points"`
    IsEnabled    int32 `gorm:"column:is_enabled"`
}
```

**c. `buildEligibility` 内消耗余额检查按 ConsumeType 分支**

line 439（`summary.RemainingLotteryTime`）：
```go
// 改为：按 consume_type 返回对应余额
if strings.TrimSpace(activity.ConsumeType) == drawConsumeTypePoints {
    summary.RemainingLotteryTime = member.Points
} else {
    summary.RemainingLotteryTime = member.LotteryTimes
}
```

line 448（`MinimumLotteryTimes` 门槛检查）：
```go
if rules.MinimumLotteryTimes > 0 {
    available := member.LotteryTimes
    if strings.TrimSpace(activity.ConsumeType) == drawConsumeTypePoints {
        available = member.Points
    }
    if available < rules.MinimumLotteryTimes {
        // ... 不足门槛
    }
}
```

line 472（余额 >= consumeAmount 检查）：
```go
available := member.LotteryTimes
if strings.TrimSpace(activity.ConsumeType) == drawConsumeTypePoints {
    available = member.Points
}
if available < activity.ConsumeAmount {
    msg := "剩余抽奖次数不足"
    if strings.TrimSpace(activity.ConsumeType) == drawConsumeTypePoints {
        msg = "剩余积分不足"
    }
    summary.Status = drawEligibilityQuotaExhausted
    summary.Code = drawEligibilityQuotaExhausted
    summary.Message = msg
    summary.NextAction = drawNextActionRetryLater
    return summary
}
```

#### 2. `rpc/sms/internal/logic/drawparticipationservice/participate_draw_logic.go`

line 138-152（扣减货币逻辑）：
```go
// 当前（硬编码）：
before := member.LotteryTimes
after := before - activity.ConsumeAmount
updateResult := tx.WithContext(l.ctx).Table(...).
    Where("member_id = ? AND lottery_times >= ?", ...).
    Updates(map[string]interface{}{"lottery_times": after, ...})

// 改为（按 consume_type）：
isPoints := strings.TrimSpace(activity.ConsumeType) == drawConsumeTypePoints
var before, after int32
var currencyField string
if isPoints {
    before = member.Points
    after = before - activity.ConsumeAmount
    currencyField = "points"
} else {
    before = member.LotteryTimes
    after = before - activity.ConsumeAmount
    currencyField = "lottery_times"
}
updateResult := tx.WithContext(l.ctx).Table(member.TableName()).
    Where("member_id = ? AND "+currencyField+" >= ?", in.MemberId, activity.ConsumeAmount).
    Updates(map[string]interface{}{currencyField: after, "update_time": time.Now()})
```

> 注：`drawParticipationRecordRow.LotteryTimesBefore/After` 列名保留原样，存储实际扣减前后值（技术债，后续可用迁移重命名为 `currency_before/after`）。

#### 3. web-admin 前端（可选，低优先级）

`web-admin/src/pages/sms/DrawActivity/` 活动表单里 `consume_type` 下拉框补充 `points` 选项及对应中文标签（当前可能只有 `lottery_times`）。

---

## 数据库无需迁移

`ums_member_info.points` 字段已存在，`sms_draw_activity.consume_type` 已存在。

---

## 测试

- 单元测试：在 `draw_participation_logic_test.go` 新增 consume_type=points 场景
- 手动验证：将活动 910001 的 `consume_type` 改为 `points`，用 David(member_id=4) 抽卡，确认 `points` 被扣减

---

## 影响范围

- `rpc/sms/internal/logic/drawparticipationservice/draw_participation_helper.go`
- `rpc/sms/internal/logic/drawparticipationservice/participate_draw_logic.go`
- （可选）`web-admin/src/pages/sms/DrawActivity/`
- 部署：sms-rpc
