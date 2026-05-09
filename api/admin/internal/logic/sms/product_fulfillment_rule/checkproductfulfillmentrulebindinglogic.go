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

// CheckProductFulfillmentRuleBindingLogic 检查发卡规则绑定情况
/*
Author: LiuFeiHua
Date: 2026/05/06 10:40:14
*/
type CheckProductFulfillmentRuleBindingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCheckProductFulfillmentRuleBindingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckProductFulfillmentRuleBindingLogic {
	return &CheckProductFulfillmentRuleBindingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CheckProductFulfillmentRuleBinding 检查发卡规则绑定情况
func (l *CheckProductFulfillmentRuleBindingLogic) CheckProductFulfillmentRuleBinding(req *types.QueryProductFulfillmentRuleDetailReq) (resp *types.CheckProductFulfillmentRuleBindingResp, err error) {
	// 解析治理范围
	readScope, err := common.ResolveQueryGovernanceScope(l.ctx, common.RequestedGovernanceScope{
		ScopeType:  req.ScopeType,
		PlatformID: req.PlatformId,
		TenantID:   req.TenantId,
		MerchantID: req.MerchantId,
	})
	if err != nil {
		return nil, err
	}

	ruleReq := &smsclient.CheckProductFulfillmentRuleBindingReq{
		Id:    req.Id,
		Scope: common.SMSGovernanceScope(readScope),
	}

	result, err := l.svcCtx.ProductFulfillmentRuleService.CheckProductFulfillmentRuleBinding(l.ctx, ruleReq)
	if err != nil {
		logc.Errorf(l.ctx, "检查发卡规则绑定情况失败,参数：%+v,响应：%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	return &types.CheckProductFulfillmentRuleBindingResp{
		Code:    "000000",
		Message: "查询成功",
		Data: types.CheckProductFulfillmentRuleData{
			BindingCount:        result.BindingCount,
			BindingProductNames: result.BindingProductNames,
			CanDisable:          result.CanDisable,
			CanDelete:           result.CanDelete,
			Message:             result.Message,
		},
	}, nil
}
