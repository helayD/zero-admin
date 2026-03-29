///
/// 订单状态相关
///
/// 作者：刘飞华
/// 日期：2023/11/21 17:17
///

/// OMS 订单状态 → 前端显示文本映射
/// OMS状态: 0=待支付,1=已支付/待发货,2=已取消,3=已完成,4=售后中
String getOmsOrderStatusTxt(int status) {
  switch (status) {
    case 0:
      return "等待付款";
    case 1:
      return "待发货";
    case 2:
      return "已取消";
    case 3:
      return "交易完成";
    case 4:
      return "售后中";
    default:
      return "未知状态";
  }
}

/// 兼容旧状态码（Story 5.x 使用）
/// 旧状态: 0->待付款；1->待发货；2->已发货；3->已完成；4->已关闭；5->无效订单
String getOrderStatusTxt(int status) {
  Map<int, String> statusMap = {
    0: "等待付款",
    1: "待发货",
    2: "等待收货",
    3: "交易完成",
    4: "已关闭",
    5: "无效订单",
  };
  return statusMap[status] ?? "未知状态";
}

/// 兼容旧状态码 → OMS状态映射（用于旧代码路径兼容）
int oldStatusToOmsStatus(int oldStatus) {
  switch (oldStatus) {
    case 0:
      return 0; // 待支付
    case 1:
      return 1; // 待发货
    case 2:
      return 1; // 已发货 → 归入待发货
    case 3:
      return 3; // 已完成
    case 4:
      return 2; // 已关闭 → 已取消
    case 5:
      return 2; // 无效订单 → 已取消
    default:
      return 0;
  }
}

/// Flutter Tab 索引 → OMS Status 映射（Story 6-1 Task 5.2）
/// Flutter Tab: 0=全部,1=待支付,2=待发货,3=已完成,4=已取消
int flutterTabToOmsStatus(int tabIndex) {
  switch (tabIndex) {
    case 0:
      return 0; // 全部
    case 1:
      return 1; // 待支付
    case 2:
      return 2; // 待发货（OMS:1=已支付=待发货）
    case 3:
      return 3; // 已完成
    case 4:
      return 4; // 已取消
    default:
      return 0;
  }
}

/// Flutter Tab 索引 → 后端请求 Status（Story 6-1 Task 5.2）
int flutterTabToBackendStatus(int tabIndex) {
  switch (tabIndex) {
    case 0:
      return 0; // 全部
    case 1:
      return 1; // 待支付
    case 2:
      return 2; // 待发货
    case 3:
      return 3; // 已完成
    case 4:
      return 4; // 已取消
    default:
      return 0;
  }
}
