package productfulfillmentruleservice

import (
	"context"
	"errors"
	"fmt"

	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type QueryProductFulfillmentRuleDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryProductFulfillmentRuleDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryProductFulfillmentRuleDetailLogic {
	return &QueryProductFulfillmentRuleDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *QueryProductFulfillmentRuleDetailLogic) QueryProductFulfillmentRuleDetail(in *smsclient.QueryProductFulfillmentRuleDetailReq) (*smsclient.QueryProductFulfillmentRuleDetailResp, error) {
	// 1. 解析治理范围
	platformId := int64(1)
	tenantId := int64(0)
	merchantId := int64(0)
	if in.Scope != nil {
		platformId = in.Scope.PlatformId
		tenantId = in.Scope.TenantId
		merchantId = in.Scope.MerchantId
	}

	// 2. 查询规则详情
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
		CreateBy            int64  `gorm:"column:create_by"`
		CreateTime          string `gorm:"column:create_time"`
		UpdateBy            int64  `gorm:"column:update_by"`
		UpdateTime          string `gorm:"column:update_time"`
	}

	var row ruleRow
	err := l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_product_fulfillment_rule").
		Select("id, rule_name, rule_status, card_template_id, expire_days, transferable, transfer_limit, claim_condition, redemption_condition, refund_policy, platform_id, tenant_id, merchant_id, create_by, create_time, update_by, update_time").
		Where("id = ? AND platform_id = ? AND tenant_id = ? AND merchant_id = ? AND is_deleted = 0",
			in.Id, platformId, tenantId, merchantId).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("发卡规则不存在")
		}
		logc.Errorf(l.ctx, "查询发卡规则详情失败: %v", err)
		return nil, err
	}

	// 3. 查询绑定商品数量
	var bindingCount int64
	l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_product_fulfillment_binding").
		Where("rule_id = ? AND is_deleted = 0", row.Id).
		Count(&bindingCount)

	// 4. 返回结果
	return &smsclient.QueryProductFulfillmentRuleDetailResp{
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
			CreateBy:            fmt.Sprintf("%d", row.CreateBy),
			CreateTime:          row.CreateTime,
			UpdateBy:            fmt.Sprintf("%d", row.UpdateBy),
			UpdateTime:          row.UpdateTime,
			BindingCount:        int32(bindingCount),
		},
	}, nil
}
