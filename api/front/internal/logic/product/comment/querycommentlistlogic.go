package comment

import (
	"context"

	"github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/pkg/errorx"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"google.golang.org/grpc/status"

	"github.com/zeromicro/go-zero/core/logx"
)

// QueryCommentListLogic 查询商品评价列表
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

// QueryCommentList 查询商品评价列表（前台仅显示已审核通过）
func (l *QueryCommentListLogic) QueryCommentList(req *types.QueryCommentListReq) (*types.QueryCommentListResp, error) {
	// Review Fix M-2: 使用 CurrentGovernanceScope 保证严格校验
	scope, err := common.CurrentGovernanceScope(l.ctx)
	if err != nil {
		return nil, err
	}

	result, err := l.svcCtx.CommentService.QueryCommentList(l.ctx, &pmsclient.QueryCommentListReq{
		ProductId:  req.ProductId,
		PlatformId: scope.PlatformID,
		TenantId:   scope.TenantID,
		MerchantId: scope.MerchantID,
		ShowStatus: 1, // 前台固定查已审核通过
		PageNum:    int64(req.PageNum),
		PageSize:   int64(req.PageSize),
	})

	if err != nil {
		logc.Errorf(l.ctx, "查询商品评价列表失败,参数:%+v,异常:%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	var list []types.CommentListItem
	for _, item := range result.List {
		list = append(list, types.CommentListItem{
			Id:               item.Id,
			ProductId:        item.ProductId,
			MemberNickName:   item.MemberNickName,
			MemberId:         item.MemberId,
			MemberIcon:       item.MemberIcon,
			Star:             int(item.Star),
			Content:          item.Content,
			Pics:             item.Pics,
			ProductAttribute: item.ProductAttribute,
			ShowStatus:       int(item.ShowStatus),
			ReplayCount:      int(item.ReplayCount),
			MemberIp:         item.MemberIp,
			CreateTime:       item.CreateTime,
		})
	}

	return &types.QueryCommentListResp{
		Code:    0,
		Message: "查询成功",
		Data:    list,
		Total:   result.Total,
	}, nil
}
