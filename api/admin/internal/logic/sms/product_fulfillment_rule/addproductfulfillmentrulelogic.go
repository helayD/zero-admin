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

// AddProductFulfillmentRuleLogic 添加发卡规则
/*
Author: LiuFeiHua
Date: 2026/05/06 10:40:14
*/
type AddProductFulfillmentRuleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddProductFulfillmentRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddProductFulfillmentRuleLogic {
	return &AddProductFulfillmentRuleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AddProductFulfillmentRule 添加发卡规则
func (l *AddProductFulfillmentRuleLogic) AddProductFulfillmentRule(req *types.AddProductFulfillmentRuleReq) (resp *types.BaseResp, err error) {
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

	logc.Infof(l.ctx, "添加发卡规则，操作人：%d，规则名：%s，治理范围：%+v", userId, req.RuleName, writeScope)

	ruleReq := &smsclient.AddProductFulfillmentRuleReq{
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

	_, err = l.svcCtx.ProductFulfillmentRuleService.AddProductFulfillmentRule(l.ctx, ruleReq)
	if err != nil {
		logc.Errorf(l.ctx, "添加发卡规则失败,参数：%+v,响应：%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	return &types.BaseResp{
		Code:    "000000",
		Message: "添加发卡规则成功",
	}, nil
}
