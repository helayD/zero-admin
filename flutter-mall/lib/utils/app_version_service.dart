import 'package:flutter_mall/config/constant_param.dart';
import 'package:flutter_mall/model/app_version_info.dart';
import 'package:package_info_plus/package_info_plus.dart';

abstract class AppVersionAdapter {
  Future<AppVersionInfo> load();
}

class PackageInfoAppVersionAdapter implements AppVersionAdapter {
  @override
  Future<AppVersionInfo> load() async {
    final packageInfo = await PackageInfo.fromPlatform();
    final installerStore = (packageInfo.installerStore ?? '').trim();
    return AppVersionInfo(
      version: packageInfo.version.trim().isEmpty
          ? appVersion
          : packageInfo.version.trim(),
      buildNumber: packageInfo.buildNumber.trim().isEmpty
          ? '1'
          : packageInfo.buildNumber.trim(),
      platform: detectAppPlatform(),
      installerStore: installerStore,
      channel: installerStore.isEmpty ? 'direct' : installerStore,
      isFallback: false,
    );
  }
}

class AppVersionService {
  static AppVersionAdapter _adapter = PackageInfoAppVersionAdapter();
  static AppVersionInfo _currentInfo =
      AppVersionInfo.fallback(fallbackVersion: appVersion);
  static bool _initialized = false;

  static AppVersionInfo get currentInfo => _currentInfo;

  static Future<void> init({bool forceRefresh = false}) async {
    if (_initialized && !forceRefresh) {
      return;
    }
    try {
      _currentInfo = await _adapter.load();
    } catch (_) {
      _currentInfo = AppVersionInfo.fallback(fallbackVersion: appVersion);
    }
    _initialized = true;
  }

  static Future<AppVersionInfo> getCurrentInfo() async {
    if (!_initialized) {
      await init();
    }
    return _currentInfo;
  }

  static void debugSetAdapter(AppVersionAdapter adapter) {
    _adapter = adapter;
    _initialized = false;
  }

  static void debugSetCurrentInfo(AppVersionInfo info) {
    _currentInfo = info;
    _initialized = true;
  }

  static void debugReset() {
    _adapter = PackageInfoAppVersionAdapter();
    _currentInfo = AppVersionInfo.fallback(fallbackVersion: appVersion);
    _initialized = false;
  }
}
