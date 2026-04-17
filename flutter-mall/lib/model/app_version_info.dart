import 'package:flutter/foundation.dart';

class AppVersionInfo {
  final String version;
  final String buildNumber;
  final String platform;
  final String installerStore;
  final String channel;
  final bool isFallback;

  const AppVersionInfo({
    required this.version,
    required this.buildNumber,
    required this.platform,
    required this.installerStore,
    required this.channel,
    required this.isFallback,
  });

  factory AppVersionInfo.fallback({
    required String fallbackVersion,
    String buildNumber = '1',
  }) {
    final platform = detectAppPlatform();
    return AppVersionInfo(
      version: fallbackVersion,
      buildNumber: buildNumber,
      platform: platform,
      installerStore: '',
      channel: 'direct',
      isFallback: true,
    );
  }

  AppVersionInfo copyWith({
    String? version,
    String? buildNumber,
    String? platform,
    String? installerStore,
    String? channel,
    bool? isFallback,
  }) {
    return AppVersionInfo(
      version: version ?? this.version,
      buildNumber: buildNumber ?? this.buildNumber,
      platform: platform ?? this.platform,
      installerStore: installerStore ?? this.installerStore,
      channel: channel ?? this.channel,
      isFallback: isFallback ?? this.isFallback,
    );
  }
}

String detectAppPlatform() {
  if (kIsWeb) {
    return 'web';
  }
  switch (defaultTargetPlatform) {
    case TargetPlatform.android:
      return 'android';
    case TargetPlatform.iOS:
      return 'ios';
    case TargetPlatform.macOS:
      return 'macos';
    case TargetPlatform.windows:
      return 'windows';
    case TargetPlatform.linux:
      return 'linux';
    case TargetPlatform.fuchsia:
      return 'fuchsia';
  }
}
