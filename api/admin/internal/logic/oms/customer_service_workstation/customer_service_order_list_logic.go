package customer_service_workstation

import (
	"context"
	"strings"

	admincommon "github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"google.golang.org/grpc/status"

	"github.com/zeromicro/go-zero/core/logx"
)

var (
	orderStatusTextMap = map[int32]string{
		1: "待支付",
		2: "已支付",
		3: "已发货",
		4: "已完成",
		5: "已取消",
		6: "已退款",
		7: "售后中",
	}
	payStatusTextMap = map[int32]string{
		0: "待支付",
		1: "支付成功",
		2: "支付失败",
	}
	returnStatusTextMap = map[int32]string{
		0: "无售后",
		1: "待审核",
		2: "审核通过",
		3: "审核拒绝",
	}
)

func buildOrderItemData(items []*omsclient.OrderItemData) []*types.OrderItemData {
	if len(items) == 0 {
		return nil
	}
	result := make([]*types.OrderItemData, 0, len(items))
	for _, item := range items {
		result = append(result, &types.OrderItemData{
			Id:              item.Id,
			OrderId:         item.OrderId,
			OrderNo:         item.OrderNo,
			OrderItemStatus: item.OrderItemStatus,
			SkuId:           item.SkuId,
			SkuName:         item.SkuName,
			SkuPic:          item.SkuPic,
			SkuPrice:        item.SkuPrice,
			SkuQuantity:     item.SkuQuantity,
			SpecData:        item.SpecData,
			SkuTotalAmount:  item.SkuTotalAmount,
			PromotionAmount: item.PromotionAmount,
			CouponAmount:    item.CouponAmount,
			PointsAmount:    item.PointsAmount,
			DiscountAmount:  item.DiscountAmount,
			RealAmount:      item.RealAmount,
			CreateTime:      item.CreateTime,
		})
	}
	return result
}

type CustomerServiceOrderListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCustomerServiceOrderListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CustomerServiceOrderListLogic {
	return &CustomerServiceOrderListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CustomerServiceOrderList 客服工作台订单列表（聚合：订单+物流状态+售后状态+收货信息）
func (l *CustomerServiceOrderListLogic) CustomerServiceOrderList(req *types.CustomerServiceOrderListReq) (resp *types.CustomerServiceOrderListResp, err error) {
	queryScope, err := admincommon.ResolveQueryGovernanceScope(l.ctx, admincommon.RequestedGovernanceScope{
		ScopeType:  req.ScopeType,
		PlatformID: req.PlatformId,
		TenantID:   req.TenantId,
		MerchantID: req.MerchantId,
	})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}

	result, err := l.svcCtx.OrderService.QueryOrderList(l.ctx, &omsclient.QueryOrderListReq{
		PageNum:    req.Current,
		PageSize:   req.PageSize,
		OrderNo:    req.OrderNo,
		OrderStatus: req.OrderStatus,
		Scope:      admincommon.OMSGovernanceScope(queryScope),
	})
	if err != nil {
		logc.Errorf(l.ctx, "客服工作台查询订单列表失败,参数：%+v,响应：%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	if len(result.List) == 0 {
		return &types.CustomerServiceOrderListResp{
			Code:     "000000",
			Message:  "查询成功",
			Data:     []*types.CustomerServiceOrderItem{},
			Total:    0,
			Current:  req.Current,
			PageSize: req.PageSize,
			Success:  true,
		}, nil
	}

	orderIds := make([]int64, len(result.List))
	for i, item := range result.List {
		orderIds[i] = item.Id
	}

	orderItemsMap := make(map[int64][]*omsclient.OrderItemData)
	for _, item := range result.List {
		if len(item.OrderItemData) > 0 {
			orderItemsMap[item.Id] = item.OrderItemData
		}
	}

	returnResult, err := l.svcCtx.OrderReturnService.QueryOrderReturnList(l.ctx, &omsclient.QueryOrderReturnListReq{
		PageNum:  1,
		PageSize: int32(len(orderIds) * 2),
	})
	if err != nil {
		logc.Errorf(l.ctx, "客服工作台查询退货列表失败: %s", err.Error())
	}

	returnMap := make(map[int64]*omsclient.OrderReturnData)
	if returnResult != nil {
		for _, item := range returnResult.List {
			if _, exists := returnMap[item.OrderId]; !exists {
				returnMap[item.OrderId] = item
			}
		}
	}

	var list []*types.CustomerServiceOrderItem
	for _, item := range result.List {
		delivery := item.DeliveryData
		returnItem := returnMap[item.Id]

		returnStatus := int32(0)
		if returnItem != nil {
			switch returnItem.Status {
			case 0:
				returnStatus = 1
			case 1:
				returnStatus = 2
			case 2:
				returnStatus = 3
			}
		}

		if req.ReturnStatus > 0 && returnStatus != req.ReturnStatus {
			continue
		}

		receiverName := ""
		receiverPhone := ""
		receiverAddress := ""
		deliveryCompany := ""
		deliverySn := ""
		if delivery != nil {
			receiverName = delivery.ReceiverName
			receiverPhone = delivery.ReceiverPhone
			receiverAddress = delivery.ReceiverProvince + delivery.ReceiverCity + delivery.ReceiverDistrict + delivery.ReceiverAddress
			deliveryCompany = delivery.DeliveryCompany
			deliverySn = delivery.DeliveryNo
		}

		if req.ReceiverName != "" && !strings.Contains(receiverName, req.ReceiverName) {
			continue
		}
		if req.ReceiverPhone != "" && !strings.Contains(receiverPhone, req.ReceiverPhone) {
			continue
		}

		payStatus := int32(0)
		if len(item.PaymentData) > 0 {
			payStatus = item.PaymentData[0].PayStatus
		}

		orderStatusText := orderStatusTextMap[item.OrderStatus]
		if orderStatusText == "" {
			orderStatusText = "未知"
		}
		payStatusText := payStatusTextMap[payStatus]
		if payStatusText == "" {
			payStatusText = "未知"
		}
		returnStatusText := returnStatusTextMap[returnStatus]
		if returnStatusText == "" {
			returnStatusText = "未知"
		}

		list = append(list, &types.CustomerServiceOrderItem{
			Id:               item.Id,
			OrderNo:          item.OrderNo,
			OrderStatus:      item.OrderStatus,
			OrderStatusText:  orderStatusText,
			PayStatus:        payStatus,
			PayStatusText:    payStatusText,
			ReturnStatus:     returnStatus,
			ReturnStatusText: returnStatusText,
			MemberId:         item.UserId,
			ReceiverName:     receiverName,
			ReceiverPhone:    receiverPhone,
			ReceiverAddress:  receiverAddress,
			TotalAmount:     item.TotalAmount,
			PayAmount:       item.PayAmount,
			DeliveryCompany: deliveryCompany,
			DeliverySn:      deliverySn,
			CreateTime:      item.CreateTime,
			PayTime:         item.PayTime,
			DeliveryTime:    item.DeliveryTime,
			ReceiveTime:     item.ReceiveTime,
			OrderItemData:   buildOrderItemData(orderItemsMap[item.Id]),
		})
	}

	return &types.CustomerServiceOrderListResp{
		Code:     "000000",
		Message:  "查询成功",
		Data:     list,
		Total:    result.Total,
		Current:  req.Current,
		PageSize: req.PageSize,
		Success:  true,
	}, nil
}
