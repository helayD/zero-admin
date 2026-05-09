package productfulfillmentruleservice

import (
	"context"
	"errors"
	"time"

	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

type AddProductFulfillmentRuleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddProductFulfillmentRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddProductFulfillmentRuleLogic {
	return &AddProductFulfillmentRuleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AddProductFulfillmentRuleLogic) AddProductFulfillmentRule(in *smsclient.AddProductFulfillmentRuleReq) (*smsclient.AddProductFulfillmentRuleResp, error) {
	// 1. 验证参数
	if in.RuleName == "" {
		return nil, errors.New("规则名称不能为空")
	}
	if in.CardTemplateId <= 0 {
		return nil, errors.New("关联卡片模板ID无效")
	}
	if in.RefundPolicy == "" {
		in.RefundPolicy = "freeze_card"
	}

	// 2. 解析治理范围
	platformId := int64(1)
	tenantId := int64(0)
	merchantId := int64(0)
	if in.Scope != nil {
		platformId = in.Scope.PlatformId
		tenantId = in.Scope.TenantId
		merchantId = in.Scope.MerchantId
	}

	// 3. 检查规则名称是否重复
	var count int64
	err := l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_product_fulfillment_rule").
		Where("platform_id = ? AND tenant_id = ? AND merchant_id = ? AND rule_name = ? AND is_deleted = 0",
			platformId, tenantId, merchantId, in.RuleName).
		Count(&count).Error
	if err != nil {
		logc.Errorf(l.ctx, "检查规则名称重复失败: %v", err)
		return nil, errors.New("检查规则名称失败")
	}
	if count > 0 {
		return nil, errors.New("规则名称已存在")
	}

	// Story 10.10 Task 8.1: 校验关联的卡片模板存在、未被禁用、与当前 scope 匹配
	if err := validateCardTemplateForRule(l.ctx, l.svcCtx.DB, in.CardTemplateId, platformId, tenantId, merchantId); err != nil {
		return nil, err
	}

	// 4. 插入数据库
	now := time.Now()
	rule := map[string]interface{}{
		"platform_id":          platformId,
		"tenant_id":            tenantId,
		"merchant_id":          merchantId,
		"rule_name":            in.RuleName,
		"rule_status":          1, // 默认启用
		"card_template_id":     in.CardTemplateId,
		"expire_days":          in.ExpireDays,
		"transferable":         in.Transferable,
		"transfer_limit":       in.TransferLimit,
		"claim_condition":      in.ClaimCondition,
		"redemption_condition": in.RedemptionCondition,
		"refund_policy":        in.RefundPolicy,
		"create_by":            0, // TODO: 从上下文获取操作人ID
		"create_time":          now,
		"is_deleted":           0,
	}

	err = l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_product_fulfillment_rule").
		Create(rule).Error
	if err != nil {
		logc.Errorf(l.ctx, "插入发卡规则失败: %v", err)
		return nil, errors.New("创建发卡规则失败")
	}

	// 5. 获取插入的ID
	var insertedRule struct {
		Id int64 `gorm:"column:id"`
	}
	err = l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_product_fulfillment_rule").
		Select("id").
		Where("platform_id = ? AND tenant_id = ? AND merchant_id = ? AND rule_name = ? AND is_deleted = 0",
			platformId, tenantId, merchantId, in.RuleName).
		Order("id DESC").
		Take(&insertedRule).Error
	if err != nil {
		logc.Errorf(l.ctx, "获取插入的规则ID失败: %v", err)
		return nil, errors.New("获取创建的规则失败")
	}

	// 6. 返回结果
	return &smsclient.AddProductFulfillmentRuleResp{
		Rule: &smsclient.ProductFulfillmentRuleData{
			Id:                  insertedRule.Id,
			RuleName:            in.RuleName,
			RuleStatus:          1,
			CardTemplateId:      in.CardTemplateId,
			ExpireDays:          in.ExpireDays,
			Transferable:        in.Transferable,
			TransferLimit:       in.TransferLimit,
			ClaimCondition:      in.ClaimCondition,
			RedemptionCondition: in.RedemptionCondition,
			RefundPolicy:        in.RefundPolicy,
			PlatformId:          platformId,
			TenantId:            tenantId,
			MerchantId:          merchantId,
			CreateTime:          now.Format("2006-01-02 15:04:05"),
		},
	}, nil
}
