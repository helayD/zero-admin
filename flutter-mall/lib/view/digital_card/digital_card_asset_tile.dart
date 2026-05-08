import 'package:flutter/material.dart';
import 'package:flutter_mall/model/digital_card/digital_card_asset_model.dart';
import 'package:flutter_mall/theme/app_theme.dart';
import 'package:flutter_mall/widgets/cached_image_widget.dart';
import 'package:flutter_mall/view/digital_card/digital_card_display_text.dart';

class DigitalCardAssetTile extends StatelessWidget {
  final DigitalCardAssetItem item;
  final VoidCallback? onTap;

  const DigitalCardAssetTile({
    super.key,
    required this.item,
    this.onTap,
  });

  Color _statusColor() {
    if (item.complianceStatus == 'compliance_recycled' ||
        item.displayStatus == 'display_recycled') {
      return const Color(0xFFD92D20);
    }
    if (item.complianceStatus == 'compliance_review' ||
        item.complianceStatus == 'compliance_restricted' ||
        item.displayStatus == 'display_hidden' ||
        item.displayStatus == 'display_offlined') {
      return const Color(0xFFB54708);
    }
    if (item.mintStatus == 'mint_success') {
      return AppColors.success;
    }
    return const Color(0xFF2563EB);
  }

  @override
  Widget build(BuildContext context) {
    final Color accentColor = _statusColor();
    final DigitalCardStatusCopy statusCopy = digitalCardAssetPrimaryCopy(item);
    final String sourceName = item.sourceDisplayName.trim().isNotEmpty
        ? item.sourceDisplayName.trim()
        : item.activityName.trim();
    return Material(
      color: Colors.transparent,
      child: InkWell(
        borderRadius: BorderRadius.circular(AppRadii.xl),
        onTap: onTap,
        child: Container(
          padding: const EdgeInsets.all(AppSpacing.md),
          decoration: BoxDecoration(
            color: AppColors.surface,
            borderRadius: BorderRadius.circular(AppRadii.xl),
            border: Border.all(color: AppColors.border),
            boxShadow: const <BoxShadow>[
              BoxShadow(
                color: Color(0x110F172A),
                blurRadius: 18,
                offset: Offset(0, 10),
              ),
            ],
          ),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              ClipRRect(
                borderRadius: BorderRadius.circular(AppRadii.lg),
                child: CachedImageWidget(
                  88,
                  116,
                  item.cardFaceImage,
                ),
              ),
              const SizedBox(width: AppSpacing.md),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    Row(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: <Widget>[
                        Expanded(
                          child: Text(
                            item.templateName.trim().isEmpty
                                ? '提货卡'
                                : item.templateName,
                            style: Theme.of(context)
                                .textTheme
                                .titleMedium
                                ?.copyWith(
                                  color: AppColors.textPrimary,
                                  fontWeight: FontWeight.w700,
                                ),
                          ),
                        ),
                        if (item.rarity.trim().isNotEmpty)
                          Container(
                            padding: const EdgeInsets.symmetric(
                              horizontal: AppSpacing.sm,
                              vertical: AppSpacing.xs,
                            ),
                            decoration: BoxDecoration(
                              color: accentColor.withValues(alpha: 0.12),
                              borderRadius: BorderRadius.circular(999),
                            ),
                            child: Text(
                              item.rarity,
                              style: Theme.of(context)
                                  .textTheme
                                  .labelMedium
                                  ?.copyWith(
                                    color: accentColor,
                                  ),
                            ),
                          ),
                      ],
                    ),
                    const SizedBox(height: AppSpacing.sm),
                    Text(
                      '编号 ${item.assetNo.isEmpty ? '待分配' : item.assetNo}',
                      style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                            color: AppColors.textPrimary,
                          ),
                    ),
                    const SizedBox(height: AppSpacing.xs),
                    Text(
                      sourceName.isEmpty ? '来源待同步' : '来源 $sourceName',
                      style: Theme.of(context).textTheme.bodySmall?.copyWith(
                            color: AppColors.textSecondary,
                          ),
                    ),
                    const SizedBox(height: AppSpacing.md),
                    Wrap(
                      spacing: AppSpacing.sm,
                      runSpacing: AppSpacing.sm,
                      children: <Widget>[
                        _buildChip(
                          context,
                          statusCopy.label,
                          accentColor,
                        ),
                        _buildChip(
                          context,
                          digitalCardDisplayStatusText(
                            item.displayStatus,
                            item.displayStatusText,
                          ),
                          const Color(0xFF2563EB),
                        ),
                      ],
                    ),
                    const SizedBox(height: AppSpacing.md),
                    Text(
                      statusCopy.description,
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                      style: Theme.of(context).textTheme.bodySmall?.copyWith(
                            color: AppColors.textSecondary,
                          ),
                    ),
                    if (statusCopy.actionHint.trim().isNotEmpty) ...<Widget>[
                      const SizedBox(height: AppSpacing.xs),
                      Text(
                        statusCopy.actionHint,
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
                              color: accentColor,
                              fontWeight: FontWeight.w600,
                            ),
                      ),
                    ],
                    const SizedBox(height: AppSpacing.md),
                    Text(
                      item.obtainedAt.trim().isEmpty
                          ? '获取时间待同步'
                          : '获取时间 ${item.obtainedAt}',
                      style: Theme.of(context).textTheme.bodySmall,
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildChip(BuildContext context, String text, Color color) {
    final String trimmed = text.trim();
    if (trimmed.isEmpty) {
      return const SizedBox.shrink();
    }

    return Container(
      padding: const EdgeInsets.symmetric(
        horizontal: AppSpacing.sm,
        vertical: AppSpacing.xs,
      ),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(999),
      ),
      child: Text(
        trimmed,
        style: Theme.of(context).textTheme.labelMedium?.copyWith(
              color: color,
            ),
      ),
    );
  }
}
