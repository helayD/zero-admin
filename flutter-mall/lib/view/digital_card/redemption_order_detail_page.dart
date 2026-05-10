import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/theme/app_theme.dart';
import 'package:flutter_mall/utils/http_util.dart';

/// 提货单详情页（Story 10.7 Review Fix MEDIUM-3）
///
/// - 展示提货单号、当前状态、收货信息、OMS 订单号
/// - 展示提货状态时间线（pending → processing → shipped → delivered）
/// - 未来可扩展：取消提货单（pending 状态）、查看物流
///
/// 注意：C 端展示严禁暴露链上 token/tokenId/链上交易号，页面不拉取链上字段。
class RedemptionOrderDetailPage extends StatefulWidget {
  final int orderId;
  final int? cardInstanceId;

  const RedemptionOrderDetailPage({
    super.key,
    required this.orderId,
    this.cardInstanceId,
  });

  @override
  State<RedemptionOrderDetailPage> createState() =>
      _RedemptionOrderDetailPageState();
}

class _RedemptionOrderDetailPageState extends State<RedemptionOrderDetailPage> {
  bool _isLoading = true;
  String? _errorMessage;
  Map<String, dynamic>? _order;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });
    try {
      final Response response = await HttpUtil.get(
        queryRedemptionOrderUrl,
        queryParameters: <String, dynamic>{
          'orderId': widget.orderId,
          if (widget.cardInstanceId != null && widget.cardInstanceId! > 0)
            'cardInstanceId': widget.cardInstanceId,
        },
      );
      final Map<String, dynamic> data = _responseData(response.data);
      if (!mounted) return;
      setState(() {
        _order = data;
        _isLoading = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _errorMessage = _parseError(e);
        _isLoading = false;
      });
    }
  }

  Map<String, dynamic> _responseData(dynamic raw) {
    if (raw is Map && raw['data'] is Map) {
      return Map<String, dynamic>.from(raw['data'] as Map);
    }
    return <String, dynamic>{};
  }

  String _parseError(dynamic error) {
    if (error is DioException) {
      final dynamic data = error.response?.data;
      if (data is Map && data['message'] != null) {
        return data['message'].toString();
      }
      if (error.response?.statusCode == 404) {
        return '提货单不存在';
      }
    }
    return '加载提货单失败，请稍后重试';
  }

  String _statusLabel(String status) {
    switch (status) {
      case 'pending':
        return '待处理';
      case 'processing':
        return '处理中';
      case 'shipped':
        return '已发货';
      case 'delivered':
        return '已签收';
      case 'cancelled':
        return '已取消';
      case 'failed':
        return '处理失败';
      default:
        return status;
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        title: const Text('提货单详情'),
      ),
      body: _buildBody(),
    );
  }

  Widget _buildBody() {
    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }
    if (_errorMessage != null || _order == null) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(AppSpacing.xl),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              const Icon(Icons.error_outline,
                  size: 52, color: Color(0xFFB42318)),
              const SizedBox(height: AppSpacing.md),
              Text(
                _errorMessage ?? '无数据',
                textAlign: TextAlign.center,
                style: Theme.of(context).textTheme.bodyMedium,
              ),
              const SizedBox(height: AppSpacing.lg),
              FilledButton(
                onPressed: _load,
                child: const Text('重新加载'),
              ),
            ],
          ),
        ),
      );
    }

    final Map<String, dynamic> order = _order!;
    final String status = order['status']?.toString() ?? '';

    return RefreshIndicator(
      onRefresh: _load,
      child: ListView(
        padding: const EdgeInsets.all(AppSpacing.lg),
        children: <Widget>[
          _buildHeroCard(order, status),
          const SizedBox(height: AppSpacing.lg),
          _buildReceiverCard(order),
          const SizedBox(height: AppSpacing.lg),
          _buildMetaCard(order),
        ],
      ),
    );
  }

  Widget _buildHeroCard(Map<String, dynamic> order, String status) {
    return Container(
      padding: const EdgeInsets.all(AppSpacing.xl),
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          colors: <Color>[Color(0xFF1F2937), Color(0xFF2563EB)],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.circular(AppRadii.xxl),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(
            '提货单号 ${order['orderNo'] ?? '-'}',
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: Colors.white.withValues(alpha: 0.85),
                ),
          ),
          const SizedBox(height: AppSpacing.sm),
          Text(
            _statusLabel(status),
            style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                  color: Colors.white,
                  fontSize: 26,
                ),
          ),
          const SizedBox(height: AppSpacing.sm),
          Text(
            _statusDescription(status),
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: Colors.white.withValues(alpha: 0.8),
                ),
          ),
        ],
      ),
    );
  }

  String _statusDescription(String status) {
    switch (status) {
      case 'pending':
        return '提货请求已提交，正在为您安排发货。';
      case 'processing':
        return '我们正在为您准备商品，请稍候。';
      case 'shipped':
        return '商品已发出，请关注物流更新。';
      case 'delivered':
        return '提货已完成，感谢使用。';
      case 'cancelled':
        return '该提货单已取消。';
      case 'failed':
        return '提货处理失败，请联系客服。';
      default:
        return '';
    }
  }

  Widget _buildReceiverCard(Map<String, dynamic> order) {
    return Container(
      padding: const EdgeInsets.all(AppSpacing.lg),
      decoration: BoxDecoration(
        color: AppColors.surface,
        borderRadius: BorderRadius.circular(AppRadii.xl),
        border: Border.all(color: AppColors.border),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text('收货信息', style: Theme.of(context).textTheme.titleLarge),
          const SizedBox(height: AppSpacing.md),
          _buildLine('收货人', order['receiverName']?.toString() ?? '-'),
          _buildLine('联系电话', order['receiverPhone']?.toString() ?? '-'),
          _buildLine('收货地址', order['receiverAddress']?.toString() ?? '-'),
        ],
      ),
    );
  }

  Widget _buildMetaCard(Map<String, dynamic> order) {
    return Container(
      padding: const EdgeInsets.all(AppSpacing.lg),
      decoration: BoxDecoration(
        color: AppColors.surface,
        borderRadius: BorderRadius.circular(AppRadii.xl),
        border: Border.all(color: AppColors.border),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text('处理记录', style: Theme.of(context).textTheme.titleLarge),
          const SizedBox(height: AppSpacing.md),
          _buildLine('创建时间', order['createTime']?.toString() ?? '-'),
          _buildLine('更新时间', order['updateTime']?.toString() ?? '-'),
          if ((order['shippedAt']?.toString() ?? '').isNotEmpty)
            _buildLine('发货时间', order['shippedAt'].toString()),
          if ((order['deliveredAt']?.toString() ?? '').isNotEmpty)
            _buildLine('签收时间', order['deliveredAt'].toString()),
          if ((order['cancelReason']?.toString() ?? '').isNotEmpty)
            _buildLine('取消原因', order['cancelReason'].toString()),
        ],
      ),
    );
  }

  Widget _buildLine(String label, String value) {
    return Padding(
      padding: const EdgeInsets.only(bottom: AppSpacing.sm),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          SizedBox(
            width: 84,
            child: Text(
              label,
              style: Theme.of(context).textTheme.bodySmall,
            ),
          ),
          Expanded(
            child: Text(
              value.isEmpty ? '-' : value,
              style: Theme.of(context).textTheme.bodyMedium,
            ),
          ),
        ],
      ),
    );
  }
}
