import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/model/message_model.dart';
import 'package:flutter_mall/provider/app_lifecycle_provider.dart';
import 'package:flutter_mall/utils/app_intent_dispatcher.dart';
import 'package:flutter_mall/utils/app_recovery_router.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/utils/commerce_state_resolver.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/view/mine/login/login.dart';
import 'package:flutter_mall/widgets/commerce_state_shell.dart';
import 'package:provider/provider.dart';

///
/// 消息页面
///
/// Story 9-3：接入统一 recall intent dispatcher，并复用 Commerce State Shell
///
/// 作者：David
/// 日期：2026/4/2
///
class Message extends StatefulWidget {
  const Message({super.key});

  @override
  State<Message> createState() => _MessageState();
}

class _MessageState extends State<Message> {
  final List<MessageData> _messages = [];
  final Map<int, List<MessageData>> _tabMessages = {
    0: [],
    1: [],
    2: [],
    3: [],
    4: [],
  };
  int _selectedTab = 0; // 0-全部 1-订单 2-售后 3-活动 4-会员
  bool _isLoading = false;
  bool _hasMore = true;
  int _pageNum = 1;
  static const int _pageSize = 20;
  Object? _loadError;

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

    setState(() {
      _isLoading = true;
      if (reset) {
        _loadError = null;
      }
    });

    int? messageType;
    if (_selectedTab != 0) {
      messageType = _selectedTab;
    }

    try {
      final params = <String, String>{
        'pageNum': _pageNum.toString(),
        'pageSize': _pageSize.toString(),
      };
      if (messageType != null) {
        params['messageType'] = messageType.toString();
      }
      final queryString =
          params.entries.map((e) => '${e.key}=${e.value}').join('&');
      final Response result =
          await HttpUtil.get('$messageListDataUrl?$queryString');

      final model = MessageModel.fromJson(result.data);
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
        _loadError = null;
        _isLoading = false;
      });
    } catch (e) {
      if (mounted) {
        setState(() {
          _loadError = e;
          _isLoading = false;
        });
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

  Future<bool> _markAsRead(int messageId) async {
    try {
      await HttpUtil.post(messageReadUrl, data: {'id': messageId});
      return true;
    } catch (_) {
      return false;
    }
  }

  void _markLocalAsRead(int messageId) {
    MessageData patch(MessageData data) =>
        data.id == messageId ? data.copyWith(status: 1) : data;

    setState(() {
      for (final entry in _tabMessages.entries) {
        _tabMessages[entry.key] = entry.value.map(patch).toList();
      }
      for (var i = 0; i < _messages.length; i++) {
        if (_messages[i].id == messageId) {
          _messages[i] = _messages[i].copyWith(status: 1);
        }
      }
    });
  }

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

  Future<void> _handleMessageTap(MessageData msg) async {
    final intent = msg.intent;
    if (intent == null) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('跳转目标无效'), backgroundColor: Colors.orange),
      );
      return;
    }

    final resolvedIntent = intent.copyWith(
      source: 'message_tap',
      lastValidatedAt: DateTime.now(),
    );
    final lifecycleProvider = context.read<AppLifecycleProvider>();
    lifecycleProvider.recordIntentReceived(resolvedIntent);
    final plan = AppIntentDispatcher.resolve(
      resolvedIntent,
      hasValidToken: AppRecoveryStore.hasValidToken(),
    );

    if (plan.shouldMarkMessageRead && msg.status == 0) {
      final success = await _markAsRead(msg.id);
      if (success) {
        _markLocalAsRead(msg.id);
      }
    }

    if (!mounted) return;

    switch (plan.action) {
      case AppIntentDispatchAction.login:
        lifecycleProvider.recordIntentLoginRequired(plan.intent);
        await AppRecoveryStore.savePendingIntent(
          plan.intent.copyWith(
            source: 'login_restore',
            lastValidatedAt: DateTime.now(),
          ),
        );
        if (!mounted) return;
        await Navigator.of(context).push(
          MaterialPageRoute(
            builder: (_) => Login(recoveryIntent: plan.intent),
          ),
        );
        await _loadMessages(reset: true);
        return;
      case AppIntentDispatchAction.target:
        lifecycleProvider.recordIntentRestored(plan.intent);
        await AppRecoveryStore.saveActiveIntentCandidate(
          plan.intent.copyWith(lastValidatedAt: DateTime.now()),
        );
        if (!mounted) return;
        await AppRecoveryRouter.pushTarget(
          context,
          plan.intent,
          message: plan.message,
        );
        return;
      case AppIntentDispatchAction.fallback:
        lifecycleProvider.recordIntentFallbackUsed(
          plan.intent,
          failureReason: plan.failureReason,
        );
        if (!mounted) return;
        await AppRecoveryRouter.pushFallback(
          context,
          recentContext: plan.intent,
          message: plan.message,
        );
        return;
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
    final data =
        _selectedTab == 0 ? _messages : (_tabMessages[_selectedTab] ?? []);
    final pageState = CommerceStateResolver.resolvePageState(
      isLoading: _isLoading,
      hasContent: data.isNotEmpty,
      isEmpty: data.isEmpty,
      error: _loadError,
    );
    final failure = CommerceStateResolver.resolveFailure(
      _loadError,
      errorSummary: '消息加载失败，请稍后重试',
      weakNetworkSummary: '当前网络较弱，消息列表可能暂时无法刷新',
    );

    return CommerceStateShell(
      state: pageState,
      title: _stateTitle(pageState),
      summary: _stateSummary(pageState, failure),
      detail: failure?.detail,
      primaryAction: pageState == CommercePageState.content
          ? null
          : CommerceStateAction(
              label: '重新加载',
              onPressed: () {
                _loadMessages(reset: true);
              },
            ),
      secondaryAction: pageState == CommercePageState.empty
          ? CommerceStateAction(
              label: '返回上一页',
              onPressed: () {
                Navigator.of(context).maybePop();
              },
            )
          : null,
      showWeakNetworkBanner:
          data.isNotEmpty && failure != null && failure.isWeakNetwork,
      weakNetworkBannerAction:
          data.isNotEmpty && failure != null && failure.isWeakNetwork
              ? CommerceStateAction(
                  label: '重试刷新',
                  onPressed: () {
                    _loadMessages(reset: true);
                  },
                )
              : null,
      child: RefreshIndicator(
        onRefresh: _onRefresh,
        child: NotificationListener<ScrollNotification>(
          onNotification: (ScrollNotification scrollInfo) {
            if (scrollInfo.metrics.pixels >=
                scrollInfo.metrics.maxScrollExtent - 100) {
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
                  child:
                      Center(child: CircularProgressIndicator(strokeWidth: 2)),
                );
              }
              final msg = data[index];
              return _buildMessageCard(msg);
            },
          ),
        ),
      ),
    );
  }

  String _stateTitle(CommercePageState state) {
    switch (state) {
      case CommercePageState.initialLoading:
        return '正在加载消息';
      case CommercePageState.empty:
        return '暂无消息';
      case CommercePageState.error:
        return '消息加载失败';
      case CommercePageState.weakNetwork:
        return '网络较弱，消息加载受阻';
      case CommercePageState.content:
        return '消息列表';
    }
  }

  String _stateSummary(CommercePageState state, CommerceFailure? failure) {
    switch (state) {
      case CommercePageState.initialLoading:
        return '正在同步订单提醒、活动通知和优惠券消息';
      case CommercePageState.empty:
        return '订单、活动和优惠券相关消息会在这里统一展示';
      case CommercePageState.error:
      case CommercePageState.weakNetwork:
        return failure?.summary ?? '请稍后重试';
      case CommercePageState.content:
        return '';
    }
  }

  Widget _buildMessageCard(MessageData msg) {
    return Padding(
      padding: const EdgeInsets.only(top: 15),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Text(
                _formatTime(msg.createTime),
                style: TextStyle(
                  fontSize: 13,
                  color: Color(int.parse('7d7d7d', radix: 16)).withAlpha(255),
                ),
              ),
              const SizedBox(width: 8),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
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
          GestureDetector(
            onTap: () {
              unawaited(_handleMessageTap(msg));
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
                      color:
                          Color(int.parse('303133', radix: 16)).withAlpha(255),
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
                      color:
                          Color(int.parse('606266', radix: 16)).withAlpha(255),
                    ),
                    maxLines: 3,
                    overflow: TextOverflow.ellipsis,
                  ),
                  if (msg.intent?.failureReason.isNotEmpty == true) ...[
                    const SizedBox(height: 8),
                    Text(
                      msg.intent!.recoveryHint,
                      style: TextStyle(
                        fontSize: 12,
                        color: Color(int.parse('fa436a', radix: 16))
                            .withAlpha(255),
                      ),
                    ),
                  ],
                  Container(
                    margin: const EdgeInsets.only(top: 10),
                    padding: const EdgeInsets.only(top: 10),
                    decoration: BoxDecoration(
                      border: Border(
                        top: BorderSide(
                          width: 1,
                          color: Color(int.parse('f5f5f5', radix: 16))
                              .withAlpha(255),
                        ),
                      ),
                    ),
                    child: Row(
                      children: [
                        Expanded(
                          child: Text(
                            "查看详情",
                            style: TextStyle(
                              fontSize: 12,
                              color: Color(int.parse('707070', radix: 16))
                                  .withAlpha(255),
                            ),
                          ),
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

  String _formatTime(DateTime? time) {
    if (time == null) return '';
    return '${time.year}-${time.month.toString().padLeft(2, '0')}-${time.day.toString().padLeft(2, '0')} '
        '${time.hour.toString().padLeft(2, '0')}:${time.minute.toString().padLeft(2, '0')}';
  }
}
