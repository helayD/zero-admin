import 'package:flutter/material.dart';
import 'package:flutter_mall/model/digital_card/physical_fulfillment_model.dart';
import 'package:flutter_mall/theme/app_theme.dart';

class PhysicalFulfillmentTimeline extends StatelessWidget {
  final List<PhysicalFulfillmentTimelineItem> timeline;

  const PhysicalFulfillmentTimeline({
    super.key,
    required this.timeline,
  });

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
          '暂无实体卡进度，稍后下拉刷新可获取最新状态。',
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
          final PhysicalFulfillmentTimelineItem item = timeline[index];
          final bool isLast = index == timeline.length - 1;
          return Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              Column(
                children: <Widget>[
                  Container(
                    width: 14,
                    height: 14,
                    decoration: const BoxDecoration(
                      color: Color(0xFF2563EB),
                      shape: BoxShape.circle,
                    ),
                  ),
                  if (!isLast)
                    Container(
                      width: 2,
                      height: 52,
                      color: const Color(0xFF2563EB).withValues(alpha: 0.16),
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
                        item.actionText.trim().isEmpty
                            ? item.action
                            : item.actionText,
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
                                    color: const Color(0xFF2563EB),
                                  ),
                        ),
                      if (item.reason.trim().isNotEmpty) ...<Widget>[
                        const SizedBox(height: AppSpacing.xs),
                        Text(
                          item.reason,
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
