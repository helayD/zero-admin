package order

import (
	"context"

	"github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/pkg/errorx"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"google.golang.org/grpc/status"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// QueryOrderListLogic 订单查询（Story 6-1 重构）
/*
Author: LiuFeiHua
Date: 2025/6/20 11:32
*/
type QueryOrderListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryOrderListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryOrderListLogic {
	return &QueryOrderListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// QueryOrderList 按状态分页获取用户订单列表
// Story 6-1 Task 1+3: Flutter Tab映射 → OMS Status映射
// Flutter Tab: 0=全部,1=待支付,2=待发货,3=已完成,4=已取消
// OMS Status: 0=全部不填,1=待支付,2=已支付(=待发货),3=已发货,4=已取消
func (l *QueryOrderListLogic) QueryOrderList(req *types.OrderListReq) (resp1 *types.OrderListResp, err error) {
	memberId, err := common.GetMemberId(l.ctx)
	if err != nil {
		return nil, err
	}

	// Story 6-1 Task 1.2: Flutter状态 → OMS状态映射
	// Flutter: 0=全部,1=待支付,2=待发货,3=已完成,4=已取消
	// OMS:    1=待支付,2=已支付(待发货),3=已发货,4=已完成,5=已取消
	var omsStatus int32
	switch req.Status {
	case 0:
		omsStatus = 0 // 全部：不填，传0让OMS返回所有
	case 1:
		omsStatus = 1 // 待支付
	case 2:
		omsStatus = 2 // 待发货（OMS:已支付）
	case 3:
		omsStatus = 4 // 已完成
	case 4:
		omsStatus = 5 // 已取消
	default:
		omsStatus = 0
	}

	// Story 6-1 Task 1.4: pageNum/pageSize 分页
	pageNum := req.Current
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}

	res, err := l.svcCtx.OrderService.QueryOrderList(l.ctx, &omsclient.QueryOrderListReq{
		PageNum:     pageNum,   // 当前页
		PageSize:    pageSize,  // 每页条数
		UserId:      memberId,  // 用户ID
		OrderStatus: omsStatus, // OMS订单状态
	})

	if err != nil {
		logc.Errorf(l.ctx, "查询订单列表失败,参数：%+v,响应：%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	orderData := make([]*types.QueryOrderData, 0)
	for _, detail := range res.List {
		// Story 6-1 Task 1.3: 从OMS OrderListData构建摘要（含首商品缩略图）
		var thumbnail string
		if len(detail.OrderItemData) > 0 {
			thumbnail = detail.OrderItemData[0].SkuPic
		}

		elems := &types.QueryOrderData{
			Id:               detail.Id,
			OrderNo:          detail.OrderNo,
			UserId:           detail.UserId,
			OrderStatus:      detail.OrderStatus,
			TotalAmount:      float64(detail.TotalAmount),
			PromotionAmount:  float64(detail.PromotionAmount),
			CouponAmount:     float64(detail.CouponAmount),
			PointsAmount:     float64(detail.PointsAmount),
			DiscountAmount:   float64(detail.DiscountAmount),
			FreightAmount:    float64(detail.FreightAmount),
			PayAmount:        float64(detail.PayAmount),
			PayType:          detail.PayType,
			PayTime:          detail.PayTime,
			DeliveryTime:     detail.DeliveryTime,
			ReceiveTime:      detail.ReceiveTime,
			CommentTime:      detail.CommentTime,
			SourceType:       detail.SourceType,
			ExpressOrderNumber: detail.ExpressOrderNumber,
			UsePoints:        detail.UsePoints,
			ReceiveStatus:    detail.ReceiveStatus,
			Remark:           detail.Remark,
			CreateTime:       detail.CreateTime,
			UpdateTime:       detail.UpdateTime,
			Thumbnail:        thumbnail,
		}

		// 支付状态从 OMS PaymentData 推断（Story 6-1 Task 1.3 注：OMS proto 无 pay_status 独立字段，用 PayType 推断）
		// PayType > 0 且 PayTime 非空 → 已支付
		if detail.PayType > 0 && detail.PayTime != "" {
			elems.PayStatus = 1
		}

		itemData := detail.OrderItemData
		orderItemData := make([]*types.OrderItemData, 0)
		for _, d := range itemData {
			orderItemData = append(orderItemData, &types.OrderItemData{
				Id:              d.Id,
				OrderId:         d.OrderId,
				OrderNo:         d.OrderNo,
				OrderItemStatus: d.OrderItemStatus,
				SkuId:           d.SkuId,
				SkuName:         d.SkuName,
				SkuPic:          d.SkuPic,
				SkuPrice:        float64(d.SkuPrice),
				SkuQuantity:     d.SkuQuantity,
				SpecData:        d.SpecData,
				SkuTotalAmount:  float64(d.SkuTotalAmount),
				PromotionAmount: float64(d.PromotionAmount),
				CouponAmount:    float64(d.CouponAmount),
				PointsAmount:    float64(d.PointsAmount),
				DiscountAmount:  float64(d.DiscountAmount),
				RealAmount:      float64(d.RealAmount),
			})
		}
		elems.OrderItemData = orderItemData

		orderData = append(orderData, elems)
	}

	// Story 6-1 Task 1.2: 返回分页信息
	return &types.OrderListResp{
		Code:     0,
		Message:  "操作成功",
		PageNum:  int64(pageNum),
		PageSize: int64(pageSize),
		Total:    res.Total,
		Data:     orderData,
	}, nil
}
