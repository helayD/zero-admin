package order

// OrderStateMachine 订单状态机
// OMS order_status: 1=待支付, 2=已支付, 3=已发货, 4=已完成, 5=已取消, 6=已退款, 7=售后中
// ⚠️ OMS order_status 从 1 开始，不存在 0
// ⚠️ OMS 跳过了 order_status=3 注释；ConfirmOrder 将 2→4，商家发货由 OMS Admin 执行
//
// Author: Claude AI
// Date: 2026-03-30
// Story: 6-5 订单状态编排与多角色视图一致

// OMS 订单主状态常量（来自 rpc/oms/proto/order_main.proto）
const (
	OrderStatusPendingPayment = 1 // 待支付（OMS 1）
	OrderStatusPaid           = 2 // 已支付/待发货（OMS 2）
	OrderStatusShipped        = 3 // 已发货（OMS 3）
	OrderStatusCompleted      = 4 // 已完成（OMS 4）
	OrderStatusCancelled      = 5 // 已取消（OMS 5）
	OrderStatusRefunded       = 6 // 已退款（OMS 6）
	OrderStatusAfterSale      = 7 // 售后中（OMS 7）
)

// OMS 支付状态常量（来自 rpc/oms/proto/order_payment.proto）
const (
	PayStatusPending = 0 // 待支付
	PayStatusSuccess = 1 // 支付成功
	PayStatusFailed  = 2 // 支付失败
)

// OMS 操作人类型常量（来自 rpc/oms/proto/order_operation_log.proto）
const (
	OperatorTypeUser   = 1 // 用户操作
	OperatorTypeSystem = 2 // 系统操作
	OperatorTypeAdmin  = 3 // 管理员操作
)

// 状态转换操作类型（业务语义，非 OMS operation_type）
const (
	OpCreateOrder    = 1 // 创建订单（初始化）
	OpPaymentSuccess = 2 // 支付成功
	OpPaymentFailed  = 3 // 支付失败
	OpDelivery       = 4 // 商家发货（OMS Admin）
	OpConfirmReceive = 5 // 确认收货
	OpCancel         = 6 // 取消订单
	OpRefund         = 7 // 退款成功
	OpApplyAfterSale = 8 // 申请售后
	OpAfterSaleClose = 9 // 售后关闭
)

// ==================== Story 7.3B: 一致性阶段模型 ====================
// 用途：表达"权益一致性阶段"，用于前后台展示库存/优惠券/积分等下游同步状态
// 约束：是 OMS 主状态的辅助视图，不替换 order_status / pay_status

// ConsistencyStage 一致性阶段枚举
// 0=无阶段(正常主状态流程), 1=支付确认中, 2=支付成功同步中, 3=支付失败
// 4=取消回退中, 5=取消回退完成, 6=售后待处理, 7=售后处理中, 8=售后完成, 9=全部同步完成
const (
	ConsistencyStageNone                = 0 // 无一致性阶段（正常主状态流程，不需要同步）
	ConsistencyStagePaying              = 1 // 支付确认中（第三方回调处理中）
	ConsistencyStagePaySuccess          = 2 // 支付成功同步中（积分/优惠券确认中）
	ConsistencyStagePayFailed           = 3 // 支付失败（无需同步下游）
	ConsistencyStageCancelling          = 4 // 取消回退中（库存释放/优惠券返还/积分返还处理中）
	ConsistencyStageCancelled           = 5 // 取消回退完成
	ConsistencyStageAfterSalePending    = 6 // 售后待处理
	ConsistencyStageAfterSaleProcessing = 7 // 售后处理中（退款/库存/优惠券/积分处理中）
	ConsistencyStageAfterSaleCompleted  = 8 // 售后完成
	ConsistencyStageCompleted           = 9 // 全部同步完成（权益一致性已达成）
)

// ConsistencyResult 一致性结果枚举
const (
	ConsistencyResultNone           = 0 // 无结果
	ConsistencyResultProcessing     = 1 // 处理中
	ConsistencyResultSucceeded      = 2 // 成功
	ConsistencyResultFailed         = 3 // 失败
	ConsistencyResultManualRequired = 4 // 需要人工介入
)

// PendingAction 待处理动作枚举（位掩码，可组合）
const (
	PendingActionNone   = 0 // 无待处理
	PendingActionStock  = 1 // 库存处理
	PendingActionCoupon = 2 // 优惠券处理
	PendingActionPoints = 4 // 积分处理
)

// validTransitions 定义合法状态转换矩阵
// key: 当前状态 → value: 允许的目标状态集合
var validTransitions = map[int][]int{
	OrderStatusPendingPayment: {OrderStatusPaid, OrderStatusCancelled},
	OrderStatusPaid:           {OrderStatusShipped, OrderStatusRefunded, OrderStatusCancelled},
	OrderStatusShipped:        {OrderStatusCompleted, OrderStatusRefunded, OrderStatusAfterSale},
	OrderStatusCompleted:      {OrderStatusRefunded, OrderStatusAfterSale},
	OrderStatusCancelled:      {}, // 终态
	OrderStatusRefunded:       {}, // 终态
	OrderStatusAfterSale:      {OrderStatusCompleted, OrderStatusRefunded, OrderStatusCancelled},
}

// GetValidTargets 返回给定状态允许转换到的目标状态列表
func GetValidTargets(currentStatus int) []int {
	if targets, ok := validTransitions[currentStatus]; ok {
		return targets
	}
	return nil
}

// IsValidTransition 校验状态转换是否合法
func IsValidTransition(from, to int) bool {
	targets, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, t := range targets {
		if t == to {
			return true
		}
	}
	return false
}

// GetTransitionDescription 返回状态转换的业务说明
func GetTransitionDescription(from, to int) string {
	descMap := map[string]string{
		"1->2": "支付成功，订单从待支付变为已支付",
		"1->5": "用户取消订单，订单从待支付变为已取消",
		"2->3": "商家发货，订单从已支付变为已发货",
		"2->5": "用户取消订单，订单从已支付变为已取消",
		"2->6": "退款完成，订单从已支付变为已退款",
		"3->4": "用户确认收货，订单从已发货变为已完成",
		"3->6": "退款完成，订单从已发货变为已退款",
		"3->7": "用户申请售后，订单从已发货变为售后中",
		"4->6": "退款完成，订单从已完成变为已退款",
		"4->7": "用户申请售后，订单从已完成变为售后中",
		"7->4": "售后关闭，订单从售后中变为已完成",
		"7->5": "售后被拒绝/取消，订单从售后中变为已取消",
		"7->6": "退款完成，订单从售后中变为已退款",
	}

	key := itoa(from) + "->" + itoa(to)
	if desc, ok := descMap[key]; ok {
		return desc
	}
	return "未知状态转换"
}

// GetStatusText 返回状态的中文描述
func GetStatusText(status int) string {
	statusMap := map[int]string{
		OrderStatusPendingPayment: "待支付",
		OrderStatusPaid:           "已支付",
		OrderStatusShipped:        "已发货",
		OrderStatusCompleted:      "已完成",
		OrderStatusCancelled:      "已取消",
		OrderStatusRefunded:       "已退款",
		OrderStatusAfterSale:      "售后中",
	}
	if text, ok := statusMap[status]; ok {
		return text
	}
	return "未知状态"
}

// GetPayStatusText 返回支付状态的中文描述
func GetPayStatusText(payStatus int) string {
	payStatusMap := map[int]string{
		PayStatusPending: "待支付",
		PayStatusSuccess: "支付成功",
		PayStatusFailed:  "支付失败",
	}
	if text, ok := payStatusMap[payStatus]; ok {
		return text
	}
	return "未知状态"
}

// GetOpTypeText 返回操作类型的中文描述
func GetOpTypeText(opType int) string {
	opTypeMap := map[int]string{
		OpCreateOrder:    "创建订单",
		OpPaymentSuccess: "支付成功",
		OpPaymentFailed:  "支付失败",
		OpDelivery:       "商家发货",
		OpConfirmReceive: "确认收货",
		OpCancel:         "取消订单",
		OpRefund:         "退款成功",
		OpApplyAfterSale: "申请售后",
		OpAfterSaleClose: "售后关闭",
	}
	if text, ok := opTypeMap[opType]; ok {
		return text
	}
	return "未知操作"
}

// IsFinalStatus 判断是否为终态（不可再转换）
func IsFinalStatus(status int) bool {
	return status == OrderStatusCancelled || status == OrderStatusRefunded
}

// ==================== Story 7.3B: 一致性阶段相关函数 ====================

// GetConsistencyStageText 返回一致性阶段的中文描述
func GetConsistencyStageText(stage int) string {
	stageMap := map[int]string{
		ConsistencyStageNone:                "正常",
		ConsistencyStagePaying:              "支付确认中",
		ConsistencyStagePaySuccess:          "权益同步中",
		ConsistencyStagePayFailed:           "支付失败",
		ConsistencyStageCancelling:          "取消回退中",
		ConsistencyStageCancelled:           "已取消",
		ConsistencyStageAfterSalePending:    "售后待处理",
		ConsistencyStageAfterSaleProcessing: "售后处理中",
		ConsistencyStageAfterSaleCompleted:  "售后已完成",
		ConsistencyStageCompleted:           "已完成",
	}
	if text, ok := stageMap[stage]; ok {
		return text
	}
	return "未知阶段"
}

// GetConsistencyResultText 返回一致性结果的中文描述
func GetConsistencyResultText(result int) string {
	resultMap := map[int]string{
		ConsistencyResultNone:           "",
		ConsistencyResultProcessing:     "处理中",
		ConsistencyResultSucceeded:      "成功",
		ConsistencyResultFailed:         "失败",
		ConsistencyResultManualRequired: "需人工处理",
	}
	if text, ok := resultMap[result]; ok {
		return text
	}
	return "未知"
}

// GetPendingActionsText 返回待处理动作的中文描述列表
func GetPendingActionsText(actions int) []string {
	var result []string
	if actions&PendingActionStock != 0 {
		result = append(result, "库存")
	}
	if actions&PendingActionCoupon != 0 {
		result = append(result, "优惠券")
	}
	if actions&PendingActionPoints != 0 {
		result = append(result, "积分")
	}
	return result
}

// CalcConsistencyStage 根据订单状态和支付状态计算一致性阶段
// 这是 Story 7.3B 的核心逻辑：基于 OMS 主状态推断当前权益一致性阶段
func CalcConsistencyStage(orderStatus, payStatus int, cancelCompensationStatus string, afterSaleStatus int32) (stage int, pendingActions int) {
	switch orderStatus {
	case OrderStatusPendingPayment:
		if payStatus == PayStatusFailed {
			return ConsistencyStagePayFailed, PendingActionNone
		}
		return ConsistencyStageNone, PendingActionNone
	case OrderStatusPaid:
		return ConsistencyStagePaySuccess, PendingActionCoupon | PendingActionPoints
	case OrderStatusShipped, OrderStatusCompleted:
		return ConsistencyStageCompleted, PendingActionNone
	case OrderStatusCancelled:
		if payStatus == PayStatusSuccess {
			if cancelCompensationStatus == "completed" {
				return ConsistencyStageCancelled, PendingActionNone
			}
			return ConsistencyStageCancelling, PendingActionStock | PendingActionCoupon | PendingActionPoints
		}
		return ConsistencyStageCancelled, PendingActionNone
	case OrderStatusRefunded:
		if afterSaleStatus != 0 {
			return ConsistencyStageAfterSaleCompleted, PendingActionNone
		}
		return ConsistencyStageCompleted, PendingActionNone
	case OrderStatusAfterSale:
		switch afterSaleStatus {
		case 0:
			return ConsistencyStageAfterSalePending, PendingActionNone
		case 1, 2:
			return ConsistencyStageAfterSaleProcessing, PendingActionStock | PendingActionCoupon | PendingActionPoints
		case 3, 4, 5:
			return ConsistencyStageAfterSaleCompleted, PendingActionNone
		default:
			return ConsistencyStageAfterSalePending, PendingActionNone
		}
	}
	return ConsistencyStageNone, PendingActionNone
}

func calcConsistencyResult(orderStatus, payStatus int, cancelCompensationStatus string, afterSaleStatus int32) int {
	switch orderStatus {
	case OrderStatusPendingPayment:
		if payStatus == PayStatusFailed {
			return ConsistencyResultFailed
		}
		return ConsistencyResultNone
	case OrderStatusPaid:
		return ConsistencyResultProcessing
	case OrderStatusCancelled:
		if payStatus == PayStatusSuccess && cancelCompensationStatus != "completed" {
			return ConsistencyResultProcessing
		}
		return ConsistencyResultSucceeded
	case OrderStatusAfterSale:
		switch afterSaleStatus {
		case 4:
			return ConsistencyResultFailed
		case 3, 5:
			return ConsistencyResultSucceeded
		default:
			return ConsistencyResultProcessing
		}
	case OrderStatusRefunded, OrderStatusShipped, OrderStatusCompleted:
		return ConsistencyResultSucceeded
	default:
		return ConsistencyResultNone
	}
}

func buildConsistencyMessage(stage, result int, afterSaleStatus int32) string {
	switch stage {
	case ConsistencyStagePaying:
		return "支付确认中，请稍候..."
	case ConsistencyStagePaySuccess:
		return "支付成功，权益同步中..."
	case ConsistencyStagePayFailed:
		return "支付失败，请重新支付"
	case ConsistencyStageCancelling:
		return "取消处理中，正在回退权益..."
	case ConsistencyStageCancelled:
		return "已取消，权益已回退"
	case ConsistencyStageAfterSalePending:
		if afterSaleStatus == 0 {
			return "售后申请已提交，等待审核..."
		}
		return "售后待处理"
	case ConsistencyStageAfterSaleProcessing:
		switch afterSaleStatus {
		case 1:
			return "售后审核已通过，等待进一步处理..."
		case 2:
			return "售后处理中，请耐心等待..."
		default:
			return "售后处理中，请耐心等待..."
		}
	case ConsistencyStageAfterSaleCompleted:
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
	case ConsistencyStageCompleted:
		if result == ConsistencyResultSucceeded {
			return "订单处理完成"
		}
	}
	return ""
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

// GetNextStatusByAction 根据操作类型返回目标状态
// 如果操作无法产生有效的状态转换，返回 0
func GetNextStatusByAction(currentStatus, action int) int {
	switch action {
	case OpPaymentSuccess:
		if currentStatus == OrderStatusPendingPayment {
			return OrderStatusPaid
		}
	case OpPaymentFailed:
		// 支付失败不改变 order_status
		return currentStatus
	case OpDelivery:
		if currentStatus == OrderStatusPaid {
			return OrderStatusShipped
		}
	case OpConfirmReceive:
		// OMS ConfirmOrder 直接 2→4，跳过 3
		if currentStatus == OrderStatusShipped {
			return OrderStatusCompleted
		}
	case OpCancel:
		if currentStatus == OrderStatusPendingPayment || currentStatus == OrderStatusPaid {
			return OrderStatusCancelled
		}
	case OpRefund:
		if currentStatus == OrderStatusPaid || currentStatus == OrderStatusShipped ||
			currentStatus == OrderStatusCompleted || currentStatus == OrderStatusAfterSale {
			return OrderStatusRefunded
		}
	case OpApplyAfterSale:
		if currentStatus == OrderStatusCompleted || currentStatus == OrderStatusAfterSale ||
			currentStatus == OrderStatusShipped {
			return OrderStatusAfterSale
		}
	case OpAfterSaleClose:
		if currentStatus == OrderStatusAfterSale {
			return OrderStatusCompleted
		}
	}
	return 0
}

// itoa 将 int 转为 string（不使用 strconv 简化依赖）
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + uitoa(uint(-n))
	}
	return uitoa(uint(n))
}

// uitoa 将 uint 转为 string
func uitoa(val uint) string {
	if val == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf) - 1
	for val >= 10 {
		q := val / 10
		buf[i] = byte('0' + byte(val-q*10))
		i--
		val = q
	}
	buf[i] = byte('0' + byte(val))
	return string(buf[i:])
}
