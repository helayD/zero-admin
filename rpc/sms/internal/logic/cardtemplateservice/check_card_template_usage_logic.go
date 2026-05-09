package cardtemplateservice

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

type CheckCardTemplateUsageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckCardTemplateUsageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckCardTemplateUsageLogic {
	return &CheckCardTemplateUsageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CheckCardTemplateUsage 检查卡片模板是否被发卡规则引用
/*
Author: Cascade (受 Story 10.10 委托)
Date: 2026-05-09

供前端在删除/禁用前调用，返回引用计数与引用规则名称列表。
*/
func (l *CheckCardTemplateUsageLogic) CheckCardTemplateUsage(in *smsclient.CheckCardTemplateUsageReq) (*smsclient.CheckCardTemplateUsageResp, error) {
	if in.Id <= 0 {
		return nil, errors.New("模板ID无效")
	}
	platformId, tenantId, merchantId := resolveScope(in.Scope)

	// 模板存在性
	var count int64
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_card_template").
		Where("id = ? AND platform_id = ? AND tenant_id = ? AND merchant_id = ? AND is_deleted = 0",
			in.Id, platformId, tenantId, merchantId).
		Count(&count).Error; err != nil {
		logc.Errorf(l.ctx, "检查卡片模板存在性失败: %v", err)
		return nil, err
	}
	if count == 0 {
		return nil, errors.New("卡片模板不存在")
	}

	// 查询引用规则
	type refRow struct {
		Id       int64  `gorm:"column:id"`
		RuleName string `gorm:"column:rule_name"`
	}
	var rows []refRow
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_product_fulfillment_rule").
		Select("id, rule_name").
		Where("card_template_id = ? AND is_deleted = 0", in.Id).
		Find(&rows).Error; err != nil {
		logc.Errorf(l.ctx, "查询模板被引用列表失败: %v", err)
		return nil, err
	}

	refRuleNames := make([]string, 0, len(rows))
	for _, r := range rows {
		refRuleNames = append(refRuleNames, r.RuleName)
	}
	refCount := int32(len(rows))
	canDisable := refCount == 0
	canDelete := refCount == 0
	message := "该模板未被任何发卡规则引用，可以禁用或删除"
	if refCount > 0 {
		message = "该模板已被发卡规则引用，无法禁用或删除"
	}

	return &smsclient.CheckCardTemplateUsageResp{
		RefRuleCount: refCount,
		RefRuleNames: refRuleNames,
		CanDisable:   canDisable,
		CanDelete:    canDelete,
		Message:      message,
	}, nil
}
