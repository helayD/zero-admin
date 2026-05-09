package cardtemplateservice

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type QueryCardTemplateDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryCardTemplateDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryCardTemplateDetailLogic {
	return &QueryCardTemplateDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// QueryCardTemplateDetail 查询卡片模板详情
/*
Author: Cascade (受 Story 10.10 委托)
Date: 2026-05-09
*/
func (l *QueryCardTemplateDetailLogic) QueryCardTemplateDetail(in *smsclient.QueryCardTemplateDetailReq) (*smsclient.QueryCardTemplateDetailResp, error) {
	if in.Id <= 0 {
		return nil, errors.New("模板ID无效")
	}
	platformId, tenantId, merchantId := resolveScope(in.Scope)

	var row cardTemplateRow
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_card_template").
		Select(cardTemplateSelectColumns).
		Where("id = ? AND platform_id = ? AND tenant_id = ? AND merchant_id = ? AND is_deleted = 0",
			in.Id, platformId, tenantId, merchantId).
		Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("卡片模板不存在")
		}
		logc.Errorf(l.ctx, "查询卡片模板详情失败: %v", err)
		return nil, err
	}

	var refRuleCount int64
	l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_product_fulfillment_rule").
		Where("card_template_id = ? AND is_deleted = 0", row.Id).
		Count(&refRuleCount)

	return &smsclient.QueryCardTemplateDetailResp{
		Template: row.toProto(int32(refRuleCount)),
	}, nil
}
