package card_template

import (
	"context"

	"github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/status"
)

type CheckCardTemplateUsageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCheckCardTemplateUsageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckCardTemplateUsageLogic {
	return &CheckCardTemplateUsageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CheckCardTemplateUsageLogic) CheckCardTemplateUsage(req *types.CheckCardTemplateUsageReq) (resp *types.CheckCardTemplateUsageResp, err error) {
	readScope, err := common.ResolveQueryGovernanceScope(l.ctx, common.RequestedGovernanceScope{
		ScopeType:  req.ScopeType,
		PlatformID: req.PlatformId,
		TenantID:   req.TenantId,
		MerchantID: req.MerchantId,
	})
	if err != nil {
		return nil, err
	}

	result, err := l.svcCtx.CardTemplateService.CheckCardTemplateUsage(l.ctx, &smsclient.CheckCardTemplateUsageReq{
		Id:    req.Id,
		Scope: common.SMSGovernanceScope(readScope),
	})
	if err != nil {
		logc.Errorf(l.ctx, "检查卡片模板被引用失败,参数：%+v,响应：%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	return &types.CheckCardTemplateUsageResp{
		Code:    "000000",
		Message: "查询成功",
		Data: types.CheckCardTemplateUsage{
			RefRuleCount: result.RefRuleCount,
			RefRuleNames: result.RefRuleNames,
			CanDisable:   result.CanDisable,
			CanDelete:    result.CanDelete,
			Message:      result.Message,
		},
	}, nil
}
