import 'dart:convert';

/// Story 10.7：把后端返回的 specData / productAttr JSON 字符串解析成
/// 「容量: 256GB · 颜色: 金色 · 网络版本: 全网」的可读格式。
///
/// 后端可能的格式：
///   - JSON 对象：`{"容量": "256GB", "颜色": "金色"}`
///   - JSON 数组：`[{"key": "容量", "value": "256GB"}]`
///   - 已格式化的纯字符串：`"容量 256GB"`（直接返回）
///
/// 解析失败时回退原字符串，避免 UI 崩溃。
String formatSpec(String raw) {
  final value = raw.trim();
  if (value.isEmpty) return "";
  if (!value.startsWith('{') && !value.startsWith('[')) return value;
  try {
    final decoded = json.decode(value);
    final pairs = <String>[];
    if (decoded is Map) {
      for (final entry in decoded.entries) {
        final key = entry.key.toString().trim();
        final v = entry.value?.toString().trim() ?? "";
        if (key.isEmpty || v.isEmpty) continue;
        pairs.add("$key: $v");
      }
    } else if (decoded is List) {
      for (final item in decoded) {
        if (item is Map) {
          final key = item["key"]?.toString().trim() ?? "";
          final v = item["value"]?.toString().trim() ?? "";
          if (key.isEmpty || v.isEmpty) continue;
          pairs.add("$key: $v");
        }
      }
    }
    if (pairs.isEmpty) return value;
    return pairs.join(' · ');
  } catch (_) {
    return value;
  }
}
