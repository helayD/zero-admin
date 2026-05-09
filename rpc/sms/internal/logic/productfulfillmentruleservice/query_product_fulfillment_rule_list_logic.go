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

	// 注意 (Story 10.10 Task 3.6): 主表别名为 r，LEFT JOIN sms_card_template 取 template_name 透出。
	query := l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_product_fulfillment_rule AS r").
		Where("r.platform_id = ? AND r.tenant_id = ? AND r.merchant_id = ? AND r.is_deleted = 0",
			platformId, tenantId, merchantId)

	// 3. 添加筛选条件
	if in.RuleName != "" {
		query = query.Where("r.rule_name LIKE ?", "%"+in.RuleName+"%")
	}
	if in.CardTemplateId > 0 {
		query = query.Where("r.card_template_id = ?", in.CardTemplateId)
	}
	if in.RuleStatus >= 0 {
		query = query.Where("r.rule_status = ?", in.RuleStatus)
	}

	// 4. 查询总数
	var total int64
	err := query.Count(&total).Error
	if err != nil {
		logc.Errorf(l.ctx, "查询发卡规则总数失败: %v", err)
		return nil, err
	}

	// 5. 查询列表（JOIN 取 cardTemplateName）
	type ruleRow struct {
		Id                  int64  `gorm:"column:id"`
		RuleName            string `gorm:"column:rule_name"`
		RuleStatus          int32  `gorm:"column:rule_status"`
		CardTemplateId      int64  `gorm:"column:card_template_id"`
		CardTemplateName    string `gorm:"column:card_template_name"`
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
		Joins("LEFT JOIN sms_card_template AS t ON t.id = r.card_template_id AND t.is_deleted = 0").
		Select("r.id, r.rule_name, r.rule_status, r.card_template_id, t.template_name AS card_template_name, " +
			"r.expire_days, r.transferable, r.transfer_limit, r.claim_condition, r.redemption_condition, r.refund_policy, " +
			"r.platform_id, r.tenant_id, r.merchant_id, r.create_by, r.create_time, r.update_by, r.update_time").
		Order("r.id DESC").
		Offset(int((page - 1) * pageSize)).
		Limit(int(pageSize)).
		Find(&rows).Error
	if err != nil {
		logc.Errorf(l.ctx, "查询发卡规则列表失败: %v", err)
		return nil, err
	}

	// 6. Story 10.10 修复 M3: 批量回填绑定商品数量，避免 N+1。
	ruleIds := make([]int64, 0, len(rows))
	for i := range rows {
		ruleIds = append(ruleIds, rows[i].Id)
	}
	bindingCounts := make(map[int64]int32, len(ruleIds))
	if len(ruleIds) > 0 {
		type bindingCountRow struct {
			RuleId int64 `gorm:"column:rule_id"`
			Cnt    int64 `gorm:"column:cnt"`
		}
		var bindingRows []bindingCountRow
		if err := l.svcCtx.DB.WithContext(l.ctx).
			Table("sms_product_fulfillment_binding").
			Select("rule_id, COUNT(*) AS cnt").
			Where("rule_id IN ? AND is_deleted = 0", ruleIds).
			Group("rule_id").
			Find(&bindingRows).Error; err != nil {
			logc.Errorf(l.ctx, "批量查询规则绑定商品数失败: %v", err)
			// 失败不阻塞列表，按 0 兜底
		} else {
			for _, b := range bindingRows {
				bindingCounts[b.RuleId] = int32(b.Cnt)
			}
		}
	}

	// 7. 转换结果
	list := make([]*smsclient.ProductFulfillmentRuleData, 0, len(rows))
	for _, row := range rows {
		list = append(list, &smsclient.ProductFulfillmentRuleData{
			Id:                  row.Id,
			RuleName:            row.RuleName,
			RuleStatus:          row.RuleStatus,
			CardTemplateId:      row.CardTemplateId,
			CardTemplateName:    row.CardTemplateName,
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
			BindingCount:        bindingCounts[row.Id],
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
