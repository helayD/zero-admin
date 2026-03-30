package order

import (
	"testing"
)

// TestIsValidTransition 测试状态转换合法性校验
func TestIsValidTransition(t *testing.T) {
	tests := []struct {
		name  string
		from  int
		to    int
		valid bool
	}{
		// 合法转换
		{"待支付→已支付", OrderStatusPendingPayment, OrderStatusPaid, true},
		{"待支付→已取消", OrderStatusPendingPayment, OrderStatusCancelled, true},
		{"已支付→已发货", OrderStatusPaid, OrderStatusShipped, true},
		{"已支付→已退款", OrderStatusPaid, OrderStatusRefunded, true},
		{"已支付→已取消", OrderStatusPaid, OrderStatusCancelled, true},
		{"已发货→已完成", OrderStatusShipped, OrderStatusCompleted, true},
		{"已发货→已退款", OrderStatusShipped, OrderStatusRefunded, true},
		{"已发货→售后中", OrderStatusShipped, OrderStatusAfterSale, true},
		{"已完成→已退款", OrderStatusCompleted, OrderStatusRefunded, true},
		{"已完成→售后中", OrderStatusCompleted, OrderStatusAfterSale, true},
		{"售后中→已完成", OrderStatusAfterSale, OrderStatusCompleted, true},
		{"售后中→已退款", OrderStatusAfterSale, OrderStatusRefunded, true},
		{"售后中→已取消", OrderStatusAfterSale, OrderStatusCancelled, true},

		// 非法转换
		{"待支付→已完成（跳过中间状态）", OrderStatusPendingPayment, OrderStatusCompleted, false},
		{"待支付→已发货（跳过）", OrderStatusPendingPayment, OrderStatusShipped, false},
		{"待支付→已退款", OrderStatusPendingPayment, OrderStatusRefunded, false},
		{"待支付→售后中", OrderStatusPendingPayment, OrderStatusAfterSale, false},
		{"已支付→已完成（跳过中间状态）", OrderStatusPaid, OrderStatusCompleted, false},
		{"已发货→已支付（反向）", OrderStatusShipped, OrderStatusPaid, false},
		{"已发货→待支付（反向）", OrderStatusShipped, OrderStatusPendingPayment, false},
		{"已完成→已发货（反向）", OrderStatusCompleted, OrderStatusShipped, false},
		{"已完成→已支付（反向）", OrderStatusCompleted, OrderStatusPaid, false},
		{"已完成→待支付（反向）", OrderStatusCompleted, OrderStatusPendingPayment, false},
		{"已取消→任意状态（终态不可逆）", OrderStatusCancelled, OrderStatusPaid, false},
		{"已取消→已完成（终态不可逆）", OrderStatusCancelled, OrderStatusCompleted, false},
		{"已退款→任意状态（终态不可逆）", OrderStatusRefunded, OrderStatusPaid, false},
		{"已退款→已完成（终态不可逆）", OrderStatusRefunded, OrderStatusCompleted, false},
		// OMS 不存在 0
		{"0→任意（OMS 不存在状态0）", 0, OrderStatusPaid, false},
		// 未知状态
		{"未知状态9", 9, OrderStatusPaid, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidTransition(tt.from, tt.to)
			if got != tt.valid {
				t.Errorf("IsValidTransition(%d, %d) = %v, want %v", tt.from, tt.to, got, tt.valid)
			}
		})
	}
}

// TestGetStatusText 测试状态文本描述
func TestGetStatusText(t *testing.T) {
	tests := []struct {
		status int
		want   string
	}{
		{OrderStatusPendingPayment, "待支付"},
		{OrderStatusPaid, "已支付"},
		{OrderStatusShipped, "已发货"},
		{OrderStatusCompleted, "已完成"},
		{OrderStatusCancelled, "已取消"},
		{OrderStatusRefunded, "已退款"},
		{OrderStatusAfterSale, "售后中"},
		{99, "未知状态"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := GetStatusText(tt.status)
			if got != tt.want {
				t.Errorf("GetStatusText(%d) = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

// TestGetPayStatusText 测试支付状态文本描述
func TestGetPayStatusText(t *testing.T) {
	tests := []struct {
		status int
		want   string
	}{
		{PayStatusPending, "待支付"},
		{PayStatusSuccess, "支付成功"},
		{PayStatusFailed, "支付失败"},
		{99, "未知状态"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := GetPayStatusText(tt.status)
			if got != tt.want {
				t.Errorf("GetPayStatusText(%d) = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

// TestIsFinalStatus 测试终态判断
func TestIsFinalStatus(t *testing.T) {
	tests := []struct {
		status  int
		isFinal bool
	}{
		{OrderStatusPendingPayment, false},
		{OrderStatusPaid, false},
		{OrderStatusShipped, false},
		{OrderStatusCompleted, false},
		{OrderStatusCancelled, true},
		{OrderStatusRefunded, true},
		{OrderStatusAfterSale, false},
	}

	for _, tt := range tests {
		t.Run(GetStatusText(tt.status), func(t *testing.T) {
			got := IsFinalStatus(tt.status)
			if got != tt.isFinal {
				t.Errorf("IsFinalStatus(%d) = %v, want %v", tt.status, got, tt.isFinal)
			}
		})
	}
}

// TestGetNextStatusByAction 测试根据操作获取目标状态
func TestGetNextStatusByAction(t *testing.T) {
	tests := []struct {
		name         string
		current      int
		action       int
		want         int
	}{
		// 支付成功
		{"支付成功-待支付→已支付", OrderStatusPendingPayment, OpPaymentSuccess, OrderStatusPaid},
		{"支付成功-已支付→不变", OrderStatusPaid, OpPaymentSuccess, 0},
		{"支付成功-已完成→不变", OrderStatusCompleted, OpPaymentSuccess, 0},
		// 支付失败
		{"支付失败-待支付→保持不变", OrderStatusPendingPayment, OpPaymentFailed, OrderStatusPendingPayment},
		// 确认收货
		{"确认收货-已发货→已完成", OrderStatusShipped, OpConfirmReceive, OrderStatusCompleted},
		{"确认收货-已支付→不变（需先发货）", OrderStatusPaid, OpConfirmReceive, 0},
		// 取消
		{"取消-待支付→已取消", OrderStatusPendingPayment, OpCancel, OrderStatusCancelled},
		{"取消-已支付→已取消", OrderStatusPaid, OpCancel, OrderStatusCancelled},
		{"取消-已发货→不变（需先收货或退款）", OrderStatusShipped, OpCancel, 0},
		{"取消-已完成→不变", OrderStatusCompleted, OpCancel, 0},
		// 申请售后
		{"申请售后-已完成→售后中", OrderStatusCompleted, OpApplyAfterSale, OrderStatusAfterSale},
		{"申请售后-售后中→保持售后中", OrderStatusAfterSale, OpApplyAfterSale, OrderStatusAfterSale},
		{"申请售后-已发货→售后中", OrderStatusShipped, OpApplyAfterSale, OrderStatusAfterSale},
		{"申请售后-待支付→不变", OrderStatusPendingPayment, OpApplyAfterSale, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetNextStatusByAction(tt.current, tt.action)
			if got != tt.want {
				t.Errorf("GetNextStatusByAction(%d, %d) = %d, want %d (%s→%s)",
					tt.current, tt.action, got, tt.want,
					GetStatusText(tt.current), GetStatusText(tt.want))
			}
		})
	}
}

// TestGetValidTargets 测试获取允许的目标状态列表
func TestGetValidTargets(t *testing.T) {
	tests := []struct {
		status   int
		expected []int
	}{
		{OrderStatusPendingPayment, []int{OrderStatusPaid, OrderStatusCancelled}},
		{OrderStatusPaid, []int{OrderStatusShipped, OrderStatusRefunded, OrderStatusCancelled}},
		{OrderStatusShipped, []int{OrderStatusCompleted, OrderStatusRefunded, OrderStatusAfterSale}},
		{OrderStatusCompleted, []int{OrderStatusRefunded, OrderStatusAfterSale}},
		{OrderStatusCancelled, []int{}},
		{OrderStatusRefunded, []int{}},
		{OrderStatusAfterSale, []int{OrderStatusCompleted, OrderStatusRefunded, OrderStatusCancelled}},
		{99, nil}, // 未知状态
	}

	for _, tt := range tests {
		t.Run(GetStatusText(tt.status), func(t *testing.T) {
			got := GetValidTargets(tt.status)
			if len(got) != len(tt.expected) {
				t.Errorf("GetValidTargets(%d) length = %d, want %d", tt.status, len(got), len(tt.expected))
				return
			}
			for i, v := range got {
				if v != tt.expected[i] {
					t.Errorf("GetValidTargets(%d)[%d] = %d, want %d", tt.status, i, v, tt.expected[i])
				}
			}
		})
	}
}

// TestGetTransitionDescription 测试状态转换描述
func TestGetTransitionDescription(t *testing.T) {
	tests := []struct {
		from  int
		to    int
		known bool // 是否已知描述
	}{
		{OrderStatusPendingPayment, OrderStatusPaid, true},
		{OrderStatusPendingPayment, OrderStatusCancelled, true},
		{OrderStatusShipped, OrderStatusCompleted, true},
		{OrderStatusCompleted, OrderStatusAfterSale, true},
		{99, 1, false}, // 未知状态
	}

	for _, tt := range tests {
		t.Run(GetStatusText(tt.from)+"→"+GetStatusText(tt.to), func(t *testing.T) {
			desc := GetTransitionDescription(tt.from, tt.to)
			if tt.known && desc == "未知状态转换" {
				t.Errorf("GetTransitionDescription(%d, %d) = %q, expected known description", tt.from, tt.to, desc)
			}
			if !tt.known && desc != "未知状态转换" {
				t.Errorf("GetTransitionDescription(%d, %d) = %q, expected unknown", tt.from, tt.to, desc)
			}
		})
	}
}
