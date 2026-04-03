package comment

import (
	"context"

	"github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/status"
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

// QueryCommentDetail 查询评价详情（包含回复列表和审核历史）
func (l *QueryCommentDetailLogic) QueryCommentDetail(req *QueryCommentDetailReq) (*types.QueryCommentDetailResp, error) {
	currentScope, err := common.CurrentGovernanceScope(l.ctx)
	if err != nil {
		return nil, err
	}

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

	replayResult, err := l.svcCtx.CommentReplayService.QueryCommentReplayList(l.ctx, &pmsclient.QueryCommentReplayListReq{
		PageNum:   1,
		PageSize:  100,
		CommentId: detail.Id,
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询评价回复列表失败,ID:%s,异常:%s", req.Id, err.Error())
	}

	auditResult, err := l.svcCtx.CommentService.QueryCommentAuditLog(l.ctx, &pmsclient.QueryCommentAuditLogReq{
		CommentId:  detail.Id,
		PageNum:    1,
		PageSize:   100,
		PlatformId: currentScope.PlatformID,
		TenantId:   currentScope.TenantID,
		MerchantId: currentScope.MerchantID,
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询评价审核历史失败,ID:%s,异常:%s", req.Id, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	replays := make([]*types.CommentReplayItemData, 0)
	if replayResult != nil {
		for _, item := range replayResult.List {
			replays = append(replays, &types.CommentReplayItemData{
				Id:             item.Id,
				CommentId:      item.CommentId,
				Type:           item.Type,
				MemberNickName: item.MemberNickName,
				MemberIcon:     item.MemberIcon,
				Content:        item.Content,
				CreateTime:     item.CreateTime,
			})
		}
	}

	auditLogs := make([]*types.CommentAuditLogData, 0, len(auditResult.List))
	for _, item := range auditResult.List {
		auditLogs = append(auditLogs, &types.CommentAuditLogData{
			Id:           item.Id,
			CommentId:    item.CommentId,
			Action:       item.Action,
			FromStatus:   item.FromStatus,
			ToStatus:     item.ToStatus,
			OperatorId:   item.OperatorId,
			OperatorName: item.OperatorName,
			Remark:       item.Remark,
			CreatedAt:    item.CreatedAt,
		})
	}

	return &types.QueryCommentDetailResp{
		Code:    "000000",
		Message: "查询成功",
		Data: types.CommentDetailData{
			Id:               detail.Id,
			ProductId:        detail.ProductId,
			ProductName:      detail.ProductName,
			MemberNickName:   detail.MemberNickName,
			MemberId:         detail.MemberId,
			MemberIcon:       detail.MemberIcon,
			Star:             detail.Star,
			Content:          detail.Content,
			Pics:             detail.Pics,
			ShowStatus:       detail.ShowStatus,
			AuditStatus:      detail.AuditStatus,
			Hidden:           detail.Hidden,
			AuditRemark:      detail.AuditRemark,
			AuditorId:        detail.AuditorId,
			AuditorName:      detail.AuditorName,
			AuditedAt:        detail.AuditedAt,
			AppealStatus:     detail.AppealStatus,
			AppealReason:     detail.AppealReason,
			AppealReply:      detail.AppealReply,
			AppealedAt:       detail.AppealedAt,
			AppealHandledAt:  detail.AppealHandledAt,
			ProductAttribute: detail.ProductAttribute,
			CollectCount:     detail.CollectCount,
			ReadCount:        detail.ReadCount,
			ReplayCount:      detail.ReplayCount,
			MemberIp:         detail.MemberIp,
			CreateTime:       detail.CreateTime,
		},
		Replays:   replays,
		AuditLogs: auditLogs,
	}, nil
}
