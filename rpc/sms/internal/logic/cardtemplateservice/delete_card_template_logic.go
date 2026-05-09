package cardtemplateservice

import (
	"context"
	"errors"
	"time"

	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteCardTemplateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteCardTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCardTemplateLogic {
	return &DeleteCardTemplateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteCardTemplate 删除卡片模板（软删除）
/*
Author: Cascade (受 Story 10.10 委托)
Date: 2026-05-09

被任何发卡规则引用时拒绝删除。
*/
func (l *DeleteCardTemplateLogic) DeleteCardTemplate(in *smsclient.DeleteCardTemplateReq) (*smsclient.DeleteCardTemplateResp, error) {
	if in.Id <= 0 {
		return nil, errors.New("模板ID无效")
	}
	platformId, tenantId, merchantId := resolveScope(in.Scope)

	// 1. 模板存在性
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

	// 2. 引用检查：被发卡规则引用则拒绝
	var refRuleCount int64
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_product_fulfillment_rule").
		Where("card_template_id = ? AND is_deleted = 0", in.Id).
		Count(&refRuleCount).Error; err != nil {
		logc.Errorf(l.ctx, "检查模板被发卡规则引用失败: %v", err)
		return nil, err
	}
	if refRuleCount > 0 {
		return nil, errors.New("该模板已被发卡规则引用，无法删除，请先解绑或删除相关发卡规则")
	}

	// 3. 软删除
	now := time.Now()
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_card_template").
		Where("id = ? AND platform_id = ? AND tenant_id = ? AND merchant_id = ? AND is_deleted = 0",
			in.Id, platformId, tenantId, merchantId).
		Updates(map[string]interface{}{
			"is_deleted":  1,
			"update_time": now,
		}).Error; err != nil {
		logc.Errorf(l.ctx, "删除卡片模板失败: %v", err)
		return nil, errors.New("删除卡片模板失败")
	}

	return &smsclient.DeleteCardTemplateResp{Code: 0, Message: "删除成功"}, nil
}
