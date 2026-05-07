package product_fulfillment_rule

import (
	"context"

	"github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

// QueryProductFulfillmentRuleDetailLogic 查询发卡规则详情
/*
Author: LiuFeiHua
Date: 2026/05/06 10:40:14
*/
type QueryProductFulfillmentRuleDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryProductFulfillmentRuleDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryProductFulfillmentRuleDetailLogic {
	return &QueryProductFulfillmentRuleDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// QueryProductFulfillmentRuleDetail 查询发卡规则详情
func (l *QueryProductFulfillmentRuleDetailLogic) QueryProductFulfillmentRuleDetail(req *types.QueryProductFulfillmentRuleDetailReq) (resp *types.QueryProductFulfillmentRuleDetailResp, err error) {
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

	// TODO: 调用 RPC 服务查询发卡规则详情
	// 当 RPC 服务创建后，替换为真实的调用
	logc.Infof(l.ctx, "查询发卡规则详情，规则ID：%d，治理范围：%+v", req.Id, readScope)

	// 临时返回空详情
	return &types.QueryProductFulfillmentRuleDetailResp{
		Code:    "000000",
		Message: "查询成功",
		Data:    types.ProductFulfillmentRuleData{},
	}, nil

	// 以下是 RPC 调用的示例代码，待 RPC 服务创建后启用
	/*
		ruleReq := &smsclient.QueryProductFulfillmentRuleDetailReq{
			Id:    req.Id,
			Scope: common.SMSGovernanceScope(readScope),
		}

		result, err := l.svcCtx.ProductFulfillmentRuleService.QueryProductFulfillmentRuleDetail(l.ctx, ruleReq)
		if err != nil {
			logc.Errorf(l.ctx, "查询发卡规则详情失败,参数：%+v,响应：%s", req, err.Error())
			return nil, err
		}

		return &types.QueryProductFulfillmentRuleDetailResp{
			Code:    "000000",
			Message: "查询成功",
			Data: types.ProductFulfillmentRuleData{
				Id:                  result.Rule.Id,
				RuleName:            result.Rule.RuleName,
				RuleStatus:          result.Rule.RuleStatus,
				CardTemplateId:      result.Rule.CardTemplateId,
				CardTemplateName:    result.Rule.CardTemplateName,
				ExpireDays:          result.Rule.ExpireDays,
				Transferable:        result.Rule.Transferable,
				TransferLimit:       result.Rule.TransferLimit,
				ClaimCondition:      result.Rule.ClaimCondition,
				RedemptionCondition: result.Rule.RedemptionCondition,
				RefundPolicy:        result.Rule.RefundPolicy,
				RefundPolicyText:    result.Rule.RefundPolicyText,
				PlatformId:          result.Rule.PlatformId,
				TenantId:            result.Rule.TenantId,
				MerchantId:          result.Rule.MerchantId,
				CreateBy:            result.Rule.CreateBy,
				UpdateBy:            result.Rule.UpdateBy,
				CreateTime:          result.Rule.CreateTime,
				UpdateTime:          result.Rule.UpdateTime,
				BindingCount:        result.Rule.BindingCount,
			},
		}, nil
	*/
}
