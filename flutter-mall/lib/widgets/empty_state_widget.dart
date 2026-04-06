import 'package:flutter/material.dart';

import 'commerce_state_shell.dart';

enum EmptyStateDisplayMode {
  inline,
  page,
}

///
/// 统一空状态组件
///
/// 用于商品列表、分类、品牌等页面在无数据时展示明确的空状态与下一步推荐入口。
/// 遵循 Commerce State Shell 规范：空态提供明确下一步推荐入口，不使用静默空白。
///
class EmptyStateWidget extends StatelessWidget {
  final String message;
  final String? actionText;
  final VoidCallback? onAction;
  final IconData icon;
  final EmptyStateDisplayMode displayMode;

  const EmptyStateWidget({
    super.key,
    this.message = '暂无数据',
    this.actionText,
    this.onAction,
    this.icon = Icons.inbox_outlined,
    this.displayMode = EmptyStateDisplayMode.inline,
  });

  @override
  Widget build(BuildContext context) {
    if (displayMode == EmptyStateDisplayMode.page) {
      return CommerceStateShell(
        state: CommercePageState.empty,
        title: '暂无内容',
        summary: message,
        icon: icon,
        primaryAction: actionText != null && onAction != null
            ? CommerceStateAction(label: actionText!, onPressed: onAction!)
            : null,
      );
    }

    return Semantics(
      container: true,
      liveRegion: true,
      label: [
        '空状态',
        message,
        if (actionText != null && actionText!.isNotEmpty) actionText!,
      ].join('，'),
      child: Center(
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: 60, horizontal: 24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(
                icon,
                size: 64,
                color: const Color(0xFFBDBDBD),
              ),
              const SizedBox(height: 16),
              Text(
                message,
                textAlign: TextAlign.center,
                style: const TextStyle(
                  fontSize: 14,
                  color: Color(0xFF909399),
                ),
              ),
              if (actionText != null && onAction != null) ...[
                const SizedBox(height: 20),
                OutlinedButton(
                  onPressed: onAction,
                  style: OutlinedButton.styleFrom(
                    foregroundColor: const Color(0xFFFA436A),
                    side: const BorderSide(color: Color(0xFFFA436A)),
                    padding: const EdgeInsets.symmetric(
                        horizontal: 24, vertical: 10),
                  ),
                  child: Text(actionText!),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}

///
/// 统一错误/重试状态组件
///
class ErrorRetryWidget extends StatelessWidget {
  final String message;
  final VoidCallback onRetry;
  final String retryText;
  final String? secondaryActionText;
  final VoidCallback? onSecondaryAction;

  const ErrorRetryWidget({
    super.key,
    this.message = '加载失败，请重试',
    required this.onRetry,
    this.retryText = '重试',
    this.secondaryActionText,
    this.onSecondaryAction,
  });

  @override
  Widget build(BuildContext context) {
    return CommerceStateShell(
      state: CommercePageState.error,
      title: '加载失败',
      summary: message,
      primaryAction: CommerceStateAction(label: retryText, onPressed: onRetry),
      secondaryAction: secondaryActionText != null && onSecondaryAction != null
          ? CommerceStateAction(
              label: secondaryActionText!,
              onPressed: onSecondaryAction!,
            )
          : null,
    );
  }
}
