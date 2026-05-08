package cardassetservicelogic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	cardminttaskservicelogic "github.com/feihua/zero-admin/rpc/sms/internal/logic/cardminttaskservice"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

const (
	// 订单购买型资产来源
	sourceTypePurchase = "purchase"
	sourceTypeDraw     = "draw"

	// 订单购买型资产日志操作
	cardAssetOperationCreateFromOrder = "asset_created_from_order"
	cardAssetReasonPurchase           = "order_purchase"
)

// orderPurchaseRecord 订单购买记录（用于创建卡片实例）
type orderPurchaseRecord struct {
	OrderId           int64  `json:"orderId"`
	OrderItemId       int64  `json:"orderItemId"`
	ProductId         int64  `json:"productId"`
	SkuId             int64  `json:"skuId"`
	MemberId          int64  `json:"memberId"`
	FulfillmentRuleId int64  `json:"fulfillmentRuleId"`
	TraceId           string `json:"traceId"`
	RequestId         string `json:"requestId"`
}

// fulfillmentRuleSnapshot 发卡规则快照
type fulfillmentRuleSnapshot struct {
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

func (fulfillmentRuleSnapshot) TableName() string {
	return "sms_product_fulfillment_rule"
}

// EnsureOrderPurchaseCardInstanceLogic 订单购买型卡片资产创建逻辑
type EnsureOrderPurchaseCardInstanceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewEnsureOrderPurchaseCardInstanceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EnsureOrderPurchaseCardInstanceLogic {
	return &EnsureOrderPurchaseCardInstanceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// EnsureOrderPurchaseCardInstance 确保订单购买型卡片资产实例存在（幂等）
func (l *EnsureOrderPurchaseCardInstanceLogic) EnsureOrderPurchaseCardInstance(in *smsclient.EnsureOrderPurchaseCardInstanceReq) (*smsclient.EnsureOrderPurchaseCardInstanceResp, error) {
	// 1. 验证参数
	if in.OrderItemId <= 0 {
		return nil, errors.New("订单明细ID无效")
	}
	if in.FulfillmentRuleId <= 0 {
		return nil, errors.New("发卡规则ID无效")
	}
	if in.MemberId <= 0 {
		return nil, errors.New("会员ID无效")
	}

	// 2. 在事务中执行
	var asset *CardInstanceSnapshot
	var dispatchTaskID int64
	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		var txErr error
		asset, txErr = ensureOrderPurchaseCardInstance(l.ctx, tx, &orderPurchaseRecord{
			OrderId:           in.OrderId,
			OrderItemId:       in.OrderItemId,
			ProductId:         in.ProductId,
			SkuId:             in.SkuId,
			MemberId:          in.MemberId,
			FulfillmentRuleId: in.FulfillmentRuleId,
			TraceId:           in.TraceId,
			RequestId:         in.RequestId,
		}, in.OperatorType, in.TraceId)
		if txErr != nil {
			return txErr
		}

		// 创建发放任务
		dispatchTaskID, txErr = cardminttaskservicelogic.EnsureCardMintTaskByAssetInstance(l.ctx, l.svcCtx, tx, asset.ID, in.OperatorType)
		return txErr
	})
	if err != nil {
		logc.Errorf(l.ctx, "创建订单购买型卡片资产失败: %v, orderId=%d, orderItemId=%d", err, in.OrderId, in.OrderItemId)
		return nil, err
	}

	// 3. 派发发放任务
	cardminttaskservicelogic.DispatchCardMintTask(l.ctx, l.svcCtx, dispatchTaskID, "订单购买型资产建账后自动派发链上发放任务")

	return &smsclient.EnsureOrderPurchaseCardInstanceResp{
		Asset: buildCardInstanceData(asset),
	}, nil
}

// ensureOrderPurchaseCardInstance 确保订单购买型卡片资产实例存在（幂等）
func ensureOrderPurchaseCardInstance(ctx context.Context, tx *gorm.DB, record *orderPurchaseRecord, operatorType string, traceID string) (*CardInstanceSnapshot, error) {
	if record == nil {
		return nil, errors.New("订单购买记录不能为空")
	}
	if tx == nil {
		return nil, errors.New("数据库事务不能为空")
	}

	// 1. 尝试加载已存在的卡片实例（幂等检查）
	existing, err := loadCardInstanceByOrderItemId(ctx, tx, record.OrderItemId)
	switch {
	case err == nil:
		// 已存在，直接返回
		return buildCardInstanceSnapshot(existing), nil
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		return nil, err
	}

	// 2. 加载发卡规则
	rule, err := loadFulfillmentRule(ctx, tx, record.FulfillmentRuleId)
	if err != nil {
		return nil, fmt.Errorf("加载发卡规则失败: %w", err)
	}

	// 3. 创建卡片实例
	instance, err := createOrderPurchaseCardInstance(ctx, tx, record, rule, normalizeOperatorType(operatorType), traceID)
	if err != nil {
		return nil, err
	}

	// 4. 创建初始资产日志
	if err := ensureOrderPurchaseAssetLog(ctx, tx, record, instance, normalizeOperatorType(operatorType), traceID); err != nil {
		return nil, err
	}

	return buildCardInstanceSnapshot(instance), nil
}

// loadCardInstanceByOrderItemId 根据订单明细ID加载卡片实例
func loadCardInstanceByOrderItemId(ctx context.Context, db *gorm.DB, orderItemId int64) (*cardInstanceRow, error) {
	var row cardInstanceRow
	query := db.WithContext(ctx).
		Table(row.TableName()).
		Where("source_type = ? AND source_id = ? AND is_deleted = 0", sourceTypePurchase, orderItemId)
	if err := query.Take(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

// loadFulfillmentRule 加载发卡规则
func loadFulfillmentRule(ctx context.Context, db *gorm.DB, ruleId int64) (*fulfillmentRuleSnapshot, error) {
	var rule fulfillmentRuleSnapshot
	err := db.WithContext(ctx).
		Table(rule.TableName()).
		Where("id = ? AND is_deleted = 0", ruleId).
		Take(&rule).Error
	if err != nil {
		return nil, err
	}
	if rule.RuleStatus != 1 {
		return nil, errors.New("发卡规则已禁用")
	}
	return &rule, nil
}

// createOrderPurchaseCardInstance 创建订单购买型卡片实例
func createOrderPurchaseCardInstance(ctx context.Context, tx *gorm.DB, record *orderPurchaseRecord, rule *fulfillmentRuleSnapshot, operatorType string, traceID string) (*cardInstanceRow, error) {
	for attempt := 0; attempt < maxAssetNoRetryCount; attempt++ {
		issuedAt := time.Now()
		row := &cardInstanceRow{
			PlatformID:            rule.PlatformID,
			TenantID:              rule.TenantID,
			MerchantID:            rule.MerchantID,
			ActivityID:            0, // 订单购买型没有活动ID
			MemberID:              record.MemberId,
			ParticipationRecordID: 0, // 订单购买型没有参与记录ID
			RequestID:             strings.TrimSpace(record.RequestId),
			TraceID:               firstNonEmpty(strings.TrimSpace(traceID), strings.TrimSpace(record.RequestId)),
			Scope:                 fmt.Sprintf("platform:%d,tenant:%d,merchant:%d", rule.PlatformID, rule.TenantID, rule.MerchantID),
			PoolID:                0, // 订单购买型没有卡池ID
			TemplateID:            rule.CardTemplateID,
			Rarity:                "", // 订单购买型没有稀有度
			AssetNo:               assetNoGenerator(issuedAt),
			AssetStatus:           cardAssetStatusCreated,
			MintStatus:            cardMintStatusPending,
			TokenID:               "",
			ChainStatus:           "",
			SourceType:            sourceTypePurchase,
			SourceID:              record.OrderItemId,
			FulfillmentRuleID:     record.FulfillmentRuleId,
			MintTaskID:            0,
			IssuedAt:              &issuedAt,
			CreateBy:              0,
		}

		err := tx.WithContext(ctx).
			Table(row.TableName()).
			Omit("create_time", "update_by", "update_time").
			Create(row).Error
		if err == nil {
			return row, nil
		}
		if !isDuplicateEntryError(err) {
			return nil, err
		}
		// 幂等重试：尝试加载已存在的记录
		existing, existingErr := loadCardInstanceByOrderItemId(ctx, tx, record.OrderItemId)
		if existingErr == nil {
			return existing, nil
		}
		if existingErr != nil && !errors.Is(existingErr, gorm.ErrRecordNotFound) {
			return nil, existingErr
		}
	}
	return nil, errors.New("生成资产编号失败，请稍后重试")
}

// ensureOrderPurchaseAssetLog 创建订单购买型资产日志
func ensureOrderPurchaseAssetLog(ctx context.Context, tx *gorm.DB, record *orderPurchaseRecord, instance *cardInstanceRow, operatorType string, traceID string) error {
	if record == nil || instance == nil {
		return errors.New("记录或资产实例不能为空")
	}

	// 检查是否已存在日志（幂等）
	var count int64
	if err := tx.WithContext(ctx).
		Table(cardAssetLogRow{}.TableName()).
		Where("asset_instance_id = ? AND operation_type = ?", instance.ID, cardAssetOperationCreateFromOrder).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"assetNo":           instance.AssetNo,
		"orderId":           record.OrderId,
		"orderItemId":       record.OrderItemId,
		"productId":         record.ProductId,
		"skuId":             record.SkuId,
		"memberId":          record.MemberId,
		"fulfillmentRuleId": record.FulfillmentRuleId,
		"requestId":         record.RequestId,
	})

	logRow := &cardAssetLogRow{
		AssetInstanceID:       instance.ID,
		ParticipationRecordID: 0, // 订单购买型没有参与记录ID
		FromStatus:            "",
		ToStatus:              instance.AssetStatus,
		OperationType:         cardAssetOperationCreateFromOrder,
		OperatorType:          operatorType,
		TraceID:               firstNonEmpty(strings.TrimSpace(traceID), strings.TrimSpace(record.RequestId)),
		ReasonCode:            cardAssetReasonPurchase,
		ReasonText:            "订单支付成功后创建提货卡实例",
		PayloadJSON:           string(payload),
	}
	return tx.WithContext(ctx).Table(logRow.TableName()).Create(logRow).Error
}
