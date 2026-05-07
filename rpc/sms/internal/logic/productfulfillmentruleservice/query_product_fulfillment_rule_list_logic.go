package productfulfillmentruleservice

import (
	"context"
	"fmt"

	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryProductFulfillmentRuleListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryProductFulfillmentRuleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryProductFulfillmentRuleListLogic {
	return &QueryProductFulfillmentRuleListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *QueryProductFulfillmentRuleListLogic) QueryProductFulfillmentRuleList(in *smsclient.QueryProductFulfillmentRuleListReq) (*smsclient.QueryProductFulfillmentRuleListResp, error) {
	// 1. 解析治理范围
	platformId := int64(1)
	tenantId := int64(0)
	merchantId := int64(0)
	if in.Scope != nil {
		platformId = in.Scope.PlatformId
		tenantId = in.Scope.TenantId
		merchantId = in.Scope.MerchantId
	}

	// 2. 构建查询
	page := in.Page
	if page <= 0 {
		page = 1
	}
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	query := l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_product_fulfillment_rule").
		Where("platform_id = ? AND tenant_id = ? AND merchant_id = ? AND is_deleted = 0",
			platformId, tenantId, merchantId)

	// 3. 添加筛选条件
	if in.RuleName != "" {
		query = query.Where("rule_name LIKE ?", "%"+in.RuleName+"%")
	}
	if in.CardTemplateId > 0 {
		query = query.Where("card_template_id = ?", in.CardTemplateId)
	}
	if in.RuleStatus >= 0 {
		query = query.Where("rule_status = ?", in.RuleStatus)
	}

	// 4. 查询总数
	var total int64
	err := query.Count(&total).Error
	if err != nil {
		logc.Errorf(l.ctx, "查询发卡规则总数失败: %v", err)
		return nil, err
	}

	// 5. 查询列表
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

	var rows []ruleRow
	err = query.
		Select("id, rule_name, rule_status, card_template_id, expire_days, transferable, transfer_limit, claim_condition, redemption_condition, refund_policy, platform_id, tenant_id, merchant_id, create_by, create_time, update_by, update_time").
		Order("id DESC").
		Offset(int((page - 1) * pageSize)).
		Limit(int(pageSize)).
		Find(&rows).Error
	if err != nil {
		logc.Errorf(l.ctx, "查询发卡规则列表失败: %v", err)
		return nil, err
	}

	// 6. 转换结果
	var list []*smsclient.ProductFulfillmentRuleData
	for _, row := range rows {
		// 查询绑定商品数量
		var bindingCount int64
		l.svcCtx.DB.WithContext(l.ctx).
			Table("sms_product_fulfillment_binding").
			Where("rule_id = ? AND is_deleted = 0", row.Id).
			Count(&bindingCount)

		list = append(list, &smsclient.ProductFulfillmentRuleData{
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
			CreateBy:            int64ToString(row.CreateBy),
			CreateTime:          row.CreateTime,
			UpdateBy:            int64ToString(row.UpdateBy),
			UpdateTime:          row.UpdateTime,
			BindingCount:        int32(bindingCount),
		})
	}

	return &smsclient.QueryProductFulfillmentRuleListResp{
		List:  list,
		Total: total,
	}, nil
}

func int64ToString(n int64) string {
	if n == 0 {
		return ""
	}
	return fmt.Sprintf("%d", n)
}
