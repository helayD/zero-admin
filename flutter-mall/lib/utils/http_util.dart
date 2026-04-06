import 'dart:convert';
import 'dart:io' show HttpClient, Platform;

import 'package:dio/dio.dart';
import 'package:dio/io.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_mall/config/constant_param.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/view/mine/login/login.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../config/nav_key.dart';

///
/// http工具类
///
/// 作者：David
/// 日期：2023/11/21 17:17
///
class HttpUtil {
  static Dio? _dio;
  static bool _isRedirectingToLogin = false;

  static Dio get dio {
    if (_dio == null) {
      // macOS desktop 模式下，强制直连绕过系统代理，解决 sandbox 下访问外网被拒绝的问题
      final IOHttpClientAdapter httpClientAdapter = IOHttpClientAdapter(
        createHttpClient: () {
          final client = HttpClient();
          if (Platform.isMacOS) {
            client.findProxy = (uri) => 'DIRECT';
          }
          return client;
        },
      );

      // 配置Dio实例
      BaseOptions options = BaseOptions(
        baseUrl: "",
        connectTimeout: const Duration(milliseconds: 5000),
        receiveTimeout: const Duration(milliseconds: 5000),
      );

      _dio = Dio(options)..httpClientAdapter = httpClientAdapter;

      _dio!.interceptors.add(InterceptorsWrapper(
        onRequest: (RequestOptions options, RequestInterceptorHandler handler) {
          if (kDebugMode) {
            print("\n\n========================请求数据===================");
            print("url=${options.uri.toString()}");
            print("params=${options.data}");
          }
          return handler.next(options);
        },
        onResponse: (Response response, ResponseInterceptorHandler handler) {
          if (kDebugMode) {
            print("\n\n========================响应数据===================");
            print("code=${response.statusCode}");
            print("response=${response.data}");
            print("==================================================\n\n\n");
          }
          return handler.next(response);
        },
        onError: (DioException e, ErrorInterceptorHandler handler) async {
          if (kDebugMode) {
            print("\n\n========================错误数据===================");
            print("code=${e.response?.statusCode}");
            print("message=${e.response?.statusMessage}");
            print("data=${e.response?.data}");
            print("==================================================\n\n\n");
          }
          if (e.response?.statusCode == 401) {
            if (_isRedirectingToLogin) {
              return handler.next(e);
            }
            _isRedirectingToLogin = true;
            try {
              final navigatorState = NavKey.navKey.currentState;
              if (navigatorState == null) {
                return handler.next(e);
              }
              final recoveryIntent =
                  AppRecoveryStore.peekCurrentIntentContext();
              if (recoveryIntent != null) {
                await AppRecoveryStore.savePendingIntent(
                  recoveryIntent.copyWith(
                    source: 'login_restore',
                    lastValidatedAt: DateTime.now(),
                  ),
                );
              }
              await navigatorState.push(
                MaterialPageRoute(
                  builder: (context) => Login(recoveryIntent: recoveryIntent),
                ),
              );
            } finally {
              _isRedirectingToLogin = false;
            }
            return handler.next(e);
          }
          return handler.next(e);
        },
      ));
    }

    return _dio!;
  }

  // 封装GET请求
  static Future<Response> get(String path,
      {Map<String, dynamic>? queryParameters}) async {
    Response response = await dio.get(
      path,
      queryParameters: queryParameters,
      options: Options(headers: await _buildHeaders()),
    );
    return response;
  }

  // 封装POST请求，数据以JSON格式发送
  static Future<Response> post(String path,
      {Map<String, dynamic>? data}) async {
    Response response = await dio.post(
      path,
      data: jsonEncode(data),
      options: Options(
        contentType: 'application/json',
        headers: await _buildHeaders(),
      ),
    );
    return response;
  }

  // 封装POST请求，支持自定义请求头（用于幂等键等场景）
  static Future<Response> postWithHeaders(
    String path, {
    Map<String, dynamic>? data,
    Map<String, String>? headers,
  }) async {
    final header = await _buildHeaders();
    if (headers != null) {
      header.addAll(headers);
    }

    Response response = await dio.post(
      path,
      data: jsonEncode(data),
      options: Options(
        contentType: 'application/json',
        headers: header,
      ),
    );
    return response;
  }

  // 封装POST请求，数据以form表单格式发送
  static Future<Response> postForm(String path,
      {Map<String, dynamic>? data}) async {
    Response response = await dio.post(
      path,
      queryParameters: data,
      options: Options(headers: await _buildHeaders()),
    );
    return response;
  }

  static Future<Map<String, dynamic>> _buildHeaders() async {
    final SharedPreferences prefs = await SharedPreferences.getInstance();
    final Map<String, dynamic> header = <String, dynamic>{};
    header["Authorization"] = prefs.getString(token) ?? "";
    header["X-App-Version"] = appVersion;
    header["X-Client-Platform"] = _currentPlatform();
    header["X-Network-State"] = "unknown";

    final currentIntent = AppRecoveryStore.peekCurrentIntentContext();
    if (currentIntent != null) {
      if (currentIntent.source.trim().isNotEmpty) {
        header["X-Intent-Source"] = currentIntent.source;
      }
      if (currentIntent.intentId.trim().isNotEmpty) {
        header["X-Intent-Id"] = currentIntent.intentId;
      }
    }

    return header;
  }

  static String _currentPlatform() {
    if (kIsWeb) {
      return "web";
    }
    if (Platform.isAndroid) {
      return "android";
    }
    if (Platform.isIOS) {
      return "ios";
    }
    if (Platform.isMacOS) {
      return "macos";
    }
    if (Platform.isWindows) {
      return "windows";
    }
    if (Platform.isLinux) {
      return "linux";
    }
    return "unknown";
  }
}
