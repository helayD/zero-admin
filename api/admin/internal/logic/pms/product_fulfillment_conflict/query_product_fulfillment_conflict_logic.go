package product_fulfillment_conflict

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// QueryProductFulfillmentConflictLogic 查询商品履约模式配置冲突
type QueryProductFulfillmentConflictLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryProductFulfillmentConflictLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryProductFulfillmentConflictLogic {
	return &QueryProductFulfillmentConflictLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// productConflictRow 商品冲突检查查询结果
type productConflictRow struct {
	ID                int64  `gorm:"column:id"`
	Name              string `gorm:"column:name"`
	FulfillmentMode   string `gorm:"column:fulfillment_mode"`
	FulfillmentRuleID int64  `gorm:"column:fulfillment_rule_id"`
	PublishStatus     int32  `gorm:"column:publish_status"`
	PlatformID        int64  `gorm:"column:platform_id"`
	TenantID          int64  `gorm:"column:tenant_id"`
	MerchantID        int64  `gorm:"column:merchant_id"`
}

// fulfillmentRuleConflictRow 发卡规则冲突检查查询结果
type fulfillmentRuleConflictRow struct {
	ID             int64  `gorm:"column:id"`
	RuleStatus     int32  `gorm:"column:rule_status"`
	CardTemplateID int64  `gorm:"column:card_template_id"`
	PlatformID     int64  `gorm:"column:platform_id"`
	TenantID       int64  `gorm:"column:tenant_id"`
	MerchantID     int64  `gorm:"column:merchant_id"`
}

// cardTemplateConflictRow 卡片模板冲突检查查询结果
type cardTemplateConflictRow struct {
	ID     int64 `gorm:"column:id"`
	Status int32 `gorm:"column:status"`
}

// QueryProductFulfillmentConflict 查询商品履约模式配置冲突
func (l *QueryProductFulfillmentConflictLogic) QueryProductFulfillmentConflict(req *types.QueryProductFulfillmentConflictReq) (resp *types.QueryProductFulfillmentConflictResp, err error) {
	if len(req.Ids) == 0 {
		return &types.QueryProductFulfillmentConflictResp{
			Code:    "000000",
			Message: "查询成功",
			Data:    []types.ProductFulfillmentConflictItem{},
		}, nil
	}

	// 解析治理范围
	readScope, err := common.ResolveQueryGovernanceScope(l.ctx, common.RequestedGovernanceScope{
		ScopeType:  req.ScopeType,
		PlatformID: req.PlatformId,
		TenantID:   req.TenantId,
		MerchantID: req.MerchantId,
	})
	if err != nil {
		return nil, err
	}

	// 查询商品信息
	var products []productConflictRow
	err = l.svcCtx.DB.WithContext(l.ctx).
		Table("pms_product_spu").
		Select("id, name, fulfillment_mode, fulfillment_rule_id, publish_status, platform_id, tenant_id, merchant_id").
		Where("id IN ? AND is_deleted = 0", req.Ids).
		Find(&products).Error
	if err != nil {
		logc.Errorf(l.ctx, "查询商品信息失败: %v", err)
		return nil, err
	}

	// 构建商品映射
	productMap := make(map[int64]*productConflictRow)
	for i := range products {
		productMap[products[i].ID] = &products[i]
	}

	// 检查每个商品的冲突
	var items []types.ProductFulfillmentConflictItem
	for _, id := range req.Ids {
		item := types.ProductFulfillmentConflictItem{
			ProductId: id,
		}

		product, exists := productMap[id]
		if !exists {
			item.ProductName = "商品不存在"
			item.HasConflict = false
			items = append(items, item)
			continue
		}

		item.ProductName = product.Name

		// 检查作用域
		if readScope.PlatformID > 0 && product.PlatformID != readScope.PlatformID {
			item.HasConflict = false // 不在当前作用域，跳过检查
			items = append(items, item)
			continue
		}
		if readScope.TenantID > 0 && product.TenantID != readScope.TenantID {
			item.HasConflict = false
			items = append(items, item)
			continue
		}
		if readScope.MerchantID > 0 && product.MerchantID != readScope.MerchantID {
			item.HasConflict = false
			items = append(items, item)
			continue
		}

		// 只检查已上架的提货卡模式商品
		if product.PublishStatus != 1 || product.FulfillmentMode != "digital_asset" {
			item.HasConflict = false
			items = append(items, item)
			continue
		}

		// 检查发卡规则
		conflictMsg := l.checkFulfillmentRuleConflict(product)
		item.HasConflict = conflictMsg != ""
		item.ConflictMsg = conflictMsg
		items = append(items, item)
	}

	return &types.QueryProductFulfillmentConflictResp{
		Code:    "000000",
		Message: "查询成功",
		Data:    items,
	}, nil
}

// checkFulfillmentRuleConflict 检查发卡规则冲突
func (l *QueryProductFulfillmentConflictLogic) checkFulfillmentRuleConflict(product *productConflictRow) string {
	// 检查是否绑定发卡规则
	if product.FulfillmentRuleID <= 0 {
		return "提货卡模式商品未绑定发卡规则"
	}

	// 查询发卡规则
	var rule fulfillmentRuleConflictRow
	err := l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_product_fulfillment_rule").
		Select("id, rule_status, card_template_id, platform_id, tenant_id, merchant_id").
		Where("id = ? AND is_deleted = 0", product.FulfillmentRuleID).
		Take(&rule).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "发卡规则不存在或已删除"
		}
		return "查询发卡规则失败"
	}

	// 检查发卡规则状态
	if rule.RuleStatus != 1 {
		return "发卡规则已禁用"
	}

	// 检查作用域一致性
	if rule.PlatformID != product.PlatformID || rule.TenantID != product.TenantID || rule.MerchantID != product.MerchantID {
		return "发卡规则的作用域与商品不一致"
	}

	// 检查卡片模板
	if rule.CardTemplateID > 0 {
		var template cardTemplateConflictRow
		err := l.svcCtx.DB.WithContext(l.ctx).
			Table("sms_card_template").
			Select("id, status").
			Where("id = ? AND is_deleted = 0", rule.CardTemplateID).
			Take(&template).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "卡片模板不存在或已删除"
			}
			return "查询卡片模板失败"
		}
		if template.Status == 0 {
			return "卡片模板已禁用"
		}
	}

	return ""
}
