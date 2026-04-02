import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/model/message_model.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/view/mine/coupon/coupon_list.dart';
import 'package:flutter_mall/view/mine/order/order_detail.dart';

///
/// 消息页面
///
/// Story 8-1 重构：实现真实 API 调用、分页、点击跳转
///
/// 作者：刘飞华
/// 日期：2026/4/2
///
class Message extends StatefulWidget {
  const Message({super.key});

  @override
  State<Message> createState() => _MessageState();
}

class _MessageState extends State<Message> {
  final List<MessageData> _messages = [];
  final Map<int, List<MessageData>> _tabMessages = {0: [], 1: [], 2: [], 3: [], 4: []};
  int _selectedTab = 0; // 0-全部 1-订单 2-售后 3-活动 4-会员
  bool _isLoading = false;
  bool _hasMore = true;
  int _pageNum = 1;
  static const int _pageSize = 20;

  @override
  void initState() {
    super.initState();
    _loadMessages(reset: true);
  }

  Future<void> _loadMessages({bool reset = false}) async {
    if (_isLoading) return;
    if (reset) {
      _pageNum = 1;
      _hasMore = true;
    }
    if (!_hasMore && !reset) return;

    setState(() => _isLoading = true);

    int? messageType;
    if (_selectedTab != 0) {
      messageType = _selectedTab; // 1-订单 2-售后 3-活动 4-会员
    }

    try {
      String url = messageListDataUrl;
      final params = <String, String>{
        'pageNum': _pageNum.toString(),
        'pageSize': _pageSize.toString(),
      };
      if (messageType != null) {
        params['messageType'] = messageType.toString();
      }
      final queryString = params.entries.map((e) => '${e.key}=${e.value}').join('&');
      Response result = await HttpUtil.get('$url?$queryString');

      MessageModel model = MessageModel.fromJson(result.data);
      if (!mounted) return;

      setState(() {
        if (reset) {
          _messages.clear();
          if (_selectedTab != 0) {
            _tabMessages[_selectedTab]?.clear();
          }
        }
        if (model.data.isEmpty) {
          _hasMore = false;
        } else {
          if (_selectedTab == 0) {
            _messages.addAll(model.data);
          } else {
            _tabMessages[_selectedTab]?.addAll(model.data);
          }
        }
        _isLoading = false;
      });
    } catch (e) {
      if (mounted) {
        setState(() => _isLoading = false);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('加载消息失败: $e'), backgroundColor: Colors.red),
        );
      }
    }
  }

  Future<void> _onRefresh() async {
    await _loadMessages(reset: true);
  }

  Future<void> _loadMore() async {
    if (!_hasMore || _isLoading) return;
    _pageNum++;
    await _loadMessages();
  }

  /// 标记单条消息已读
  Future<void> _markAsRead(int messageId) async {
    try {
      await HttpUtil.post(messageReadUrl, data: {'id': messageId});
    } catch (_) {}
  }

  /// 标记全部已读
  Future<void> _markAllAsRead() async {
    try {
      await HttpUtil.post(markAllReadUrl);
      await _loadMessages(reset: true);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('全部已读'), backgroundColor: Colors.green),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('操作失败: $e'), backgroundColor: Colors.red),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        backgroundColor: Colors.white,
        title: const Text("通知"),
        centerTitle: true,
        actions: [
          TextButton(
            onPressed: _markAllAsRead,
            child: Text(
              "全部已读",
              style: TextStyle(
                fontSize: 14,
                color: Color(int.parse('fa436a', radix: 16)).withAlpha(255),
              ),
            ),
          ),
        ],
      ),
      body: Column(
        children: [
          _buildTabBar(),
          Expanded(child: _buildMessageList()),
        ],
      ),
    );
  }

  Widget _buildTabBar() {
    final tabs = ['全部', '订单', '售后', '活动', '会员'];
    return Container(
      color: Colors.white,
      child: Row(
        children: List.generate(tabs.length, (index) {
          final isSelected = _selectedTab == index;
          return Expanded(
            child: GestureDetector(
              onTap: () {
                if (_selectedTab == index) return;
                setState(() => _selectedTab = index);
                _loadMessages(reset: true);
              },
              child: Container(
                padding: const EdgeInsets.symmetric(vertical: 12),
                decoration: BoxDecoration(
                  border: Border(
                    bottom: BorderSide(
                      width: 2,
                      color: isSelected
                          ? Color(int.parse('fa436a', radix: 16)).withAlpha(255)
                          : Colors.transparent,
                    ),
                  ),
                ),
                child: Text(
                  tabs[index],
                  textAlign: TextAlign.center,
                  style: TextStyle(
                    fontSize: 14,
                    color: isSelected
                        ? Color(int.parse('fa436a', radix: 16)).withAlpha(255)
                        : Color(int.parse('303133', radix: 16)).withAlpha(255),
                  ),
                ),
              ),
            ),
          );
        }),
      ),
    );
  }

  Widget _buildMessageList() {
    final data = _selectedTab == 0 ? _messages : (_tabMessages[_selectedTab] ?? []);

    if (_isLoading && data.isEmpty) {
      return const Center(child: CircularProgressIndicator());
    }

    if (data.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.notifications_none,
                size: 64, color: Color(int.parse('c0c0c0', radix: 16)).withAlpha(255)),
            const SizedBox(height: 16),
            Text("暂无消息",
                style: TextStyle(
                    fontSize: 14,
                    color: Color(int.parse('909399', radix: 16)).withAlpha(255))),
          ],
        ),
      );
    }

    return RefreshIndicator(
      onRefresh: _onRefresh,
      child: NotificationListener<ScrollNotification>(
        onNotification: (ScrollNotification scrollInfo) {
          if (scrollInfo.metrics.pixels >= scrollInfo.metrics.maxScrollExtent - 100) {
            _loadMore();
          }
          return false;
        },
        child: ListView.builder(
          padding: const EdgeInsets.symmetric(horizontal: 15),
          itemCount: data.length + (_hasMore ? 1 : 0),
          itemBuilder: (context, index) {
            if (index == data.length) {
              return const Padding(
                padding: EdgeInsets.symmetric(vertical: 16),
                child: Center(child: CircularProgressIndicator(strokeWidth: 2)),
              );
            }
            final msg = data[index];
            return _buildMessageCard(msg);
          },
        ),
      ),
    );
  }

  Widget _buildMessageCard(MessageData msg) {
    return Padding(
      padding: const EdgeInsets.only(top: 15),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // 时间 + 类型标签
          Row(
            children: [
              Text(
                _formatTime(msg.createTime),
                style: TextStyle(
                    fontSize: 13,
                    color: Color(int.parse('7d7d7d', radix: 16)).withAlpha(255)),
              ),
              const SizedBox(width: 8),
              Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                decoration: BoxDecoration(
                  color: Color(msg.typeColor).withAlpha(40),
                  borderRadius: BorderRadius.circular(4),
                ),
                child: Text(
                  msg.typeName,
                  style: TextStyle(
                    fontSize: 11,
                    color: Color(msg.typeColor),
                  ),
                ),
              ),
              const Spacer(),
              if (msg.status == 0)
                Container(
                  width: 8,
                  height: 8,
                  decoration: BoxDecoration(
                    color: Color(int.parse('fa436a', radix: 16)).withAlpha(255),
                    shape: BoxShape.circle,
                  ),
                ),
            ],
          ),
          // 消息卡片
          GestureDetector(
            onTap: () {
              if (msg.status == 0) {
                _markAsRead(msg.id);
              }
              _handleMessageTap(msg);
            },
            child: Container(
              color: Colors.white,
              margin: const EdgeInsets.only(top: 10),
              padding: const EdgeInsets.all(15),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    msg.title,
                    style: TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.w500,
                      color: Color(int.parse('303133', radix: 16)).withAlpha(255),
                    ),
                  ),
                  if (msg.imageUrl != null && msg.imageUrl!.isNotEmpty) ...[
                    const SizedBox(height: 10),
                    Image.network(
                      msg.imageUrl!,
                      width: 80,
                      height: 80,
                      fit: BoxFit.cover,
                      errorBuilder: (_, __, ___) => const SizedBox(),
                    ),
                  ],
                  const SizedBox(height: 8),
                  Text(
                    msg.content,
                    style: TextStyle(
                      fontSize: 14,
                      color: Color(int.parse('606266', radix: 16)).withAlpha(255),
                    ),
                    maxLines: 3,
                    overflow: TextOverflow.ellipsis,
                  ),
                  Container(
                    margin: const EdgeInsets.only(top: 10),
                    padding: const EdgeInsets.only(top: 10),
                    decoration: BoxDecoration(
                      border: Border(
                        top: BorderSide(
                          width: 1,
                          color: Color(int.parse('f5f5f5', radix: 16)).withAlpha(255),
                        ),
                      ),
                    ),
                    child: Row(
                      children: [
                        Expanded(
                          child: Text("查看详情",
                              style: TextStyle(
                                fontSize: 12,
                                color: Color(int.parse('707070', radix: 16))
                                    .withAlpha(255),
                              )),
                        ),
                        Image.asset(
                          "images/right_arrow.png",
                          height: 16,
                          width: 17,
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  void _handleMessageTap(MessageData msg) {
    // Story 8-1 Fix #3: 根据 linkType 跳转到对应页面
    final linkType = msg.linkType ?? '';
    final linkId = msg.linkId ?? '';

    if (linkType.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('跳转目标无效'), backgroundColor: Colors.orange),
      );
      return;
    }

    Widget targetPage;
    switch (linkType) {
      case 'order':
        // 跳转订单详情
        final orderId = int.tryParse(linkId) ?? 0;
        if (orderId <= 0) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('订单ID无效'), backgroundColor: Colors.orange),
          );
          return;
        }
        targetPage = OrderDetail(orderId: orderId, intentSource: 'message');
        break;
      case 'coupon':
        // 跳转优惠券列表
        targetPage = const CouponList();
        break;
      case 'product':
        // 跳转商品详情（如果有商品详情页面的话，用 message 作为来源）
        Navigator.of(context).pushNamed('/product', arguments: {
          'productId': linkId,
          'source': 'message',
        });
        return;
      case 'activity':
        // 跳转活动详情（如果没有活动详情页面，显示提示）
        Navigator.of(context).pushNamed('/activity', arguments: {
          'activityId': linkId,
          'source': 'message',
        });
        return;
      default:
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('暂不支持跳转类型: $linkType'),
            backgroundColor: Colors.orange,
          ),
        );
        return;
    }

    Navigator.of(context).push(
      MaterialPageRoute(builder: (context) => targetPage),
    );
  }

  String _formatTime(DateTime? time) {
    if (time == null) return '';
    return '${time.year}-${time.month.toString().padLeft(2, '0')}-${time.day.toString().padLeft(2, '0')} '
        '${time.hour.toString().padLeft(2, '0')}:${time.minute.toString().padLeft(2, '0')}';
  }
}
