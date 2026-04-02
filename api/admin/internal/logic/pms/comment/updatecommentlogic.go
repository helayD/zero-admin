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
	"google.golang.org/grpc/status"

	"github.com/zeromicro/go-zero/core/logx"
)

// UpdateCommentLogic 审核/屏蔽/恢复评价
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

// UpdateComment 审核/屏蔽/恢复评价
func (l *UpdateCommentLogic) UpdateComment(req *types.UpdateCommentReq) (*types.BaseResp, error) {
	if req.Id == "" {
		return nil, errors.New("评价ID不能为空")
	}

	userName, err := common.GetUserName(l.ctx)
	if err != nil {
		return nil, err
	}

	currentScope, err := common.CurrentGovernanceScope(l.ctx)
	if err != nil {
		return nil, err
	}

	// 先查询评价详情，获取 productId 和现有字段（带 scope 校验）
	detail, err := l.svcCtx.CommentService.QueryCommentDetail(l.ctx, &pmsclient.QueryCommentDetailReq{
		Id:         req.Id,
		PlatformId: currentScope.PlatformID,
		TenantId:   currentScope.TenantID,
		MerchantId: currentScope.MerchantID,
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询评价详情失败,ID:%s,异常:%s", req.Id, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	_, err = l.svcCtx.CommentService.UpdateComment(l.ctx, &pmsclient.UpdateCommentReq{
		Id:               req.Id,
		ProductId:        detail.ProductId,
		MemberNickName:   detail.MemberNickName,
		ProductName:      detail.ProductName,
		Star:             detail.Star,
		MemberIp:         detail.MemberIp,
		ShowStatus:       req.ShowStatus,
		ProductAttribute: detail.ProductAttribute,
		CollectCount:     detail.CollectCount,
		ReadCount:        detail.ReadCount,
		Content:          detail.Content,
		Pics:             detail.Pics,
		MemberIcon:       detail.MemberIcon,
		ReplayCount:      detail.ReplayCount,
		UpdateBy:         userName,
		// Review Fix H-NEW-2: 注入治理 scope 防止越权
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
