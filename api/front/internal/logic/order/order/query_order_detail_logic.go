package order

import (
	"context"
	"strings"

	"github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	"github.com/feihua/zero-admin/pkg/errorx"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"google.golang.org/grpc/status"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/trace"
)

// maskPhone 手机号脱敏（中间4位用*替代）
// 例: 13812345678 → 138****5678
func maskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

// QueryOrderDetailLogic 订单详情（Story 6-1 重构）
/*
Author: LiuFeiHua
Date: 2025/6/20 11:35
*/
type QueryOrderDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryOrderDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryOrderDetailLogic {
	return &QueryOrderDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// buildTimeline 根据OMS订单状态构建时间线节点（Story 6-1 Task 2.2）
// OMS order_status: 1=待支付, 2=已支付/待发货, 3=已发货, 4=已完成, 5=已取消, 6=已退款, 7=售后中
// deliveryStatus: 0=未发货, 1=已发货, 2=已收货
// aftersaleStatus: 0=无售后, 1=售后申请中, 2=售后完成
func (l *QueryOrderDetailLogic) buildTimeline(detail *omsclient.OrderListData) []types.TimelineNode {
	var nodes []types.TimelineNode

	// 节点1: 订单创建
	nodes = append(nodes, types.TimelineNode{
		Status: "completed",
		Title:  "订单创建",
		Time:   formatTimelineTime(detail.CreateTime),
	})

	// 节点2: 支付成功（仅当已支付时）
	if detail.PayTime != "" {
		nodes = append(nodes, types.TimelineNode{
			Status: "completed",
			Title:  "支付成功",
			Time:   formatTimelineTime(detail.PayTime),
		})
	}

	// 节点3: 商家发货（当已发货时）
	if detail.DeliveryTime != "" && detail.DeliveryTime != "0001-01-01T00:00:00Z" {
		detailInfo := ""
		if detail.ExpressOrderNumber != "" {
			detailInfo = detail.ExpressOrderNumber
		}
		nodes = append(nodes, types.TimelineNode{
			Status: "completed",
			Title:  "商家发货",
			Time:   formatTimelineTime(detail.DeliveryTime),
			Detail: detailInfo,
		})
	}

	// 节点4: 确认收货
	if detail.ReceiveTime != "" && detail.ReceiveTime != "0001-01-01T00:00:00Z" {
		nodes = append(nodes, types.TimelineNode{
			Status: "completed",
			Title:  "确认收货",
			Time:   formatTimelineTime(detail.ReceiveTime),
		})
	}

	// 节点5: 订单完成（当状态为已完成）
	// 注：OMS OrderListData proto 无 finish_time 字段，OrderStatus==4 即表示已完成
	if detail.OrderStatus == 4 {
		nodes = append(nodes, types.TimelineNode{
			Status: "completed",
			Title:  "订单完成",
			Time:   formatTimelineTime(detail.ReceiveTime), // 用收货时间标记
		})
	}

	// 当前进行中节点判断（用于高亮）
	// OMS 真实值：0=待支付, 1=已支付(待发货), 2=已发货, 3=?, 4=已完成, 5=已取消, 7=售后中
	currentNodeIdx := 0
	if detail.OrderStatus == 0 {
		// 等待付款 → 高亮"订单创建"
		currentNodeIdx = 0
	} else if detail.OrderStatus == 1 {
		// 已支付(待发货) → 高亮"支付成功"
		currentNodeIdx = 1
	} else if detail.OrderStatus == 5 {
		// 已取消 → 时间线在"支付"后中断（OMS 5=已取消）
		currentNodeIdx = -1
	} else if detail.OrderStatus == 4 {
		// 已完成 → 高亮"订单完成"
		currentNodeIdx = len(nodes) - 1
	} else if detail.OrderStatus == 7 {
		// 售后中（OMS 7=售后中）
		currentNodeIdx = -1
	}

	// 标记当前节点
	for i := range nodes {
		if i == currentNodeIdx && currentNodeIdx >= 0 {
			nodes[i].Status = "current"
		}
	}

	// 中断节点（取消/售后）
	if detail.OrderStatus == 5 {
		// 已取消：追加中断节点
		nodes = append(nodes, types.TimelineNode{
			Status: "interrupted",
			Title:  "订单已取消",
			Time:   "",
			Detail: "系统自动关闭或用户取消",
		})
	} else if detail.OrderStatus == 7 {
		// 售后中：追加中断节点
		nodes = append(nodes, types.TimelineNode{
			Status: "interrupted",
			Title:  "售后处理中",
			Time:   "",
			Detail: "客服正在处理您的售后申请",
		})
	}

	return nodes
}

// formatTimelineTime 将ISO时间格式化为 HH:mm（Story 6-1 Review Fix: LOW-2）
func formatTimelineTime(isoTime string) string {
	if isoTime == "" || isoTime == "0001-01-01T00:00:00Z" {
		return ""
	}
	// 格式: 2025-03-29T10:30:00+08:00 → 10:30
	parts := strings.Split(isoTime, "T")
	if len(parts) < 2 {
		return "" // 格式不符时返回空而非原文，避免用户看到 "2025-03-29T10:30:00+08:00"
	}
	timePart := parts[1]
	hourMin := strings.Split(timePart, ":")
	if len(hourMin) >= 2 {
		return hourMin[0] + ":" + hourMin[1]
	}
	return ""
}

// QueryOrderDetail 获取订单详情（Story 6-1 Task 2）
func (l *QueryOrderDetailLogic) QueryOrderDetail(req *types.OrderDetailReq) (resp1 *types.OrderDetailResp, err error) {
	memberId, err := common.GetMemberId(l.ctx)
	if err != nil {
		return nil, err
	}

	// Story 6-1 Task 2.1: 归属校验
	res, err := l.svcCtx.OrderService.QueryOrderDetail(l.ctx, &omsclient.QueryOrderDetailReq{
		Id:     req.OrderId,
		UserId: memberId,
	})

	if err != nil {
		logc.Errorf(l.ctx, "查询订单详情失败,参数：%+v,响应：%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	detail := res.Data

	// Story 6-1 Task 2.3: 构建金额拆分
	pb := &types.PriceBreakdown{
		OrderAmount:     float64(detail.TotalAmount),
		FreightAmount:   float64(detail.FreightAmount),
		PromotionAmount: float64(detail.PromotionAmount),
		CouponAmount:    float64(detail.CouponAmount),
		PointsAmount:    float64(detail.PointsAmount),
		DiscountAmount:  float64(detail.DiscountAmount),
		PayAmount:       float64(detail.PayAmount),
	}

	// Story 6-1 Task 2.2: 时间线构建
	timeline := l.buildTimeline(detail)

	// 构建商品明细
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

	// Story 6-1 Task 2.4: 收货地址
	var recvAddr types.MemberReceiveAddressList
	if detail.DeliveryData != nil {
		recvAddr = types.MemberReceiveAddressList{
			Id:            detail.DeliveryData.Id,
			MemberId:      memberId,
			ReceiverName:  detail.DeliveryData.ReceiverName,
			ReceiverPhone: maskPhone(detail.DeliveryData.ReceiverPhone),
			Province:      detail.DeliveryData.ReceiverProvince,
			City:          detail.DeliveryData.ReceiverCity,
			District:      detail.DeliveryData.ReceiverDistrict,
			DetailAddress: detail.DeliveryData.ReceiverAddress,
			PostalCode:    "",
			Tag:           "",
			IsDefault:     1,
		}
	}

	// Story 10.7：附加提货卡摘要（C 端安全字段，仅订单包含数字卡商品时返回）
	digitalCards := l.loadDigitalCardSummaries(detail.Id, memberId)

	data := types.QueryOrderData{
		Id:                       detail.Id,
		OrderNo:                  detail.OrderNo,
		UserId:                   memberId,
		RequestTraceId:           trace.TraceIDFromContext(l.ctx),
		OrderStatus:              detail.OrderStatus,
		TotalAmount:              float64(detail.TotalAmount),
		PromotionAmount:          float64(detail.PromotionAmount),
		CouponAmount:             float64(detail.CouponAmount),
		PointsAmount:             float64(detail.PointsAmount),
		DiscountAmount:           float64(detail.DiscountAmount),
		FreightAmount:            float64(detail.FreightAmount),
		PayAmount:                float64(detail.PayAmount),
		PayType:                  detail.PayType,
		PayTime:                  detail.PayTime,
		DeliveryTime:             detail.DeliveryTime,
		ReceiveTime:              detail.ReceiveTime,
		CommentTime:              detail.CommentTime,
		SourceType:               detail.SourceType,
		ExpressOrderNumber:       detail.ExpressOrderNumber,
		UsePoints:                detail.UsePoints,
		ReceiveStatus:            detail.ReceiveStatus,
		Remark:                   detail.Remark,
		CreateTime:               detail.CreateTime,
		UpdateTime:               detail.UpdateTime,
		OrderItemData:            orderItemData,
		MemberReceiveAddressList: recvAddr,
		// Story 6-1 新增
		Timeline:       timeline,
		PriceBreakdown: pb,
		// Story 10.7 新增
		DigitalCards: digitalCards,
	}

	return &types.OrderDetailResp{
		Code:    0,
		Message: "操作成功",
		Data:    data,
	}, nil
}

// loadDigitalCardSummaries 查询订单关联的提货卡（Story 10.7）。
//
// 失败不阻塞订单详情主流程：拿不到时仅打日志返回 nil，让前端只展示常规订单信息。
func (l *QueryOrderDetailLogic) loadDigitalCardSummaries(orderID int64, memberID int64) []types.OrderDigitalCardItem {
	if l == nil || l.svcCtx == nil || l.svcCtx.DB == nil || orderID <= 0 {
		return nil
	}
	service := digitalcardmint.NewService(l.svcCtx.DB, nil, nil)
	summaries, err := service.LoadOrderCardSummaries(l.ctx, orderID, memberID)
	if err != nil {
		logc.Errorf(l.ctx, "loadDigitalCardSummaries failed orderId=%d memberId=%d err=%v", orderID, memberID, err)
		return nil
	}
	if len(summaries) == 0 {
		return nil
	}
	items := make([]types.OrderDigitalCardItem, 0, len(summaries))
	for _, summary := range summaries {
		items = append(items, types.OrderDigitalCardItem{
			AssetInstanceId:  summary.AssetInstanceID,
			AssetNoMasked:    summary.AssetNoMasked,
			TemplateId:       summary.TemplateID,
			TemplateName:     summary.TemplateName,
			TemplateImage:    summary.TemplateImage,
			MintStatus:       summary.MintStatus,
			MintStatusText:   summary.MintStatusText,
			DisplayStatus:    summary.DisplayStatus,
			ComplianceStatus: summary.ComplianceStatus,
			ComplianceTip:    summary.ComplianceTip,
			OrderItemId:      summary.OrderItemID,
			IssuedAt:         summary.IssuedAt,
		})
	}
	return items
}
