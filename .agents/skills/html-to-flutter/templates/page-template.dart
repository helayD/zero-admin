// ============================================================================
// Flutter 页面骨架模板 — HTML → Flutter 迁移
// 使用: 复制此模板并替换 {{}} 占位符
// 规范: AMNT Design System + Riverpod + GoRouter
// ============================================================================

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// {{PAGE_TITLE}} 页面
/// 
/// 对应原型: wib/ModelAi/app-prototype/{{PAGE_NAME}}.html
/// 模块: features/{{MODULE}}/presentation/{{PAGE_FILE}}
class {{PageClass}}Page extends ConsumerWidget {
  const {{PageClass}}Page({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final colorScheme = Theme.of(context).colorScheme;
    final textTheme = Theme.of(context).textTheme;

    // TODO: 替换为真实的 provider
    // final asyncData = ref.watch({{module}}Provider);

    return Scaffold(
      appBar: AppBar(
        title: Text('{{PAGE_TITLE}}'),
        // 如果是二级页面，添加返回按钮（GoRouter 自动处理）
        // 如果需要右侧操作按钮:
        // actions: [
        //   IconButton(
        //     icon: const Icon(Icons.notifications_outlined),
        //     onPressed: () {},
        //   ),
        // ],
      ),
      body: SafeArea(
        child: RefreshIndicator(
          onRefresh: () async {
            // TODO: 刷新数据
            // ref.invalidate({{module}}Provider);
          },
          child: SingleChildScrollView(
            physics: const AlwaysScrollableScrollPhysics(),
            padding: const EdgeInsets.symmetric(horizontal: 16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const SizedBox(height: 8),

                // ============================
                // Section 1: {{SECTION_1_NAME}}
                // ============================
                // TODO: 从 HTML 原型提取并实现

                // ============================
                // Section 2: {{SECTION_2_NAME}}
                // ============================
                // TODO: 从 HTML 原型提取并实现

                const SizedBox(height: 24), // 底部安全间距
              ],
            ),
          ),
        ),
      ),
    );
  }

  // === 加载态 ===
  Widget _buildLoading(BuildContext context) {
    return const Center(
      child: Padding(
        padding: EdgeInsets.all(48),
        child: CircularProgressIndicator(),
      ),
    );
    // TODO: 替换为 Shimmer/Skeleton 骨架屏
  }

  // === 空状态 ===
  Widget _buildEmpty(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final textTheme = Theme.of(context).textTheme;

    return Center(
      child: Padding(
        padding: const EdgeInsets.all(48),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              Icons.inbox_outlined, // TODO: 替换为语义图标
              size: 64,
              color: colorScheme.onSurfaceVariant.withOpacity(0.5),
            ),
            const SizedBox(height: 16),
            Text(
              '暂无数据', // TODO: 替换为具体文案
              style: textTheme.titleMedium?.copyWith(
                color: colorScheme.onSurfaceVariant,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              '这里还没有内容', // TODO: 替换为引导文案
              style: textTheme.bodySmall?.copyWith(
                color: colorScheme.onSurfaceVariant.withOpacity(0.7),
              ),
            ),
            const SizedBox(height: 24),
            FilledButton.icon(
              onPressed: () {
                // TODO: CTA 操作
              },
              icon: const Icon(Icons.add, size: 18),
              label: const Text('开始'),
            ),
          ],
        ),
      ),
    );
  }

  // === 错误态 ===
  Widget _buildError(BuildContext context, Object error, StackTrace stack) {
    final colorScheme = Theme.of(context).colorScheme;
    final textTheme = Theme.of(context).textTheme;

    return Center(
      child: Padding(
        padding: const EdgeInsets.all(48),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              Icons.error_outline,
              size: 48,
              color: colorScheme.error,
            ),
            const SizedBox(height: 16),
            Text(
              '加载失败',
              style: textTheme.titleMedium?.copyWith(
                color: colorScheme.onSurface,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              error.toString(),
              style: textTheme.bodySmall?.copyWith(
                color: colorScheme.onSurfaceVariant,
              ),
              textAlign: TextAlign.center,
              maxLines: 3,
              overflow: TextOverflow.ellipsis,
            ),
            const SizedBox(height: 24),
            OutlinedButton.icon(
              onPressed: () {
                // TODO: 重试逻辑
                // ref.invalidate({{module}}Provider);
              },
              icon: const Icon(Icons.refresh, size: 18),
              label: const Text('重试'),
            ),
          ],
        ),
      ),
    );
  }
}
