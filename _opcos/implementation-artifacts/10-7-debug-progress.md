# Story 10.7 调试进度（2026-05-10 16:20 找到根因 + 已修复）

## 🎯 真正根因（2026-05-10 16:14 复测捕获）

新 wrap 抓到的精确错误：

```
ensureOrderPurchaseAssetTx: loadOrderPurchaseAsset after duplicate(
  itemId=30, assetNo=CARD2026051016135504DCE68A,
  dupErr=Error 1062 (23000): Duplicate entry '0-0' for key
         'sms_card_instance.uk_card_instance_participation'
): record not found
```

**问题：** `sms_card_instance.uk_card_instance_participation` 唯一索引建在
`(participation_record_id, is_deleted)` 上。订单购买流程总是写
`participation_record_id=0`，导致第二张订单卡建账时撞 `'0-0'` 重复键。

**为何 sku=33 (rule_id=1) 能成功而 sku=20 (rule_id=2) 失败？**
sku=33 是历史上第一张 purchase 卡（占了 `(0, 0)` 唯一槽位），
sku=20 是第二张，必撞重复键。这跟 rule_id 完全无关，只是顺序问题。

## ✅ 修复（commit 待提交）

1. **`rpc/sms/internal/logic/cardassetservice/card_asset_helper.go`**：
   抽卡流程 `createCardInstanceWithRetry` 补设 `SourceType="draw", SourceID=record.ID`，
   让新 draw 行也走 `uk_source_type_id` 唯一性保护。
2. **`script/sql/sms/migration_20260510_drop_card_instance_participation_unique.sql`**：
   drop 老的 `uk_card_instance_participation` 索引（已被 `uk_source_type_id` 等价覆盖）。
3. **测试 schema 同步**：`card_asset_logic_test.go` / `draw_participation_logic_test.go`
   去掉旧 unique index，加上 `uk_source_type_id`。

---

# Story 10.7 调试进度（2026-05-10 16:12 更新）

> 下次继续从这里读起。已修完的不再重复，只列**当前未解决的问题 + 下次第一步该做什么**。

---

## 当前手机端实测结论

| 链路 | 状态 |
|---|---|
| 商品下单 → 模拟支付（小米手机 sku=33 / rule_id=1） | ✅ 走通过一次（建出卡 970012, source_id=29, template=920002, mint_manual_review） |
| 商品下单 → 模拟支付（苏泊尔压力锅 sku=20 / rule_id=2） | ❌ **`订单明细[30]提货卡建账失败: ensureOrderPurchaseAssetTx: record not found`** |
| 卡包页 / 卡详情页 | ✅ |
| 我要提货 / 分享给朋友 / H5 领取页 | ⏳ 还没在手机上实测（被支付失败卡住） |
| 支付成功页 → "查看订单详情" | ⏳ 已修但因支付失败未验证 |

最新失败截屏：`/tmp/phone-screen-4.png`

---

## 当前问题：sku=20 的支付链路 record not found

### 已知数据
- `order_item.id=30` → `sku_id=20` → `spu_id=10`（苏泊尔压力锅）
- `pms_product_sku.id=20`：`sku_mode=NULL, sku_rule=NULL`
- `pms_product_spu.id=10`：`fulfillment_mode='digital_asset', fulfillment_rule_id=2`
- `sms_product_fulfillment_rule.id=2`：rule_status=1, card_template_id=920004, platform_id=1, tenant_id=0, merchant_id=0
- `sms_card_template.id=920004`：status=1, display_status=1, audit_status=2, content_audit_status=2, is_deleted=0
- `sms_card_instance` 中**没有** source_id=30 的记录（INSERT 没成功就回滚了）
- 对比：成功的 sku=33 走 rule_id=1，失败的 sku=20 走 rule_id=2

### 已经 wrap 但没复现具体步骤的代码
本轮 commit `9f6bd1fb` 已部署，但失败信息只到 `ensureOrderPurchaseAssetTx: record not found`，**没有出现 `loadOrderPurchaseAsset` / `loadProductFulfillmentRule` / `loadCardTemplate` 等更深层 wrap**。说明 record not found 来自 **`ensureOrderPurchaseAssetTx` 内还没加 wrap 的位置**：

```@/Users/helay/Documents/GitHub/zero-admin/pkg/digitalcardmint/order_purchase_asset.go:103-173
func (s *Service) ensureOrderPurchaseAssetTx(ctx context.Context, tx *gorm.DB, input EnsureOrderPurchaseAssetInput) (*CardInstanceRow, error) {
	if existing, err := s.loadOrderPurchaseAsset(ctx, tx, input.OrderItemID); err == nil {
		return existing, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("loadOrderPurchaseAsset(itemId=%d): %w", input.OrderItemID, err)
	}

	rule, err := s.loadProductFulfillmentRule(ctx, tx, input.FulfillmentRuleID)
	if err != nil {
		return nil, fmt.Errorf("loadProductFulfillmentRule(ruleId=%d): %w", input.FulfillmentRuleID, err)
	}
	if rule.RuleStatus != 1 {
		return nil, errors.New("发卡规则已禁用")
	}
	if input.PlatformID > 0 && rule.PlatformID != input.PlatformID || ... { ... }

	issuedAt := s.now()
	instance := &CardInstanceRow{ ... }
	if err = tx.WithContext(ctx).Table(instance.TableName()).Create(instance).Error; err != nil {
		if isDuplicateEntryError(err) {
			return s.loadOrderPurchaseAsset(ctx, tx, input.OrderItemID)  // <-- 嫌疑 #1
		}
		return nil, err  // <-- 嫌疑 #2：raw err 没 wrap
	}

	if err = s.appendAssetLogTx(ctx, tx, instance, ...); err != nil {
		return nil, err  // <-- 嫌疑 #3
	}
	return instance, nil
}
```

### 嫌疑路径排序

1. **嫌疑 #1（最可能）**：INSERT 触发 duplicate（虽然 source_id=30 数据库看着没有），fallback 到 `loadOrderPurchaseAsset` 又找不到 → 直接返回 `gorm.ErrRecordNotFound`，没 wrap。
   - 需要排查：是否存在某个唯一索引被违反？比如 `uk_card_instance_asset_no` 的 asset_no 因为 `generateOrderPurchaseAssetNo` 太快产生了重复？
   - `generateOrderPurchaseAssetNo` 用的是 `time.Now().UnixNano()` + 随机字节，理论上不会冲突，但要看具体实现。

2. **嫌疑 #2**：INSERT 报别的错（比如某 NOT NULL 字段没填、外键不存在），但 raw err 不是 record not found —— 实际上不太可能直接显示成 record not found。

3. **嫌疑 #3**：`appendAssetLogTx` 内的 `Create` 报了 record not found（基本不可能，Create 不会返回 ErrRecordNotFound）。

### 已完成（commit `f230ac9b`，已部署到 47.107.224.56）

把 `ensureOrderPurchaseAssetTx` 剩余 raw err 全部 wrap：

- ✅ `tx.Create(instance)` → `createCardInstance(itemId=%d, assetNo=%s, templateId=%d): %w`
- ✅ dup → reload 失败 → `loadOrderPurchaseAsset after duplicate(itemId=%d, assetNo=%s, dupErr=%v): %w`
- ✅ `appendAssetLogTx` → `appendAssetLog(assetId=%d, itemId=%d): %w`
- ✅ `loadOrderPurchaseAsset` 非 ErrRecordNotFound → `loadOrderPurchaseAsset query(itemId=%d): %w`
- ✅ `ensureTaskTxForAsset` 内的 `loadTaskByAssetInstance` / `EnsureTaskTx` 都 wrap
- ✅ `EnsurePaidOrderPurchaseAssets` caller 错误信息附加 `skuId/productId/ruleId/scope=p%d/t%d/m%d`

### 远程数据库现状（2026-05-10 16:11 直查）

```
sms_card_instance source_id=30: 无记录
sms_card_instance source_type='purchase' member_id=4: 仅 970012(source_id=29)
sms_product_fulfillment_rule id=2: rule_status=1, card_template_id=920004, platform=1/t=0/m=0, is_deleted=0 ✅
sms_card_template id=920004: status=1, display=1, audit=2, content_audit=2, is_deleted=0, p=1/t=0/m=0 ✅
```

**结论：** 嫌疑 #1（duplicate-entry → reload not found）几乎不成立——数据库就没有 source_id=30 的脏数据。
真正的 record not found 来源更可能是 `tx.Create(instance)` 报了某个非 dup 错误（比如外键、NOT NULL 字段、字段长度），但 raw err 包成 `record not found` 字面量。等下次复测看新 wrap 内容就明白。

### 下次第一步要做的

**让用户重新模拟支付 sku=20（订单 28 / item 30）一次**，新错误信息会精准告诉我们是哪个 step 抛的：

- 如果是 `订单明细[30](skuId=20, productId=10, ruleId=2, scope=p1/t0/m0)提货卡建账失败: ensureOrderPurchaseAssetTx: createCardInstance(itemId=30, assetNo=CARD..., templateId=920004): <真实错误>` → 字段问题，看真实错误修
- 如果是 `... appendAssetLog(assetId=..., itemId=30): <真实错误>` → 资产已经建出来了，问题在日志写入
- 如果是 `... ensureTaskTxForAsset(assetId=..., itemId=30): <真实错误>` → 资产 + 日志都好，问题在任务建账

---

## 已修但未在手机上验证的内容

1. **OrderPay "查看订单详情" 按钮真跳 OrderDetail**（commit `9e909ab6`）
2. **`appendAssetLogTx` `CreateTime` 零值修复**（commit `5a3c04cd`）
3. **`app_recent_context.dart` 5 个 pre-existing 编译 bug**（同上）
4. **分享链路 H5 落地页 + QR + 白名单兜底**（commit `a881236d` + `5a3c04cd`）
5. **`ensureOrderPurchaseAssetTx` 部分 step wrap**（commit `9f6bd1fb`）
6. **`ensureOrderPurchaseAssetTx` 全 step wrap + ensureTaskTxForAsset wrap + caller 信息扩充**（commit `f230ac9b`，已部署）

---

## 远程环境状态

- 当前分支：`dev` @ `f230ac9b`
- 已部署服务：`sms-rpc` / `front-api` / `consumer`（最近一次部署 backup: `/root/zero-admin/deploy-backup/20260510-160916`）
- Flutter App：在一加 9 (4661d9aa) 上运行（command 330 RUNNING，hot reload 待命）
- 测试账号：David / member_id=4 / mobile=16698129676

## 失败的订单遗留

- 订单 27 / item 29：第一次失败（create_time bug），第二次成功 → 卡 970012 已建
- 订单 28 / item 30：本次失败（record not found 未定位）
- 订单状态都已被 ensurePaidOrderPurchaseAssets fallback 回滚到待支付

## 下次启动检查清单

- [ ] 把 `ensureOrderPurchaseAssetTx` 剩余 raw err 全 wrap
- [ ] 部署 `sms-rpc` + `front-api`
- [ ] 让用户重新模拟支付 sku=20 一次，看新的错误链路
- [ ] 根据 wrap 信息修真正的根因
- [ ] 验证支付成功 → "查看订单详情" 按钮跳转
- [ ] 在卡详情页测 "我要提货"（地址簿预填 + 提货单详情页）
- [ ] 在卡详情页测 "分享给朋友"（QR + 复制链接 + H5 领取页）
- [ ] 朋友手机扫码或浏览器打开 H5 链接，验证未注册/已注册两条登录路径都能领取
