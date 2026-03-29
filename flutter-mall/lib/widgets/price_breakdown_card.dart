import 'package:flutter/material.dart';
import 'package:flutter_mall/model/order_detail.dart';

///
/// 订单金额拆分卡片组件
///
/// Story 6-1 Task 8.2 首次建设
///
class PriceBreakdownCard extends StatelessWidget {
  final PriceBreakdown priceBreakdown;

  const PriceBreakdownCard({
    super.key,
    required this.priceBreakdown,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      color: Colors.white,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // 标题
          Container(
            height: 42,
            padding: const EdgeInsets.symmetric(horizontal: 15),
            decoration: BoxDecoration(
              border: Border(
                bottom: BorderSide(
                  width: 1,
                  color: Colors.grey[200]!,
                ),
              ),
            ),
            alignment: Alignment.centerLeft,
            child: const Text(
              "金额明细",
              style: TextStyle(
                fontSize: 15,
                fontWeight: FontWeight.w600,
                color: Color(0xFF606266),
              ),
            ),
          ),
          // 金额项
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 15),
            child: Column(
              children: [
                _PriceRow(
                  label: "商品金额",
                  value: priceBreakdown.orderAmount,
                ),
                _PriceRow(
                  label: "运费",
                  value: priceBreakdown.freightAmount,
                ),
                if (priceBreakdown.promotionAmount > 0)
                  _PriceRow(
                    label: "促销优惠",
                    value: -priceBreakdown.promotionAmount,
                    valueColor: const Color(0xFFFA436A),
                  ),
                if (priceBreakdown.couponAmount > 0)
                  _PriceRow(
                    label: "优惠券",
                    value: -priceBreakdown.couponAmount,
                    valueColor: const Color(0xFFFA436A),
                  ),
                if (priceBreakdown.pointsAmount > 0)
                  _PriceRow(
                    label: "积分抵扣",
                    value: -priceBreakdown.pointsAmount,
                    valueColor: const Color(0xFFFA436A),
                  ),
                if (priceBreakdown.discountAmount != 0)
                  _PriceRow(
                    label: "管理员调整",
                    value: priceBreakdown.discountAmount,
                    valueColor: const Color(0xFFFA436A),
                  ),
                // 实付（高亮）
                Container(
                  padding: const EdgeInsets.symmetric(vertical: 10),
                  decoration: BoxDecoration(
                    border: Border(
                      top: BorderSide(
                        width: 1,
                        color: Colors.grey[200]!,
                      ),
                    ),
                  ),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      const Text(
                        "实付款",
                        style: TextStyle(
                          fontSize: 14,
                          fontWeight: FontWeight.w600,
                          color: Color(0xFF303133),
                        ),
                      ),
                      Text(
                        "￥${priceBreakdown.payAmount.toStringAsFixed(2)}",
                        style: const TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.bold,
                          color: Color(0xFFFA436A),
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 12),
        ],
      ),
    );
  }
}

class _PriceRow extends StatelessWidget {
  final String label;
  final double value;
  final Color? valueColor;

  const _PriceRow({
    required this.label,
    required this.value,
    this.valueColor,
  });

  @override
  Widget build(BuildContext context) {
    final isNegative = value < 0;
    final displayValue = isNegative
        ? "-￥${(-value).toStringAsFixed(2)}"
        : "￥${value.toStringAsFixed(2)}";

    return Container(
      height: 36,
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(
            label,
            style: const TextStyle(
              fontSize: 13,
              color: Color(0xFF909399),
            ),
          ),
          Text(
            displayValue,
            style: TextStyle(
              fontSize: 13,
              color: valueColor ?? const Color(0xFF303133),
            ),
          ),
        ],
      ),
    );
  }
}
