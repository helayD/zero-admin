import 'package:flutter/material.dart';
import 'package:flutter_mall/model/digital_card/digital_card_asset_model.dart';
import 'package:flutter_mall/theme/app_theme.dart';

class MintStatusTimeline extends StatelessWidget {
  final List<DigitalCardAssetTimelineItem> timeline;

  const MintStatusTimeline({
    super.key,
    required this.timeline,
  });

  Color _colorForItem(DigitalCardAssetTimelineItem item) {
    final String text = '${item.operationType} ${item.statusText}';
    if (text.contains('成功') || text.contains('到账')) {
      return AppColors.success;
    }
    if (text.contains('回收')) {
      return const Color(0xFFD92D20);
    }
    if (text.contains('复核') || text.contains('限制') || text.contains('下线')) {
      return const Color(0xFFB54708);
    }
    return const Color(0xFF2563EB);
  }

  @override
  Widget build(BuildContext context) {
    if (timeline.isEmpty) {
      return Container(
        width: double.infinity,
        padding: const EdgeInsets.all(AppSpacing.lg),
        decoration: BoxDecoration(
          color: AppColors.surface,
          borderRadius: BorderRadius.circular(AppRadii.xl),
          border: Border.all(color: AppColors.border),
        ),
        child: Text(
          '暂无进度时间线，稍后下拉刷新可获取最新状态。',
          style: Theme.of(context).textTheme.bodyMedium,
        ),
      );
    }

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(AppSpacing.lg),
      decoration: BoxDecoration(
        color: AppColors.surface,
        borderRadius: BorderRadius.circular(AppRadii.xl),
        border: Border.all(color: AppColors.border),
      ),
      child: Column(
        children: List<Widget>.generate(timeline.length, (int index) {
          final DigitalCardAssetTimelineItem item = timeline[index];
          final bool isLast = index == timeline.length - 1;
          final Color accentColor = _colorForItem(item);
          return Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              Column(
                children: <Widget>[
                  Container(
                    width: 14,
                    height: 14,
                    decoration: BoxDecoration(
                      color: accentColor,
                      shape: BoxShape.circle,
                    ),
                  ),
                  if (!isLast)
                    Container(
                      width: 2,
                      height: 52,
                      color: accentColor.withValues(alpha: 0.18),
                    ),
                ],
              ),
              const SizedBox(width: AppSpacing.md),
              Expanded(
                child: Padding(
                  padding: EdgeInsets.only(bottom: isLast ? 0 : AppSpacing.md),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: <Widget>[
                      Text(
                        item.operationText.trim().isEmpty
                            ? item.operationType
                            : item.operationText,
                        style: Theme.of(context).textTheme.titleSmall?.copyWith(
                              color: AppColors.textPrimary,
                            ),
                      ),
                      const SizedBox(height: AppSpacing.xs),
                      if (item.statusText.trim().isNotEmpty)
                        Text(
                          item.statusText,
                          style:
                              Theme.of(context).textTheme.labelMedium?.copyWith(
                                    color: accentColor,
                                  ),
                        ),
                      if (item.reasonText.trim().isNotEmpty) ...<Widget>[
                        const SizedBox(height: AppSpacing.xs),
                        Text(
                          item.reasonText,
                          style: Theme.of(context).textTheme.bodyMedium,
                        ),
                      ],
                      const SizedBox(height: AppSpacing.xs),
                      Text(
                        item.createTime.trim().isEmpty
                            ? '时间待同步'
                            : item.createTime,
                        style: Theme.of(context).textTheme.bodySmall,
                      ),
                    ],
                  ),
                ),
              ),
            ],
          );
        }),
      ),
    );
  }
}
