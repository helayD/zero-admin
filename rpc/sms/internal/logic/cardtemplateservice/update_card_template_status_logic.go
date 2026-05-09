package cardtemplateservice

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UpdateCardTemplateStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCardTemplateStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCardTemplateStatusLogic {
	return &UpdateCardTemplateStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateCardTemplateStatus 启停卡片模板
/*
Author: Cascade (受 Story 10.10 委托)
Date: 2026-05-09

注意：禁用模板只是一个软开关，不会影响已经创建的卡片实例，
但会拦截"新发卡规则保存时引用此模板"以及上层"提货卡商品发布"。
*/
func (l *UpdateCardTemplateStatusLogic) UpdateCardTemplateStatus(in *smsclient.UpdateCardTemplateStatusReq) (*smsclient.UpdateCardTemplateStatusResp, error) {
	if in.Id <= 0 {
		return nil, errors.New("模板ID无效")
	}
	if in.Status != 0 && in.Status != 1 {
		return nil, errors.New("模板状态无效，只能是 0（禁用）或 1（启用）")
	}

	platformId, tenantId, merchantId := resolveScope(in.Scope)

	// 1. 模板存在性校验
	var existing cardTemplateRow
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_card_template").
		Select("id, status").
		Where("id = ? AND platform_id = ? AND tenant_id = ? AND merchant_id = ? AND is_deleted = 0",
			in.Id, platformId, tenantId, merchantId).
		Take(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("卡片模板不存在")
		}
		logc.Errorf(l.ctx, "查询卡片模板失败: %v", err)
		return nil, err
	}

	// 2. Story 10.10 修复 H1: 禁用模板时硬拦截"被启用发卡规则引用"。
	//    前端 checkCardTemplateUsage 是软提示，可被 curl 绕过；后端必须自校验避免不一致。
	if in.Status == 0 {
		var refRuleCount int64
		if err := l.svcCtx.DB.WithContext(l.ctx).
			Table("sms_product_fulfillment_rule").
			Where("card_template_id = ? AND is_deleted = 0", in.Id).
			Count(&refRuleCount).Error; err != nil {
			logc.Errorf(l.ctx, "校验模板引用关系失败: %v", err)
			return nil, errors.New("校验模板引用关系失败")
		}
		if refRuleCount > 0 {
			return nil, fmt.Errorf("该模板已被 %d 条发卡规则引用，无法禁用，请先解绑或删除相关发卡规则", refRuleCount)
		}
	}

	// 3. UPDATE
	now := time.Now()
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_card_template").
		Where("id = ? AND platform_id = ? AND tenant_id = ? AND merchant_id = ? AND is_deleted = 0",
			in.Id, platformId, tenantId, merchantId).
		Updates(map[string]interface{}{
			"status":      in.Status,
			"update_time": now,
		}).Error; err != nil {
		logc.Errorf(l.ctx, "更新卡片模板状态失败: %v", err)
		return nil, errors.New("更新模板状态失败")
	}

	// 4. 取最新结果
	var row cardTemplateRow
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_card_template").
		Select(cardTemplateSelectColumns).
		Where("id = ?", in.Id).
		Take(&row).Error; err != nil {
		logc.Errorf(l.ctx, "查询更新后的模板失败: %v", err)
		return nil, errors.New("查询模板失败")
	}

	return &smsclient.UpdateCardTemplateStatusResp{Template: row.toProto(0)}, nil
}
