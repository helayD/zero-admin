import 'package:flutter/material.dart';
import 'package:flutter_mall/model/app_version_policy.dart';
import 'package:flutter_mall/model/upgrade_gate_context.dart';
import 'package:flutter_mall/provider/app_lifecycle_provider.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/utils/upgrade_gate_service.dart';
import 'package:flutter_mall/widgets/upgrade_gate_dialog.dart';
import 'package:provider/provider.dart';

typedef UpgradeGateNavigationCallback = Future<void> Function(
  BuildContext context,
  PendingUpgradeContext pendingContext,
);

class UpgradeGatePage extends StatefulWidget {
  final PendingUpgradeContext pendingContext;
  final AppVersionPolicy initialPolicy;
  final UpgradeGateNavigationCallback onResolved;
  final UpgradeGateNavigationCallback? onContinueLater;
  final bool autoResolveWhenSatisfied;

  const UpgradeGatePage({
    super.key,
    required this.pendingContext,
    required this.initialPolicy,
    required this.onResolved,
    this.onContinueLater,
    this.autoResolveWhenSatisfied = true,
  });

  @override
  State<UpgradeGatePage> createState() => _UpgradeGatePageState();
}

class _UpgradeGatePageState extends State<UpgradeGatePage>
    with WidgetsBindingObserver {
  late PendingUpgradeContext _pendingContext;
  late AppVersionPolicy _policy;
  bool _launching = false;
  bool _checking = false;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    _pendingContext = widget.pendingContext;
    _policy = widget.initialPolicy;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _recordGateShown();
      if (!_policy.hasUpgradeGate && widget.autoResolveWhenSatisfied) {
        _resolveAndExit();
      }
    });
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed) {
      _refreshPolicy();
    }
  }

  void _recordGateShown() {
    if (!mounted) {
      return;
    }
    final lifecycleProvider = context.read<AppLifecycleProvider>();
    lifecycleProvider.recordUpgradeGateShown(
      scene: _pendingContext.scene,
      source: _pendingContext.recoveryContext?.source ?? '',
      intentId: _pendingContext.recoveryContext?.intentId ?? '',
      targetType: _pendingContext.recoveryContext?.targetTypeValue ??
          _pendingContext.targetType.name,
      blocking: _policy.blocking,
      currentVersion: _policy.currentVersion,
      requiredVersion: _policy.requiredDisplayVersion,
      platform: _policy.platform,
      channel: _policy.channel,
    );
    if (_policy.blocking) {
      lifecycleProvider.recordUpgradeGateBlocked(
        scene: _pendingContext.scene,
        source: _pendingContext.recoveryContext?.source ?? '',
        intentId: _pendingContext.recoveryContext?.intentId ?? '',
        targetType: _pendingContext.recoveryContext?.targetTypeValue ??
            _pendingContext.targetType.name,
        blocking: _policy.blocking,
        currentVersion: _policy.currentVersion,
        requiredVersion: _policy.requiredDisplayVersion,
        platform: _policy.platform,
        channel: _policy.channel,
      );
    }
  }

  Future<void> _refreshPolicy() async {
    if (_checking) {
      return;
    }
    setState(() {
      _checking = true;
    });

    try {
      final refreshed = await UpgradeGateService.queryPendingPolicy(
        _pendingContext,
      );
      if (!mounted) {
        return;
      }
      setState(() {
        _policy = refreshed;
        _errorMessage = null;
      });
      if (!_policy.hasUpgradeGate) {
        context.read<AppLifecycleProvider>().recordUpgradeCompleted(
              scene: _pendingContext.scene,
              source: _pendingContext.recoveryContext?.source ?? '',
              intentId: _pendingContext.recoveryContext?.intentId ?? '',
              targetType: _pendingContext.recoveryContext?.targetTypeValue ??
                  _pendingContext.targetType.name,
              blocking: _pendingContext.blocking,
              currentVersion: refreshed.currentVersion,
              requiredVersion: _pendingContext.requiredVersion,
              platform: refreshed.platform,
              channel: refreshed.channel,
            );
        await _resolveAndExit();
      }
    } catch (_) {
      if (!mounted) {
        return;
      }
      setState(() {
        _errorMessage = '暂时无法确认升级结果，你也可以稍后再次返回继续。';
      });
    } finally {
      if (mounted) {
        setState(() {
          _checking = false;
        });
      }
    }
  }

  Future<void> _resolveAndExit() async {
    await AppRecoveryStore.clearPendingUpgradeContext();
    if (!mounted) {
      return;
    }
    context.read<AppLifecycleProvider>().recordUpgradeRestoreSucceeded(
          scene: _pendingContext.scene,
          source: _pendingContext.recoveryContext?.source ?? '',
          intentId: _pendingContext.recoveryContext?.intentId ?? '',
          targetType: _pendingContext.recoveryContext?.targetTypeValue ??
              _pendingContext.targetType.name,
          blocking: _pendingContext.blocking,
          currentVersion: _policy.currentVersion,
          requiredVersion: _pendingContext.requiredVersion,
          platform: _policy.platform,
          channel: _policy.channel,
        );
    await widget.onResolved(context, _pendingContext);
  }

  Future<void> _handlePrimaryAction() async {
    if (_policy.isUpToDate) {
      if (widget.autoResolveWhenSatisfied) {
        await _resolveAndExit();
      } else if (Navigator.of(context).canPop()) {
        Navigator.of(context).pop(true);
      }
      return;
    }

    context.read<AppLifecycleProvider>().recordUpgradeActionTapped(
          scene: _pendingContext.scene,
          source: _pendingContext.recoveryContext?.source ?? '',
          intentId: _pendingContext.recoveryContext?.intentId ?? '',
          targetType: _pendingContext.recoveryContext?.targetTypeValue ??
              _pendingContext.targetType.name,
          blocking: _policy.blocking,
          currentVersion: _policy.currentVersion,
          requiredVersion: _policy.requiredDisplayVersion,
          platform: _policy.platform,
          channel: _policy.channel,
        );

    setState(() {
      _launching = true;
      _errorMessage = null;
    });

    final nextContext =
        _pendingContext.copyWith(actionTaken: 'upgrade_started');
    await AppRecoveryStore.savePendingUpgradeContext(nextContext);
    _pendingContext = nextContext;

    final launched = await UpgradeGateService.launchUpgrade(_policy);
    if (!mounted) {
      return;
    }

    setState(() {
      _launching = false;
      if (!launched) {
        _errorMessage = '当前渠道暂时无法自动打开升级入口，请稍后重试或联系支持。';
      }
    });
  }

  Future<void> _handleSecondaryAction() async {
    await AppRecoveryStore.clearPendingUpgradeContext();
    if (!mounted) {
      return;
    }
    context.read<AppLifecycleProvider>().recordUpgradeRestoreFallbackUsed(
          scene: _pendingContext.scene,
          source: _pendingContext.recoveryContext?.source ?? '',
          intentId: _pendingContext.recoveryContext?.intentId ?? '',
          targetType: _pendingContext.recoveryContext?.targetTypeValue ??
              _pendingContext.targetType.name,
          blocking: _pendingContext.blocking,
          currentVersion: _policy.currentVersion,
          requiredVersion: _pendingContext.requiredVersion,
          platform: _policy.platform,
          channel: _policy.channel,
        );
    if (widget.onContinueLater != null) {
      await widget.onContinueLater!(context, _pendingContext);
      return;
    }
    if (Navigator.of(context).canPop()) {
      Navigator.of(context).pop(false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xCC101828),
      body: Center(
        child: UpgradeGateDialog(
          policy: _policy,
          launching: _launching,
          checking: _checking,
          errorMessage: _errorMessage,
          onPrimaryAction: _handlePrimaryAction,
          onSecondaryAction:
              _policy.canContinueLater ? _handleSecondaryAction : null,
        ),
      ),
    );
  }
}
