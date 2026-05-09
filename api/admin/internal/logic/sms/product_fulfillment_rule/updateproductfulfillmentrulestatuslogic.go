package product_fulfillment_rule

import (
	"context"

	"github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"google.golang.org/grpc/status"

	"github.com/zeromicro/go-zero/core/logx"
)

// UpdateProductFulfillmentRuleStatusLogic 更新发卡规则状态
/*
Author: LiuFeiHua
Date: 2026/05/06 10:40:14
*/
type UpdateProductFulfillmentRuleStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateProductFulfillmentRuleStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProductFulfillmentRuleStatusLogic {
	return &UpdateProductFulfillmentRuleStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UpdateProductFulfillmentRuleStatus 更新发卡规则状态
func (l *UpdateProductFulfillmentRuleStatusLogic) UpdateProductFulfillmentRuleStatus(req *types.UpdateProductFulfillmentRuleStatusReq) (resp *types.BaseResp, err error) {
	userId, err := common.GetUserId(l.ctx)
	if err != nil {
		return nil, err
	}

	// 解析治理范围
	writeScope, err := common.ResolveWriteGovernanceScope(l.ctx, common.RequestedGovernanceScope{
		ScopeType:  req.ScopeType,
		PlatformID: req.PlatformId,
		TenantID:   req.TenantId,
		MerchantID: req.MerchantId,
	})
	if err != nil {
		return nil, err
	}

	logc.Infof(l.ctx, "更新发卡规则状态，操作人：%d，规则ID：%d，新状态：%d，治理范围：%+v", userId, req.Id, req.RuleStatus, writeScope)

	ruleReq := &smsclient.UpdateProductFulfillmentRuleStatusReq{
		Id:           req.Id,
		RuleStatus:   req.RuleStatus,
		Scope:        common.SMSGovernanceScope(writeScope),
		OperatorType: "admin",
	}

	_, err = l.svcCtx.ProductFulfillmentRuleService.UpdateProductFulfillmentRuleStatus(l.ctx, ruleReq)
	if err != nil {
		logc.Errorf(l.ctx, "更新发卡规则状态失败,参数：%+v,响应：%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	return &types.BaseResp{
		Code:    "000000",
		Message: "更新发卡规则状态成功",
	}, nil
}
