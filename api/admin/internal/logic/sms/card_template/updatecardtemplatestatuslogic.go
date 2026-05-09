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

type UpdateCardTemplateStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateCardTemplateStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCardTemplateStatusLogic {
	return &UpdateCardTemplateStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateCardTemplateStatusLogic) UpdateCardTemplateStatus(req *types.UpdateCardTemplateStatusReq) (resp *types.BaseResp, err error) {
	userId, err := common.GetUserId(l.ctx)
	if err != nil {
		return nil, err
	}

	writeScope, err := common.ResolveWriteGovernanceScope(l.ctx, common.RequestedGovernanceScope{
		ScopeType:  req.ScopeType,
		PlatformID: req.PlatformId,
		TenantID:   req.TenantId,
		MerchantID: req.MerchantId,
	})
	if err != nil {
		return nil, err
	}

	logc.Infof(l.ctx, "更新卡片模板状态，操作人：%d，模板ID：%d，状态：%d", userId, req.Id, req.Status)

	if _, err = l.svcCtx.CardTemplateService.UpdateCardTemplateStatus(l.ctx, &smsclient.UpdateCardTemplateStatusReq{
		Id:           req.Id,
		Status:       req.Status,
		Scope:        common.SMSGovernanceScope(writeScope),
		OperatorType: "admin",
	}); err != nil {
		logc.Errorf(l.ctx, "更新卡片模板状态失败,参数：%+v,响应：%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	return &types.BaseResp{Code: "000000", Message: "更新卡片模板状态成功"}, nil
}
