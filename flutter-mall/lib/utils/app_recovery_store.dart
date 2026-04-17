import 'dart:async';
import 'dart:convert';

import 'package:flutter_mall/config/constant_param.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/model/permission_flow_context.dart';
import 'package:flutter_mall/model/upgrade_gate_context.dart';
import 'package:flutter_mall/utils/shared_preferences_util.dart';

class AppRecoveryStore {
  static const String recentContextKey = 'app_recent_context';
  static const String pendingIntentKey = 'app_pending_recovery_intent';
  static const String activeIntentCandidateKey =
      'app_active_recovery_candidate';
  static const String pendingPermissionContextKey =
      'app_pending_permission_context';
  static const String pendingLostMediaKey = 'app_pending_permission_lost_media';
  static const String pendingUpgradeContextKey = 'app_pending_upgrade_context';
  static const String notificationPreferenceKey =
      'app_notification_preference_enabled';
  static const String commentDraftPrefix = 'app_comment_draft_';
  static const String afterSalesDraftPrefix = 'app_after_sales_draft_';

  static Future<void> saveRecentContext(AppRecentContext context) async {
    final normalized = _attachCurrentMember(context);
    if (!normalized.isRecoverable) {
      return;
    }
    await SharedPreferencesUtil.saveJsonString(
        recentContextKey, normalized.toJson());
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
    await SharedPreferencesUtil.saveJsonString(
        pendingIntentKey, normalized.toJson());
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

  static Future<void> savePendingPermissionContext(
    PermissionFlowContext context,
  ) async {
    await SharedPreferencesUtil.saveJsonString(
      pendingPermissionContextKey,
      context.toJson(),
    );
  }

  static PermissionFlowContext? peekPendingPermissionContext() {
    return PermissionFlowContext.tryParse(
      SharedPreferencesUtil.getJsonString(pendingPermissionContextKey),
    );
  }

  static Future<PermissionFlowContext?>
      consumePendingPermissionContext() async {
    final context = peekPendingPermissionContext();
    await clearPendingPermissionContext();
    return context;
  }

  static Future<void> clearPendingPermissionContext() async {
    await SharedPreferencesUtil.remove(pendingPermissionContextKey);
  }

  static Future<void> savePendingLostMedia(
    PermissionLostMediaSnapshot snapshot,
  ) async {
    await SharedPreferencesUtil.saveJsonString(
      pendingLostMediaKey,
      snapshot.toJson(),
    );
  }

  static PermissionLostMediaSnapshot? peekPendingLostMedia() {
    return PermissionLostMediaSnapshot.tryParse(
      SharedPreferencesUtil.getJsonString(pendingLostMediaKey),
    );
  }

  static Future<PermissionLostMediaSnapshot?> consumePendingLostMedia() async {
    final snapshot = peekPendingLostMedia();
    await clearPendingLostMedia();
    return snapshot;
  }

  static Future<void> clearPendingLostMedia() async {
    await SharedPreferencesUtil.remove(pendingLostMediaKey);
  }

  static Future<void> saveActiveIntentCandidate(
      AppRecentContext context) async {
    final normalized = _attachCurrentMember(context);
    if (!normalized.isRecoverable) {
      return;
    }
    await SharedPreferencesUtil.saveJsonString(
        activeIntentCandidateKey, normalized.toJson());
  }

  static AppRecentContext? peekActiveIntentCandidate() {
    return _loadContext(activeIntentCandidateKey, clearOnMismatch: true);
  }

  static AppRecentContext? peekCurrentIntentContext() {
    final pendingUpgrade = peekPendingUpgradeContext();
    return pendingUpgrade?.recoveryContext ??
        peekPendingIntent() ??
        peekActiveIntentCandidate() ??
        getRecentContext();
  }

  static Future<void> savePendingUpgradeContext(
    PendingUpgradeContext context,
  ) async {
    await SharedPreferencesUtil.saveJsonString(
      pendingUpgradeContextKey,
      context.toJson(),
    );
  }

  static PendingUpgradeContext? peekPendingUpgradeContext() {
    return PendingUpgradeContext.tryParse(
      SharedPreferencesUtil.getJsonString(pendingUpgradeContextKey),
    );
  }

  static Future<PendingUpgradeContext?> consumePendingUpgradeContext() async {
    final context = peekPendingUpgradeContext();
    await clearPendingUpgradeContext();
    return context;
  }

  static Future<void> clearPendingUpgradeContext() async {
    await SharedPreferencesUtil.remove(pendingUpgradeContextKey);
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
    if (!_matchesContext(
      candidate,
      targetType: targetType,
      targetId: targetId,
      tabIndex: tabIndex,
    )) {
      return;
    }
    await clearActiveIntentCandidate();
  }

  static Future<void> clearRecentContextIfMatches(
    AppRecentTargetType targetType, {
    int? targetId,
    int? tabIndex,
  }) async {
    final context = getRecentContext();
    if (!_matchesContext(
      context,
      targetType: targetType,
      targetId: targetId,
      tabIndex: tabIndex,
    )) {
      return;
    }
    await clearRecentContext();
  }

  static Future<void> clearRecoveryState() async {
    await clearPendingIntent();
    await clearPendingPermissionContext();
    await clearPendingLostMedia();
    await clearPendingUpgradeContext();
    await clearActiveIntentCandidate();
    await clearRecentContext();
    await clearNotificationPreference();
    await _clearEntriesWithPrefixes(
      const <String>[commentDraftPrefix, afterSalesDraftPrefix],
    );
  }

  static Future<void> saveNotificationPreferenceEnabled(bool value) async {
    await SharedPreferencesUtil.saveBool(notificationPreferenceKey, value);
  }

  static bool getNotificationPreferenceEnabled() {
    return SharedPreferencesUtil.getBool(notificationPreferenceKey) ?? false;
  }

  static Future<void> clearNotificationPreference() async {
    await SharedPreferencesUtil.remove(notificationPreferenceKey);
  }

  static Future<void> saveCommentDraft(CommentDraftSnapshot draft) async {
    await SharedPreferencesUtil.saveJsonString(
      _commentDraftKey(draft.orderId),
      draft.toJson(),
    );
  }

  static CommentDraftSnapshot? getCommentDraft(int orderId) {
    return CommentDraftSnapshot.tryParse(
      SharedPreferencesUtil.getJsonString(_commentDraftKey(orderId)),
    );
  }

  static Future<void> clearCommentDraft(int orderId) async {
    await SharedPreferencesUtil.remove(_commentDraftKey(orderId));
  }

  static Future<void> saveAfterSalesDraft(AfterSalesDraftSnapshot draft) async {
    await SharedPreferencesUtil.saveJsonString(
      _afterSalesDraftKey(draft.orderId),
      draft.toJson(),
    );
  }

  static AfterSalesDraftSnapshot? getAfterSalesDraft(int orderId) {
    return AfterSalesDraftSnapshot.tryParse(
      SharedPreferencesUtil.getJsonString(_afterSalesDraftKey(orderId)),
    );
  }

  static Future<void> clearAfterSalesDraft(int orderId) async {
    await SharedPreferencesUtil.remove(_afterSalesDraftKey(orderId));
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

  static AppRecentContext? _loadContext(String key,
      {required bool clearOnMismatch}) {
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
    if (!context.requiresAuth ||
        (context.ownerMemberId != null && context.ownerMemberId! > 0)) {
      return context;
    }
    final currentMemberId = getCurrentMemberId();
    if (currentMemberId == null || currentMemberId <= 0) {
      return context;
    }
    return context.copyWith(ownerMemberId: currentMemberId);
  }

  static bool _matchesContext(
    AppRecentContext? context, {
    required AppRecentTargetType targetType,
    int? targetId,
    int? tabIndex,
  }) {
    if (context == null || context.targetType != targetType) {
      return false;
    }
    if (targetId != null && context.targetId != targetId) {
      return false;
    }
    if (tabIndex != null && context.tabIndex != tabIndex) {
      return false;
    }
    return true;
  }

  static Future<void> _clearEntriesWithPrefixes(List<String> prefixes) async {
    final keys = SharedPreferencesUtil.getKeys()
        .where((key) => prefixes.any((prefix) => key.startsWith(prefix)))
        .toList();
    if (keys.isEmpty) {
      return;
    }
    await Future.wait(
      keys.map(SharedPreferencesUtil.remove),
    );
  }

  static String _commentDraftKey(int orderId) {
    return '$commentDraftPrefix$orderId';
  }

  static String _afterSalesDraftKey(int orderId) {
    return '$afterSalesDraftPrefix$orderId';
  }
}
