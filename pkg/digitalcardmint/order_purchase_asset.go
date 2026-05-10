package digitalcardmint

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

const sourceTypePurchase = "purchase"

type EnsureOrderPurchaseAssetInput struct {
	OrderID           int64
	OrderItemID       int64
	ProductID         int64
	SkuID             int64
	MemberID          int64
	FulfillmentRuleID int64
	PlatformID        int64
	TenantID          int64
	MerchantID        int64
	RequestID         string
	TraceID           string
	OperatorType      string
}

type EnsureOrderPurchaseAssetResult struct {
	AssetInstanceID int64
	AssetNo         string
	MintStatus      string
	MintTaskID      int64
}

type productFulfillmentRuleRow struct {
	ID                  int64  `gorm:"column:id"`
	RuleStatus          int32  `gorm:"column:rule_status"`
	CardTemplateID      int64  `gorm:"column:card_template_id"`
	ExpireDays          int32  `gorm:"column:expire_days"`
	Transferable        int32  `gorm:"column:transferable"`
	TransferLimit       int32  `gorm:"column:transfer_limit"`
	ClaimCondition      string `gorm:"column:claim_condition"`
	RedemptionCondition string `gorm:"column:redemption_condition"`
	RefundPolicy        string `gorm:"column:refund_policy"`
	PlatformID          int64  `gorm:"column:platform_id"`
	TenantID            int64  `gorm:"column:tenant_id"`
	MerchantID          int64  `gorm:"column:merchant_id"`
}

func (productFulfillmentRuleRow) TableName() string {
	return "sms_product_fulfillment_rule"
}

func (s *Service) EnsureOrderPurchaseAsset(ctx context.Context, input EnsureOrderPurchaseAssetInput) (*EnsureOrderPurchaseAssetResult, error) {
	if s == nil || s.DB == nil {
		return nil, errors.New("数据库未初始化")
	}
	if input.OrderItemID <= 0 {
		return nil, errors.New("订单明细ID不能为空")
	}
	if input.MemberID <= 0 {
		return nil, errors.New("会员ID不能为空")
	}
	if input.FulfillmentRuleID <= 0 {
		return nil, errors.New("发卡规则ID不能为空")
	}

	var result *EnsureOrderPurchaseAssetResult
	var taskID int64
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		instance, txErr := s.ensureOrderPurchaseAssetTx(ctx, tx, input)
		if txErr != nil {
			return fmt.Errorf("ensureOrderPurchaseAssetTx: %w", txErr)
		}
		task, txErr := s.ensureTaskTxForAsset(ctx, tx, instance.ID, normalizeOperatorType(input.OperatorType))
		if txErr != nil {
			return fmt.Errorf("ensureTaskTxForAsset(assetId=%d): %w", instance.ID, txErr)
		}
		if task != nil {
			taskID = task.ID
		}
		result = &EnsureOrderPurchaseAssetResult{
			AssetInstanceID: instance.ID,
			AssetNo:         instance.AssetNo,
			MintStatus:      instance.MintStatus,
			MintTaskID:      taskID,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if taskID > 0 {
		_ = s.DispatchTask(ctx, taskID, "订单购买型资产建账后自动派发发放任务")
	}
	return result, nil
}

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
	if input.PlatformID > 0 && rule.PlatformID != input.PlatformID || input.TenantID > 0 && rule.TenantID != input.TenantID || input.MerchantID > 0 && rule.MerchantID != input.MerchantID {
		return nil, errors.New("发卡规则作用域与订单不一致")
	}

	issuedAt := s.now()
	instance := &CardInstanceRow{
		PlatformID:            rule.PlatformID,
		TenantID:              rule.TenantID,
		MerchantID:            rule.MerchantID,
		MemberID:              input.MemberID,
		RequestID:             strings.TrimSpace(input.RequestID),
		TraceID:               firstNonEmpty(strings.TrimSpace(input.TraceID), strings.TrimSpace(input.RequestID)),
		Scope:                 fmt.Sprintf("platform:%d,tenant:%d,merchant:%d", rule.PlatformID, rule.TenantID, rule.MerchantID),
		TemplateID:            rule.CardTemplateID,
		AssetNo:               generateOrderPurchaseAssetNo(issuedAt),
		AssetStatus:           AssetStatusCreated,
		MintStatus:            MintStatusPending,
		ChainStatus:           ChainStatusUnknown,
		SourceType:            sourceTypePurchase,
		SourceID:              input.OrderItemID,
		FulfillmentRuleID:     input.FulfillmentRuleID,
		Transferable:          rule.Transferable,
		TransferLimit:         rule.TransferLimit,
		ClaimCondition:        rule.ClaimCondition,
		RedemptionCondition:   rule.RedemptionCondition,
		RefundPolicy:          rule.RefundPolicy,
		DisplayStatus:         DisplayStatusVisible,
		ComplianceStatus:      ComplianceStatusClear,
		IssuedAt:              &issuedAt,
		CreateTime:            &issuedAt,
		ParticipationRecordID: 0,
	}
	if err = tx.WithContext(ctx).Table(instance.TableName()).Create(instance).Error; err != nil {
		if isDuplicateEntryError(err) {
			reload, reloadErr := s.loadOrderPurchaseAsset(ctx, tx, input.OrderItemID)
			if reloadErr != nil {
				return nil, fmt.Errorf("loadOrderPurchaseAsset after duplicate(itemId=%d, assetNo=%s, dupErr=%v): %w", input.OrderItemID, instance.AssetNo, err, reloadErr)
			}
			return reload, nil
		}
		return nil, fmt.Errorf("createCardInstance(itemId=%d, assetNo=%s, templateId=%d): %w", input.OrderItemID, instance.AssetNo, instance.TemplateID, err)
	}

	if err = s.appendAssetLogTx(ctx, tx, instance, "", instance.AssetStatus, OperationAssetCreatedFromOrder, normalizeOperatorType(input.OperatorType), instance.TraceID, "order_purchase", "订单支付成功后创建数字卡片资产", map[string]interface{}{
		"orderId":             input.OrderID,
		"orderItemId":         input.OrderItemID,
		"productId":           input.ProductID,
		"skuId":               input.SkuID,
		"memberId":            input.MemberID,
		"fulfillmentRuleId":   input.FulfillmentRuleID,
		"expireDays":          rule.ExpireDays,
		"transferable":        rule.Transferable,
		"transferLimit":       rule.TransferLimit,
		"claimCondition":      rule.ClaimCondition,
		"redemptionCondition": rule.RedemptionCondition,
		"refundPolicy":        rule.RefundPolicy,
	}); err != nil {
		return nil, fmt.Errorf("appendAssetLog(assetId=%d, itemId=%d): %w", instance.ID, input.OrderItemID, err)
	}
	return instance, nil
}

func (s *Service) ensureTaskTxForAsset(ctx context.Context, tx *gorm.DB, assetInstanceID int64, operatorType string) (*CardMintTaskRow, error) {
	if existing, err := s.loadTaskByAssetInstance(ctx, tx, assetInstanceID, true); err == nil {
		return existing, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("loadTaskByAssetInstance(assetId=%d): %w", assetInstanceID, err)
	}
	task, err := s.EnsureTaskTx(ctx, tx, assetInstanceID, operatorType)
	if err != nil {
		return nil, fmt.Errorf("EnsureTaskTx(assetId=%d): %w", assetInstanceID, err)
	}
	return task, nil
}

func (s *Service) loadOrderPurchaseAsset(ctx context.Context, db *gorm.DB, orderItemID int64) (*CardInstanceRow, error) {
	var row CardInstanceRow
	err := db.WithContext(ctx).
		Table(row.TableName()).
		Where("source_type = ? AND source_id = ? AND is_deleted = 0", sourceTypePurchase, orderItemID).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("loadOrderPurchaseAsset query(itemId=%d): %w", orderItemID, err)
	}
	return &row, nil
}

func (s *Service) loadProductFulfillmentRule(ctx context.Context, db *gorm.DB, ruleID int64) (*productFulfillmentRuleRow, error) {
	var row productFulfillmentRuleRow
	if err := db.WithContext(ctx).Table(row.TableName()).Where("id = ? AND is_deleted = 0", ruleID).Take(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func generateOrderPurchaseAssetNo(now time.Time) string {
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("CARD%s%06d", now.Format("20060102150405"), now.UnixNano()%1000000)
	}
	return fmt.Sprintf("CARD%s%s", now.Format("20060102150405"), strings.ToUpper(hex.EncodeToString(buf)))
}

func isDuplicateEntryError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate entry") || strings.Contains(message, "unique constraint failed")
}
