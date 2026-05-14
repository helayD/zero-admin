import 'package:flutter/material.dart';
import 'package:flutter_mall/model/digital_card/digital_card_asset_model.dart';
import 'package:flutter_mall/theme/app_theme.dart';
import 'package:flutter_mall/view/digital_card/digital_card_display_text.dart';

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
      padding: const EdgeInsets.symmetric(
          horizontal: AppSpacing.lg, vertical: AppSpacing.md),
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
          final String operationText = digitalCardOperationLabel(
            item.operationText.trim().isEmpty
                ? item.operationType
                : item.operationText,
          );
          final String statusText = digitalCardStatusLabel(item.statusText);
          final String reasonText = digitalCardUserFacingText(item.reasonText);
          return IntrinsicHeight(
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: <Widget>[
                // 左侧时间轴
                SizedBox(
                  width: 20,
                  child: Column(
                    children: <Widget>[
                      Container(
                        width: 10,
                        height: 10,
                        margin: const EdgeInsets.only(top: 3),
                        decoration: BoxDecoration(
                          color: accentColor,
                          shape: BoxShape.circle,
                        ),
                      ),
                      if (!isLast)
                        Expanded(
                          child: Container(
                            width: 2,
                            margin: const EdgeInsets.symmetric(vertical: 2),
                            color: accentColor.withValues(alpha: 0.2),
                          ),
                        ),
                    ],
                  ),
                ),
                const SizedBox(width: AppSpacing.sm),
                // 右侧内容
                Expanded(
                  child: Padding(
                    padding: EdgeInsets.only(
                        bottom: isLast ? 0 : AppSpacing.sm),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: <Widget>[
                        Row(
                          children: <Widget>[
                            Expanded(
                              child: Text(
                                operationText,
                                style: Theme.of(context)
                                    .textTheme
                                    .bodyMedium
                                    ?.copyWith(
                                      color: AppColors.textPrimary,
                                      fontWeight: FontWeight.w500,
                                    ),
                              ),
                            ),
                            if (statusText.trim().isNotEmpty)
                              Text(
                                statusText,
                                style: Theme.of(context)
                                    .textTheme
                                    .labelSmall
                                    ?.copyWith(color: accentColor),
                              ),
                          ],
                        ),
                        if (reasonText.trim().isNotEmpty) ...<Widget>[
                          const SizedBox(height: 2),
                          Text(
                            reasonText,
                            style:
                                Theme.of(context).textTheme.bodySmall?.copyWith(
                                      color: AppColors.textSecondary,
                                    ),
                          ),
                        ],
                        const SizedBox(height: 2),
                        Text(
                          item.createTime.trim().isEmpty
                              ? '时间待同步'
                              : item.createTime,
                          style:
                              Theme.of(context).textTheme.bodySmall?.copyWith(
                                    color: AppColors.textSecondary,
                                    fontSize: 11,
                                  ),
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            ),
          );
        }),
      ),
    );
  }
}
