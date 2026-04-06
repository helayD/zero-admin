import 'package:flutter/material.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';

class IntentTelemetrySnapshot {
  final String eventName;
  final String? failureReason;
  final String targetType;
  final String source;
  final String intentId;
  final DateTime recordedAt;

  const IntentTelemetrySnapshot({
    required this.eventName,
    required this.targetType,
    required this.source,
    required this.intentId,
    required this.recordedAt,
    this.failureReason,
  });
}

class PermissionTelemetrySnapshot {
  final String eventName;
  final String scene;
  final String source;
  final String intentId;
  final String recoveryId;
  final DateTime recordedAt;

  const PermissionTelemetrySnapshot({
    required this.eventName,
    required this.scene,
    required this.source,
    required this.intentId,
    required this.recoveryId,
    required this.recordedAt,
  });
}

class AppLifecycleProvider extends ChangeNotifier with WidgetsBindingObserver {
  AppLifecycleState _currentState = AppLifecycleState.resumed;
  AppRecentContext? _pausedCandidate;
  int _resumeTick = 0;
  bool _restoreAttempted = false;
  bool _restoreSucceeded = false;
  bool _fallbackUsed = false;
  String? _restoreFailedReason;
  int _intentReceivedCount = 0;
  int _intentLoginRequiredCount = 0;
  int _intentRestoredCount = 0;
  int _intentFallbackUsedCount = 0;
  IntentTelemetrySnapshot? _latestIntentTelemetry;
  PermissionTelemetrySnapshot? _latestPermissionTelemetry;

  AppLifecycleState get currentState => _currentState;

  AppRecentContext? get pausedCandidate => _pausedCandidate;

  int get resumeTick => _resumeTick;

  bool get restoreAttempted => _restoreAttempted;

  bool get restoreSucceeded => _restoreSucceeded;

  bool get fallbackUsed => _fallbackUsed;

  String? get restoreFailedReason => _restoreFailedReason;

  int get intentReceivedCount => _intentReceivedCount;

  int get intentLoginRequiredCount => _intentLoginRequiredCount;

  int get intentRestoredCount => _intentRestoredCount;

  int get intentFallbackUsedCount => _intentFallbackUsedCount;

  IntentTelemetrySnapshot? get latestIntentTelemetry => _latestIntentTelemetry;

  PermissionTelemetrySnapshot? get latestPermissionTelemetry =>
      _latestPermissionTelemetry;

  void startObserving() {
    WidgetsBinding.instance.addObserver(this);
  }

  void markRestoreStarted() {
    _restoreAttempted = true;
    _restoreSucceeded = false;
    _fallbackUsed = false;
    _restoreFailedReason = null;
    notifyListeners();
  }

  void markRestoreSucceeded() {
    _restoreSucceeded = true;
    _fallbackUsed = false;
    _restoreFailedReason = null;
    notifyListeners();
  }

  void markFallbackUsed([String? reason]) {
    _fallbackUsed = true;
    _restoreSucceeded = false;
    _restoreFailedReason = reason;
    notifyListeners();
  }

  void markRestoreFailed(String reason) {
    _restoreSucceeded = false;
    _restoreFailedReason = reason;
    notifyListeners();
  }

  void recordIntentReceived(AppRecentContext intent) {
    _intentReceivedCount++;
    _recordIntentEvent('intentReceived', intent);
  }

  void recordIntentLoginRequired(AppRecentContext intent) {
    _intentLoginRequiredCount++;
    _recordIntentEvent(
      'intentLoginRequired',
      intent,
      failureReason: 'login_required',
    );
  }

  void recordIntentRestored(AppRecentContext intent) {
    _intentRestoredCount++;
    _recordIntentEvent('intentRestored', intent);
  }

  void recordIntentFallbackUsed(
    AppRecentContext intent, {
    String? failureReason,
  }) {
    _intentFallbackUsedCount++;
    _recordIntentEvent(
      'intentFallbackUsed',
      intent,
      failureReason: failureReason,
    );
  }

  void _recordIntentEvent(
    String eventName,
    AppRecentContext intent, {
    String? failureReason,
  }) {
    _latestIntentTelemetry = IntentTelemetrySnapshot(
      eventName: eventName,
      failureReason: failureReason,
      targetType: intent.targetTypeValue,
      source: intent.source,
      intentId: intent.intentId,
      recordedAt: DateTime.now(),
    );
    notifyListeners();
  }

  void recordPermissionPromptShown({
    required String scene,
    required String source,
    required String intentId,
    required String recoveryId,
  }) {
    _recordPermissionEvent(
      'permissionPromptShown',
      scene: scene,
      source: source,
      intentId: intentId,
      recoveryId: recoveryId,
    );
  }

  void recordPermissionGranted({
    required String scene,
    required String source,
    required String intentId,
    required String recoveryId,
  }) {
    _recordPermissionEvent(
      'permissionGranted',
      scene: scene,
      source: source,
      intentId: intentId,
      recoveryId: recoveryId,
    );
  }

  void recordPermissionDenied({
    required String scene,
    required String source,
    required String intentId,
    required String recoveryId,
  }) {
    _recordPermissionEvent(
      'permissionDenied',
      scene: scene,
      source: source,
      intentId: intentId,
      recoveryId: recoveryId,
    );
  }

  void recordPermissionSettingsRedirected({
    required String scene,
    required String source,
    required String intentId,
    required String recoveryId,
  }) {
    _recordPermissionEvent(
      'permissionSettingsRedirected',
      scene: scene,
      source: source,
      intentId: intentId,
      recoveryId: recoveryId,
    );
  }

  void recordPermissionFallbackUsed({
    required String scene,
    required String source,
    required String intentId,
    required String recoveryId,
  }) {
    _recordPermissionEvent(
      'permissionFallbackUsed',
      scene: scene,
      source: source,
      intentId: intentId,
      recoveryId: recoveryId,
    );
  }

  void recordPermissionLostDataRecovered({
    required String scene,
    required String source,
    required String intentId,
    required String recoveryId,
  }) {
    _recordPermissionEvent(
      'permissionLostDataRecovered',
      scene: scene,
      source: source,
      intentId: intentId,
      recoveryId: recoveryId,
    );
  }

  void _recordPermissionEvent(
    String eventName, {
    required String scene,
    required String source,
    required String intentId,
    required String recoveryId,
  }) {
    _latestPermissionTelemetry = PermissionTelemetrySnapshot(
      eventName: eventName,
      scene: scene,
      source: source,
      intentId: intentId,
      recoveryId: recoveryId,
      recordedAt: DateTime.now(),
    );
    notifyListeners();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    _currentState = state;
    if (state == AppLifecycleState.paused) {
      _pausedCandidate = AppRecoveryStore.getRecentContext();
      if (_pausedCandidate != null) {
        AppRecoveryStore.saveActiveIntentCandidate(
          _pausedCandidate!.copyWith(
            source: 'resume',
            lastValidatedAt: DateTime.now(),
          ),
        );
      }
    }
    if (state == AppLifecycleState.resumed) {
      _resumeTick++;
      if (_pausedCandidate != null) {
        AppRecoveryStore.saveActiveIntentCandidate(
          _pausedCandidate!.copyWith(
            source: 'resume',
            lastValidatedAt: DateTime.now(),
          ),
        );
      }
    }
    notifyListeners();
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }
}
