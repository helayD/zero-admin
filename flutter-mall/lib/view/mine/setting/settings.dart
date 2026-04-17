import 'package:flutter/material.dart';
import 'package:flutter_mall/layout/upgrade_gate_page.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/model/permission_flow_context.dart';
import 'package:flutter_mall/model/upgrade_gate_context.dart';
import 'package:flutter_mall/provider/app_lifecycle_provider.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/utils/app_version_service.dart';
import 'package:flutter_mall/utils/permission_broker.dart';
import 'package:flutter_mall/utils/upgrade_gate_service.dart';
import 'package:flutter_mall/view/mine/message/message.dart';
import 'package:flutter_mall/view/mine/profile/profile_edit.dart';
import 'package:flutter_mall/widgets/permission_prompt_sheet.dart';
import 'package:provider/provider.dart';

///
/// 设置页面
///
/// 作者：David
/// 日期：2023/11/21 17:17
///
class Settings extends StatefulWidget {
  final PermissionBroker permissionBroker;

  const Settings({
    super.key,
    this.permissionBroker = const PermissionBroker(),
  });

  @override
  State<Settings> createState() => _SettingsState();
}

class _SettingsState extends State<Settings> {
  NotificationPreferenceSnapshot? _notificationSnapshot;
  bool _isLoadingNotification = true;
  AppLifecycleProvider? _lifecycleProvider;
  int _lastResumeTick = 0;
  String _currentVersionLabel = AppVersionService.currentInfo.version;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) async {
      await _refreshNotificationSnapshot();
      await _refreshVersionLabel();
      await _resumePendingNotificationFlow();
      await AppRecoveryStore.saveRecentContext(
        AppRecentContext.create(
          targetType: AppRecentTargetType.settings,
          source: 'manual_open',
          requiresAuth: false,
          fallbackType: AppRecentTargetType.home,
          fallbackTabIndex: 4,
        ),
      );
    });
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final provider = context.read<AppLifecycleProvider>();
    if (_lifecycleProvider == provider) {
      return;
    }
    _lifecycleProvider?.removeListener(_handleLifecycleChanged);
    _lifecycleProvider = provider;
    _lastResumeTick = provider.resumeTick;
    provider.addListener(_handleLifecycleChanged);
  }

  @override
  void dispose() {
    _lifecycleProvider?.removeListener(_handleLifecycleChanged);
    super.dispose();
  }

  Future<void> _refreshNotificationSnapshot() async {
    final snapshot =
        await widget.permissionBroker.getNotificationPreferenceSnapshot(
      flow: _buildNotificationFlow(),
    );
    if (!mounted) {
      return;
    }
    setState(() {
      _notificationSnapshot = snapshot;
      _isLoadingNotification = false;
    });
  }

  Future<void> _refreshVersionLabel() async {
    final versionInfo = await AppVersionService.getCurrentInfo();
    if (!mounted) {
      return;
    }
    setState(() {
      _currentVersionLabel = versionInfo.version;
    });
  }

  Future<void> _resumePendingNotificationFlow() async {
    final pending = AppRecoveryStore.peekPendingPermissionContext();
    if (pending == null ||
        !pending.matchesReturnTarget(AppRecentTargetType.settings)) {
      return;
    }

    await AppRecoveryStore.clearPendingPermissionContext();
    final inspection = await widget.permissionBroker.inspect(pending);
    if (inspection.isUsable && pending.requestedToggleEnabled) {
      await widget.permissionBroker.persistNotificationPreference(true);
      _recordPermissionGranted(pending);
      if (!mounted) {
        return;
      }
      _showToast('系统通知已允许，消息提醒已开启');
    } else if (!inspection.isUsable) {
      _recordPermissionDenied(pending);
      if (!mounted) {
        return;
      }
      _showToast('系统通知仍未开启，你仍可通过站内消息查看提醒');
    }
    await _refreshNotificationSnapshot();
  }

  void _handleLifecycleChanged() {
    final provider = _lifecycleProvider;
    if (provider == null) {
      return;
    }
    if (_lastResumeTick == provider.resumeTick) {
      return;
    }
    _lastResumeTick = provider.resumeTick;
    _resumePendingNotificationFlow();
  }

  PermissionFlowContext _buildNotificationFlow() {
    return PermissionFlowContext.create(
      scene: PermissionScene.notificationSubscription,
      permissionType: PermissionType.notification,
      source: PermissionFlowSource.notificationToggle,
      returnTarget: AppRecentTargetType.settings,
      fallbackAction: PermissionFallbackAction.openMessageCenter,
      intentId: 'settings_notification_toggle',
      recoveryId: 'settings_notification_toggle',
      requestedToggleEnabled: true,
      retryOnResume: true,
    );
  }

  Future<void> _handleNotificationChanged(bool value) async {
    if (_isLoadingNotification) {
      return;
    }

    if (!value) {
      await widget.permissionBroker.persistNotificationPreference(false);
      await _refreshNotificationSnapshot();
      if (!mounted) {
        return;
      }
      _showToast('已关闭 App 内提醒偏好，仍可在消息中心查看站内消息');
      return;
    }

    final flow = _buildNotificationFlow();
    final inspection = await widget.permissionBroker.inspect(flow);
    if (inspection.isUsable) {
      await widget.permissionBroker.persistNotificationPreference(true);
      _recordPermissionGranted(flow);
      await _refreshNotificationSnapshot();
      if (!mounted) {
        return;
      }
      _showToast('消息提醒已开启');
      return;
    }

    final action = await _showPermissionPrompt(
      flow,
      inspection.status,
      alternativeLabel: '查看站内消息',
      statusCheckOnly: inspection.statusCheckOnly,
    );
    if (!mounted || action == null) {
      return;
    }
    await _handleNotificationPromptAction(
      flow,
      action,
      currentStatus: inspection.status,
    );
  }

  Future<void> _handleNotificationPromptAction(
    PermissionFlowContext flow,
    PermissionPromptAction action, {
    required PermissionFlowStatus currentStatus,
  }) async {
    switch (action) {
      case PermissionPromptAction.request:
        final result = await widget.permissionBroker.request(flow);
        if (result.isUsable) {
          await widget.permissionBroker.persistNotificationPreference(true);
          _recordPermissionGranted(flow);
          await _refreshNotificationSnapshot();
          if (!mounted) {
            return;
          }
          _showToast('消息提醒已开启');
          return;
        }
        _recordPermissionDenied(flow);
        final followUpAction = await _showPermissionPrompt(
          flow,
          result.status,
          alternativeLabel: '查看站内消息',
          statusCheckOnly: result.statusCheckOnly,
        );
        if (!mounted || followUpAction == null) {
          return;
        }
        await _handleNotificationPromptAction(
          flow,
          followUpAction,
          currentStatus: result.status,
        );
        return;
      case PermissionPromptAction.openSettings:
        _recordPermissionSettingsRedirected(flow);
        await widget.permissionBroker.openSettingsForFlow(
          flow.copyWith(status: currentStatus),
        );
        return;
      case PermissionPromptAction.alternative:
        await widget.permissionBroker.persistNotificationPreference(false);
        _recordPermissionFallbackUsed(flow);
        await _refreshNotificationSnapshot();
        if (!mounted) {
          return;
        }
        await Navigator.of(context).push(
          MaterialPageRoute(builder: (_) => const Message()),
        );
        return;
      case PermissionPromptAction.dismiss:
        return;
    }
  }

  Future<PermissionPromptAction?> _showPermissionPrompt(
    PermissionFlowContext flow,
    PermissionFlowStatus status, {
    required String alternativeLabel,
    bool statusCheckOnly = false,
  }) {
    _recordPermissionPromptShown(flow);
    final shouldOpenSystemNotificationSettings =
        statusCheckOnly && status == PermissionFlowStatus.denied;
    return PermissionPromptSheet.show(
      context,
      flow: flow,
      status: status,
      alternativeLabel: alternativeLabel,
      detailOverride: shouldOpenSystemNotificationSettings
          ? '当前系统版本不会在应用内弹出通知授权，请前往系统通知设置开启提醒；未开启时你仍可通过站内消息中心和未读角标查看提醒。'
          : null,
      primaryActionOverride: shouldOpenSystemNotificationSettings
          ? PermissionPromptAction.openSettings
          : null,
      primaryLabelOverride:
          shouldOpenSystemNotificationSettings ? '去系统设置' : null,
    );
  }

  void _recordPermissionPromptShown(PermissionFlowContext flow) {
    _lifecycleProvider?.recordPermissionPromptShown(
      scene: permissionSceneToValue(flow.scene),
      source: permissionFlowSourceToValue(flow.source),
      intentId: flow.intentId,
      recoveryId: flow.recoveryId,
    );
  }

  void _recordPermissionGranted(PermissionFlowContext flow) {
    _lifecycleProvider?.recordPermissionGranted(
      scene: permissionSceneToValue(flow.scene),
      source: permissionFlowSourceToValue(flow.source),
      intentId: flow.intentId,
      recoveryId: flow.recoveryId,
    );
  }

  void _recordPermissionDenied(PermissionFlowContext flow) {
    _lifecycleProvider?.recordPermissionDenied(
      scene: permissionSceneToValue(flow.scene),
      source: permissionFlowSourceToValue(flow.source),
      intentId: flow.intentId,
      recoveryId: flow.recoveryId,
    );
  }

  void _recordPermissionSettingsRedirected(PermissionFlowContext flow) {
    _lifecycleProvider?.recordPermissionSettingsRedirected(
      scene: permissionSceneToValue(flow.scene),
      source: permissionFlowSourceToValue(flow.source),
      intentId: flow.intentId,
      recoveryId: flow.recoveryId,
    );
  }

  void _recordPermissionFallbackUsed(PermissionFlowContext flow) {
    _lifecycleProvider?.recordPermissionFallbackUsed(
      scene: permissionSceneToValue(flow.scene),
      source: permissionFlowSourceToValue(flow.source),
      intentId: flow.intentId,
      recoveryId: flow.recoveryId,
    );
  }

  void _showToast(String msg) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(msg)),
    );
  }

  Future<void> _openUpgradeGate() async {
    final recoveryContext = AppRecentContext.create(
      targetType: AppRecentTargetType.settings,
      source: 'settings_check',
      requiresAuth: false,
      fallbackType: AppRecentTargetType.home,
      fallbackTabIndex: 4,
    );
    final policy = await UpgradeGateService.queryPolicy(
      scene: 'settings_check',
      recoveryContext: recoveryContext,
    );
    if (!mounted) {
      return;
    }
    final pendingUpgrade = PendingUpgradeContext.forRecentContext(
      scene: 'settings_check',
      recoveryContext: recoveryContext,
      policy: policy,
    );
    await Navigator.of(context).push<bool>(
      MaterialPageRoute(
        builder: (_) => UpgradeGatePage(
          pendingContext: pendingUpgrade,
          initialPolicy: policy,
          autoResolveWhenSatisfied: false,
          onResolved: (gateContext, _) async {
            Navigator.of(gateContext).pop(true);
          },
          onContinueLater: policy.canContinueLater
              ? (gateContext, _) async {
                  Navigator.of(gateContext).pop(false);
                }
              : null,
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        backgroundColor: Colors.white,
        title: const Text('设置'),
        centerTitle: true,
      ),
      body: Container(
        color: const Color(0xFFF5F5F5),
        child: ListView(
          children: [
            const SizedBox(height: 8),
            _buildSettingsTile(
              title: '个人资料',
              onTap: () {
                Navigator.of(context).push(
                  MaterialPageRoute(
                    builder: (_) => const ProfileEdit(),
                  ),
                );
              },
            ),
            _buildSettingsTile(title: '收货地址'),
            _buildSettingsTile(title: '实名认证'),
            _buildNotificationTile(),
            _buildSettingsTile(
              title: '清除缓存',
              onTap: () async {
                await AppRecoveryStore.clearRecoveryState();
                if (!mounted) {
                  return;
                }
                _showToast('缓存已清理');
              },
            ),
            _buildSettingsTile(title: '关于九克城'),
            _buildSettingsTile(
              title: '检查更新',
              onTap: _openUpgradeGate,
              trailing: Text(
                '当前版本 $_currentVersionLabel',
                style: const TextStyle(fontSize: 12, color: Color(0xFF707070)),
              ),
            ),
            const SizedBox(height: 14),
            InkWell(
              onTap: () async {
                await AppRecoveryStore.clearAuthToken();
                await AppRecoveryStore.clearRecoveryState();
                if (context.mounted) {
                  Navigator.of(context).pop();
                }
              },
              child: Container(
                alignment: Alignment.center,
                height: 50,
                color: Colors.white,
                child: const Text(
                  '退出登录',
                  style: TextStyle(color: Color(0xFFFA436A), fontSize: 15),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildNotificationTile() {
    final snapshot = _notificationSnapshot;
    final enabled = snapshot?.isEnabled ?? false;
    final subtitle = _notificationSubtitle(snapshot);

    return Container(
      margin: const EdgeInsets.symmetric(vertical: 6),
      color: Colors.white,
      child: ListTile(
        title: const Text(
          '消息推送',
          style: TextStyle(fontSize: 15, color: Color(0xFF303133)),
        ),
        subtitle: Text(
          subtitle,
          style: const TextStyle(fontSize: 12, color: Color(0xFF909399)),
        ),
        trailing: Switch(
          value: enabled,
          materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
          onChanged: _isLoadingNotification ? null : _handleNotificationChanged,
          activeThumbColor: Colors.white,
          activeTrackColor: const Color(0xFFFA436A),
        ),
      ),
    );
  }

  Widget _buildSettingsTile({
    required String title,
    VoidCallback? onTap,
    Widget? trailing,
  }) {
    return InkWell(
      onTap: onTap,
      child: Container(
        height: 56,
        padding: const EdgeInsets.symmetric(horizontal: 16),
        decoration: const BoxDecoration(
          color: Colors.white,
          border: Border(
            bottom: BorderSide(color: Color(0xFFF5F5F5), width: 1),
          ),
        ),
        child: Row(
          children: [
            Expanded(
              child: Text(
                title,
                style: const TextStyle(fontSize: 15, color: Color(0xFF303133)),
              ),
            ),
            trailing ??
                Image.asset(
                  'images/right_arrow.png',
                  height: 16,
                  width: 17,
                ),
          ],
        ),
      ),
    );
  }

  String _notificationSubtitle(NotificationPreferenceSnapshot? snapshot) {
    if (_isLoadingNotification || snapshot == null) {
      return '正在同步系统通知状态...';
    }
    switch (snapshot.presentation) {
      case NotificationPreferencePresentation.enabled:
        return snapshot.statusCheckOnly
            ? '系统通知已允许，App 内提醒偏好已开启'
            : '系统通知已允许，重要提醒会优先通过系统通知展示';
      case NotificationPreferencePresentation.mutedLocally:
        return '系统通知可用，但 App 内提醒偏好当前关闭';
      case NotificationPreferencePresentation.blockedBySystem:
        return '系统通知未开启，可继续通过站内消息中心和未读角标查看提醒';
    }
  }
}
