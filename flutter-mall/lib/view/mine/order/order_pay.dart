import 'package:flutter/material.dart';
import 'package:flutter_mall/view/mine/order/pay_test.dart';

///
/// 订单支付页面（Story 5.4 重构）
///
/// 重构内容:
/// - 增加 orderId / orderSn / payType 参数，支持幂等命中后跳转已有订单
/// - 展示订单摘要（订单号 / 金额 / 支付方式）
///
/// 作者：LiuFeiHua
/// 日期：2023/11/21 17:17
/// Story 5.4: Task 8 重构
///
class OrderPay extends StatefulWidget {
  /// 订单ID（Story 5.4 新增）
  final int? orderId;

  /// 订单编号（Story 5.4 新增）
  final String? orderSn;

  /// 支付方式：1=支付宝，2=微信（Story 5.4 新增）
  final int? payType;

  /// 支付金额（元）
  final double amount;

  const OrderPay({
    super.key,
    this.orderId,
    this.orderSn,
    this.payType,
    required this.amount,
  });

  @override
  State<OrderPay> createState() => _OrderPayState();
}

class _OrderPayState extends State<OrderPay> {
  /// 当前选中的支付方式（默认使用下单时选择的）
  late int _selectedPayType;

  @override
  void initState() {
    super.initState();
    _selectedPayType = widget.payType ?? 2; // 默认微信
  }

  @override
  Widget build(BuildContext context) {
    final border = BorderSide(
      width: 1,
      color: const Color(0xFFF5F5F5),
    );
    final textPrimary = const Color(0xFF303133);
    final textSecondary = const Color(0xFF909399);
    final themeColor = const Color(0xFFFA436A);

    return Scaffold(
      appBar: AppBar(
        backgroundColor: Colors.white,
        title: const Text("支付"),
        titleTextStyle: const TextStyle(fontSize: 16, color: Colors.black),
        centerTitle: true,
      ),
      body: Container(
        color: Colors.white,
        width: MediaQuery.of(context).size.width,
        padding: const EdgeInsets.only(top: 44),
        child: Column(
          children: [
            // === Story 5.4 Task 8.2: 订单摘要卡片 ===
            if (widget.orderSn != null && widget.orderSn!.isNotEmpty)
              Container(
                width: double.infinity,
                margin: const EdgeInsets.symmetric(horizontal: 16),
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: themeColor.withAlpha(25),
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(
                    color: themeColor.withAlpha(51),
                    width: 1,
                  ),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Icon(Icons.receipt_long,
                            size: 16, color: themeColor),
                        const SizedBox(width: 6),
                        Text(
                          "订单号：${widget.orderSn}",
                          style: TextStyle(
                            fontSize: 13,
                            color: textPrimary,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                      ],
                    ),
                    if (widget.orderId != null && widget.orderId! > 0) ...[
                      const SizedBox(height: 4),
                      Text(
                        "订单ID：${widget.orderId}",
                        style: TextStyle(
                          fontSize: 12,
                          color: textSecondary,
                        ),
                      ),
                    ],
                  ],
                ),
              ),

            const SizedBox(height: 30),

            // 支付金额展示
            Column(
              children: [
                Text(
                  "支付金额",
                  style: TextStyle(fontSize: 12, color: textSecondary),
                ),
                const SizedBox(height: 5),
                Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Text(
                      "￥",
                      style: TextStyle(fontSize: 20, color: textPrimary),
                    ),
                    Text(
                      widget.amount.toString(),
                      style: TextStyle(fontSize: 25, color: textPrimary),
                    ),
                  ],
                ),
              ],
            ),

            const SizedBox(height: 30),

            // 支付方式选择
            Container(
              height: 60,
              margin: const EdgeInsets.only(left: 30),
              padding: const EdgeInsets.only(right: 18),
              decoration: BoxDecoration(border: Border(bottom: border)),
              child: Row(
                children: [
                  Image.asset(
                    "images/wx_pay.png",
                    height: 27,
                    width: 26,
                  ),
                  const SizedBox(width: 16),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Text(
                          "微信支付",
                          style: TextStyle(fontSize: 16, color: textPrimary),
                        ),
                        const SizedBox(height: 2),
                        Text(
                          "推荐使用微信支付",
                          style:
                              TextStyle(fontSize: 12, color: textSecondary),
                        ),
                      ],
                    ),
                  ),
                  Transform.scale(
                    scale: 1.3,
                    child: Checkbox(
                      value: _selectedPayType == 2,
                      activeColor: themeColor,
                      shape: const CircleBorder(),
                      side: const BorderSide(width: 1, color: Colors.black26),
                      onChanged: (value) {
                        if (value == true) {
                          setState(() => _selectedPayType = 2);
                        }
                      },
                    ),
                  ),
                ],
              ),
            ),

            Container(
              height: 60,
              margin: const EdgeInsets.only(left: 30, bottom: 30),
              padding: const EdgeInsets.only(right: 18),
              child: Row(
                children: [
                  Image.asset(
                    "images/ali_pay.png",
                    height: 27,
                    width: 26,
                  ),
                  const SizedBox(width: 16),
                  Expanded(
                    child: Text(
                      "支付宝支付",
                      style: TextStyle(fontSize: 16, color: textPrimary),
                    ),
                  ),
                  Transform.scale(
                    scale: 1.3,
                    child: Checkbox(
                      value: _selectedPayType == 1,
                      activeColor: themeColor,
                      shape: const CircleBorder(),
                      side: const BorderSide(width: 1, color: Colors.black26),
                      onChanged: (value) {
                        if (value == true) {
                          setState(() => _selectedPayType = 1);
                        }
                      },
                    ),
                  ),
                ],
              ),
            ),

            // 确认支付按钮
            InkWell(
              onTap: () {
                Navigator.of(context).push(
                  MaterialPageRoute(
                    builder: (context) => const PayTest(),
                  ),
                );
              },
              child: Container(
                alignment: Alignment.center,
                width: MediaQuery.of(context).size.width,
                height: 40,
                margin: const EdgeInsets.only(top: 15, left: 30, right: 30),
                decoration: BoxDecoration(
                  color: themeColor,
                  borderRadius: BorderRadius.circular(5),
                ),
                child: const Text(
                  '确认支付',
                  style: TextStyle(color: Colors.white, fontSize: 16),
                ),
              ),
            ),

            // Story 5.4 Task 8.4: Saga 失败后提示（由上层页面处理，此处展示兜底）
            if (widget.orderId == null || widget.orderId == 0)
              Padding(
                padding: const EdgeInsets.only(top: 16),
                child: Text(
                  "如支付遇到问题，请在订单详情页重新发起支付",
                  style: TextStyle(fontSize: 12, color: textSecondary),
                ),
              ),
          ],
        ),
      ),
    );
  }
}
