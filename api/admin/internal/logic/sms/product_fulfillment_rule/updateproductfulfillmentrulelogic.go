package product_fulfillment_rule

import (
	"context"

	"github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

// UpdateProductFulfillmentRuleLogic 更新发卡规则
/*
Author: LiuFeiHua
Date: 2026/05/06 10:40:14
*/
type UpdateProductFulfillmentRuleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateProductFulfillmentRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProductFulfillmentRuleLogic {
	return &UpdateProductFulfillmentRuleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UpdateProductFulfillmentRule 更新发卡规则
func (l *UpdateProductFulfillmentRuleLogic) UpdateProductFulfillmentRule(req *types.UpdateProductFulfillmentRuleReq) (resp *types.BaseResp, err error) {
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

	// TODO: 调用 RPC 服务更新发卡规则
	// 当 RPC 服务创建后，替换为真实的调用
	logc.Infof(l.ctx, "更新发卡规则，操作人：%d，规则ID：%d，治理范围：%+v", userId, req.Id, writeScope)

	// 临时返回成功响应
	return &types.BaseResp{
		Code:    "000000",
		Message: "更新发卡规则成功",
	}, nil

	// 以下是 RPC 调用的示例代码，待 RPC 服务创建后启用
	/*
		ruleReq := &smsclient.UpdateProductFulfillmentRuleReq{
			Id:                  req.Id,
			RuleName:            req.RuleName,
			CardTemplateId:      req.CardTemplateId,
			ExpireDays:          req.ExpireDays,
			Transferable:        req.Transferable,
			TransferLimit:       req.TransferLimit,
			ClaimCondition:      req.ClaimCondition,
			RedemptionCondition: req.RedemptionCondition,
			RefundPolicy:        req.RefundPolicy,
			Scope:               common.SMSGovernanceScope(writeScope),
			OperatorType:        "admin",
		}

		_, err = l.svcCtx.ProductFulfillmentRuleService.UpdateProductFulfillmentRule(l.ctx, ruleReq)
		if err != nil {
			logc.Errorf(l.ctx, "更新发卡规则失败,参数：%+v,响应：%s", req, err.Error())
			s, _ := status.FromError(err)
			return nil, errorx.NewDefaultError(s.Message())
		}

		return &types.BaseResp{
			Code:    "000000",
			Message: "更新发卡规则成功",
		}, nil
	*/
}
