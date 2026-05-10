import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/config/order_status.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/view/mine/order/order_detail.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/widgets/cached_image_widget.dart';
import 'package:flutter_mall/widgets/empty_state_widget.dart';

import '../../../model/order_list_model.dart';

///
/// 订单列表页面
///
/// Story 6-1 重构：接入真实 API、分页加载、刷新、State Shell
///
/// 作者：David
/// 日期：2023/11/21 17:17
///
class OrderList extends StatefulWidget {
  final int initialTab;
  final String? intentSource;

  const OrderList({super.key, this.initialTab = 0, this.intentSource});

  @override
  State<OrderList> createState() => _OrderListState();
}

class _OrderListState extends State<OrderList> {
  final List<String> _orderStatus = ['全部', '待支付', '待发货', '已完成', '已取消', '售后'];

  /// 每个 tab 的数据独立管理
  final Map<int, List<OrderListData>> _orderDataCache = {};
  final Map<int, int> _currentPageCache = {};
  final Map<int, bool> _hasMoreCache = {};
  final Map<int, bool> _loadingCache = {};

  late int _currentTab;

  AppRecentContext _buildOrderListRecoveryContext(int tab, [String? source]) {
    return AppRecentContext.create(
      targetType: AppRecentTargetType.orderList,
      tabIndex: tab,
      source: source ?? widget.intentSource ?? 'manual_open',
      requiresAuth: true,
      fallbackType: AppRecentTargetType.home,
      fallbackTabIndex: 0,
    );
  }

  @override
  void initState() {
    super.initState();
    _currentTab = widget.initialTab;
    AppRecoveryStore.saveActiveIntentCandidate(
      _buildOrderListRecoveryContext(_currentTab),
    );
    _initTab(_currentTab);
  }

  @override
  void dispose() {
    AppRecoveryStore.clearActiveIntentCandidateIfMatches(
      AppRecentTargetType.orderList,
      tabIndex: _currentTab,
    );
    super.dispose();
  }

  void _initTab(int tab) {
    if (_orderDataCache[tab] == null) {
      _queryOrderList(tab, refresh: true);
    }
  }

  /// 获取当前页码
  int _getCurrentPage() {
    return _currentPageCache[_currentTab] ?? 1;
  }

  /// 是否有更多数据
  bool _hasMore() {
    return _hasMoreCache[_currentTab] ?? true;
  }

  /// 是否正在加载
  bool _isLoading() {
    return _loadingCache[_currentTab] ?? false;
  }

  /// 查询订单列表（Story 6-1 Task 5.1/5.2/5.5/5.6）
  Future<void> _queryOrderList(int tab,
      {bool refresh = false, bool loadMore = false}) async {
    final page = refresh ? 1 : _getCurrentPage() + 1;

    // 缓存加载状态
    _loadingCache[tab] = true;

    try {
      final status = flutterTabToBackendStatus(tab);
      final url = "$orderListDataUrl$status&current=$page&pageSize=10";

      final Response result = await HttpUtil.get(url);
      final OrderListModel model = OrderListModel.fromJson(result.data);

      setState(() {
        if (refresh || loadMore) {
          // 刷新或加载更多
          final existing = _orderDataCache[tab] ?? [];
          if (loadMore) {
            _orderDataCache[tab] = [...existing, ...model.data];
          } else {
            _orderDataCache[tab] = model.data;
          }
        } else {
          _orderDataCache[tab] = model.data;
        }
        _currentPageCache[tab] = page;
        _hasMoreCache[tab] = model.hasMore;
        _loadingCache[tab] = false;
      });
      await AppRecoveryStore.saveRecentContext(
        _buildOrderListRecoveryContext(tab),
      );
    } catch (e) {
      setState(() {
        _loadingCache[tab] = false;
      });
      // 错误由 State Shell 处理
    }
  }

  void _onTabChanged(int index) {
    setState(() {
      _currentTab = index;
    });
    AppRecoveryStore.saveActiveIntentCandidate(
      _buildOrderListRecoveryContext(index),
    );
    _initTab(index);
  }

  void _onRefresh() {
    _queryOrderList(_currentTab, refresh: true);
  }

  void _onLoadMore() {
    if (!_hasMore() || _isLoading()) return;
    _queryOrderList(_currentTab, loadMore: true);
  }

  @override
  Widget build(BuildContext context) {
    return DefaultTabController(
      length: _orderStatus.length,
      initialIndex: _currentTab,
      child: Scaffold(
        appBar: AppBar(
          title: const Text(' 我的订单 '),
          titleTextStyle: const TextStyle(fontSize: 16, color: Colors.black),
          centerTitle: true,
          backgroundColor: Colors.white,
          elevation: 0,
          bottom: TabBar(
            isScrollable: true,
            indicatorColor: const Color(0xFFFA436A),
            labelColor: const Color(0xFFFA436A),
            unselectedLabelColor: const Color(0xFF606266),
            onTap: _onTabChanged,
            tabs: _orderStatus.map((status) => Tab(text: status)).toList(),
          ),
        ),
        body: TabBarView(
          children: _orderStatus.asMap().entries.map((entry) {
            final tab = entry.key;
            return _OrderListBody(
              tab: tab,
              data: _orderDataCache[tab] ?? [],
              isLoading: _loadingCache[tab] ?? false,
              hasMore: _hasMoreCache[tab] ?? true,
              onRefresh: _onRefresh,
              onLoadMore: _onLoadMore,
            );
          }).toList(),
        ),
      ),
    );
  }
}

/// 订单列表内容区（Story 6-1 Task 7.1: Commerce State Shell）
class _OrderListBody extends StatelessWidget {
  final int tab;
  final List<OrderListData> data;
  final bool isLoading;
  final bool hasMore;
  final VoidCallback onRefresh;
  final VoidCallback onLoadMore;

  const _OrderListBody({
    required this.tab,
    required this.data,
    required this.isLoading,
    required this.hasMore,
    required this.onRefresh,
    required this.onLoadMore,
  });

  @override
  Widget build(BuildContext context) {
    // 首次加载中
    if (isLoading && data.isEmpty) {
      return const Center(child: CircularProgressIndicator());
    }

    // 空态（Story 6-1 Task 7.1）
    if (!isLoading && data.isEmpty) {
      return EmptyStateWidget(
        message: "暂无相关订单",
        actionText: "去逛逛",
        icon: Icons.shopping_bag_outlined,
        onAction: () {
          Navigator.of(context).pop();
        },
      );
    }

    // 刷新/加载更多列表
    return RefreshIndicator(
      onRefresh: () async => onRefresh(),
      color: const Color(0xFFFA436A),
      child: NotificationListener<ScrollNotification>(
        onNotification: (notification) {
          if (notification is ScrollEndNotification) {
            final metrics = notification.metrics;
            if (metrics.pixels >= metrics.maxScrollExtent - 100) {
              onLoadMore();
            }
          }
          return false;
        },
        child: ListView.builder(
          padding: const EdgeInsets.only(bottom: 20),
          itemCount: data.length + (hasMore ? 1 : 0),
          itemBuilder: (context, index) {
            if (index >= data.length) {
              return const Padding(
                padding: EdgeInsets.all(16),
                child: Center(
                  child: Text(
                    "加载中...",
                    style: TextStyle(color: Colors.grey, fontSize: 13),
                  ),
                ),
              );
            }
            return _OrderListItem(data: data[index]);
          },
        ),
      ),
    );
  }
}

/// 订单列表项（Story 6-1 Task 5.3/5.4）
class _OrderListItem extends StatelessWidget {
  final OrderListData data;

  const _OrderListItem({required this.data});

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: () {
        Navigator.of(context).push(
          MaterialPageRoute(
            builder: (context) => OrderDetail(orderId: data.id),
          ),
        );
      },
      child: Container(
        margin: const EdgeInsets.only(top: 5),
        color: Colors.white,
        child: Column(
          children: [
            // 头部：订单号 + 状态（Story 6-1 Task 5.3）
            _buildHeader(),
            // 商品列表（最多展示1个）
            if (data.orderItemData.isNotEmpty) _buildProductPreview(),
            // 底部：实付款
            _buildAmountFooter(),
          ],
        ),
      ),
    );
  }

  Widget _buildHeader() {
    final statusText = getOmsOrderStatusTxt(data.status);
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 15, vertical: 10),
      decoration: BoxDecoration(
        border: Border(
          bottom: BorderSide(
            width: 1,
            color: Colors.grey[200]!,
          ),
        ),
      ),
      child: Row(
        children: [
          Expanded(
            child: Row(
              children: [
                Flexible(
                  child: Text(
                    data.orderNo.isNotEmpty ? data.orderSn : "订单号: ${data.id}",
                    style: const TextStyle(
                      fontSize: 13,
                      color: Color(0xFF303133),
                    ),
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
                // Story 10.7：含提货卡徽标
                if (data.hasDigitalCards) ...[
                  const SizedBox(width: 6),
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                    decoration: BoxDecoration(
                      color: const Color(0xFFFFF3E0),
                      borderRadius: BorderRadius.circular(3),
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: const [
                        Icon(Icons.card_giftcard,
                            size: 11, color: Color(0xFFE65100)),
                        SizedBox(width: 2),
                        Text(
                          "含提货卡",
                          style: TextStyle(
                            fontSize: 10,
                            color: Color(0xFFE65100),
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ],
            ),
          ),
          // Story 10.7：状态文字按状态分色（让用户一眼区分待付/待发货/已完成/已取消）
          Text(
            statusText,
            style: TextStyle(
              fontSize: 14,
              fontWeight: FontWeight.w600,
              color: _orderStatusColor(data.status),
            ),
          ),
        ],
      ),
    );
  }

  // 订单状态文字颜色（Story 10.7）
  // OMS: 1=待支付, 2=已支付/待发货, 3=已发货, 4=已完成, 5=已取消, 7=售后中
  Color _orderStatusColor(int status) {
    switch (status) {
      case 1:
        return const Color(0xFFFA436A); // 待支付 - 品牌红（最需行动）
      case 2:
      case 3:
        return const Color(0xFFE65100); // 已支付/已发货 - 橙色（进行中）
      case 4:
        return const Color(0xFF2E7D32); // 已完成 - 绿色
      case 5:
        return const Color(0xFF9E9E9E); // 已取消 - 灰色
      case 7:
        return const Color(0xFFFFA000); // 售后中 - 黄色
      default:
        return const Color(0xFF606266); // 未知 - 中性灰
    }
  }

  Widget _buildProductPreview() {
    final firstItem = data.orderItemData.first;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 15, vertical: 10),
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
            firstItem.skuPic,
            fit: BoxFit.cover,
          ),
          const SizedBox(width: 10),
          // 商品信息
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  firstItem.skuName,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    fontSize: 14,
                    color: Color(0xFF303133),
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  firstItem.formattedSpec.isNotEmpty ? firstItem.formattedSpec : " ",
                  maxLines: 1,
                  style: const TextStyle(
                    fontSize: 12,
                    color: Color(0xFF909399),
                  ),
                ),
                const SizedBox(height: 4),
                Row(
                  children: [
                    Text(
                      "￥${firstItem.skuPrice.toStringAsFixed(2)}",
                      style: const TextStyle(
                        fontSize: 14,
                        color: Color(0xFF303133),
                      ),
                    ),
                    const SizedBox(width: 8),
                    Text(
                      "x${firstItem.skuQuantity}",
                      style: const TextStyle(
                        fontSize: 12,
                        color: Color(0xFF909399),
                      ),
                    ),
                    if (data.orderItemData.length > 1) ...[
                      const SizedBox(width: 8),
                      Text(
                        "+${data.orderItemData.length - 1} 件",
                        style: const TextStyle(
                          fontSize: 12,
                          color: Color(0xFF909399),
                        ),
                      ),
                    ],
                  ],
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildAmountFooter() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 15, vertical: 10),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.center,
        mainAxisAlignment: MainAxisAlignment.end,
        children: [
          Text(
            "共 ${data.orderItemData.length} 件商品  实付款 ",
            style: const TextStyle(
              fontSize: 13,
              color: Color(0xFF707070),
            ),
          ),
          Text(
            " ￥",
            style: const TextStyle(
              fontSize: 12,
              color: Color(0xFF707070),
            ),
          ),
          Text(
            data.payAmount.toStringAsFixed(2),
            style: const TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.bold,
              color: Color(0xFF303133),
            ),
          ),
        ],
      ),
    );
  }
}
