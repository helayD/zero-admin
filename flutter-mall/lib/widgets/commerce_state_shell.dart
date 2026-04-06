import 'package:flutter/material.dart';
import 'package:flutter_spinkit/flutter_spinkit.dart';
import 'package:flutter_mall/theme/app_theme.dart';

enum CommercePageState {
  initialLoading,
  empty,
  error,
  weakNetwork,
  content,
}

class CommerceStateAction {
  final String label;
  final VoidCallback onPressed;

  const CommerceStateAction({
    required this.label,
    required this.onPressed,
  });
}

class CommerceStateShell extends StatelessWidget {
  final CommercePageState state;
  final Widget? child;
  final String title;
  final String? summary;
  final String? detail;
  final CommerceStateAction? primaryAction;
  final CommerceStateAction? secondaryAction;
  final bool showWeakNetworkBanner;
  final String? weakNetworkBannerText;
  final CommerceStateAction? weakNetworkBannerAction;
  final IconData? weakNetworkBannerIcon;
  final Color? weakNetworkBannerBackgroundColor;
  final Color? weakNetworkBannerForegroundColor;
  final EdgeInsetsGeometry contentPadding;
  final IconData? icon;

  const CommerceStateShell({
    super.key,
    required this.state,
    required this.title,
    this.child,
    this.summary,
    this.detail,
    this.primaryAction,
    this.secondaryAction,
    this.showWeakNetworkBanner = false,
    this.weakNetworkBannerText,
    this.weakNetworkBannerAction,
    this.weakNetworkBannerIcon,
    this.weakNetworkBannerBackgroundColor,
    this.weakNetworkBannerForegroundColor,
    this.contentPadding = const EdgeInsets.all(16),
    this.icon,
  });

  @override
  Widget build(BuildContext context) {
    if (state == CommercePageState.content) {
      return Column(
        children: [
          if (showWeakNetworkBanner)
            _WeakNetworkBanner(
              text: weakNetworkBannerText ?? '当前网络较弱，已保留最近一次有效内容',
              action: weakNetworkBannerAction,
              icon: weakNetworkBannerIcon,
              backgroundColor: weakNetworkBannerBackgroundColor,
              foregroundColor: weakNetworkBannerForegroundColor,
            ),
          Expanded(
            child: child ?? const SizedBox.shrink(),
          ),
        ],
      );
    }

    return SafeArea(
      child: Padding(
        padding: contentPadding,
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 420),
            child: _StateCard(
              state: state,
              title: title,
              summary: summary,
              detail: detail,
              primaryAction: primaryAction,
              secondaryAction: secondaryAction,
              icon: icon,
            ),
          ),
        ),
      ),
    );
  }
}

class _WeakNetworkBanner extends StatelessWidget {
  final String text;
  final CommerceStateAction? action;
  final IconData? icon;
  final Color? backgroundColor;
  final Color? foregroundColor;

  const _WeakNetworkBanner({
    required this.text,
    this.action,
    this.icon,
    this.backgroundColor,
    this.foregroundColor,
  });

  @override
  Widget build(BuildContext context) {
    final Color resolvedForegroundColor =
        foregroundColor ?? const Color(0xFFD46B08);
    return Semantics(
      label: text,
      child: Container(
        width: double.infinity,
        color: backgroundColor ?? const Color(0xFFFFF7E6),
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
        child: Row(
          children: [
            Icon(
              icon ?? Icons.wifi_tethering_error_rounded,
              size: 18,
              color: resolvedForegroundColor,
            ),
            const SizedBox(width: 8),
            Expanded(
              child: Text(
                text,
                style: TextStyle(fontSize: 13, color: resolvedForegroundColor),
              ),
            ),
            if (action != null) ...[
              const SizedBox(width: 8),
              TextButton(
                onPressed: action!.onPressed,
                style: TextButton.styleFrom(
                  foregroundColor: resolvedForegroundColor,
                  padding:
                      const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                  minimumSize: const Size(0, 0),
                  tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                ),
                child: Text(action!.label),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

class _StateCard extends StatelessWidget {
  final CommercePageState state;
  final String title;
  final String? summary;
  final String? detail;
  final CommerceStateAction? primaryAction;
  final CommerceStateAction? secondaryAction;
  final IconData? icon;

  const _StateCard({
    required this.state,
    required this.title,
    this.summary,
    this.detail,
    this.primaryAction,
    this.secondaryAction,
    this.icon,
  });

  @override
  Widget build(BuildContext context) {
    final resolvedIcon = icon ?? _resolveIcon();
    final color = _resolveColor();
    final bool isLoading = state == CommercePageState.initialLoading;

    return Semantics(
      container: true,
      liveRegion: true,
      label: [title, summary, detail]
          .whereType<String>()
          .where((value) => value.isNotEmpty)
          .join('，'),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 28),
        decoration: BoxDecoration(
          color: AppColors.surface,
          borderRadius: BorderRadius.circular(20),
          boxShadow: const [
            BoxShadow(
              color: Color(0x14000000),
              blurRadius: 24,
              offset: Offset(0, 12),
            ),
          ],
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            if (isLoading)
              const SpinKitCircle(
                color: AppColors.primary,
                size: 40,
              )
            else
              Container(
                width: 72,
                height: 72,
                decoration: BoxDecoration(
                  color: color.withValues(alpha: 0.12),
                  borderRadius: BorderRadius.circular(36),
                ),
                child: Icon(resolvedIcon, size: 36, color: color),
              ),
            const SizedBox(height: 18),
            Text(
              title,
              textAlign: TextAlign.center,
              style: const TextStyle(
                fontSize: 18,
                fontWeight: FontWeight.w600,
                color: AppColors.textPrimary,
              ),
            ),
            if (summary != null && summary!.isNotEmpty) ...[
              const SizedBox(height: 10),
              Text(
                summary!,
                textAlign: TextAlign.center,
                style: const TextStyle(
                  fontSize: 14,
                  color: AppColors.textSecondary,
                  height: 1.5,
                ),
              ),
            ],
            if (detail != null && detail!.isNotEmpty) ...[
              const SizedBox(height: 8),
              Text(
                detail!,
                textAlign: TextAlign.center,
                style: const TextStyle(
                  fontSize: 13,
                  color: AppColors.textHint,
                  height: 1.5,
                ),
              ),
            ],
            if (primaryAction != null || secondaryAction != null) ...[
              const SizedBox(height: 20),
              Wrap(
                alignment: WrapAlignment.center,
                spacing: 12,
                runSpacing: 12,
                children: [
                  if (primaryAction != null)
                    Semantics(
                      button: true,
                      label: primaryAction!.label,
                      child: ElevatedButton(
                        onPressed: primaryAction!.onPressed,
                        style: ElevatedButton.styleFrom(
                          backgroundColor: AppColors.primary,
                          foregroundColor: Colors.white,
                          padding: const EdgeInsets.symmetric(
                              horizontal: 24, vertical: 12),
                        ),
                        child: Text(primaryAction!.label),
                      ),
                    ),
                  if (secondaryAction != null)
                    Semantics(
                      button: true,
                      label: secondaryAction!.label,
                      child: OutlinedButton(
                        onPressed: secondaryAction!.onPressed,
                        style: OutlinedButton.styleFrom(
                          foregroundColor: AppColors.primary,
                          side: const BorderSide(color: AppColors.primary),
                          padding: const EdgeInsets.symmetric(
                              horizontal: 24, vertical: 12),
                        ),
                        child: Text(secondaryAction!.label),
                      ),
                    ),
                ],
              ),
            ],
          ],
        ),
      ),
    );
  }

  IconData _resolveIcon() {
    switch (state) {
      case CommercePageState.initialLoading:
        return Icons.hourglass_bottom_rounded;
      case CommercePageState.empty:
        return Icons.inbox_outlined;
      case CommercePageState.error:
        return Icons.error_outline_rounded;
      case CommercePageState.weakNetwork:
        return Icons.wifi_tethering_error_rounded;
      case CommercePageState.content:
        return Icons.check_circle_outline;
    }
  }

  Color _resolveColor() {
    switch (state) {
      case CommercePageState.initialLoading:
        return AppColors.primary;
      case CommercePageState.empty:
        return AppColors.textHint;
      case CommercePageState.error:
        return const Color(0xFFF56C6C);
      case CommercePageState.weakNetwork:
        return const Color(0xFFE6A23C);
      case CommercePageState.content:
        return const Color(0xFF67C23A);
    }
  }
}
