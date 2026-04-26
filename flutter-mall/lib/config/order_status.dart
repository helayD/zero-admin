//
// 订单状态相关
//
// 作者：David
// 日期：2023/11/21 17:17
//

/// OMS 订单状态 → 前端显示文本映射（以 OMS proto 真实值为准）
/// OMS proto: 1=待支付, 2=已支付(待发货), 3=已发货, 4=已完成, 5=已取消, 7=售后中
/// 注意：后端 API 返回的 status 字段值即 OMS 真实值
String getOmsOrderStatusTxt(int status) {
  switch (status) {
    case 1:
      return "待发货"; // OMS 1=已支付（等待商家发货）
    case 2:
      return "已支付"; // OMS 2=已支付（与后端返回的 Flutter 语义对齐）
    case 3:
      return "已发货"; // OMS 3=已发货
    case 4:
      return "交易完成"; // OMS 4=已完成
    case 5:
      return "已取消"; // OMS 5=已取消
    case 7:
      return "售后中"; // OMS 7=售后中
    default:
      return "等待付款"; // OMS 0 或未知 → 显示为等待付款
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

/// Flutter Tab 索引 → OMS Status 映射
///
/// 重要约束（以 OMS proto 真实值为准）：
/// - OMS:     1=待支付, 2=已支付(待发货), 3=已发货, 4=已完成, 5=已取消, 7=售后中
/// - Flutter: 0=全部,   1=待支付,  2=待发货,    3=已完成,  4=已取消,  5=售后
///
/// 直接返回 OMS 真实值（等价于 flutterTabToOmsStatus，保留以兼容旧调用）
int flutterTabToOmsStatus(int tabIndex) {
  switch (tabIndex) {
    case 0:
      return 0; // 全部（后端 OMS 层不填条件，返回所有）
    case 1:
      return 1; // 待支付
    case 2:
      return 2; // 待发货（OMS: 2=已支付=待发货）
    case 3:
      return 4; // 已完成（OMS: 4=已完成）
    case 4:
      return 5; // 已取消（OMS: 5=已取消）
    case 5:
      return 7; // 售后（OMS: 7=售后中）
    default:
      return 0;
  }
}

/// Flutter Tab 索引 → 后端请求 Status
/// 后端 OrderListReq.Status 值与 OMS 真实值一致：
/// 0=全部, 1=待支付, 2=已支付(待发货), 4=已完成, 5=已取消, 7=售后中
int flutterTabToBackendStatus(int tabIndex) {
  switch (tabIndex) {
    case 0:
      return 0; // 全部
    case 1:
      return 1; // 待支付
    case 2:
      return 2; // 已支付(待发货)
    case 3:
      return 4; // 已完成
    case 4:
      return 5; // 已取消
    case 5:
      return 7; // 售后中
    default:
      return 0;
  }
}
