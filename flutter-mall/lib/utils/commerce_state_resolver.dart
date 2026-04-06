import 'package:dio/dio.dart';

import '../widgets/commerce_state_shell.dart';

class CommerceFailure {
  final CommercePageState state;
  final String summary;
  final String? detail;

  const CommerceFailure({
    required this.state,
    required this.summary,
    this.detail,
  });

  bool get isWeakNetwork => state == CommercePageState.weakNetwork;
}

class CommerceStateResolver {
  static CommercePageState resolvePageState({
    required bool isLoading,
    required bool hasContent,
    required bool isEmpty,
    Object? error,
  }) {
    if (isLoading && !hasContent) {
      return CommercePageState.initialLoading;
    }
    if (!hasContent && error != null) {
      return isWeakNetworkError(error)
          ? CommercePageState.weakNetwork
          : CommercePageState.error;
    }
    if (!hasContent && isEmpty) {
      return CommercePageState.empty;
    }
    return CommercePageState.content;
  }

  static CommerceFailure? resolveFailure(
    Object? error, {
    String errorSummary = '加载失败，请稍后重试',
    String weakNetworkSummary = '当前网络较弱，请检查网络后重试',
  }) {
    if (error == null) {
      return null;
    }

    final detail = extractDetail(error);
    if (isWeakNetworkError(error)) {
      return CommerceFailure(
        state: CommercePageState.weakNetwork,
        summary: weakNetworkSummary,
        detail: detail,
      );
    }

    return CommerceFailure(
      state: CommercePageState.error,
      summary: errorSummary,
      detail: detail,
    );
  }

  static bool isWeakNetworkError(Object? error) {
    if (error is DioException) {
      switch (error.type) {
        case DioExceptionType.connectionTimeout:
        case DioExceptionType.sendTimeout:
        case DioExceptionType.receiveTimeout:
        case DioExceptionType.connectionError:
        case DioExceptionType.unknown:
          return error.response == null;
        case DioExceptionType.badCertificate:
        case DioExceptionType.badResponse:
        case DioExceptionType.cancel:
          return false;
      }
    }

    final text = error.toString().toLowerCase();
    return text.contains('timeout') ||
        text.contains('timed out') ||
        text.contains('connection error') ||
        text.contains('socketexception') ||
        text.contains('network is unreachable') ||
        text.contains('failed host lookup');
  }

  static String? extractDetail(Object? error) {
    if (error == null) {
      return null;
    }
    if (error is DioException) {
      final data = error.response?.data;
      if (data is Map<String, dynamic>) {
        final message = data['message']?.toString();
        if (message != null && message.isNotEmpty) {
          return message;
        }
      }
      if (error.message != null && error.message!.isNotEmpty) {
        return error.message;
      }
      if (error.response?.statusMessage != null &&
          error.response!.statusMessage!.isNotEmpty) {
        return error.response!.statusMessage;
      }
      return null;
    }

    final message = error.toString().trim();
    if (message.isEmpty || message == 'null') {
      return null;
    }
    return message;
  }
}
