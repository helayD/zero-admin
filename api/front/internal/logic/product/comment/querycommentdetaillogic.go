package comment

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/pkg/errorx"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"google.golang.org/grpc/status"

	"github.com/zeromicro/go-zero/core/logx"
)

// QueryCommentDetailLogic 查询商品评价详情
/*
Author: LiuFeiHua
Date: 2026/04/02
*/
type QueryCommentDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryCommentDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryCommentDetailLogic {
	return &QueryCommentDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// QueryCommentDetail 查询商品评价详情（包含回复列表）
func (l *QueryCommentDetailLogic) QueryCommentDetail(req *types.QueryCommentDetailReq) (*types.QueryCommentDetailResp, error) {
	scope, err := common.CurrentGovernanceScope(l.ctx)
	if err != nil {
		return nil, err
	}

	// 查询评价详情（带 scope 校验，防止越权访问）
	detail, err := l.svcCtx.CommentService.QueryCommentDetail(l.ctx, &pmsclient.QueryCommentDetailReq{
		Id:         req.Id,
		PlatformId: scope.PlatformID,
		TenantId:   scope.TenantID,
		MerchantId: scope.MerchantID,
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询评价详情失败,ID:%s,异常:%s", req.Id, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	// Review Fix R-2: productId 归属校验（AC-2: 不显示不属于当前商品的评价记录）
	if req.ProductId > 0 && detail.ProductId != req.ProductId {
		return nil, errors.New("无权访问该评价")
	}

	// 查询回复列表（使用评论的 ObjectID）
	replayResult, err := l.svcCtx.CommentReplayService.QueryCommentReplayList(l.ctx, &pmsclient.QueryCommentReplayListReq{
		PageNum:   1,
		PageSize:  100, // 回复一般不多
		CommentId:  detail.Id, // 传入评论的 ObjectID（非 req.Id）
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询评价回复列表失败,ID:%s,异常:%s", req.Id, err.Error())
		// 回复查询失败不影响主数据
	}

	var replays []types.CommentReplayItem
	if replayResult != nil {
		for _, item := range replayResult.List {
			replays = append(replays, types.CommentReplayItem{
				Id:             item.Id,
				CommentId:      item.CommentId,
				Type:           int(item.Type),
				MemberNickName: item.MemberNickName,
				MemberIcon:     item.MemberIcon,
				Content:        item.Content,
				CreateTime:     item.CreateTime,
			})
		}
	}

	return &types.QueryCommentDetailResp{
		Code:    0,
		Message: "查询成功",
		Data: types.CommentListItem{
			Id:               detail.Id,
			ProductId:        detail.ProductId,
			MemberId:         detail.MemberId,
			MemberNickName:   detail.MemberNickName,
			MemberIcon:       detail.MemberIcon,
			Star:             int(detail.Star),
			Content:          detail.Content,
			Pics:             detail.Pics,
			ProductAttribute: detail.ProductAttribute,
			ShowStatus:       int(detail.ShowStatus),
			ReplayCount:      int(detail.ReplayCount),
			MemberIp:         detail.MemberIp,
			CreateTime:       detail.CreateTime,
		},
		Replays: replays,
	}, nil
}
