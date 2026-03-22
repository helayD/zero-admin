package coupon

import (
	"context"
	"time"

	frontcommon "github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/api/front/internal/logic/order/cart"
	"github.com/feihua/zero-admin/pkg/errorx"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"google.golang.org/grpc/status"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// QueryCouponListByCartLogic 获取登录会员购物车的相关优惠券
/*
Author: LiuFeiHua
Date: 2025/6/19 11:34
*/
type QueryCouponListByCartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryCouponListByCartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryCouponListByCartLogic {
	return &QueryCouponListByCartLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// QueryCouponListByCart 获取登录会员购物车的相关优惠券
func (l *QueryCouponListByCartLogic) QueryCouponListByCart(req *types.CouponListByCartReq) (resp *types.CouponListByCartResp, err error) {
	// 1.获取购物车信息
	cartPromotionItemList, err := cart.QueryCartListPromotion(nil, l.ctx, l.svcCtx)

	if err != nil {
		return nil, err
	}
	// 获取该用户所有优惠券
	enableList, disableList, err := QueryCouponList(l.svcCtx, l.ctx, cartPromotionItemList)
	if err != nil {
		return nil, err
	}
	return &types.CouponListByCartResp{
		Data: types.CouponListByCartData{
			EnableList:  enableList,
			DisableList: disableList,
		},
		Code:    0,
		Message: "查询会员优惠券成功",
	}, nil
}

func QueryCouponList(svcCtx *svc.ServiceContext, ctx context.Context, cartPromotionItemList []types.CarItemtPromotionListData) ([]types.CouponData, []types.CouponData, error) {
	if len(cartPromotionItemList) == 0 {
		return []types.CouponData{}, []types.CouponData{}, nil
	}

	memberId, err := frontcommon.GetMemberId(ctx)
	if err != nil {
		return nil, nil, err
	}

	couponList, err := svcCtx.CouponRecordService.QueryMemberCouponList(ctx, &smsclient.QueryMemberCouponListReq{
		MemberId: memberId,
		Status:   0,
	})
	if err != nil {
		logc.Errorf(ctx, "获取会员优惠券失败,memberId:%d,异常:%s", memberId, err.Error())
		s, _ := status.FromError(err)
		return nil, nil, errorx.NewDefaultError(s.Message())
	}

	currentScope := frontcommon.ResolveEffectiveGovernanceScope(ctx)
	availableResp, err := svcCtx.CouponService.QueryCouponByScopeId(ctx, &smsclient.QueryCouponByScopeIdReq{
		ScopeIds: buildCouponScopeIDs(cartPromotionItemList),
		Scope:    frontcommon.SMSGovernanceScope(currentScope),
	})
	if err != nil {
		logc.Errorf(ctx, "查询购物车可用优惠券失败,scope:%+v,异常:%s", currentScope, err.Error())
		s, _ := status.FromError(err)
		return nil, nil, errorx.NewDefaultError(s.Message())
	}

	availableMap := make(map[int64]*smsclient.CouponListData, len(availableResp.List))
	for _, item := range availableResp.List {
		availableMap[item.Id] = item
	}

	enableList := make([]types.CouponData, 0)
	disableList := make([]types.CouponData, 0)
	for _, item := range couponList.List {
		available, ok := availableMap[item.Id]
		if !ok {
			continue
		}

		scopeResp, err := svcCtx.CouponScopeService.QueryCouponScopeList(ctx, &smsclient.QueryCouponScopeListReq{
			CouponId: item.Id,
			PageNum:  1,
			PageSize: 200,
		})
		if err != nil {
			logc.Errorf(ctx, "查询优惠券适用范围失败,couponId:%d,异常:%s", item.Id, err.Error())
			continue
		}

		couponData := toCouponData(available, item.ScopeType)
		if couponSubtotal(couponData.ScopeType, scopeResp.List, cartPromotionItemList) >= float64(couponData.MinAmount) &&
			couponUsableNow(couponData.StartTime, couponData.EndTime) {
			enableList = append(enableList, couponData)
		} else {
			disableList = append(disableList, couponData)
		}
	}

	return enableList, disableList, nil
}

func buildCouponScopeIDs(cartPromotionItemList []types.CarItemtPromotionListData) []int64 {
	ids := make([]int64, 0, len(cartPromotionItemList)*2)
	seen := make(map[int64]struct{}, len(cartPromotionItemList)*2)
	for _, item := range cartPromotionItemList {
		for _, id := range []int64{item.ProductId, item.ProductCategoryId} {
			if id <= 0 {
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			ids = append(ids, id)
		}
	}
	return ids
}

func couponSubtotal(scopeType int32, scopeRows []*smsclient.CouponScopeListData, cartPromotionItemList []types.CarItemtPromotionListData) float64 {
	if scopeType == 0 {
		total := 0.0
		for _, item := range cartPromotionItemList {
			total += cartItemAmount(item)
		}
		return total
	}

	scopeIDs := make(map[int64]struct{}, len(scopeRows))
	for _, row := range scopeRows {
		if row.ScopeId <= 0 {
			continue
		}
		scopeIDs[row.ScopeId] = struct{}{}
	}

	total := 0.0
	for _, item := range cartPromotionItemList {
		switch scopeType {
		case 1:
			if _, ok := scopeIDs[item.ProductCategoryId]; ok {
				total += cartItemAmount(item)
			}
		case 2:
			if _, ok := scopeIDs[item.ProductId]; ok {
				total += cartItemAmount(item)
			}
		}
	}
	return total
}

func cartItemAmount(item types.CarItemtPromotionListData) float64 {
	price := float64(item.Price) - float64(item.ReduceAmount)
	if price < 0 {
		price = 0
	}
	return price * float64(item.Quantity)
}

func couponUsableNow(startTime, endTime string) bool {
	start, err := time.ParseInLocation("2006-01-02 15:04:05", startTime, time.Local)
	if err != nil {
		return false
	}
	end, err := time.ParseInLocation("2006-01-02 15:04:05", endTime, time.Local)
	if err != nil {
		return false
	}
	now := time.Now()
	return !now.Before(start) && !now.After(end)
}

func toCouponData(item *smsclient.CouponListData, scopeType int32) types.CouponData {
	return types.CouponData{
		Id:          item.Id,
		TypeId:      item.TypeId,
		Name:        item.Name,
		Code:        item.Code,
		Amount:      item.Amount,
		MinAmount:   item.MinAmount,
		StartTime:   item.StartTime,
		EndTime:     item.EndTime,
		PerLimit:    item.PerLimit,
		Status:      item.Status,
		Description: item.Description,
		ScopeType:   scopeType,
	}
}
