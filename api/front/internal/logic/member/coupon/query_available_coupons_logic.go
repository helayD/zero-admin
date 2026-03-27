// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package coupon

import (
	"context"

	"github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"google.golang.org/grpc/status"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryAvailableCouponsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryAvailableCouponsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryAvailableCouponsLogic {
	return &QueryAvailableCouponsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// QueryAvailableCoupons 获取当前会员可领取的优惠券列表
func (l *QueryAvailableCouponsLogic) QueryAvailableCoupons(req *types.QueryAvailableCouponsReq) (resp *types.QueryAvailableCouponsResp, err error) {
	memberId, err := common.GetMemberId(l.ctx)
	if err != nil {
		return nil, err
	}

	pageNum := req.PageNum
	pageSize := req.PageSize
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	scope := common.ResolveEffectiveGovernanceScope(l.ctx)
	respData, err := l.svcCtx.CouponRecordService.QueryAvailableCoupons(l.ctx, &smsclient.QueryAvailableCouponsReq{
		MemberId: memberId,
		PageNum:  pageNum,
		PageSize: pageSize,
		Scope: &smsclient.GovernanceScope{
			PlatformId: scope.PlatformID,
			TenantId:   scope.TenantID,
			MerchantId: scope.MerchantID,
		},
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询可领取优惠券列表失败, memberId:%d, 异常:%s", memberId, err.Error())
	_, _ = status.FromError(err)
	}

	list := make([]*types.CouponData, 0, len(respData.List))
	for _, item := range respData.List {
		list = append(list, &types.CouponData{
			Id:            item.Id,
			TypeId:        item.TypeId,
			Name:          item.Name,
			Code:          item.Code,
			Amount:        item.Amount,
			MinAmount:     item.MinAmount,
			StartTime:     item.StartTime,
			EndTime:       item.EndTime,
			PerLimit:      item.PerLimit,
			Status:        item.Status,
			Description:   item.Description,
			ScopeType:     item.ScopeType,
			ReceiveStatus: item.ReceiveStatus,
			TotalCount:    item.TotalCount,
			ReceivedCount: item.ReceivedCount,
		})
	}

	return &types.QueryAvailableCouponsResp{
		Code:    0,
		Message: "查询可领取优惠券成功",
		Data:    list,
	}, nil
}
