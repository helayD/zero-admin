package comment

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/pkg/errorx"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"google.golang.org/grpc/status"

	"github.com/zeromicro/go-zero/core/logx"
)

// AddCommentLogic 提交商品评价
/*
Author: LiuFeiHua
Date: 2026/04/02
*/
type AddCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddCommentLogic {
	return &AddCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AddComment 提交商品评价
func (l *AddCommentLogic) AddComment(req *types.AddCommentReq) (*types.AddCommentResp, error) {
	memberId, err := common.GetMemberId(l.ctx)
	if err != nil {
		return nil, err
	}

	// Review Fix R-1: 校验订单完成状态（AC-1: 订单已完成且商品满足评价条件）
	orderDetail, err := l.svcCtx.OrderService.QueryOrderDetail(l.ctx, &omsclient.QueryOrderDetailReq{
		Id:     req.OrderId,
		UserId: memberId,
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询订单详情失败,orderId:%d,异常:%s", req.OrderId, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	// 订单归属校验：确保订单属于当前用户
	if orderDetail.Data == nil {
		return nil, errors.New("订单不存在")
	}
	if orderDetail.Data.UserId != memberId {
		return nil, errors.New("无权操作此订单")
	}
	// 订单状态校验：OMS order_status: 1-待支付,2-已支付,3-已发货,4-已完成,5-已取消,6-已退款,7-售后中
	if orderDetail.Data.OrderStatus != 4 {
		return nil, errors.New("只有已完成订单才能提交评价")
	}

	// Review Fix M-2: 使用 CurrentGovernanceScope 而非 ResolveEffectiveGovernanceScope
	// 前台接口必须登录后访问，未登录会返回错误，不允许空 scope 默认化
	scope, err := common.CurrentGovernanceScope(l.ctx)
	if err != nil {
		return nil, err
	}

	_, err = l.svcCtx.CommentService.AddComment(l.ctx, &pmsclient.AddCommentReq{
		ProductId:        req.ProductId,
		MemberNickName:   req.MemberNickName,
		ProductName:      req.ProductName,
		Star:             int32(req.Star),
		ShowStatus:       0, // 默认待审核
		ProductAttribute: req.ProductAttribute,
		Content:          req.Content,
		Pics:             req.Pics,
		MemberIcon:       req.MemberIcon,
		MemberId:         memberId,
		PlatformId:       scope.PlatformID,
		TenantId:         scope.TenantID,
		MerchantId:       scope.MerchantID,
		OrderId:          req.OrderId,
	})

	if err != nil {
		logc.Errorf(l.ctx, "提交商品评价失败,参数:%+v,异常:%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	return &types.AddCommentResp{
		Code:    0,
		Message: "评价提交成功",
	}, nil
}
