import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/layout/upgrade_gate_page.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/model/app_version_policy.dart';
import 'package:flutter_mall/model/message_model.dart';
import 'package:flutter_mall/model/upgrade_gate_context.dart';
import 'package:flutter_mall/provider/app_lifecycle_provider.dart';
import 'package:flutter_mall/theme/app_theme.dart';
import 'package:flutter_mall/utils/app_intent_dispatcher.dart';
import 'package:flutter_mall/utils/app_recovery_router.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/utils/commerce_state_resolver.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/utils/upgrade_gate_service.dart';
import 'package:flutter_mall/view/digital_card/digital_card_display_text.dart';
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
    5: [],
  };
  static const List<_MessageTab> _messageTabs = [
    _MessageTab(label: '全部', messageType: 0),
    _MessageTab(label: '订单', messageType: 1),
    _MessageTab(label: '支付', messageType: 2),
    _MessageTab(label: '售后', messageType: 3),
    _MessageTab(label: '活动', messageType: 4),
    _MessageTab(label: '会员', messageType: 5),
  ];

  int _selectedTab = 0;
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

    final int? messageType = _selectedTab == 0 ? null : _selectedTab;

    try {
      final params = <String, dynamic>{
        'pageNum': _pageNum,
        'pageSize': _pageSize,
      };
      if (messageType != null) {
        params['messageType'] = messageType;
      }
      final Response result = await HttpUtil.get(
        messageListDataUrl,
        queryParameters: params,
      );

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
        final int loadedCount =
            (_selectedTab == 0 ? _messages : _tabMessages[_selectedTab] ?? [])
                .length;
        _hasMore = model.total > 0
            ? loadedCount < model.total
            : model.data.length >= _pageSize;
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
      await HttpUtil.post(messageReadDataUrl(messageId));
      return true;
    } catch (_) {
      try {
        await HttpUtil.post(messageReadUrl, data: {'id': messageId});
        return true;
      } catch (_) {
        return false;
      }
    }
  }

  Future<bool> _deleteMessage(int messageId) async {
    try {
      await HttpUtil.delete(messageDeleteDataUrl(messageId));
      return true;
    } catch (_) {
      try {
        await HttpUtil.post(messageDeleteUrl, data: {'id': messageId});
        return true;
      } catch (_) {
        return false;
      }
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

  void _removeLocalMessage(int messageId) {
    setState(() {
      _messages.removeWhere((item) => item.id == messageId);
      for (final entry in _tabMessages.entries) {
        _tabMessages[entry.key] =
            entry.value.where((item) => item.id != messageId).toList();
      }
    });
  }

  Future<void> _openMessageDetail(MessageData msg) async {
    final _MessageDetailResult? result = await Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => _MessageDetailPage(
          initialMessage: msg,
          markAsRead: _markAsRead,
          deleteMessage: _deleteMessage,
        ),
      ),
    );
    if (!mounted || result == null) {
      return;
    }
    if (result.markedRead) {
      _markLocalAsRead(result.message.id);
    }
    if (result.deleted) {
      _removeLocalMessage(result.message.id);
      return;
    }
    if (result.openTarget) {
      await _handleMessageTap(result.message);
    }
  }

  Future<void> _markMessageReadIfNeeded(
    MessageData msg, {
    required bool shouldMarkMessageRead,
  }) async {
    if (!shouldMarkMessageRead || msg.status != 0) {
      return;
    }
    final success = await _markAsRead(msg.id);
    if (success) {
      _markLocalAsRead(msg.id);
    }
  }

  void _showUpgradePolicyUnavailableMessage() {
    if (!mounted) {
      return;
    }
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('暂时无法确认升级策略，请稍后重试'),
        backgroundColor: Colors.orange,
      ),
    );
  }

  Future<AppVersionPolicy?> _queryUpgradePolicyForMessage(
    AppRecentContext intent,
  ) async {
    final scene = UpgradeGateService.recoverySceneFor(intent);
    try {
      if (intent.isRecallIntent) {
        return await UpgradeGateService.queryRecallPolicy(intent, scene: scene);
      }
      return await UpgradeGateService.queryPolicy(
        scene: scene,
        recoveryContext: intent,
      );
    } catch (_) {
      _showUpgradePolicyUnavailableMessage();
      return null;
    }
  }

  Future<void> _restoreMessageTarget(
    MessageData msg,
    AppIntentDispatchPlan plan,
  ) async {
    final lifecycleProvider = context.read<AppLifecycleProvider>();
    lifecycleProvider.recordIntentRestored(plan.intent);
    await AppRecoveryStore.saveActiveIntentCandidate(
      plan.intent.copyWith(lastValidatedAt: DateTime.now()),
    );
    await _markMessageReadIfNeeded(
      msg,
      shouldMarkMessageRead: plan.shouldMarkMessageRead,
    );
    if (!mounted) return;
    await AppRecoveryRouter.pushTarget(
      context,
      plan.intent,
      message: plan.message,
    );
  }

  Future<void> _openMessageUpgradeGate(
    MessageData msg,
    AppIntentDispatchPlan plan,
    AppVersionPolicy upgradePolicy,
  ) async {
    final scene = UpgradeGateService.recoverySceneFor(plan.intent);
    final pendingUpgrade = PendingUpgradeContext.forRecentContext(
      scene: scene,
      recoveryContext: plan.intent,
      policy: upgradePolicy,
    );
    await AppRecoveryStore.savePendingUpgradeContext(pendingUpgrade);
    if (!mounted) return;
    await Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => UpgradeGatePage(
          pendingContext: pendingUpgrade,
          initialPolicy: upgradePolicy,
          onResolved: (gateContext, pending) async {
            await _markMessageReadIfNeeded(
              msg,
              shouldMarkMessageRead: plan.shouldMarkMessageRead,
            );
            if (!gateContext.mounted) {
              return;
            }
            await AppRecoveryRouter.replaceWithTarget(
              gateContext,
              pending.recoveryContext!,
              message: pending.recoveryHint.isNotEmpty
                  ? pending.recoveryHint
                  : plan.message,
            );
          },
          onContinueLater: upgradePolicy.canContinueLater
              ? (gateContext, _) async {
                  Navigator.of(gateContext).pop(false);
                }
              : null,
        ),
      ),
    );
  }

  Future<void> _markAllAsRead() async {
    try {
      await HttpUtil.post(markAllReadUrl);
      await _loadMessages(reset: true);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('已将全部消息标记为已读')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('标记已读失败，请稍后重试')),
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
        final upgradePolicy = await _queryUpgradePolicyForMessage(plan.intent);
        if (upgradePolicy == null) {
          return;
        }
        if (upgradePolicy.hasUpgradeGate) {
          await _openMessageUpgradeGate(msg, plan, upgradePolicy);
          return;
        }
        await _restoreMessageTarget(msg, plan);
        return;
      case AppIntentDispatchAction.fallback:
        lifecycleProvider.recordIntentFallbackUsed(
          plan.intent,
          failureReason: plan.failureReason,
        );
        await _markMessageReadIfNeeded(
          msg,
          shouldMarkMessageRead: plan.shouldMarkMessageRead,
        );
        if (!mounted) return;
        await AppRecoveryRouter.pushFallback(
          context,
          recentContext: plan.intent,
          message: plan.message,
        );
        return;
      case AppIntentDispatchAction.upgradeGate:
        final upgradePolicy = await _queryUpgradePolicyForMessage(plan.intent);
        if (upgradePolicy == null) {
          return;
        }
        if (!upgradePolicy.hasUpgradeGate) {
          await _restoreMessageTarget(msg, plan);
          return;
        }
        await _openMessageUpgradeGate(msg, plan, upgradePolicy);
        return;
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        backgroundColor: AppColors.surface,
        foregroundColor: AppColors.textPrimary,
        elevation: 0,
        title: const Text("消息中心"),
        centerTitle: true,
        actions: [
          TextButton(
            onPressed: _isLoading ? null : _markAllAsRead,
            child: Text(
              "全部已读",
              style: TextStyle(
                fontSize: 14,
                color: _isLoading ? AppColors.textHint : AppColors.accent,
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
    return Container(
      color: AppColors.surface,
      padding: const EdgeInsets.fromLTRB(
        AppSpacing.md,
        AppSpacing.sm,
        AppSpacing.md,
        AppSpacing.md,
      ),
      child: SingleChildScrollView(
        scrollDirection: Axis.horizontal,
        child: Row(
          children: _messageTabs.map((tab) {
            final isSelected = _selectedTab == tab.messageType;
            return Padding(
              padding: const EdgeInsets.only(right: AppSpacing.sm),
              child: ChoiceChip(
                label: Text(tab.label),
                selected: isSelected,
                onSelected: (_) {
                  if (_selectedTab == tab.messageType) return;
                  setState(() => _selectedTab = tab.messageType);
                  _loadMessages(reset: true);
                },
                selectedColor: AppColors.accentSoft,
                backgroundColor: AppColors.surfaceMuted,
                side: BorderSide(
                  color: isSelected ? AppColors.accent : AppColors.border,
                ),
                labelStyle: TextStyle(
                  color:
                      isSelected ? AppColors.accent : AppColors.textSecondary,
                  fontWeight: isSelected ? FontWeight.w700 : FontWeight.w500,
                ),
              ),
            );
          }).toList(),
        ),
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
              unawaited(_openMessageDetail(msg));
            },
            child: Container(
              color: Colors.white,
              margin: const EdgeInsets.only(top: 10),
              padding: const EdgeInsets.all(15),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    _safeMessageText(msg.title),
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
                    _safeMessageText(msg.content),
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
                      _safeMessageText(msg.intent!.recoveryHint),
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

String _safeMessageText(String text) {
  return digitalCardUserFacingText(text);
}

class _MessageTab {
  final String label;
  final int messageType;

  const _MessageTab({
    required this.label,
    required this.messageType,
  });
}

class _MessageDetailResult {
  final MessageData message;
  final bool markedRead;
  final bool deleted;
  final bool openTarget;

  const _MessageDetailResult({
    required this.message,
    this.markedRead = false,
    this.deleted = false,
    this.openTarget = false,
  });
}

class _MessageDetailPage extends StatefulWidget {
  final MessageData initialMessage;
  final Future<bool> Function(int messageId) markAsRead;
  final Future<bool> Function(int messageId) deleteMessage;

  const _MessageDetailPage({
    required this.initialMessage,
    required this.markAsRead,
    required this.deleteMessage,
  });

  @override
  State<_MessageDetailPage> createState() => _MessageDetailPageState();
}

class _MessageDetailPageState extends State<_MessageDetailPage> {
  late MessageData _message = widget.initialMessage;
  bool _isLoading = true;
  bool _isDeleting = false;
  bool _markedRead = false;
  String? _errorText;

  @override
  void initState() {
    super.initState();
    _loadDetail();
    _markReadIfNeeded();
  }

  Future<void> _loadDetail() async {
    try {
      final Response result = await HttpUtil.get(
        messageDetailDataUrl(widget.initialMessage.id),
      );
      final MessageDetailModel model = MessageDetailModel.fromJson(result.data);
      if (!mounted) {
        return;
      }
      setState(() {
        final MessageData nextMessage = model.data ?? _message;
        _message = _markedRead ? nextMessage.copyWith(status: 1) : nextMessage;
        _isLoading = false;
        _errorText = null;
      });
    } catch (_) {
      if (!mounted) {
        return;
      }
      setState(() {
        _isLoading = false;
        _errorText = '消息详情刷新失败，已展示本地消息内容';
      });
    }
  }

  Future<void> _markReadIfNeeded() async {
    if (_message.status != 0) {
      return;
    }
    final bool success = await widget.markAsRead(_message.id);
    if (!mounted || !success) {
      return;
    }
    setState(() {
      _message = _message.copyWith(status: 1);
      _markedRead = true;
    });
  }

  Future<void> _confirmDelete() async {
    final bool? confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('删除消息'),
        content: const Text('删除后将不再显示这条消息，是否继续？'),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(false),
            child: const Text('取消'),
          ),
          FilledButton(
            onPressed: () => Navigator.of(context).pop(true),
            child: const Text('删除'),
          ),
        ],
      ),
    );
    if (confirmed != true || !mounted) {
      return;
    }

    setState(() {
      _isDeleting = true;
    });
    final bool success = await widget.deleteMessage(_message.id);
    if (!mounted) {
      return;
    }
    if (!success) {
      setState(() {
        _isDeleting = false;
      });
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('删除失败，请稍后重试')),
      );
      return;
    }
    Navigator.of(context).pop(
      _MessageDetailResult(
        message: _message,
        markedRead: _markedRead,
        deleted: true,
      ),
    );
  }

  void _openTarget() {
    Navigator.of(context).pop(
      _MessageDetailResult(
        message: _message,
        markedRead: _markedRead,
        openTarget: true,
      ),
    );
  }

  void _closePage() {
    Navigator.of(context).pop(
      _markedRead
          ? _MessageDetailResult(message: _message, markedRead: true)
          : null,
    );
  }

  @override
  Widget build(BuildContext context) {
    final ThemeData theme = Theme.of(context);
    final bool hasTarget = _message.intent != null;

    return PopScope(
      canPop: false,
      onPopInvokedWithResult: (didPop, result) {
        if (!didPop) {
          _closePage();
        }
      },
      child: Scaffold(
        backgroundColor: AppColors.background,
        appBar: AppBar(
          backgroundColor: AppColors.surface,
          foregroundColor: AppColors.textPrimary,
          elevation: 0,
          leading: IconButton(
            onPressed: _closePage,
            icon: const Icon(Icons.arrow_back_rounded),
            tooltip: '返回',
          ),
          title: const Text('消息详情'),
          centerTitle: true,
          actions: [
            IconButton(
              onPressed: _isDeleting ? null : _confirmDelete,
              icon: const Icon(Icons.delete_outline_rounded),
              tooltip: '删除',
            ),
          ],
        ),
        body: ListView(
          padding: const EdgeInsets.all(AppSpacing.lg),
          children: [
            if (_errorText != null) ...[
              Container(
                padding: const EdgeInsets.all(AppSpacing.md),
                decoration: BoxDecoration(
                  color: const Color(0xFFFFF7E6),
                  borderRadius: BorderRadius.circular(AppRadii.md),
                  border: Border.all(color: const Color(0xFFFFD591)),
                ),
                child: Text(
                  _errorText!,
                  style: theme.textTheme.bodySmall?.copyWith(
                    color: const Color(0xFFD46B08),
                  ),
                ),
              ),
              const SizedBox(height: AppSpacing.md),
            ],
            Container(
              decoration: BoxDecoration(
                color: AppColors.surface,
                borderRadius: BorderRadius.circular(AppRadii.xl),
                border: Border.all(color: AppColors.border),
                boxShadow: const [
                  BoxShadow(
                    color: Color(0x0D101828),
                    blurRadius: 14,
                    offset: Offset(0, 6),
                  ),
                ],
              ),
              padding: const EdgeInsets.all(AppSpacing.lg),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      _MessageTypePill(message: _message),
                      const SizedBox(width: AppSpacing.sm),
                      Expanded(
                        child: Text(
                          _formatMessageTime(_message.createTime),
                          textAlign: TextAlign.right,
                          style: theme.textTheme.bodySmall?.copyWith(
                            color: AppColors.textHint,
                          ),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: AppSpacing.lg),
                  Text(
                    _safeMessageText(_message.title),
                    style: theme.textTheme.titleLarge?.copyWith(
                      color: AppColors.textPrimary,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                  if (_message.imageUrl != null &&
                      _message.imageUrl!.trim().isNotEmpty) ...[
                    const SizedBox(height: AppSpacing.lg),
                    ClipRRect(
                      borderRadius: BorderRadius.circular(AppRadii.md),
                      child: Image.network(
                        _message.imageUrl!,
                        width: double.infinity,
                        height: 180,
                        fit: BoxFit.cover,
                        errorBuilder: (_, __, ___) => const SizedBox.shrink(),
                      ),
                    ),
                  ],
                  const SizedBox(height: AppSpacing.lg),
                  Text(
                    _safeMessageText(_message.content),
                    style: theme.textTheme.bodyLarge?.copyWith(
                      color: AppColors.textSecondary,
                      height: 1.65,
                    ),
                  ),
                  if (_message.intent?.failureReason.isNotEmpty == true) ...[
                    const SizedBox(height: AppSpacing.lg),
                    Container(
                      width: double.infinity,
                      padding: const EdgeInsets.all(AppSpacing.md),
                      decoration: BoxDecoration(
                        color: AppColors.accentSoft,
                        borderRadius: BorderRadius.circular(AppRadii.md),
                      ),
                      child: Text(
                        _safeMessageText(_message.intent!.recoveryHint),
                        style: theme.textTheme.bodySmall?.copyWith(
                          color: AppColors.accent,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ),
                  ],
                ],
              ),
            ),
            const SizedBox(height: AppSpacing.lg),
            FilledButton.icon(
              onPressed: hasTarget && !_isLoading ? _openTarget : null,
              icon: const Icon(Icons.open_in_new_rounded),
              label: Text(hasTarget ? '查看相关内容' : '暂无可跳转内容'),
              style: FilledButton.styleFrom(
                minimumSize: const Size.fromHeight(48),
                backgroundColor: AppColors.primary,
                foregroundColor: Colors.white,
              ),
            ),
            const SizedBox(height: AppSpacing.sm),
            OutlinedButton.icon(
              onPressed: _isDeleting ? null : _confirmDelete,
              icon: _isDeleting
                  ? const SizedBox(
                      width: 16,
                      height: 16,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : const Icon(Icons.delete_outline_rounded),
              label: const Text('删除消息'),
              style: OutlinedButton.styleFrom(
                minimumSize: const Size.fromHeight(46),
                foregroundColor: AppColors.textSecondary,
                side: const BorderSide(color: AppColors.border),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _MessageTypePill extends StatelessWidget {
  final MessageData message;

  const _MessageTypePill({required this.message});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(
        horizontal: AppSpacing.sm,
        vertical: 4,
      ),
      decoration: BoxDecoration(
        color: Color(message.typeColor).withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(999),
      ),
      child: Text(
        message.typeName,
        style: Theme.of(context).textTheme.labelSmall?.copyWith(
              color: Color(message.typeColor),
              fontWeight: FontWeight.w700,
            ),
      ),
    );
  }
}

String _formatMessageTime(DateTime? time) {
  if (time == null) return '';
  return '${time.year}-${time.month.toString().padLeft(2, '0')}-${time.day.toString().padLeft(2, '0')} '
      '${time.hour.toString().padLeft(2, '0')}:${time.minute.toString().padLeft(2, '0')}';
}
