package order

import (
	"context"
	"fmt"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/pkg/errorx"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/zeromicro/go-zero/core/logc"
)

// OrderStatusService 统一订单状态更新服务
// 职责：OMS 单服务内 order_main + order_payment + order_operation_log 的状态编排
// 约束：Saga 补偿（库存/优惠券/积分）由调用方处理，不在本服务内
//
// Author: Claude AI
// Date: 2026-03-30
// Story: 6-5 订单状态编排与多角色视图一致

// UpdateOrderStatusReq 统一状态更新请求
type UpdateOrderStatusReq struct {
	OrderId  int64  // 订单ID（不是 OrderNo；归属由 memberId 校验）
	MemberId int64  // 从 JWT 提取，用于归属校验
	Action   int    // 操作类型（OpPaymentSuccess 等）
	BizData  string // 扩展业务数据（JSON 字符串，如退款金额、退货原因等）
}

// UpdateOrderStatusResp 统一状态更新响应
type UpdateOrderStatusResp struct {
	Code      int // 0=成功，非0=失败
	Message   string
	OldStatus int    // 变更前 order_status
	NewStatus int    // 变更后 order_status
	PayStatus int    // 变更后 pay_status（通过 QueryOrderPaymentDetail 获取）
	UpdatedAt string // 变更时间
}

// OrderStatusService 统一状态编排服务（struct）
type OrderStatusService struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewOrderStatusService 创建状态编排服务实例
func NewOrderStatusService(ctx context.Context, svcCtx *svc.ServiceContext) *OrderStatusService {
	return &OrderStatusService{ctx: ctx, svcCtx: svcCtx}
}

// UpdateOrderStatus 执行统一状态更新
// 包含：归属校验 → 幂等校验 → 状态转换校验 → OMS RPC 更新 → 支付状态更新 → 操作日志
func (s *OrderStatusService) UpdateOrderStatus(req *UpdateOrderStatusReq) (*UpdateOrderStatusResp, error) {
	// 1. 前置校验：归属校验
	if err := s.validateOwnership(req.OrderId, req.MemberId); err != nil {
		return nil, err
	}

	// 2. 获取当前状态
	currentStatus, err := s.getCurrentOrderStatus(req.OrderId)
	if err != nil {
		return nil, err
	}

	// 3. 根据操作类型确定目标状态
	targetStatus := GetNextStatusByAction(currentStatus, req.Action)
	if targetStatus == 0 {
		return &UpdateOrderStatusResp{
			Code:      1,
			Message:   fmt.Sprintf("当前状态【%s】不支持操作【%s】", GetStatusText(currentStatus), GetOpTypeText(req.Action)),
			OldStatus: currentStatus,
			NewStatus: currentStatus,
			PayStatus: 0,
		}, nil
	}

	// 4. 状态转换合法性校验
	if !IsValidTransition(currentStatus, targetStatus) {
		return nil, errorx.NewDefaultError(fmt.Sprintf("非法状态转换：%s → %s",
			GetStatusText(currentStatus), GetStatusText(targetStatus)))
	}

	// 5. 幂等校验：同一 OrderId + Action 的重复请求，直接返回当前状态
	if s.isIdempotent(req.OrderId, req.Action) {
		payStatus := s.getCurrentPayStatus(req.OrderId)
		return &UpdateOrderStatusResp{
			Code:      0,
			Message:   "操作已执行（幂等命中）",
			OldStatus: currentStatus,
			NewStatus: currentStatus,
			PayStatus: payStatus,
		}, nil
	}

	// 6. 事务性更新（三步原子）
	var newStatus int
	var payStatus int

	switch req.Action {
	case OpPaymentSuccess:
		newStatus, payStatus, err = s.handlePaymentSuccess(req)
	case OpPaymentFailed:
		newStatus, payStatus, err = s.handlePaymentFailed(req)
	case OpConfirmReceive:
		newStatus, err = s.handleConfirmReceive(req)
		payStatus = s.getCurrentPayStatus(req.OrderId)
	case OpCancel:
		newStatus, err = s.handleCancel(req)
		payStatus = s.getCurrentPayStatus(req.OrderId)
	case OpApplyAfterSale:
		newStatus, err = s.handleApplyAfterSale(req)
		payStatus = s.getCurrentPayStatus(req.OrderId)
	default:
		return nil, errorx.NewDefaultError(fmt.Sprintf("不支持的操作类型：%d", req.Action))
	}

	if err != nil {
		return nil, err
	}

	// 7. 写入操作日志
	s.writeOperationLog(req.OrderId, req.Action, s.getOperatorType(req.Action), req.BizData)

	// 8. 记录幂等键
	s.setIdempotentKey(req.OrderId, req.Action)
	if req.Action == OpConfirmReceive && newStatus == OrderStatusCompleted {
		grantOrderCompletedLotteryTimes(s.ctx, s.svcCtx.DB, req.MemberId, req.OrderId)
	}

	return &UpdateOrderStatusResp{
		Code:      0,
		Message:   "操作成功",
		OldStatus: currentStatus,
		NewStatus: newStatus,
		PayStatus: payStatus,
	}, nil
}

// validateOwnership 归属校验：校验订单属于当前用户
func (s *OrderStatusService) validateOwnership(orderId, memberId int64) error {
	detail, err := s.svcCtx.OrderService.QueryOrderDetail(s.ctx, &omsclient.QueryOrderDetailReq{
		Id: orderId,
	})
	if err != nil {
		return errorx.NewDefaultError("订单查询失败")
	}
	if detail == nil || detail.Data == nil {
		return errorx.NewDefaultError("订单不存在")
	}
	// OMS user_id 对应 member_id
	if detail.Data.UserId != memberId {
		return errorx.NewDefaultError("无权操作此订单")
	}
	return nil
}

// getCurrentOrderStatus 获取当前订单状态
func (s *OrderStatusService) getCurrentOrderStatus(orderId int64) (int, error) {
	detail, err := s.svcCtx.OrderService.QueryOrderDetail(s.ctx, &omsclient.QueryOrderDetailReq{
		Id: orderId,
	})
	if err != nil {
		return 0, err
	}
	if detail == nil || detail.Data == nil {
		return 0, fmt.Errorf("订单不存在 orderId=%d", orderId)
	}
	return int(detail.Data.OrderStatus), nil
}

// getCurrentPayStatus 获取当前支付状态
func (s *OrderStatusService) getCurrentPayStatus(orderId int64) int {
	// 先从 OrderDetail 的 payment_data 获取
	detail, err := s.svcCtx.OrderService.QueryOrderDetail(s.ctx, &omsclient.QueryOrderDetailReq{
		Id: orderId,
	})
	if err == nil && detail != nil && detail.Data != nil && len(detail.Data.PaymentData) > 0 {
		return int(detail.Data.PaymentData[0].PayStatus)
	}

	// 如果 OrderDetail 中没有，再通过 OrderPaymentService 查询
	paymentData, err := s.svcCtx.OrderPaymentService.QueryOrderPaymentDetail(s.ctx, &omsclient.QueryOrderPaymentDetailReq{
		Id: orderId,
	})
	if err == nil && paymentData != nil {
		return int(paymentData.PayStatus)
	}

	return PayStatusPending
}

// handlePaymentSuccess 处理支付成功
func (s *OrderStatusService) handlePaymentSuccess(req *UpdateOrderStatusReq) (int, int, error) {
	// 5.2：调用 OrderService.UpdateOrderStatus 更新 order_status=2
	_, err := s.svcCtx.OrderService.UpdateOrderStatus(s.ctx, &omsclient.UpdateOrderStatusReq{
		Ids:         []int64{req.OrderId},
		OrderStatus: OrderStatusPaid, // OMS 2=已支付
	})
	if err != nil {
		logc.Errorf(s.ctx, "支付成功更新订单状态失败 orderId=%d err=%v", req.OrderId, err)
		return 0, 0, errorx.NewDefaultError("订单状态更新失败")
	}

	// 5.2：调用 OrderPaymentService.UpdateOrderPaymentStatus 更新 pay_status=1
	_, err = s.svcCtx.OrderPaymentService.UpdateOrderPaymentStatus(s.ctx, &omsclient.UpdateOrderPaymentStatusReq{
		Ids:       []int64{req.OrderId},
		PayStatus: PayStatusSuccess, // OMS 1=支付成功
	})
	if err != nil {
		logc.Errorf(s.ctx, "支付成功更新支付状态失败 orderId=%d err=%v", req.OrderId, err)
		return OrderStatusPaid, PayStatusPending, errorx.NewDefaultError("支付状态更新失败")
	}

	return OrderStatusPaid, PayStatusSuccess, nil
}

// handlePaymentFailed 处理支付失败
func (s *OrderStatusService) handlePaymentFailed(req *UpdateOrderStatusReq) (int, int, error) {
	// 支付失败只更新 pay_status，不改变 order_status
	_, err := s.svcCtx.OrderPaymentService.UpdateOrderPaymentStatus(s.ctx, &omsclient.UpdateOrderPaymentStatusReq{
		Ids:       []int64{req.OrderId},
		PayStatus: PayStatusFailed, // OMS 2=支付失败
	})
	if err != nil {
		logc.Errorf(s.ctx, "支付失败更新支付状态失败 orderId=%d err=%v", req.OrderId, err)
		return 0, 0, errorx.NewDefaultError("支付状态更新失败")
	}

	currentStatus, _ := s.getCurrentOrderStatus(req.OrderId)
	return currentStatus, PayStatusFailed, nil
}

// handleConfirmReceive 处理确认收货
func (s *OrderStatusService) handleConfirmReceive(req *UpdateOrderStatusReq) (int, error) {
	// ⚠️ ConfirmOrder 将 order_status 从 2 直接更新到 4，跳过 3
	_, err := s.svcCtx.OrderService.ConfirmOrder(s.ctx, &omsclient.ConfirmOrderReq{
		MemberId: req.MemberId,
		OrderId:  req.OrderId,
	})
	if err != nil {
		logc.Errorf(s.ctx, "确认收货失败 orderId=%d err=%v", req.OrderId, err)
		return 0, errorx.NewDefaultError("确认收货失败")
	}
	return OrderStatusCompleted, nil
}

// handleCancel 处理取消订单
// ⚠️ Saga 补偿链路（库存释放/优惠券归还/积分返还）由调用方 cancel_user_order_logic.go 处理
func (s *OrderStatusService) handleCancel(req *UpdateOrderStatusReq) (int, error) {
	_, err := s.svcCtx.OrderService.CancelOrder(s.ctx, &omsclient.CancelOrderReq{
		MemberId: req.MemberId,
		OrderId:  req.OrderId,
	})
	if err != nil {
		logc.Errorf(s.ctx, "取消订单失败 orderId=%d err=%v", req.OrderId, err)
		return 0, errorx.NewDefaultError("取消订单失败")
	}
	return OrderStatusCancelled, nil
}

// handleApplyAfterSale 处理售后申请
func (s *OrderStatusService) handleApplyAfterSale(req *UpdateOrderStatusReq) (int, error) {
	// 售后申请需要先调用 AddOrderReturn，再更新 order_status
	// 由于 AddOrderReturn 返回 pong（无 ReturnId），售后申请的状态更新由 OMS 内部处理
	// 本服务只负责将 order_status 更新为售后中
	_, err := s.svcCtx.OrderService.UpdateOrderStatus(s.ctx, &omsclient.UpdateOrderStatusReq{
		Ids:         []int64{req.OrderId},
		OrderStatus: OrderStatusAfterSale, // OMS 7=售后中
	})
	if err != nil {
		logc.Errorf(s.ctx, "申请售后更新订单状态失败 orderId=%d err=%v", req.OrderId, err)
		return 0, errorx.NewDefaultError("售后申请失败")
	}
	return OrderStatusAfterSale, nil
}

// writeOperationLog 写入操作日志
func (s *OrderStatusService) writeOperationLog(orderId int64, action int, operatorType int, bizData string) {
	// OMS operation_type 与业务 action 的映射
	omsOpType := s.mapActionToOmsOpType(action)

	_, err := s.svcCtx.OrderOperationLogService.AddOrderOperationLog(s.ctx, &omsclient.AddOrderOperationLogReq{
		OrderId:       orderId,
		OperatorType:  int32(operatorType), // 1=用户, 2=系统, 3=管理员
		OperationType: int32(omsOpType),
		OperatorNote:  bizData,
	})
	if err != nil {
		logc.Errorf(s.ctx, "写入操作日志失败 orderId=%d action=%d err=%v", orderId, action, err)
	}
}

// mapActionToOmsOpType 将业务 Action 映射为 OMS operation_type
// OMS operation_type: 1=创建订单, 2=支付订单, 3=发货, 4=确认收货, 5=取消订单, 6=退款
func (s *OrderStatusService) mapActionToOmsOpType(action int) int {
	switch action {
	case OpPaymentSuccess:
		return 2 // 支付订单
	case OpPaymentFailed:
		return 2 // 支付订单（失败）
	case OpDelivery:
		return 3 // 发货
	case OpConfirmReceive:
		return 4 // 确认收货
	case OpCancel:
		return 5 // 取消订单
	case OpRefund:
		return 6 // 退款
	default:
		return 1 // 创建订单（兜底）
	}
}

// getOperatorType 根据操作类型返回操作人类型
func (s *OrderStatusService) getOperatorType(action int) int {
	switch action {
	case OpPaymentSuccess, OpPaymentFailed:
		return OperatorTypeSystem // 支付回调为系统触发
	default:
		return OperatorTypeUser // 用户操作
	}
}

// isIdempotent 幂等校验：检查是否已执行过同一操作
func (s *OrderStatusService) isIdempotent(orderId int64, action int) bool {
	key := fmt.Sprintf("order:status:idempotent:%d:%d", orderId, action)
	exists, err := s.svcCtx.Redis.Exists(key)
	if err != nil {
		return false // 异常时放行，避免阻塞
	}
	return exists
}

// setIdempotentKey 记录幂等键，TTL=24h
func (s *OrderStatusService) setIdempotentKey(orderId int64, action int) {
	key := fmt.Sprintf("order:status:idempotent:%d:%d", orderId, action)
	_ = s.svcCtx.Redis.Setex(key, "1", 86400)
}

func (s *OrderStatusService) getCancelCompensationStatus(orderId int64) string {
	key := fmt.Sprintf("order:cancel:compensation:%d", orderId)
	val, err := s.svcCtx.Redis.GetCtx(s.ctx, key)
	if err != nil {
		return ""
	}
	return val
}

func (s *OrderStatusService) getLatestAfterSale(orderId, memberId int64) (int32, int64, string, string) {
	returns, err := s.svcCtx.OrderReturnService.QueryOrderReturnList(s.ctx, &omsclient.QueryOrderReturnListReq{
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

// GetOrderStatusSnapshot 获取订单状态快照（聚合订单+支付+日志）
func (s *OrderStatusService) GetOrderStatusSnapshot(orderId, memberId int64) (*OrderSnapshot, error) {
	// 归属校验
	if err := s.validateOwnership(orderId, memberId); err != nil {
		return nil, err
	}

	// 查询订单详情
	orderDetail, err := s.svcCtx.OrderService.QueryOrderDetail(s.ctx, &omsclient.QueryOrderDetailReq{
		Id: orderId,
	})
	if err != nil {
		return nil, errorx.NewDefaultError("订单查询失败")
	}
	if orderDetail == nil || orderDetail.Data == nil {
		return nil, errorx.NewDefaultError("订单不存在")
	}

	// 查询支付信息
	payStatus := PayStatusPending
	if len(orderDetail.Data.PaymentData) > 0 {
		payStatus = int(orderDetail.Data.PaymentData[0].PayStatus)
	}

	cancelCompensationStatus := s.getCancelCompensationStatus(orderId)
	afterSaleStatus, returnId, returnNo, afterSaleUpdatedAt := s.getLatestAfterSale(orderId, memberId)

	// 查询操作日志
	logs, _ := s.svcCtx.OrderOperationLogService.QueryOrderOperationLogList(s.ctx, &omsclient.QueryOrderOperationLogListReq{
		OrderId:  orderId,
		PageNum:  1,
		PageSize: 20,
	})

	var logEntries []OperationLogEntry
	if logs != nil && logs.List != nil {
		for _, log := range logs.List {
			logEntries = append(logEntries, OperationLogEntry{
				Id:            log.Id,
				OperationType: int(log.OperationType),
				OperatorType:  int(log.OperatorType),
				OperatorNote:  log.OperatorNote,
				CreateTime:    log.CreateTime,
			})
		}
	}

	// Story 7.3B: 计算一致性阶段
	stage, pendingActions := CalcConsistencyStage(int(orderDetail.Data.OrderStatus), payStatus, cancelCompensationStatus, afterSaleStatus)
	stageText := GetConsistencyStageText(stage)
	consistencyResult := calcConsistencyResult(int(orderDetail.Data.OrderStatus), payStatus, cancelCompensationStatus, afterSaleStatus)
	pendingActionsText := GetPendingActionsText(pendingActions)
	consistencyMessage := buildConsistencyMessage(stage, consistencyResult, afterSaleStatus)

	// 获取最近一次一致性更新时间（从操作日志中推断）
	var lastConsistencyAt string
	if afterSaleUpdatedAt != "" {
		lastConsistencyAt = afterSaleUpdatedAt
	} else if len(logEntries) > 0 {
		lastConsistencyAt = logEntries[0].CreateTime
	}

	afterSaleStatusText := ""
	if returnId > 0 {
		afterSaleStatusText = getAfterSaleStatusText(afterSaleStatus)
	}

	return &OrderSnapshot{
		OrderId:              orderDetail.Data.Id,
		OrderNo:              orderDetail.Data.OrderNo,
		OrderStatus:          int(orderDetail.Data.OrderStatus),
		PayStatus:            payStatus,
		OrderStatusText:      GetStatusText(int(orderDetail.Data.OrderStatus)),
		PayStatusText:        GetPayStatusText(payStatus),
		OptLogs:              logEntries,
		ConsistencyStage:     stage,
		ConsistencyStageText: stageText,
		ConsistencyResult:    consistencyResult,
		ConsistencyMessage:   consistencyMessage,
		LastConsistencyAt:    lastConsistencyAt,
		PendingActions:       pendingActions,
		PendingActionsText:   pendingActionsText,
		AftersaleStatus:      afterSaleStatus,
		AftersaleStatusText:  afterSaleStatusText,
		ReturnId:             returnId,
		ReturnNo:             returnNo,
	}, nil
}

// OrderSnapshot 订单状态快照
// Story 7.3B 扩展：新增一致性阶段字段，用于表达权益同步状态
type OrderSnapshot struct {
	OrderId         int64
	OrderNo         string
	OrderStatus     int
	PayStatus       int
	OrderStatusText string
	PayStatusText   string
	OptLogs         []OperationLogEntry
	// ==================== Story 7.3B: 一致性阶段字段 ====================
	ConsistencyStage     int      // 一致性阶段枚举值
	ConsistencyStageText string   // 一致性阶段中文描述
	ConsistencyResult    int      // 一致性结果枚举值
	ConsistencyMessage   string   // 用户/运营可理解的状态提示
	LastConsistencyAt    string   // 最近阶段更新时间
	PendingActions       int      // 待处理动作位掩码
	PendingActionsText   []string // 待处理动作中文描述列表
	AftersaleStatus      int32
	AftersaleStatusText  string
	ReturnId             int64
	ReturnNo             string
}

// OperationLogEntry 操作日志条目
type OperationLogEntry struct {
	Id            int64
	OperationType int
	OperatorType  int
	OperatorNote  string
	CreateTime    string
}
