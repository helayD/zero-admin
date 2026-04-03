package comment

import (
	"context"
	"errors"
	"strings"

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

type HandleCommentAppealLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHandleCommentAppealLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleCommentAppealLogic {
	return &HandleCommentAppealLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HandleCommentAppealLogic) HandleCommentAppeal(req *types.HandleCommentAppealReq) (resp *types.BaseResp, err error) {
	if req.Id == "" {
		return nil, errors.New("评价ID不能为空")
	}
	if req.AppealStatus != 2 && req.AppealStatus != 3 {
		return nil, errors.New("申诉处理状态非法")
	}
	if strings.TrimSpace(req.AppealReply) == "" {
		return nil, errors.New("申诉处理回复不能为空")
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

	_, err = l.svcCtx.CommentService.HandleCommentAppeal(l.ctx, &pmsclient.HandleCommentAppealReq{
		Id:           req.Id,
		AppealStatus: req.AppealStatus,
		AppealReply:  strings.TrimSpace(req.AppealReply),
		OperatorId:   userID,
		OperatorName: userName,
		PlatformId:   currentScope.PlatformID,
		TenantId:     currentScope.TenantID,
		MerchantId:   currentScope.MerchantID,
	})
	if err != nil {
		logc.Errorf(l.ctx, "处理评价申诉失败,参数:%+v,异常:%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	return res.Success()
}
