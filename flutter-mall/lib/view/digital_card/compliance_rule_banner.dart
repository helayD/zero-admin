import 'package:flutter/material.dart';
import 'package:flutter_mall/theme/app_theme.dart';

class ComplianceRuleBanner extends StatelessWidget {
  final String title;
  final String summary;
  final String statusText;

  const ComplianceRuleBanner({
    super.key,
    required this.title,
    required this.summary,
    required this.statusText,
  });

  Color get _accentColor {
    if (statusText.contains('回收')) {
      return const Color(0xFFB42318);
    }
    if (statusText.contains('复核') || statusText.contains('限制')) {
      return const Color(0xFFB54708);
    }
    return AppColors.primaryDark;
  }

  @override
  Widget build(BuildContext context) {
    final String trimmedSummary = summary.trim();
    if (trimmedSummary.isEmpty) {
      return const SizedBox.shrink();
    }

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(AppSpacing.lg),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: <Color>[
            _accentColor.withValues(alpha: 0.12),
            Colors.white,
          ],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.circular(AppRadii.xl),
        border: Border.all(color: _accentColor.withValues(alpha: 0.28)),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Container(
            width: 40,
            height: 40,
            decoration: BoxDecoration(
              color: _accentColor.withValues(alpha: 0.14),
              borderRadius: BorderRadius.circular(AppRadii.md),
            ),
            child: Icon(
              Icons.verified_user_outlined,
              color: _accentColor,
            ),
          ),
          const SizedBox(width: AppSpacing.md),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Text(
                  title,
                  style: Theme.of(context).textTheme.titleSmall?.copyWith(
                        color: AppColors.textPrimary,
                      ),
                ),
                if (statusText.trim().isNotEmpty) ...<Widget>[
                  const SizedBox(height: AppSpacing.xs),
                  Text(
                    statusText,
                    style: Theme.of(context).textTheme.labelMedium?.copyWith(
                          color: _accentColor,
                        ),
                  ),
                ],
                const SizedBox(height: AppSpacing.sm),
                Text(
                  trimmedSummary,
                  style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                        color: AppColors.textSecondary,
                      ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
