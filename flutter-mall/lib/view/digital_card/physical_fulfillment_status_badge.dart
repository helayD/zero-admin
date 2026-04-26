import 'package:flutter/material.dart';
import 'package:flutter_mall/theme/app_theme.dart';

class PhysicalFulfillmentStatusBadge extends StatelessWidget {
  final String status;
  final String text;

  const PhysicalFulfillmentStatusBadge({
    super.key,
    required this.status,
    required this.text,
  });

  Color get _color {
    switch (status) {
      case 'signed':
        return AppColors.success;
      case 'exception':
      case 'cancelled':
        return const Color(0xFFD92D20);
      case 'pending_address':
      case 'pending_real_name':
      case 'pending_digital_confirmation':
      case 'reissue_pending':
        return const Color(0xFFB54708);
      default:
        return const Color(0xFF2563EB);
    }
  }

  @override
  Widget build(BuildContext context) {
    final String label = text.trim().isEmpty ? '状态待同步' : text.trim();
    return Container(
      padding: const EdgeInsets.symmetric(
        horizontal: AppSpacing.sm,
        vertical: AppSpacing.xs,
      ),
      decoration: BoxDecoration(
        color: _color.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(999),
      ),
      child: Text(
        label,
        style: Theme.of(context).textTheme.labelMedium?.copyWith(
              color: _color,
              fontWeight: FontWeight.w700,
            ),
      ),
    );
  }
}
