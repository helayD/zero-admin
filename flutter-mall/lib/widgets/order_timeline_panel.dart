import 'package:flutter/material.dart';
import 'package:flutter_mall/model/order_detail.dart';

///
/// 订单时间线面板组件
///
/// 垂直时间线：左侧图标 + 竖线，右侧内容
/// Story 6-1 Task 8.1 首次建设
///
class OrderTimelinePanel extends StatelessWidget {
  final List<TimelineNode> timeline;
  final int orderStatus;

  const OrderTimelinePanel({
    super.key,
    required this.timeline,
    required this.orderStatus,
  });

  @override
  Widget build(BuildContext context) {
    if (timeline.isEmpty) {
      return _buildEmptyTimeline();
    }
    return Container(
      color: Colors.white,
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Padding(
            padding: EdgeInsets.only(bottom: 12),
            child: Text(
              "订单进度",
              style: TextStyle(
                fontSize: 15,
                fontWeight: FontWeight.bold,
                color: Color(0xFF303133),
              ),
            ),
          ),
          ...timeline.asMap().entries.map((entry) {
            final index = entry.key;
            final node = entry.value;
            final isLast = index == timeline.length - 1;
            return _TimelineNodeWidget(
              node: node,
              isLast: isLast,
            );
          }),
        ],
      ),
    );
  }

  Widget _buildEmptyTimeline() {
    return Container(
      color: Colors.white,
      padding: const EdgeInsets.all(16),
      child: Center(
        child: Text(
          "暂无时间线信息",
          style: TextStyle(
            fontSize: 13,
            color: Colors.grey[500],
          ),
        ),
      ),
    );
  }
}

class _TimelineNodeWidget extends StatelessWidget {
  final TimelineNode node;
  final bool isLast;

  const _TimelineNodeWidget({
    required this.node,
    required this.isLast,
  });

  Color get _iconColor {
    switch (node.status) {
      case "completed":
        return const Color(0xFF4CAF50); // 绿色 - 已完成
      case "current":
        return const Color(0xFF2196F3); // 蓝色 - 进行中
      case "interrupted":
        return const Color(0xFFF44336); // 红色 - 异常/中断
      case "pending":
      default:
        return const Color(0xFFBDBDBD); // 灰色 - 未到达
    }
  }

  IconData get _icon {
    switch (node.status) {
      case "completed":
        return Icons.check_circle;
      case "current":
        return Icons.radio_button_checked;
      case "interrupted":
        return Icons.cancel;
      case "pending":
      default:
        return Icons.radio_button_unchecked;
    }
  }

  Color get _lineColor {
    if (node.status == "completed" && !isLast) {
      return const Color(0xFF4CAF50);
    }
    return const Color(0xFFBDBDBD);
  }

  @override
  Widget build(BuildContext context) {
    final isPending = node.status == "pending";
    final isInterrupted = node.status == "interrupted";

    return IntrinsicHeight(
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // 左侧图标 + 竖线
          SizedBox(
            width: 24,
            child: Column(
              children: [
                Icon(
                  _icon,
                  size: 18,
                  color: _iconColor,
                ),
                if (!isLast)
                  Expanded(
                    child: Container(
                      width: isPending ? 1 : 2,
                      color: _lineColor,
                    ),
                  ),
              ],
            ),
          ),
          const SizedBox(width: 10),
          // 右侧内容
          Expanded(
            child: Padding(
              padding: EdgeInsets.only(bottom: isLast ? 0 : 16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    node.title,
                    style: TextStyle(
                      fontSize: 14,
                      fontWeight: node.status == "current" ? FontWeight.bold : FontWeight.normal,
                      color: isInterrupted
                          ? const Color(0xFFF44336)
                          : const Color(0xFF303133),
                    ),
                  ),
                  if (node.time.isNotEmpty) ...[
                    const SizedBox(height: 4),
                    Text(
                      node.time,
                      style: TextStyle(
                        fontSize: 12,
                        color: Colors.grey[600],
                      ),
                    ),
                  ],
                  if (node.detail.isNotEmpty) ...[
                    const SizedBox(height: 4),
                    Text(
                      node.detail,
                      style: TextStyle(
                        fontSize: 12,
                        color: isInterrupted
                            ? const Color(0xFFF44336)
                            : const Color(0xFF909399),
                      ),
                    ),
                  ],
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}
