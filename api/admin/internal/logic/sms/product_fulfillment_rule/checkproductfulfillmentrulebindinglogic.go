package product_fulfillment_rule

import (
	"context"

	"github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/zeromicro/go-zero/core/logc"
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

	// TODO: 调用 RPC 服务检查发卡规则绑定情况
	// 当 RPC 服务创建后，替换为真实的调用
	logc.Infof(l.ctx, "检查发卡规则绑定情况，规则ID：%d，治理范围：%+v", req.Id, readScope)

	// 临时返回无绑定
	return &types.CheckProductFulfillmentRuleBindingResp{
		Code:    "000000",
		Message: "查询成功",
		Data: types.CheckProductFulfillmentRuleData{
			BindingCount:        0,
			BindingProductNames: []string{},
			CanDisable:          true,
			CanDelete:           true,
			Message:             "该规则未绑定任何商品，可以禁用或删除",
		},
	}, nil

	// 以下是 RPC 调用的示例代码，待 RPC 服务创建后启用
	/*
		ruleReq := &smsclient.CheckProductFulfillmentRuleBindingReq{
			Id:    req.Id,
			Scope: common.SMSGovernanceScope(readScope),
		}

		result, err := l.svcCtx.ProductFulfillmentRuleService.CheckProductFulfillmentRuleBinding(l.ctx, ruleReq)
		if err != nil {
			logc.Errorf(l.ctx, "检查发卡规则绑定情况失败,参数：%+v,响应：%s", req, err.Error())
			return nil, err
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
	*/
}
