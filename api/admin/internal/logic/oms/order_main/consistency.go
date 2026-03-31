package order_main

import (
	"context"
	"fmt"

	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
)

type consistencyView struct {
	ConsistencyStage     int32
	ConsistencyStageText string
	ConsistencyResult    int32
	ConsistencyMessage   string
	LastConsistencyAt    string
	PendingActions       int32
	PendingActionsText   []string
	AftersaleStatus      int32
	AftersaleStatusText  string
	ReturnId             int64
	ReturnNo             string
}

func buildConsistencyView(ctx context.Context, svcCtx *svc.ServiceContext, detail *omsclient.OrderListData) consistencyView {
	if detail == nil {
		return consistencyView{}
	}

	payStatus := 0
	if len(detail.PaymentData) > 0 {
		payStatus = int(detail.PaymentData[0].PayStatus)
	}

	cancelCompensationStatus := getCancelCompensationStatus(ctx, svcCtx, detail.Id)
	afterSaleStatus, returnId, returnNo, afterSaleUpdatedAt := getLatestAfterSale(ctx, svcCtx, detail.Id, detail.UserId)
	stage, pendingActions := calcConsistencyStage(int(detail.OrderStatus), payStatus, cancelCompensationStatus, afterSaleStatus)
	result := calcConsistencyResult(int(detail.OrderStatus), payStatus, cancelCompensationStatus, afterSaleStatus)
	lastConsistencyAt := afterSaleUpdatedAt
	if lastConsistencyAt == "" && len(detail.OptLogData) > 0 {
		lastConsistencyAt = detail.OptLogData[0].CreateTime
	}
	afterSaleStatusText := ""
	if returnId > 0 {
		afterSaleStatusText = getAfterSaleStatusText(afterSaleStatus)
	}

	return consistencyView{
		ConsistencyStage:     int32(stage),
		ConsistencyStageText: getConsistencyStageText(stage),
		ConsistencyResult:    int32(result),
		ConsistencyMessage:   buildConsistencyMessage(stage, result, afterSaleStatus),
		LastConsistencyAt:    lastConsistencyAt,
		PendingActions:       int32(pendingActions),
		PendingActionsText:   getPendingActionsText(pendingActions),
		AftersaleStatus:      afterSaleStatus,
		AftersaleStatusText:  afterSaleStatusText,
		ReturnId:             returnId,
		ReturnNo:             returnNo,
	}
}

func getCancelCompensationStatus(ctx context.Context, svcCtx *svc.ServiceContext, orderId int64) string {
	key := fmt.Sprintf("order:cancel:compensation:%d", orderId)
	val, err := svcCtx.Redis.GetCtx(ctx, key)
	if err != nil {
		return ""
	}
	return val
}

func getLatestAfterSale(ctx context.Context, svcCtx *svc.ServiceContext, orderId, memberId int64) (int32, int64, string, string) {
	returns, err := svcCtx.OrderReturnService.QueryOrderReturnList(ctx, &omsclient.QueryOrderReturnListReq{
		OrderId:  orderId,
		MemberId: memberId,
		Status:   2,
		PageNum:  1,
		PageSize: 20,
	})
	if err != nil || returns == nil || len(returns.List) == 0 {
		return 0, 0, "", ""
	}

	latest := returns.List[0]
	for _, item := range returns.List[1:] {
		if item.Id > latest.Id {
			latest = item
		}
	}

	lastAt := latest.CreateTime
	switch latest.Status {
	case 1:
		if latest.HandleTime != "" {
			lastAt = latest.HandleTime
		}
	case 2:
		if latest.ReceiveTime != "" {
			lastAt = latest.ReceiveTime
		}
	case 3:
		if latest.RefundTime != "" {
			lastAt = latest.RefundTime
		}
	case 4, 5:
		if latest.CloseTime != "" {
			lastAt = latest.CloseTime
		} else if latest.HandleTime != "" {
			lastAt = latest.HandleTime
		}
	}

	return latest.Status, latest.Id, latest.ReturnNo, lastAt
}

func calcConsistencyStage(orderStatus, payStatus int, cancelCompensationStatus string, afterSaleStatus int32) (int, int) {
	switch orderStatus {
	case 1:
		if payStatus == 2 {
			return 3, 0
		}
		return 0, 0
	case 2:
		return 2, 2 | 4
	case 3, 4:
		return 9, 0
	case 5:
		if payStatus == 1 {
			if cancelCompensationStatus == "completed" {
				return 5, 0
			}
			return 4, 1 | 2 | 4
		}
		return 5, 0
	case 6:
		if afterSaleStatus != 0 {
			return 8, 0
		}
		return 9, 0
	case 7:
		switch afterSaleStatus {
		case 0:
			return 6, 0
		case 1, 2:
			return 7, 1 | 2 | 4
		case 3, 4, 5:
			return 8, 0
		default:
			return 6, 0
		}
	default:
		return 0, 0
	}
}

func calcConsistencyResult(orderStatus, payStatus int, cancelCompensationStatus string, afterSaleStatus int32) int {
	switch orderStatus {
	case 1:
		if payStatus == 2 {
			return 3
		}
		return 0
	case 2:
		return 1
	case 5:
		if payStatus == 1 && cancelCompensationStatus != "completed" {
			return 1
		}
		return 2
	case 7:
		switch afterSaleStatus {
		case 4:
			return 3
		case 3, 5:
			return 2
		default:
			return 1
		}
	case 3, 4, 6:
		return 2
	default:
		return 0
	}
}

func buildConsistencyMessage(stage, result int, afterSaleStatus int32) string {
	switch stage {
	case 1:
		return "支付确认中，请稍候..."
	case 2:
		return "支付成功，权益同步中..."
	case 3:
		return "支付失败，请重新支付"
	case 4:
		return "取消处理中，正在回退权益..."
	case 5:
		return "已取消，权益已回退"
	case 6:
		if afterSaleStatus == 0 {
			return "售后申请已提交，等待审核..."
		}
		return "售后待处理"
	case 7:
		switch afterSaleStatus {
		case 1:
			return "售后审核已通过，等待进一步处理..."
		case 2:
			return "售后处理中，请耐心等待..."
		default:
			return "售后处理中，请耐心等待..."
		}
	case 8:
		switch afterSaleStatus {
		case 3:
			return "售后退款完成"
		case 4:
			return "售后申请已拒绝"
		case 5:
			return "售后已关闭"
		default:
			return "售后处理完成"
		}
	case 9:
		if result == 2 {
			return "订单处理完成"
		}
	}
	return ""
}

func getConsistencyStageText(stage int) string {
	switch stage {
	case 0:
		return "正常"
	case 1:
		return "支付确认中"
	case 2:
		return "权益同步中"
	case 3:
		return "支付失败"
	case 4:
		return "取消回退中"
	case 5:
		return "已取消"
	case 6:
		return "售后待处理"
	case 7:
		return "售后处理中"
	case 8:
		return "售后已完成"
	case 9:
		return "已完成"
	default:
		return "未知阶段"
	}
}

func getPendingActionsText(actions int) []string {
	var result []string
	if actions&1 != 0 {
		result = append(result, "库存")
	}
	if actions&2 != 0 {
		result = append(result, "优惠券")
	}
	if actions&4 != 0 {
		result = append(result, "积分")
	}
	return result
}

func getAfterSaleStatusText(status int32) string {
	switch status {
	case 0:
		return "待审核"
	case 1:
		return "审核通过"
	case 2:
		return "已收货"
	case 3:
		return "已退款"
	case 4:
		return "已拒绝"
	case 5:
		return "已关闭"
	default:
		return ""
	}
}
