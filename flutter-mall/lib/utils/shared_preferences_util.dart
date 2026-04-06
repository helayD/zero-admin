import 'dart:convert';

import 'package:shared_preferences/shared_preferences.dart';

///
/// 本地缓存工具类
///
/// 作者：David
/// 日期：2023/11/21 17:17
///
class SharedPreferencesUtil {
  static late SharedPreferences _prefs;

  static Future<void> init() async {
    _prefs = await SharedPreferences.getInstance();
  }

  static Future<void> saveString(String key, String value) async {
    await _prefs.setString(key, value);
  }

  static String? getString(String key) {
    return _prefs.getString(key);
  }

  static Future<void> saveInt(String key, int value) async {
    await _prefs.setInt(key, value);
  }

  static int? getInt(String key) {
    return _prefs.getInt(key);
  }

  static Future<void> saveDouble(String key, double value) async {
    await _prefs.setDouble(key, value);
  }

  static double? getDouble(String key) {
    return _prefs.getDouble(key);
  }

  static Future<void> saveBool(String key, bool value) async {
    await _prefs.setBool(key, value);
  }

  static bool? getBool(String key) {
    return _prefs.getBool(key);
  }

  static Future<void> saveJsonString(
    String key,
    Map<String, dynamic> value,
  ) async {
    await _prefs.setString(key, jsonEncode(value));
  }

  static dynamic getJsonString(String key) {
    final raw = _prefs.getString(key);
    if (raw == null || raw.isEmpty) {
      return null;
    }
    try {
      return jsonDecode(raw);
    } catch (_) {
      return null;
    }
  }

  static Future<void> remove(String key) async {
    await _prefs.remove(key);
  }

  static Set<String> getKeys() {
    return _prefs.getKeys();
  }

  static Future<void> clear() async {
    await _prefs.clear();
  }
}
