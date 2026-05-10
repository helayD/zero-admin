# Story 10.7 调试进度（2026-05-10 16:00 暂停点）

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

### 下次第一步要做的

继续把 `ensureOrderPurchaseAssetTx` 内 **每个 raw return 都 wrap 上 step 名**，特别是：
- `tx.Create(instance)` 失败的 raw err → wrap `Create(instance source_id=%d, asset_no=%s): %w`
- `isDuplicateEntryError` 后 reload 的失败 → wrap `loadOrderPurchaseAsset after dup(itemId=%d): %w`
- `appendAssetLogTx` 后 → wrap `appendAssetLog: %w`

然后部署，让用户重新支付，根据具体 wrap 信息精准定位。

也可以**直接连远程 sms-rpc 加 zerolog logc.Errorf 打印更详细的 step + asset_no + error**。

### 备选直接修：`asset_no` 唯一索引冲突的可能性

```@/Users/helay/Documents/GitHub/zero-admin/pkg/digitalcardmint/order_purchase_asset.go:200-219
// 查 generateOrderPurchaseAssetNo 看具体实现
```

读完确认是否高并发场景下会撞 asset_no。如果是，把 generation 改用 `crypto/rand + time + sequence`。

---

## 已修但未在手机上验证的内容

1. **OrderPay "查看订单详情" 按钮真跳 OrderDetail**（commit `9e909ab6`）
2. **`appendAssetLogTx` `CreateTime` 零值修复**（commit `5a3c04cd`）
3. **`app_recent_context.dart` 5 个 pre-existing 编译 bug**（同上）
4. **分享链路 H5 落地页 + QR + 白名单兜底**（commit `a881236d` + `5a3c04cd`）
5. **`ensureOrderPurchaseAssetTx` 部分 step wrap**（commit `9f6bd1fb`）

---

## 远程环境状态

- 当前分支：`dev` @ `9f6bd1fb`
- 已部署服务：`sms-rpc` / `front-api` / `consumer`（最近一次部署 backup: `/root/zero-admin/deploy-backup/20260510-155815`）
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
