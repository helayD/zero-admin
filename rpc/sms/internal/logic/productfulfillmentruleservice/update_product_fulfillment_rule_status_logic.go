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

type UpdateProductFulfillmentRuleStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateProductFulfillmentRuleStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProductFulfillmentRuleStatusLogic {
	return &UpdateProductFulfillmentRuleStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateProductFulfillmentRuleStatusLogic) UpdateProductFulfillmentRuleStatus(in *smsclient.UpdateProductFulfillmentRuleStatusReq) (*smsclient.UpdateProductFulfillmentRuleStatusResp, error) {
	// 1. 验证参数
	if in.Id <= 0 {
		return nil, errors.New("规则ID无效")
	}
	if in.RuleStatus != 0 && in.RuleStatus != 1 {
		return nil, errors.New("规则状态无效，只能是 0（禁用）或 1（启用）")
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

	// 3. 如果要禁用规则，检查是否绑定商品
	if in.RuleStatus == 0 {
		var bindingCount int64
		err := l.svcCtx.DB.WithContext(l.ctx).
			Table("sms_product_fulfillment_binding").
			Where("rule_id = ? AND is_deleted = 0", in.Id).
			Count(&bindingCount).Error
		if err != nil {
			logc.Errorf(l.ctx, "检查规则绑定商品失败: %v", err)
			return nil, err
		}
		if bindingCount > 0 {
			return nil, errors.New("该规则已绑定商品，无法禁用")
		}
	}

	// 3.1 Story 10.10 第二轮 Review 修复 H2:
	//     启用规则时，必须再次校验当前规则关联的卡片模板仍然合法（未被删除/未被禁用/scope 仍可见），
	//     否则会出现"规则被禁→模板被禁→规则重新启用"的脏闭环，违反 AC7 的一致性 gate。
	if in.RuleStatus == 1 {
		var ruleRow struct {
			CardTemplateId int64 `gorm:"column:card_template_id"`
		}
		if err := l.svcCtx.DB.WithContext(l.ctx).
			Table("sms_product_fulfillment_rule").
			Select("card_template_id").
			Where("id = ? AND platform_id = ? AND tenant_id = ? AND merchant_id = ? AND is_deleted = 0",
				in.Id, platformId, tenantId, merchantId).
			Take(&ruleRow).Error; err != nil {
			logc.Errorf(l.ctx, "启用规则前查询关联卡片模板失败: %v", err)
			return nil, errors.New("启用前校验关联卡片模板失败，请稍后重试")
		}
		if err := validateCardTemplateForRule(l.ctx, l.svcCtx.DB, ruleRow.CardTemplateId, platformId, tenantId, merchantId); err != nil {
			return nil, err
		}
	}

	// 4. 执行更新
	now := time.Now()
	err := l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_product_fulfillment_rule").
		Where("id = ? AND platform_id = ? AND tenant_id = ? AND merchant_id = ? AND is_deleted = 0",
			in.Id, platformId, tenantId, merchantId).
		Updates(map[string]interface{}{
			"rule_status": in.RuleStatus,
			"update_time": now,
		}).Error
	if err != nil {
		logc.Errorf(l.ctx, "更新发卡规则状态失败: %v", err)
		return nil, errors.New("更新发卡规则状态失败")
	}

	// 5. 查询更新后的规则
	type ruleRow struct {
		Id         int64  `gorm:"column:id"`
		RuleName   string `gorm:"column:rule_name"`
		RuleStatus int32  `gorm:"column:rule_status"`
		UpdateTime string `gorm:"column:update_time"`
	}

	var row ruleRow
	err = l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_product_fulfillment_rule").
		Select("id, rule_name, rule_status, update_time").
		Where("id = ?", in.Id).
		Take(&row).Error
	if err != nil {
		logc.Errorf(l.ctx, "查询更新后的规则失败: %v", err)
		return nil, err
	}

	return &smsclient.UpdateProductFulfillmentRuleStatusResp{
		Rule: &smsclient.ProductFulfillmentRuleData{
			Id:         row.Id,
			RuleName:   row.RuleName,
			RuleStatus: row.RuleStatus,
			UpdateTime: row.UpdateTime,
		},
	}, nil
}
