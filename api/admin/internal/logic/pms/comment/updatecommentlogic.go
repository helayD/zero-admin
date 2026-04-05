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

// UpdateCommentLogic 审核/屏蔽评价（兼容旧批量入口）
/*
Author: LiuFeiHua
Date: 2026/04/02
*/
type UpdateCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCommentLogic {
	return &UpdateCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UpdateComment 审核/屏蔽评价（兼容旧 UI 批量操作）
func (l *UpdateCommentLogic) UpdateComment(req *types.UpdateCommentReq) (*types.BaseResp, error) {
	if req.Id == "" && req.Ids == "" {
		return nil, errors.New("评价ID不能为空")
	}
	if req.ShowStatus != 0 && req.ShowStatus != 1 {
		return nil, errors.New("评价状态非法")
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

	_, err = l.svcCtx.CommentService.UpdateComment(l.ctx, &pmsclient.UpdateCommentReq{
		Id:         req.Id,
		Ids:        req.Ids,
		ShowStatus: req.ShowStatus,
		AuditorId:  userID,
		UpdateBy:   userName,
		PlatformId: currentScope.PlatformID,
		TenantId:   currentScope.TenantID,
		MerchantId: currentScope.MerchantID,
	})
	if err != nil {
		logc.Errorf(l.ctx, "更新商品评价状态失败,参数:%+v,异常:%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	return res.Success()
}
