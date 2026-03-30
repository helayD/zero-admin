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
	OrderStatusPaid          = 2 // 已支付/待发货（OMS 2）
	OrderStatusShipped       = 3 // 已发货（OMS 3）
	OrderStatusCompleted     = 4 // 已完成（OMS 4）
	OrderStatusCancelled     = 5 // 已取消（OMS 5）
	OrderStatusRefunded      = 6 // 已退款（OMS 6）
	OrderStatusAfterSale     = 7 // 售后中（OMS 7）
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
	OpCreateOrder    = 1  // 创建订单（初始化）
	OpPaymentSuccess = 2  // 支付成功
	OpPaymentFailed  = 3  // 支付失败
	OpDelivery       = 4  // 商家发货（OMS Admin）
	OpConfirmReceive = 5  // 确认收货
	OpCancel         = 6  // 取消订单
	OpRefund         = 7  // 退款成功
	OpApplyAfterSale = 8 // 申请售后
	OpAfterSaleClose = 9  // 售后关闭
)

// validTransitions 定义合法状态转换矩阵
// key: 当前状态 → value: 允许的目标状态集合
var validTransitions = map[int][]int{
	OrderStatusPendingPayment: {OrderStatusPaid, OrderStatusCancelled},
	OrderStatusPaid:           {OrderStatusShipped, OrderStatusRefunded, OrderStatusCancelled},
	OrderStatusShipped:        {OrderStatusCompleted, OrderStatusRefunded, OrderStatusAfterSale},
	OrderStatusCompleted:      {OrderStatusRefunded, OrderStatusAfterSale},
	OrderStatusCancelled:      {},   // 终态
	OrderStatusRefunded:       {},   // 终态
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
		OrderStatusCompleted:       "已完成",
		OrderStatusCancelled:      "已取消",
		OrderStatusRefunded:        "已退款",
		OrderStatusAfterSale:       "售后中",
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
		PayStatusSuccess:  "支付成功",
		PayStatusFailed:   "支付失败",
	}
	if text, ok := payStatusMap[payStatus]; ok {
		return text
	}
	return "未知状态"
}

// GetOpTypeText 返回操作类型的中文描述
func GetOpTypeText(opType int) string {
	opTypeMap := map[int]string{
		OpCreateOrder:     "创建订单",
		OpPaymentSuccess:  "支付成功",
		OpPaymentFailed:   "支付失败",
		OpDelivery:        "商家发货",
		OpConfirmReceive:  "确认收货",
		OpCancel:          "取消订单",
		OpRefund:          "退款成功",
		OpApplyAfterSale:  "申请售后",
		OpAfterSaleClose:  "售后关闭",
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
