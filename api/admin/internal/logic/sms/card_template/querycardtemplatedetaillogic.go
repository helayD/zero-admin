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

type QueryCardTemplateDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryCardTemplateDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryCardTemplateDetailLogic {
	return &QueryCardTemplateDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryCardTemplateDetailLogic) QueryCardTemplateDetail(req *types.QueryCardTemplateDetailReq) (resp *types.QueryCardTemplateDetailResp, err error) {
	readScope, err := common.ResolveQueryGovernanceScope(l.ctx, common.RequestedGovernanceScope{
		ScopeType:  req.ScopeType,
		PlatformID: req.PlatformId,
		TenantID:   req.TenantId,
		MerchantID: req.MerchantId,
	})
	if err != nil {
		return nil, err
	}

	result, err := l.svcCtx.CardTemplateService.QueryCardTemplateDetail(l.ctx, &smsclient.QueryCardTemplateDetailReq{
		Id:    req.Id,
		Scope: common.SMSGovernanceScope(readScope),
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询卡片模板详情失败,参数：%+v,响应：%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	return &types.QueryCardTemplateDetailResp{
		Code:    "000000",
		Message: "查询成功",
		Data:    toCardTemplateData(result.Template),
	}, nil
}
