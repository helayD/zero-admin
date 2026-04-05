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

// QueryCommentListLogic 查询评价列表
/*
Author: LiuFeiHua
Date: 2026/04/02
*/
type QueryCommentListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryCommentListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryCommentListLogic {
	return &QueryCommentListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// QueryCommentList 查询评价列表
func (l *QueryCommentListLogic) QueryCommentList(req *types.QueryCommentListReq) (*types.QueryCommentListResp, error) {
	currentScope, err := common.CurrentGovernanceScope(l.ctx)
	if err != nil {
		return nil, err
	}

	result, err := l.svcCtx.CommentService.QueryCommentList(l.ctx, &pmsclient.QueryCommentListReq{
		ProductId:   req.ProductId,
		PlatformId:  currentScope.PlatformID,
		TenantId:    currentScope.TenantID,
		MerchantId:  currentScope.MerchantID,
		ShowStatus:  req.ShowStatus,
		AuditStatus: req.AuditStatus,
		Hidden:      req.Hidden,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		ProductName: req.ProductName,
		MemberName:  req.MemberName,
		PageNum:     int64(req.Current),
		PageSize:    int64(req.PageSize),
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询评价列表失败,参数:%+v,异常:%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	list := make([]*types.CommentListData, 0, len(result.List))
	for _, item := range result.List {
		list = append(list, &types.CommentListData{
			Id:               item.Id,
			ProductId:        item.ProductId,
			ProductName:      item.ProductName,
			MemberNickName:   item.MemberNickName,
			MemberId:         item.MemberId,
			Star:             item.Star,
			Content:          item.Content,
			Pics:             item.Pics,
			MemberIcon:       item.MemberIcon,
			ShowStatus:       item.ShowStatus,
			AuditStatus:      item.AuditStatus,
			Hidden:           item.Hidden,
			AuditRemark:      item.AuditRemark,
			AuditorId:        item.AuditorId,
			AuditorName:      item.AuditorName,
			AuditedAt:        item.AuditedAt,
			AppealStatus:     item.AppealStatus,
			AppealReason:     item.AppealReason,
			AppealReply:      item.AppealReply,
			AppealedAt:       item.AppealedAt,
			AppealHandledAt:  item.AppealHandledAt,
			ProductAttribute: item.ProductAttribute,
			ReplayCount:      item.ReplayCount,
			MemberIp:         item.MemberIp,
			CreateTime:       item.CreateTime,
		})
	}

	return &types.QueryCommentListResp{
		Code:     "000000",
		Message:  "查询成功",
		Current:  req.Current,
		Data:     list,
		PageSize: req.PageSize,
		Success:  true,
		Total:    result.Total,
	}, nil
}
