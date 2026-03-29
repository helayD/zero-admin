import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/model/order_detail.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/widgets/empty_state_widget.dart';
import 'package:flutter_mall/widgets/order_timeline_panel.dart';

///
/// 物流轨迹页面（Story 6-3 Task 6）
///
/// 页面结构：
/// - 顶部物流状态摘要卡片（物流公司 + 单号 + 当前状态）
/// - 物流轨迹时间线（复用 OrderTimelinePanel）
///
/// Commerce State Shell：
/// - 骨架屏（加载中）
/// - 错误态（ErrorRetryWidget + 重试按钮）
/// - 空态（EmptyStateWidget + "暂无物流信息，包裹正在准备中"）
/// - 成功态（物流摘要卡片 + 轨迹时间线）
///
/// Intent Recovery 预留（Task 10）：接收 intentSource 参数，当前 Story 不实现恢复逻辑。
///
/// 作者：Story 6-3 Dev
/// 日期：2026-03-29
///
class OrderLogistics extends StatefulWidget {
  /// 订单ID（必填）
  final int orderId;

  /// Intent 来源标识（Story 9-3 Intent Recovery 预留参数）
  final String? intentSource;

  const OrderLogistics({
    super.key,
    required this.orderId,
    this.intentSource,
  });

  @override
  State<OrderLogistics> createState() => _OrderLogisticsState();
}

// ==================== 页面状态枚举 ====================
enum LogisticsPageState {
  initial,
  loading,
  success,
  empty,
  error,
}

// ==================== 物流响应包装 ====================
// LogisticsResp 物流查询响应（复用 model/order_detail.dart 中的 LogisticsData/LogisticsNodeData）
class LogisticsResp {
  final int code;
  final String message;
  final LogisticsData? data;

  LogisticsResp({
    required this.code,
    required this.message,
    this.data,
  });

  factory LogisticsResp.fromJson(Map<String, dynamic> json) {
    return LogisticsResp(
      code: json['code'] ?? 0,
      message: json['message'] ?? '',
      data: json['data'] != null ? LogisticsData.fromJson(json['data']) : null,
    );
  }
}


// ==================== 页面状态 ====================
class _OrderLogisticsState extends State<OrderLogistics> {
  LogisticsPageState _pageState = LogisticsPageState.initial;
  LogisticsData? _logisticsData;
  String _errorMessage = '';

  @override
  void initState() {
    super.initState();
    _loadLogistics();
  }

  Future<void> _loadLogistics() async {
    setState(() {
      _pageState = LogisticsPageState.loading;
    });

    try {
      final resp = await HttpUtil.get(
        '$logisticsDataUrl${widget.orderId}',
      );
      final result = LogisticsResp.fromJson(resp.data);

      if (!mounted) return;

      // 区分业务错误（code != 0）和无数据（code == 0 但无物流信息）
      if (result.code != 0) {
        // 外部依赖异常或归属校验失败 → 错误态 + 展示后端错误信息
        setState(() {
          _errorMessage = result.message.isNotEmpty ? result.message : '加载失败，请稍后重试';
          _pageState = LogisticsPageState.error;
        });
      } else if (result.data != null &&
                 result.data!.deliveryNo.isEmpty &&
                 result.data!.logisticsNodes.isEmpty) {
        // OMS 无物流数据（未发货/物流未填写）→ 空态
        setState(() {
          _logisticsData = null;
          _pageState = LogisticsPageState.empty;
        });
      } else {
        // 有物流数据 → 成功态
        setState(() {
          _logisticsData = result.data;
          _pageState = LogisticsPageState.success;
        });
      }
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _errorMessage = '加载失败，请稍后重试';
        _pageState = LogisticsPageState.error;
      });
    }
  }

  void _copyTrackingNumber() {
    if (_logisticsData == null || _logisticsData!.deliveryNo.isEmpty) return;
    Clipboard.setData(ClipboardData(text: _logisticsData!.deliveryNo));
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('单号已复制'),
        duration: Duration(seconds: 1),
        behavior: SnackBarBehavior.floating,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF5F5F5),
      appBar: AppBar(
        title: const Text('物流详情'),
        backgroundColor: Colors.white,
        foregroundColor: const Color(0xFF303133),
        elevation: 0,
        centerTitle: true,
      ),
      body: _buildBody(),
    );
  }

  Widget _buildBody() {
    switch (_pageState) {
      case LogisticsPageState.loading:
        return _buildSkeleton();
      case LogisticsPageState.success:
        return _buildSuccessContent();
      case LogisticsPageState.empty:
        return EmptyStateWidget(
          message: '暂无物流信息，包裹正在准备中',
          icon: Icons.local_shipping_outlined,
          actionText: '返回订单',
          onAction: () => Navigator.pop(context),
        );
      case LogisticsPageState.error:
        return ErrorRetryWidget(
          message: _errorMessage,
          onRetry: _loadLogistics,
        );
      case LogisticsPageState.initial:
        return const SizedBox.shrink();
    }
  }

  // ==================== 骨架屏 ====================
  Widget _buildSkeleton() {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(12),
      child: Column(
        children: [
          // 物流摘要卡片骨架
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(8),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Container(
                  width: 120,
                  height: 16,
                  decoration: BoxDecoration(
                    color: Colors.grey[200],
                    borderRadius: BorderRadius.circular(4),
                  ),
                ),
                const SizedBox(height: 12),
                Container(
                  width: double.infinity,
                  height: 12,
                  decoration: BoxDecoration(
                    color: Colors.grey[200],
                    borderRadius: BorderRadius.circular(4),
                  ),
                ),
                const SizedBox(height: 8),
                Container(
                  width: 180,
                  height: 12,
                  decoration: BoxDecoration(
                    color: Colors.grey[200],
                    borderRadius: BorderRadius.circular(4),
                  ),
                ),
                const SizedBox(height: 16),
                Container(
                  width: 200,
                  height: 20,
                  decoration: BoxDecoration(
                    color: Colors.grey[200],
                    borderRadius: BorderRadius.circular(4),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 12),
          // 时间线骨架
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(8),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: List.generate(
                3,
                (index) => Padding(
                  padding: EdgeInsets.only(bottom: index < 2 ? 16 : 0),
                  child: Row(
                    children: [
                      Container(
                        width: 18,
                        height: 18,
                        decoration: BoxDecoration(
                          color: Colors.grey[200],
                          shape: BoxShape.circle,
                        ),
                      ),
                      const SizedBox(width: 10),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Container(
                              width: 120,
                              height: 14,
                              decoration: BoxDecoration(
                                color: Colors.grey[200],
                                borderRadius: BorderRadius.circular(4),
                              ),
                            ),
                            const SizedBox(height: 4),
                            Container(
                              width: 80,
                              height: 10,
                              decoration: BoxDecoration(
                                color: Colors.grey[200],
                                borderRadius: BorderRadius.circular(4),
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  // ==================== 成功内容 ====================
  Widget _buildSuccessContent() {
    final data = _logisticsData!;

    return SingleChildScrollView(
      padding: const EdgeInsets.all(12),
      child: Column(
        children: [
          // 物流摘要卡片
          _buildSummaryCard(data),
          const SizedBox(height: 12),
          // 物流轨迹时间线
          _buildTimelineCard(data),
        ],
      ),
    );
  }

  // 物流摘要卡片
  Widget _buildSummaryCard(LogisticsData data) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // 物流公司
          Row(
            children: [
              const Icon(
                Icons.local_shipping,
                size: 20,
                color: Color(0xFF606266),
              ),
              const SizedBox(width: 8),
              Text(
                data.deliveryCompany.isNotEmpty ? data.deliveryCompany : '物流公司',
                style: const TextStyle(
                  fontSize: 15,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF303133),
                ),
              ),
            ],
          ),
          const SizedBox(height: 10),
          // 物流单号 + 复制按钮
          Row(
            children: [
              Text(
                '运单号：${data.deliveryNo}',
                style: const TextStyle(
                  fontSize: 13,
                  color: Color(0xFF909399),
                ),
              ),
              const SizedBox(width: 8),
              GestureDetector(
                onTap: _copyTrackingNumber,
                child: Container(
                  padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                  decoration: BoxDecoration(
                    border: Border.all(color: const Color(0xFFDCDFE6)),
                    borderRadius: BorderRadius.circular(4),
                  ),
                  child: const Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(
                        Icons.copy,
                        size: 12,
                        color: Color(0xFF909399),
                      ),
                      SizedBox(width: 2),
                      Text(
                        '复制',
                        style: TextStyle(
                          fontSize: 11,
                          color: Color(0xFF909399),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 14),
          // 当前状态（高亮）
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            decoration: BoxDecoration(
              color: const Color(0xFFFDF6EC),
              borderRadius: BorderRadius.circular(6),
              border: Border.all(color: const Color(0xFFE6A23C)),
            ),
            child: Row(
              children: [
                const Icon(
                  Icons.info_outline,
                  size: 16,
                  color: Color(0xFFE6A23C),
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    data.currentStatus,
                    style: const TextStyle(
                      fontSize: 13,
                      color: Color(0xFFB8820A),
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  // 物流轨迹时间线
  Widget _buildTimelineCard(LogisticsData data) {
    final timelineNodes = data.logisticsNodes
        .map((n) => n.toTimelineNode())
        .toList();

    if (timelineNodes.isEmpty) {
      return Container(
        padding: const EdgeInsets.all(24),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(8),
        ),
        child: const Center(
          child: Text(
            '暂无轨迹信息',
            style: TextStyle(
              fontSize: 13,
              color: Color(0xFF909399),
            ),
          ),
        ),
      );
    }

    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
      ),
      child: OrderTimelinePanel(
        timeline: timelineNodes,
        orderStatus: 2, // 已发货
      ),
    );
  }
}
