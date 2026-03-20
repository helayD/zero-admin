// ============================================================================
// Flutter 子组件模板 — HTML → Flutter 迁移
// 使用: 复制此模板并替换 {{}} 占位符
// 规范: AMNT Design System，所有颜色/字体从 Theme 获取
// ============================================================================

import 'package:flutter/material.dart';

/// {{WIDGET_NAME}} 组件
///
/// 对应原型 CSS class: {{CSS_CLASS}}
/// 模块: features/{{MODULE}}/presentation/widgets/{{WIDGET_FILE}}
class {{WidgetClass}} extends StatelessWidget {
  const {{WidgetClass}}({
    super.key,
    // TODO: 定义必需参数
    // required this.title,
    // required this.onTap,
  });

  // TODO: 定义属性
  // final String title;
  // final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final textTheme = Theme.of(context).textTheme;

    return Container(
      // === Card 样式 (对应 .card) ===
      decoration: BoxDecoration(
        color: colorScheme.surfaceContainer,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(
          color: colorScheme.outline.withOpacity(0.5),
        ),
      ),
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // TODO: 实现组件内容
        ],
      ),
    );
  }
}

// ============================================================================
// 常用子组件片段参考
// ============================================================================

/// === OPC 语义标签色常量 ===
/// 仅用于 .tag-green/.tag-blue 等语义标签，其他场景使用 colorScheme
class OPCTagColors {
  OPCTagColors._();
  static const Color green = Color(0xFF22C55E);
  static const Color blue = Color(0xFF3B82F6);
  static const Color purple = Color(0xFFA855F7);
  static const Color amber = Color(0xFFF59E0B);
  static const Color rose = Color(0xFFF43F5E);
  static const Color cyan = Color(0xFF06B6D4);
}

/// === 语义标签 Widget (对应 .tag.tag-xxx) ===
/// 用法: OPCTag(label: '获客', color: OPCTagColors.green)
class OPCTag extends StatelessWidget {
  const OPCTag({super.key, required this.label, required this.color});
  final String label;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
      decoration: BoxDecoration(
        color: color.withOpacity(0.12),
        borderRadius: BorderRadius.circular(999),
      ),
      child: Text(
        label,
        style: TextStyle(
          fontSize: 10.5,
          fontWeight: FontWeight.w600,
          color: color,
        ),
      ),
    );
  }
}

/// === 方形圆角头像 (对应 .avatar / .avatar-sm / .avatar-lg) ===
/// 用法: OPCAvatar(text: '获', color: Color(0xFF22C55E), size: OPCAvatarSize.md)
enum OPCAvatarSize { sm, md, lg }

class OPCAvatar extends StatelessWidget {
  const OPCAvatar({
    super.key,
    required this.text,
    required this.color,
    this.size = OPCAvatarSize.md,
  });
  final String text;
  final Color color;
  final OPCAvatarSize size;

  @override
  Widget build(BuildContext context) {
    final double dim;
    final double radius;
    final double fontSize;

    switch (size) {
      case OPCAvatarSize.sm:
        dim = 32; radius = 10; fontSize = 12;
      case OPCAvatarSize.md:
        dim = 40; radius = 12; fontSize = 14;
      case OPCAvatarSize.lg:
        dim = 52; radius = 14; fontSize = 16;
    }

    return Container(
      width: dim,
      height: dim,
      decoration: BoxDecoration(
        color: color.withOpacity(0.15),
        borderRadius: BorderRadius.circular(radius),
      ),
      alignment: Alignment.center,
      child: Text(
        text,
        style: TextStyle(
          fontSize: fontSize,
          fontWeight: FontWeight.w700,
          color: color,
        ),
      ),
    );
  }
}

/// === 统计数值 (对应 .stat-val + .stat-label) ===
/// 用法: OPCStatValue(value: '¥12,340', label: '本月收入', color: colorScheme.primary)
class OPCStatValue extends StatelessWidget {
  const OPCStatValue({
    super.key,
    required this.value,
    required this.label,
    this.valueColor,
    this.valueFontSize = 24,
  });
  final String value;
  final String label;
  final Color? valueColor;
  final double valueFontSize;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(
          value,
          style: TextStyle(
            fontSize: valueFontSize,
            fontWeight: FontWeight.w800,
            color: valueColor ?? colorScheme.onSurface,
            fontFeatures: const [FontFeature.tabularFigures()],
            height: 1.0,
          ),
        ),
        const SizedBox(height: 2),
        Text(
          label,
          style: TextStyle(
            fontSize: 11,
            color: colorScheme.onSurfaceVariant,
          ),
        ),
      ],
    );
  }
}

/// === 分区标题 (对应 .sec) ===
/// 用法: OPCSectionHeader(title: '活跃员工', trailing: '查看全部', onTrailingTap: () {})
class OPCSectionHeader extends StatelessWidget {
  const OPCSectionHeader({
    super.key,
    required this.title,
    this.trailing,
    this.onTrailingTap,
  });
  final String title;
  final String? trailing;
  final VoidCallback? onTrailingTap;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final textTheme = Theme.of(context).textTheme;

    return Padding(
      padding: const EdgeInsets.only(top: 20, bottom: 10),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(
            title,
            style: textTheme.titleSmall?.copyWith(fontWeight: FontWeight.w700),
          ),
          if (trailing != null)
            GestureDetector(
              onTap: onTrailingTap,
              child: Text(
                trailing!,
                style: textTheme.bodyMedium?.copyWith(
                  color: colorScheme.primary,
                ),
              ),
            ),
        ],
      ),
    );
  }
}

/// === 状态指示灯 (对应 工作中/离线 圆点) ===
/// 用法: OPCStatusDot(isOnline: true, label: '工作中')
class OPCStatusDot extends StatelessWidget {
  const OPCStatusDot({
    super.key,
    required this.isOnline,
    this.label,
  });
  final bool isOnline;
  final String? label;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final color = isOnline ? colorScheme.primary : colorScheme.onSurfaceVariant;

    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          width: 8,
          height: 8,
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            color: color,
          ),
        ),
        if (label != null) ...[
          const SizedBox(width: 4),
          Text(
            label!,
            style: TextStyle(
              fontSize: 12,
              color: color,
            ),
          ),
        ],
      ],
    );
  }
}
