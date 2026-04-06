import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_mall/widgets/empty_state_widget.dart';

void main() {
  testWidgets('EmptyStateWidget 默认保持局部空态布局', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: EmptyStateWidget(
            message: '当前分类下暂无商品',
            actionText: '浏览其他分类',
            onAction: () {},
          ),
        ),
      ),
    );

    expect(find.text('当前分类下暂无商品'), findsOneWidget);
    expect(find.text('浏览其他分类'), findsOneWidget);
    expect(find.text('暂无内容'), findsNothing);
    expect(find.byType(SafeArea), findsNothing);
  });

  testWidgets('EmptyStateWidget page 模式复用 CommerceStateShell', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: EmptyStateWidget(
            message: '首页暂无推荐内容',
            actionText: '重新加载',
            onAction: () {},
            displayMode: EmptyStateDisplayMode.page,
          ),
        ),
      ),
    );

    expect(find.text('暂无内容'), findsOneWidget);
    expect(find.text('首页暂无推荐内容'), findsOneWidget);
    expect(find.text('重新加载'), findsOneWidget);
    expect(find.byType(SafeArea), findsOneWidget);
  });
}
