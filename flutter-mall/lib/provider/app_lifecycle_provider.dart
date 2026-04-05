import 'package:flutter/material.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';

class AppLifecycleProvider extends ChangeNotifier with WidgetsBindingObserver {
  AppLifecycleState _currentState = AppLifecycleState.resumed;
  AppRecentContext? _pausedCandidate;
  int _resumeTick = 0;
  bool _restoreAttempted = false;
  bool _restoreSucceeded = false;
  bool _fallbackUsed = false;
  String? _restoreFailedReason;

  AppLifecycleState get currentState => _currentState;

  AppRecentContext? get pausedCandidate => _pausedCandidate;

  int get resumeTick => _resumeTick;

  bool get restoreAttempted => _restoreAttempted;

  bool get restoreSucceeded => _restoreSucceeded;

  bool get fallbackUsed => _fallbackUsed;

  String? get restoreFailedReason => _restoreFailedReason;

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
