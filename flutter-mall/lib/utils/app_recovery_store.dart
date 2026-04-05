import 'dart:async';
import 'dart:convert';

import 'package:flutter_mall/config/constant_param.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/utils/shared_preferences_util.dart';

class AppRecoveryStore {
  static const String recentContextKey = 'app_recent_context';
  static const String pendingIntentKey = 'app_pending_recovery_intent';
  static const String activeIntentCandidateKey = 'app_active_recovery_candidate';

  static Future<void> saveRecentContext(AppRecentContext context) async {
    final normalized = _attachCurrentMember(context);
    if (!normalized.isRecoverable) {
      return;
    }
    await SharedPreferencesUtil.saveJsonString(recentContextKey, normalized.toJson());
    await clearActiveIntentCandidate();
  }

  static AppRecentContext? getRecentContext() {
    return _loadContext(recentContextKey, clearOnMismatch: true);
  }

  static Future<void> clearRecentContext() async {
    await SharedPreferencesUtil.remove(recentContextKey);
  }

  static Future<void> savePendingIntent(AppRecentContext context) async {
    final normalized = _attachCurrentMember(context);
    if (!normalized.isRecoverable) {
      return;
    }
    await SharedPreferencesUtil.saveJsonString(pendingIntentKey, normalized.toJson());
  }

  static AppRecentContext? peekPendingIntent() {
    return _loadContext(pendingIntentKey, clearOnMismatch: true);
  }

  static Future<AppRecentContext?> consumePendingIntent() async {
    final context = peekPendingIntent();
    await clearPendingIntent();
    return context;
  }

  static Future<void> clearPendingIntent() async {
    await SharedPreferencesUtil.remove(pendingIntentKey);
  }

  static Future<void> saveActiveIntentCandidate(AppRecentContext context) async {
    final normalized = _attachCurrentMember(context);
    if (!normalized.isRecoverable) {
      return;
    }
    await SharedPreferencesUtil.saveJsonString(activeIntentCandidateKey, normalized.toJson());
  }

  static AppRecentContext? peekActiveIntentCandidate() {
    return _loadContext(activeIntentCandidateKey, clearOnMismatch: true);
  }

  static Future<void> clearActiveIntentCandidate() async {
    await SharedPreferencesUtil.remove(activeIntentCandidateKey);
  }

  static Future<void> clearActiveIntentCandidateIfMatches(
    AppRecentTargetType targetType, {
    int? targetId,
    int? tabIndex,
  }) async {
    final candidate = peekActiveIntentCandidate();
    if (candidate == null) {
      return;
    }
    if (candidate.targetType != targetType) {
      return;
    }
    if (targetId != null && candidate.targetId != targetId) {
      return;
    }
    if (tabIndex != null && candidate.tabIndex != tabIndex) {
      return;
    }
    await clearActiveIntentCandidate();
  }

  static Future<void> clearRecoveryState() async {
    await clearPendingIntent();
    await clearActiveIntentCandidate();
    await clearRecentContext();
  }

  static bool hasValidToken() {
    final currentToken = SharedPreferencesUtil.getString(token);
    return currentToken != null && currentToken.trim().isNotEmpty;
  }

  static String? getTokenSnapshot() {
    final currentToken = SharedPreferencesUtil.getString(token);
    if (currentToken == null || currentToken.trim().isEmpty) {
      return null;
    }
    return currentToken;
  }

  static int? getCurrentMemberId() {
    final snapshot = getTokenSnapshot();
    if (snapshot == null) {
      return null;
    }
    final tokenParts = snapshot.trim().split(' ');
    final jwt = tokenParts.isEmpty ? snapshot.trim() : tokenParts.last.trim();
    final segments = jwt.split('.');
    if (segments.length < 2) {
      return null;
    }
    try {
      final normalized = base64Url.normalize(segments[1]);
      final payload = utf8.decode(base64Url.decode(normalized));
      final json = jsonDecode(payload);
      if (json is! Map) {
        return null;
      }
      final rawMemberId = json['memberId'];
      if (rawMemberId is int) {
        return rawMemberId > 0 ? rawMemberId : null;
      }
      final memberId = int.tryParse(rawMemberId?.toString() ?? '');
      if (memberId == null || memberId <= 0) {
        return null;
      }
      return memberId;
    } catch (_) {
      return null;
    }
  }

  static bool isContextAllowedForCurrentMember(AppRecentContext context) {
    if (!context.requiresAuth) {
      return true;
    }
    final currentMemberId = getCurrentMemberId();
    if (currentMemberId == null || currentMemberId <= 0) {
      return true;
    }
    if (context.ownerMemberId == null || context.ownerMemberId! <= 0) {
      return true;
    }
    return context.ownerMemberId == currentMemberId;
  }

  static Future<void> persistAuthToken(String value) async {
    await SharedPreferencesUtil.saveString(token, value);
  }

  static Future<void> clearAuthToken() async {
    await SharedPreferencesUtil.remove(token);
  }

  static String? encodeContext(AppRecentContext? context) {
    if (context == null) {
      return null;
    }
    return jsonEncode(context.toJson());
  }

  static AppRecentContext? decodeContext(String? value) {
    if (value == null || value.isEmpty) {
      return null;
    }
    try {
      return AppRecentContext.tryParse(jsonDecode(value));
    } catch (_) {
      return null;
    }
  }

  static AppRecentContext? _loadContext(String key, {required bool clearOnMismatch}) {
    final raw = SharedPreferencesUtil.getJsonString(key);
    final context = AppRecentContext.tryParse(raw);
    if (context == null || !context.isRecoverable) {
      return null;
    }
    if (!isContextAllowedForCurrentMember(context)) {
      if (clearOnMismatch) {
        unawaited(SharedPreferencesUtil.remove(key));
      }
      return null;
    }
    return context;
  }

  static AppRecentContext _attachCurrentMember(AppRecentContext context) {
    if (!context.requiresAuth || (context.ownerMemberId != null && context.ownerMemberId! > 0)) {
      return context;
    }
    final currentMemberId = getCurrentMemberId();
    if (currentMemberId == null || currentMemberId <= 0) {
      return context;
    }
    return context.copyWith(ownerMemberId: currentMemberId);
  }
}
