package product_fulfillment_rule

import (
	"context"

	"github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

// QueryProductFulfillmentRuleListLogic 查询发卡规则列表
/*
Author: LiuFeiHua
Date: 2026/05/06 10:40:14
*/
type QueryProductFulfillmentRuleListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryProductFulfillmentRuleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryProductFulfillmentRuleListLogic {
	return &QueryProductFulfillmentRuleListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// QueryProductFulfillmentRuleList 查询发卡规则列表
func (l *QueryProductFulfillmentRuleListLogic) QueryProductFulfillmentRuleList(req *types.QueryProductFulfillmentRuleListReq) (resp *types.QueryProductFulfillmentRuleListResp, err error) {
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

	// TODO: 调用 RPC 服务查询发卡规则列表
	// 当 RPC 服务创建后，替换为真实的调用
	logc.Infof(l.ctx, "查询发卡规则列表，治理范围：%+v，查询条件：%+v", readScope, req)

	// 临时返回空列表
	return &types.QueryProductFulfillmentRuleListResp{
		Code:    "000000",
		Message: "查询成功",
		Data: types.ProductFulfillmentRuleListData{
			List:     []types.ProductFulfillmentRuleData{},
			Total:    0,
			Page:     req.Page,
			PageSize: req.PageSize,
		},
	}, nil

	// 以下是 RPC 调用的示例代码，待 RPC 服务创建后启用
	/*
		ruleReq := &smsclient.QueryProductFulfillmentRuleListReq{
			RuleName:       req.RuleName,
			CardTemplateId: req.CardTemplateId,
			RuleStatus:     req.RuleStatus,
			Page:           req.Page,
			PageSize:       req.PageSize,
			Scope:          common.SMSGovernanceScope(readScope),
		}

		result, err := l.svcCtx.ProductFulfillmentRuleService.QueryProductFulfillmentRuleList(l.ctx, ruleReq)
		if err != nil {
			logc.Errorf(l.ctx, "查询发卡规则列表失败,参数：%+v,响应：%s", req, err.Error())
			return nil, err
		}

		var list []types.ProductFulfillmentRuleData
		for _, item := range result.List {
			list = append(list, types.ProductFulfillmentRuleData{
				Id:                  item.Id,
				RuleName:            item.RuleName,
				RuleStatus:          item.RuleStatus,
				CardTemplateId:      item.CardTemplateId,
				CardTemplateName:    item.CardTemplateName,
				ExpireDays:          item.ExpireDays,
				Transferable:        item.Transferable,
				TransferLimit:       item.TransferLimit,
				ClaimCondition:      item.ClaimCondition,
				RedemptionCondition: item.RedemptionCondition,
				RefundPolicy:        item.RefundPolicy,
				RefundPolicyText:    item.RefundPolicyText,
				PlatformId:          item.PlatformId,
				TenantId:            item.TenantId,
				MerchantId:          item.MerchantId,
				CreateBy:            item.CreateBy,
				UpdateBy:            item.UpdateBy,
				CreateTime:          item.CreateTime,
				UpdateTime:          item.UpdateTime,
				BindingCount:        item.BindingCount,
			})
		}

		return &types.QueryProductFulfillmentRuleListResp{
			Code:    "000000",
			Message: "查询成功",
			Data: types.ProductFulfillmentRuleListData{
				List:     list,
				Total:    result.Total,
				Page:     req.Page,
				PageSize: req.PageSize,
			},
		}, nil
	*/
}
