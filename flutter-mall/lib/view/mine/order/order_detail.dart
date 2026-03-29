import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/widgets/cached_image_widget.dart';
import 'package:flutter_mall/widgets/empty_state_widget.dart';
import 'package:flutter_mall/widgets/order_timeline_panel.dart';
import 'package:flutter_mall/widgets/price_breakdown_card.dart';

import '../../../config/order_status.dart';
import '../../../model/order_item.dart'; // OrderItemList canonical
import '../../../model/order_detail.dart';

///
/// 订单详情页面
///
/// Story 6-1 重构：接入真实 API、Order Timeline Panel、Price Breakdown Card、State Shell
///
/// 作者：刘飞华
/// 日期：2023/11/21 17:17
///
class OrderDetail extends StatefulWidget {
  final int orderId;
  final String? intentSource; // Story 6-1 Task 10.2: Intent Recovery 预留

  const OrderDetail({
    super.key,
    required this.orderId,
    this.intentSource,
  });

  @override
  State<OrderDetail> createState() => _OrderDetailState();
}

class _OrderDetailState extends State<OrderDetail> with SingleTickerProviderStateMixin {
  OrderDetailData? orderDetailData;
  bool _isLoading = true;
  bool _isOperating = false; // Story 6.2 Task 7: 操作防抖
  bool _localTimelineAppended = false; // Story 6.2 Review Fix: 防止重复追加时间线节点
  late AnimationController _skeletonController;
  late Animation<double> _skeletonAnimation;

  @override
  void initState() {
    super.initState();
    _skeletonController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1200),
    )..repeat(reverse: true);
    _skeletonAnimation = Tween<double>(begin: 0.3, end: 0.7).animate(
      CurvedAnimation(parent: _skeletonController, curve: Curves.easeInOut),
    );
    _queryOrderDetail();
  }

  @override
  void dispose() {
    _skeletonController.dispose();
    super.dispose();
  }

  // Story 6-1 Review Fix: LOW-3 — 未实现功能显示友好提示
  void _showComingSoon(String feature) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text('$feature 功能即将上线，敬请期待'),
        duration: const Duration(seconds: 2),
        behavior: SnackBarBehavior.floating,
        backgroundColor: Colors.grey[700],
      ),
    );
  }

  // Story 6.2 Task 8.1+8.2+8.4: 本地追加时间线节点（status==5 取消 / status==3 确认收货）
  // Review Fix: 防止 _queryOrderDetail 重入时重复追加（flag 在 _queryOrderDetail 成功后重置）
  void _appendLocalTimelineNodes(OrderDetailData d) {
    if (_localTimelineAppended) return;
    // 取消节点：status==5 且 timeline 中无"取消"字样节点
    if (d.orderStatus == 5 && !d.timeline.any((n) => n.title.contains('取消'))) {
      d.timeline.add(TimelineNode(
        status: "interrupted",
        title: "已取消",
        time: _formatTime(d.updateTime),
        detail: "用户主动取消",
      ));
    }
    // 确认收货节点：status==3 且 timeline 中无"确认收货"节点
    if (d.orderStatus == 3 && !d.timeline.any((n) => n.title.contains('确认收货'))) {
      d.timeline.add(TimelineNode(
        status: "completed",
        title: "确认收货",
        time: _formatTime(d.receiveTime),
        detail: "",
      ));
    }
    _localTimelineAppended = true;
  }

  // Story 6.2 Task 5.1: 取消订单
  Future<void> _cancelOrder() async {
    if (_isOperating) return;
    final confirmed = await _showCancelConfirmDialog();
    if (confirmed != true) return;

    setState(() => _isOperating = true);
    try {
      final Response resp = await HttpUtil.get(cancelOrderUrl + widget.orderId.toString());
      if (resp.data['code'] == 0) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('订单已取消'),
              behavior: SnackBarBehavior.floating,
              backgroundColor: Colors.green,
            ),
          );
        }
        await _queryOrderDetail();
      } else {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(resp.data['message'] ?? '取消失败'),
              behavior: SnackBarBehavior.floating,
              backgroundColor: Colors.red,
            ),
          );
        }
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('取消失败: $e'),
            behavior: SnackBarBehavior.floating,
            backgroundColor: Colors.red,
          ),
        );
      }
    } finally {
      if (mounted) setState(() => _isOperating = false);
    }
  }

  Future<bool?> _showCancelConfirmDialog() {
    return showDialog<bool>(
      context: context,
      barrierDismissible: false,
      builder: (ctx) => AlertDialog(
        title: const Text('取消订单'),
        content: const Text('确定要取消该订单吗？取消后库存将释放，优惠券和积分将返还。'),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(false),
            child: const Text('再想想'),
          ),
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(true),
            style: TextButton.styleFrom(foregroundColor: Colors.red),
            child: const Text('确认取消'),
          ),
        ],
      ),
    );
  }

  // Story 6.2 Task 6.1+6.2+6.3+6.4: 确认收货
  Future<void> _confirmReceive() async {
    if (_isOperating) return;
    final confirmed = await _showConfirmReceiveDialog();
    if (confirmed != true) return;

    setState(() => _isOperating = true);
    try {
      final Response resp = await HttpUtil.get(confirmReceiveUrl + widget.orderId.toString());
      if (resp.data['code'] == 0) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('确认收货成功'),
              behavior: SnackBarBehavior.floating,
              backgroundColor: Colors.green,
            ),
          );
        }
        await _queryOrderDetail();
      } else {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(resp.data['message'] ?? '确认收货失败'),
              behavior: SnackBarBehavior.floating,
              backgroundColor: Colors.red,
            ),
          );
        }
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('确认收货失败: $e'),
            behavior: SnackBarBehavior.floating,
            backgroundColor: Colors.red,
          ),
        );
      }
    } finally {
      if (mounted) setState(() => _isOperating = false);
    }
  }

  Future<bool?> _showConfirmReceiveDialog() {
    return showDialog<bool>(
      context: context,
      barrierDismissible: false,
      builder: (ctx) => AlertDialog(
        title: const Text('确认收货'),
        content: const Text('请确认您已收到商品且商品完好。确认后订单将完成。'),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(false),
            child: const Text('还没收到'),
          ),
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(true),
            style: TextButton.styleFrom(foregroundColor: const Color(0xFFFA436A)),
            child: const Text('确认收货'),
          ),
        ],
      ),
    );
  }

  // Story 6-1 Task 6.1: 真实 API 接入
  Future<void> _queryOrderDetail() async {
    setState(() {
      _isLoading = true;
    });

    try {
      final Response result = await HttpUtil.get(
        orderDetailDataUrl + widget.orderId.toString(),
      );
      final OrderDetailModel model = OrderDetailModel.fromJson(result.data);

      setState(() {
        orderDetailData = model.data;
        // Story 6.2 Task 8: 本地追加时间线节点（取消/确认收货）
        // Review Fix: flag 在每次成功拉取后重置，防止重复追加
        _localTimelineAppended = false;
        _appendLocalTimelineNodes(model.data);
        _isLoading = false;
      });
    } catch (e) {
      debugPrint('[OrderDetail] _queryOrderDetail error: $e');
      setState(() {
        _isLoading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        backgroundColor: Colors.white,
        title: const Text("订单详情"),
        titleTextStyle: const TextStyle(fontSize: 16, color: Colors.black),
        centerTitle: true,
        elevation: 0,
      ),
      body: _buildBody(),
    );
  }

  Widget _buildBody() {
    // 加载态：骨架屏（Story 6-1 Task 7.2）
    if (_isLoading) {
      return _buildSkeletonScreen();
    }

    // 错误态（Story 6-1 Task 7.2）
    if (orderDetailData == null) {
      return ErrorRetryWidget(
        message: "加载失败，请重试",
        onRetry: _queryOrderDetail,
      );
    }

    // 订单不存在（空态兜底）
    if (orderDetailData!.id == 0) {
      return EmptyStateWidget(
        message: "订单不存在",
        actionText: "返回",
        icon: Icons.receipt_long_outlined,
        onAction: () => Navigator.of(context).pop(),
      );
    }

    // 正常内容
    return Stack(
      children: [
        _buildContent(),
        Positioned(
          bottom: 0,
          left: 0,
          right: 0,
          child: _buildBottomActions(),
        ),
      ],
    );
  }

  // Story 6-1 Review Fix: 骨架屏微动画（LOW-1）
  Widget _buildSkeletonBox(double width, double height) {
    return AnimatedBuilder(
      animation: _skeletonAnimation,
      builder: (context, child) {
        return Container(
          width: width,
          height: height,
          decoration: BoxDecoration(
            color: Colors.grey[300]!.withOpacity(_skeletonAnimation.value),
            borderRadius: BorderRadius.circular(4),
          ),
        );
      },
    );
  }

  // Story 6-1 Task 7.2: 骨架屏
  Widget _buildSkeletonScreen() {
    return SingleChildScrollView(
      padding: const EdgeInsets.only(bottom: 60),
      child: Column(
        children: [
          // 状态骨架
          Container(
            height: 100,
            color: Colors.grey[200],
          ),
          const SizedBox(height: 5),
          // 地址骨架
          Container(
            height: 78,
            color: Colors.white,
            margin: const EdgeInsets.only(top: 5),
            padding: const EdgeInsets.all(15),
            child: Row(
              children: [
                _buildSkeletonBox(24, 24),
                const SizedBox(width: 10),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      _buildSkeletonBox(120, 14),
                      const SizedBox(height: 6),
                      _buildSkeletonBox(200, 12),
                    ],
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 5),
          // 商品骨架
          Container(
            height: 100,
            color: Colors.white,
            margin: const EdgeInsets.only(top: 5),
            padding: const EdgeInsets.all(15),
            child: Row(
              children: [
                _buildSkeletonBox(70, 70),
                const SizedBox(width: 10),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      _buildSkeletonBox(double.infinity, 14),
                      const SizedBox(height: 6),
                      _buildSkeletonBox(100, 12),
                      const SizedBox(height: 6),
                      _buildSkeletonBox(80, 14),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  // Story 6-1 Task 6: 主内容区
  Widget _buildContent() {
    final d = orderDetailData!;
    return SingleChildScrollView(
      padding: const EdgeInsets.only(bottom: 80),
      child: Column(
        children: [
          // 1. 订单状态条（Story 6-1 Task 6 - 订单状态）
          _buildOrderStatusBar(d),
          // 2. Order Timeline Panel（Story 6-1 Task 6.2）
          if (d.timeline.isNotEmpty)
            OrderTimelinePanel(timeline: d.timeline, orderStatus: d.orderStatus),
          const SizedBox(height: 5),
          // 3. 收货信息卡片（Story 6-1 Task 6.5）
          _buildAddressCard(d),
          const SizedBox(height: 5),
          // 4. 商品明细（Story 6-1 Task 6.4）
          _buildProductSection(d),
          const SizedBox(height: 5),
          // 5. 金额拆分卡片（Story 6-1 Task 6.3）
          PriceBreakdownCard(priceBreakdown: d.priceBreakdown),
          const SizedBox(height: 5),
          // 6. 订单基本信息（Story 6-1 Task 6.6）
          _buildOrderInfo(d),
        ],
      ),
    );
  }

  // 订单状态条（Story 6-1 Task 6）
  Widget _buildOrderStatusBar(OrderDetailData d) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.symmetric(horizontal: 15, vertical: 20),
      decoration: const BoxDecoration(
        color: Color(0xFFFA436A),
      ),
      child: Row(
        children: [
          Image.asset(
            _getStatusIcon(d.orderStatus),
            height: 24,
            width: 24,
            color: Colors.white,
          ),
          const SizedBox(width: 12),
          Text(
            getOmsOrderStatusTxt(d.orderStatus),
            style: const TextStyle(
              fontSize: 16,
              color: Colors.white,
              fontWeight: FontWeight.bold,
            ),
          ),
        ],
      ),
    );
  }

  // OMS 真实值图标映射（Review Fix: OMS status code alignment）
  // OMS: 0=等待付款, 1=已支付(待发货), 2=已发货, 3=已发货(?), 4=已完成, 5=已取消, 7=售后中
  String _getStatusIcon(int status) {
    switch (status) {
      case 0:
        return "images/daifukuan.png"; // 等待付款
      case 1:
      case 2:
      case 3:
        return "images/daifahuo.png"; // 已支付/已发货 → 都是等收货
      case 4:
        return "images/tick.png"; // 交易完成
      case 5:
        return "images/delete.png"; // 已取消
      case 7:
        return "images/tuihuo.png"; // 售后中
      default:
        return "images/daifukuan.png";
    }
  }

  // 收货信息卡片（Story 6-1 Task 6.5）
  Widget _buildAddressCard(OrderDetailData d) {
    final addr = d.memberReceiveAddress;
    return Container(
      color: Colors.white,
      child: Column(
        children: [
          // 地址图标 + 信息
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 15, vertical: 14),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Icon(
                  Icons.location_on_outlined,
                  size: 22,
                  color: Color(0xFF606266),
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        "${addr.receiverName}  ${addr.maskedPhone}",
                        style: const TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.w600,
                          color: Color(0xFF303133),
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        addr.fullAddress,
                        style: const TextStyle(
                          fontSize: 13,
                          color: Color(0xFF909399),
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
          // 分割线
          Container(
            height: 5,
            color: const Color(0xFFF5F5F5),
          ),
        ],
      ),
    );
  }

  // 商品明细（Story 6-1 Task 6.4）
  Widget _buildProductSection(OrderDetailData d) {
    return Container(
      color: Colors.white,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // 标题
          Container(
            height: 42,
            padding: const EdgeInsets.symmetric(horizontal: 15),
            alignment: Alignment.centerLeft,
            decoration: BoxDecoration(
              border: Border(
                bottom: BorderSide(width: 1, color: Colors.grey[200]!),
              ),
            ),
            child: const Text(
              "商品信息",
              style: TextStyle(
                fontSize: 15,
                fontWeight: FontWeight.w600,
                color: Color(0xFF606266),
              ),
            ),
          ),
          // 商品列表
          ...d.orderItemData.map((item) => _buildProductItem(item)),
        ],
      ),
    );
  }

  Widget _buildProductItem(OrderItemList item) {
    return Container(
      padding: const EdgeInsets.all(15),
      decoration: BoxDecoration(
        border: Border(
          bottom: BorderSide(width: 1, color: Colors.grey[200]!),
        ),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // 商品图片
          CachedImageWidget(
            70,
            70,
            item.skuPic,
            fit: BoxFit.cover,
          ),
          const SizedBox(width: 10),
          // 商品信息
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  item.skuName,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    fontSize: 14,
                    color: Color(0xFF303133),
                  ),
                ),
                const SizedBox(height: 6),
                Text(
                  item.specData,
                  maxLines: 1,
                  style: const TextStyle(
                    fontSize: 12,
                    color: Color(0xFF909399),
                  ),
                ),
                const SizedBox(height: 6),
                Row(
                  children: [
                    Text(
                      "￥${item.skuPrice.toStringAsFixed(2)}",
                      style: const TextStyle(
                        fontSize: 14,
                        color: Color(0xFF303133),
                      ),
                    ),
                    const SizedBox(width: 8),
                    Text(
                      "x${item.skuQuantity}",
                      style: const TextStyle(
                        fontSize: 12,
                        color: Color(0xFF909399),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  // 订单基本信息（Story 6-1 Task 6.6）
  Widget _buildOrderInfo(OrderDetailData d) {
    return Container(
      color: Colors.white,
      padding: const EdgeInsets.symmetric(horizontal: 15),
      child: Column(
        children: [
          _buildInfoRow("订单编号", d.orderNo),
          _buildInfoRow("下单时间", _formatTime(d.createTime)),
          _buildInfoRow("支付方式", _getPayTypeTxt(d.payType)),
          if (d.payTime.isNotEmpty)
            _buildInfoRow("付款时间", _formatTime(d.payTime)),
          if (d.deliveryTime.isNotEmpty)
            _buildInfoRow("发货时间", _formatTime(d.deliveryTime)),
          if (d.receiveTime.isNotEmpty)
            _buildInfoRow("收货时间", _formatTime(d.receiveTime)),
          const SizedBox(height: 12),
        ],
      ),
    );
  }

  Widget _buildInfoRow(String label, String value) {
    return Container(
      height: 40,
      decoration: BoxDecoration(
        border: Border(
          bottom: BorderSide(width: 1, color: Colors.grey[200]!),
        ),
      ),
      child: Row(
        children: [
          Text(
            label,
            style: const TextStyle(
              fontSize: 13,
              color: Color(0xFF909399),
            ),
          ),
          const Spacer(),
          Text(
            value,
            style: const TextStyle(
              fontSize: 13,
              color: Color(0xFF303133),
            ),
          ),
        ],
      ),
    );
  }

  // 底部操作按钮（Story 6-1 Task 6 - MVP 不实现操作，仅展示状态对应按钮占位）
  Widget _buildBottomActions() {
    if (orderDetailData == null) return const SizedBox.shrink();

    final status = orderDetailData!.orderStatus;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 15, vertical: 8),
      decoration: BoxDecoration(
        color: Colors.white,
        boxShadow: [
          BoxShadow(
            blurRadius: 4,
            spreadRadius: 1,
            color: Colors.grey[300]!,
          ),
        ],
      ),
      child: SafeArea(
        child: Row(
          mainAxisAlignment: MainAxisAlignment.end,
          children: [
            // Story 6-2 实现：取消订单（OMS 0=待支付）
            if (status == 0)
              _ActionButton(label: "取消订单", isPrimary: false, onTap: _cancelOrder),
            if (status == 0) const SizedBox(width: 10),
            // Story 6-2 实现：立即付款（OMS 0=待支付）
            if (status == 0)
              _ActionButton(label: "立即付款", isPrimary: true, onTap: () => _showComingSoon("立即付款")),
            // Story 6-3 实现：查看物流（OMS 1=已支付/待发货, 2=已发货）
            if (status == 1 || status == 2)
              _ActionButton(label: "查看物流", isPrimary: false, onTap: () => _showComingSoon("查看物流")),
            if (status == 1 || status == 2) const SizedBox(width: 10),
            // Story 6-2 实现：确认收货（OMS 2=已发货）
            if (status == 2)
              _ActionButton(label: "确认收货", isPrimary: true, onTap: _confirmReceive),
            // Story 6-4 实现：申请售后（OMS 4=已完成, 7=售后中）
            if (status == 4 || status == 7)
              _ActionButton(label: "申请售后", isPrimary: false, onTap: () => _showComingSoon("申请售后")),
          ],
        ),
      ),
    );
  }

  String _getPayTypeTxt(int type) {
    switch (type) {
      case 1:
        return "支付宝支付";
      case 2:
        return "微信支付";
      default:
        return "未支付";
    }
  }

  String _formatTime(String isoTime) {
    if (isoTime.isEmpty) return "";
    try {
      // 格式: 2025-03-29T10:30:00+08:00 → 2025-03-29 10:30
      final parts = isoTime.split("T");
      if (parts.length >= 2) {
        final datePart = parts[0];
        final timePart = parts[1].substring(0, 5);
        return "$datePart $timePart";
      }
      return isoTime;
    } catch (_) {
      return isoTime;
    }
  }
}

class _ActionButton extends StatelessWidget {
  final String label;
  final bool isPrimary;
  final VoidCallback onTap;

  const _ActionButton({
    required this.label,
    required this.isPrimary,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    if (isPrimary) {
      return InkWell(
        onTap: onTap,
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
          decoration: BoxDecoration(
            border: Border.all(color: const Color(0xFFFA436A)),
            borderRadius: BorderRadius.circular(15),
          ),
          child: Text(
            label,
            style: const TextStyle(
              fontSize: 13,
              color: Color(0xFFFA436A),
            ),
          ),
        ),
      );
    } else {
      return InkWell(
        onTap: onTap,
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
          decoration: BoxDecoration(
            border: Border.all(color: Colors.grey),
            borderRadius: BorderRadius.circular(15),
          ),
          child: Text(
            label,
            style: const TextStyle(
              fontSize: 13,
              color: Color(0xFF303133),
            ),
          ),
        ),
      );
    }
  }
}
