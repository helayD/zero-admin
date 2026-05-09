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

type UpdateProductFulfillmentRuleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateProductFulfillmentRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProductFulfillmentRuleLogic {
	return &UpdateProductFulfillmentRuleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateProductFulfillmentRuleLogic) UpdateProductFulfillmentRule(in *smsclient.UpdateProductFulfillmentRuleReq) (*smsclient.UpdateProductFulfillmentRuleResp, error) {
	// 1. 验证参数
	if in.Id <= 0 {
		return nil, errors.New("规则ID无效")
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

	// 3. 检查规则是否存在
	var count int64
	err := l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_product_fulfillment_rule").
		Where("id = ? AND platform_id = ? AND tenant_id = ? AND merchant_id = ? AND is_deleted = 0",
			in.Id, platformId, tenantId, merchantId).
		Count(&count).Error
	if err != nil {
		logc.Errorf(l.ctx, "检查发卡规则是否存在失败: %v", err)
		return nil, err
	}
	if count == 0 {
		return nil, errors.New("发卡规则不存在")
	}

	// 4. 构建更新字段
	now := time.Now()
	updates := map[string]interface{}{
		"update_time": now,
	}
	if in.RuleName != "" {
		updates["rule_name"] = in.RuleName
	}
	if in.CardTemplateId > 0 {
		// Story 10.10 Task 8.1: 切换 cardTemplateId 时也要校验
		if err := validateCardTemplateForRule(l.ctx, l.svcCtx.DB, in.CardTemplateId, platformId, tenantId, merchantId); err != nil {
			return nil, err
		}
		updates["card_template_id"] = in.CardTemplateId
	}
	if in.ExpireDays > 0 {
		updates["expire_days"] = in.ExpireDays
	}
	if in.Transferable >= 0 {
		updates["transferable"] = in.Transferable
	}
	if in.TransferLimit >= 0 {
		updates["transfer_limit"] = in.TransferLimit
	}
	if in.ClaimCondition != "" {
		updates["claim_condition"] = in.ClaimCondition
	}
	if in.RedemptionCondition != "" {
		updates["redemption_condition"] = in.RedemptionCondition
	}
	if in.RefundPolicy != "" {
		updates["refund_policy"] = in.RefundPolicy
	}

	// 5. 执行更新
	err = l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_product_fulfillment_rule").
		Where("id = ? AND platform_id = ? AND tenant_id = ? AND merchant_id = ? AND is_deleted = 0",
			in.Id, platformId, tenantId, merchantId).
		Updates(updates).Error
	if err != nil {
		logc.Errorf(l.ctx, "更新发卡规则失败: %v", err)
		return nil, errors.New("更新发卡规则失败")
	}

	// 6. 查询更新后的规则
	type ruleRow struct {
		Id                  int64  `gorm:"column:id"`
		RuleName            string `gorm:"column:rule_name"`
		RuleStatus          int32  `gorm:"column:rule_status"`
		CardTemplateId      int64  `gorm:"column:card_template_id"`
		ExpireDays          int32  `gorm:"column:expire_days"`
		Transferable        int32  `gorm:"column:transferable"`
		TransferLimit       int32  `gorm:"column:transfer_limit"`
		ClaimCondition      string `gorm:"column:claim_condition"`
		RedemptionCondition string `gorm:"column:redemption_condition"`
		RefundPolicy        string `gorm:"column:refund_policy"`
		PlatformId          int64  `gorm:"column:platform_id"`
		TenantId            int64  `gorm:"column:tenant_id"`
		MerchantId          int64  `gorm:"column:merchant_id"`
		CreateTime          string `gorm:"column:create_time"`
		UpdateTime          string `gorm:"column:update_time"`
	}

	var row ruleRow
	err = l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_product_fulfillment_rule").
		Select("id, rule_name, rule_status, card_template_id, expire_days, transferable, transfer_limit, claim_condition, redemption_condition, refund_policy, platform_id, tenant_id, merchant_id, create_time, update_time").
		Where("id = ?", in.Id).
		Take(&row).Error
	if err != nil {
		logc.Errorf(l.ctx, "查询更新后的规则失败: %v", err)
		return nil, err
	}

	return &smsclient.UpdateProductFulfillmentRuleResp{
		Rule: &smsclient.ProductFulfillmentRuleData{
			Id:                  row.Id,
			RuleName:            row.RuleName,
			RuleStatus:          row.RuleStatus,
			CardTemplateId:      row.CardTemplateId,
			ExpireDays:          row.ExpireDays,
			Transferable:        row.Transferable,
			TransferLimit:       row.TransferLimit,
			ClaimCondition:      row.ClaimCondition,
			RedemptionCondition: row.RedemptionCondition,
			RefundPolicy:        row.RefundPolicy,
			PlatformId:          row.PlatformId,
			TenantId:            row.TenantId,
			MerchantId:          row.MerchantId,
			CreateTime:          row.CreateTime,
			UpdateTime:          row.UpdateTime,
		},
	}, nil
}
