package comment

import (
	"context"

	"github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"google.golang.org/grpc/status"

	"github.com/zeromicro/go-zero/core/logx"
)

// QueryCommentDetailLogic 查询评价详情
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

type QueryCommentDetailReq = types.QueryCommentDetailReq

// QueryCommentDetail 查询评价详情（包含回复列表）
func (l *QueryCommentDetailLogic) QueryCommentDetail(req *QueryCommentDetailReq) (*types.QueryCommentDetailResp, error) {
	currentScope, err := common.CurrentGovernanceScope(l.ctx)
	if err != nil {
		return nil, err
	}

	// 查询评价详情（带 scope 校验）
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

	// 查询回复列表
	replayResult, err := l.svcCtx.CommentReplayService.QueryCommentReplayList(l.ctx, &pmsclient.QueryCommentReplayListReq{
		PageNum:   1,
		PageSize:  100,
		CommentId: detail.Id, // 传入评论的 ObjectID
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询评价回复列表失败,ID:%s,异常:%s", req.Id, err.Error())
	}

	var replays []types.CommentReplayItemData
	if replayResult != nil {
		for _, item := range replayResult.List {
			replays = append(replays, types.CommentReplayItemData{
				Id:             item.Id,
				CommentId:      detail.Id, // 传入评论的 ObjectID
				Type:           int(item.Type),
				MemberNickName: item.MemberNickName,
				MemberIcon:     item.MemberIcon,
				Content:        item.Content,
				CreateTime:     item.CreateTime,
			})
		}
	}

	return &types.QueryCommentDetailResp{
		Code:    "000000",
		Message: "查询成功",
		Data: types.CommentDetailData{
			Id:                detail.Id,
			ProductId:         detail.ProductId,
			ProductName:       detail.ProductName,
			MemberNickName:    detail.MemberNickName,
			MemberId:         detail.MemberId,
			MemberIcon:        detail.MemberIcon,
			Star:              detail.Star,
			Content:           detail.Content,
			Pics:              detail.Pics,
			ShowStatus:        detail.ShowStatus,
			ProductAttribute:  detail.ProductAttribute,
			ReplayCount:       detail.ReplayCount,
			MemberIp:          detail.MemberIp,
			CreateTime:        detail.CreateTime,
		},
		Replays: replays,
	}, nil
}
