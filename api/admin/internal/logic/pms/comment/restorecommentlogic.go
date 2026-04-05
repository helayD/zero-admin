package comment

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/common/res"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/status"
)

type RestoreCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRestoreCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RestoreCommentLogic {
	return &RestoreCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RestoreCommentLogic) RestoreComment(req *types.RestoreCommentReq) (resp *types.BaseResp, err error) {
	if req.Id == "" {
		return nil, errors.New("评价ID不能为空")
	}

	userID, err := common.GetUserId(l.ctx)
	if err != nil {
		return nil, err
	}
	userName, err := common.GetUserName(l.ctx)
	if err != nil {
		return nil, err
	}
	currentScope, err := common.CurrentGovernanceScope(l.ctx)
	if err != nil {
		return nil, err
	}

	_, err = l.svcCtx.CommentService.RestoreComment(l.ctx, &pmsclient.RestoreCommentReq{
		Id:           req.Id,
		PlatformId:   currentScope.PlatformID,
		TenantId:     currentScope.TenantID,
		MerchantId:   currentScope.MerchantID,
		AuditorId:    userID,
		OperatorName: userName,
	})
	if err != nil {
		logc.Errorf(l.ctx, "恢复评价失败,参数:%+v,异常:%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	return res.Success()
}
