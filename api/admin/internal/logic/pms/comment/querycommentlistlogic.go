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

	// showStatus=-1 查全部；传了则使用该值
	showStatus := int32(-1)
	if req.ShowStatus >= 0 {
		showStatus = req.ShowStatus
	}

	result, err := l.svcCtx.CommentService.QueryCommentList(l.ctx, &pmsclient.QueryCommentListReq{
		ProductId:  req.ProductId,
		PlatformId: currentScope.PlatformID,
		TenantId:   currentScope.TenantID,
		MerchantId: currentScope.MerchantID,
		ShowStatus: showStatus,
		PageNum:    int64(req.Current),
		PageSize:   int64(req.PageSize),
	})

	if err != nil {
		logc.Errorf(l.ctx, "查询评价列表失败,参数:%+v,异常:%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	var list []types.CommentListData
	for _, item := range result.List {
		list = append(list, types.CommentListData{
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
			ProductAttribute: item.ProductAttribute,
			ReplayCount:      item.ReplayCount,
			MemberIp:         item.MemberIp,
			CreateTime:       item.CreateTime,
		})
	}

	return &types.QueryCommentListResp{
		Code:    "000000",
		Message: "查询成功",
		Data: types.CommentListPage{
			List:  list,
			Total: result.Total,
		},
	}, nil
}

// TODO(M-1): Admin 评价列表时间范围过滤
// 当前 FindPage 未支持 createTime 范围查询，待 Story 9-x 统一评价管理增强时实现：
// 1. 修改 rpc/pms/gen/model/product_comment_model.go FindPage 签名添加 startTime/endTime 参数
// 2. 修改 pms.proto QueryCommentListReq 添加 start_time / end_time 字段
// 3. Admin API types.QueryCommentListReq 添加 BeginTime / EndTime 字段
// 4. Admin logic 将时间参数透传至 RPC
