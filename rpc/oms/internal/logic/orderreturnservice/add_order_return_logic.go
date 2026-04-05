package orderreturnservicelogic

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/feihua/zero-admin/pkg/audit"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/oms/gen/model"
	"github.com/feihua/zero-admin/rpc/oms/gen/query"
	"github.com/feihua/zero-admin/rpc/oms/internal/svc"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// ReturnEventPayload 售后申请事件 payload（复用 orderservice 的统一契约模式）
type ReturnEventPayload struct {
	EventID    string                 `json:"eventId"`
	OccurredAt int64                  `json:"occurredAt"`
	TraceID    string                 `json:"traceId"`
	PlatformID int64                  `json:"platformId"`
	TenantID   int64                  `json:"tenantId"`
	MerchantID int64                  `json:"merchantId"`
	ActorID    int64                  `json:"actorId"`
	EntityID   int64                  `json:"entityId"`
	Action     string                 `json:"action"`
	Version    string                 `json:"version"`
	ScopeType  string                 `json:"scopeType"`
	Source     string                 `json:"source,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
}

// AddOrderReturnLogic 添加退货/售后
/*
Author: LiuFeiHua
Date: 2025/06/30 15:49:28
*/
type AddOrderReturnLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddOrderReturnLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddOrderReturnLogic {
	return &AddOrderReturnLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AddOrderReturn 添加退货/售后
func (l *AddOrderReturnLogic) AddOrderReturn(in *omsclient.OrderReturnReq) (*omsclient.OrderReturnResp, error) {
	order, err := query.OmsOrderMain.WithContext(l.ctx).Where(query.OmsOrderMain.ID.Eq(in.OrderId), query.OmsOrderMain.IsDeleted.Eq(0)).First()

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		logc.Errorf(l.ctx, "订单不存在, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errors.New("订单不存在")
	case err != nil:
		logc.Errorf(l.ctx, "查询订单异常, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errors.New("查询订单异常")
	}

	if order.OrderStatus != 4 {
		logc.Errorf(l.ctx, "订单状态异常, 订单ID：%d, 订单状态：%d", in.OrderId, order.OrderStatus)
		return nil, errors.New("订单状态异常,申请退货/售后明细失败")
	}

	q := query.OmsOrderReturn

	item := &model.OmsOrderReturn{
		OrderID:        in.OrderId,               // 关联订单ID
		ReturnNo:       in.ReturnNo,              // 退货单号
		MemberID:       in.MemberId,              // 会员ID
		Status:         in.Status,                // 退货状态（0待审核 1审核通过 2已收货 3已退款 4已拒绝 5已关闭）
		Type:           in.Type,                  // 售后类型（0退货退款 1仅退款 2换货）
		Reason:         in.Reason,                // 退货原因
		Description:    in.Description,           // 问题描述
		ProofPic:       in.ProofPic,              // 凭证图片，逗号分隔
		RefundAmount:   float64(in.RefundAmount), // 退款金额
		ReturnName:     in.ReturnName,            // 退货人姓名
		ReturnPhone:    in.ReturnPhone,           // 退货人电话
		CompanyAddress: in.CompanyAddress,        // 退货收货地址
		Remark:         in.Remark,                // 备注
	}

	err = q.WithContext(l.ctx).Create(item)
	if err != nil {
		logc.Errorf(l.ctx, "添加退货/售后失败,参数:%+v,异常:%s", item, err.Error())
		return nil, errors.New("添加退货/售后失败")
	}

	orderReturnItem := query.OmsOrderReturnItem
	_, _ = orderReturnItem.WithContext(l.ctx).Where(orderReturnItem.ReturnID.Eq(item.ID)).Delete()
	for _, data := range in.OrderReturnItem {

		returnItem := &model.OmsOrderReturnItem{
			ReturnID:     item.ID,                    // 退货单ID（关联oms_order_return.id）
			OrderID:      data.OrderId,               // 订单ID
			OrderItemID:  data.OrderItemId,           // 订单明细ID
			SkuID:        data.SkuId,                 // 商品SKU ID
			SkuName:      data.SkuName,               // 商品名称
			SkuPic:       data.SkuPic,                // 商品图片
			SkuAttrs:     data.SkuAttrs,              // 商品销售属性
			Quantity:     data.Quantity,              // 退货数量
			ProductPrice: float64(data.ProductPrice), // 商品单价
			RealAmount:   float64(data.RealAmount),   // 实际退款金额
			Reason:       data.Reason,                // 退货原因
			Remark:       data.Remark,                // 备注
		}

		err = orderReturnItem.WithContext(l.ctx).Create(returnItem)
		if err != nil {
			logc.Errorf(l.ctx, "添加退货/售后明细失败,参数:%+v,异常:%s", item, err.Error())
			return nil, errors.New("添加退货/售后明细失败")
		}
	}

	// 发布售后申请消息事件（统一事件契约模式）
	// Story 8-1 Fix #2: 改用 sendReturnEvent 确保 payload 包含 platformId/tenantId/merchantId
	sendReturnEvent(l.ctx, l.svcCtx, in.OrderId, order)

	return &omsclient.OrderReturnResp{}, nil
}

// sendReturnEvent 发布售后申请事件（统一事件契约模式）
// Story 8-1 Fix #2: 修复 payload 缺少 platformId/tenantId/merchantId
func sendReturnEvent(ctx context.Context, svcCtx *svc.ServiceContext, orderID int64, order *model.OmsOrderMain) {
	eventID := uuid.New().String()
	traceID := audit.NewTraceID("order.return", orderID)
	occurredAt := time.Now().UnixMilli()

	current := pkgscope.GovernanceScope{
		PlatformID: order.PlatformID,
		TenantID:   order.TenantID,
		MerchantID: order.MerchantID,
		ScopeType:  "tenant",
	}

	payload := ReturnEventPayload{
		EventID:    eventID,
		OccurredAt: occurredAt,
		TraceID:    traceID,
		PlatformID: current.PlatformID,
		TenantID:   current.TenantID,
		MerchantID: current.MerchantID,
		ActorID:    order.UserID,
		EntityID:   orderID,
		Action:     "order.return",
		Version:    "v1",
		ScopeType:  current.ScopeType,
		Data: map[string]interface{}{
			"orderNo": order.OrderNo,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		logc.Errorf(ctx, "序列化售后事件失败,orderId:%d,异常:%s", orderID, err.Error())
		return
	}

	if err := svcCtx.RabbitMQ.SendMessage("order.event.exchange", "direct", "order.return.queue", "order.return.key", body); err != nil {
		logc.Errorf(ctx, "发送售后异步消息失败,orderId:%d,scope:%+v,异常:%s", orderID, current, err.Error())
	}
	logc.Infof(ctx, "发送售后申请事件成功,eventId:%s,orderId:%d,traceId:%s", eventID, orderID, traceID)
}
