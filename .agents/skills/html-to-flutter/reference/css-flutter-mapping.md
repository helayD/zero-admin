# CSS → Flutter 组件映射表

> 将 HTML 原型中的 CSS class 映射为 Flutter Widget + AMNT Design Token。
> 参考文件: `wib/ModelAi/app-prototype/app.css`

---

## 布局组件

| CSS Class | Flutter Widget | AMNT Token / 说明 |
|-----------|---------------|-------------------|
| `.app-shell` | `Scaffold` | 页面根容器 |
| `.app-header` | `AppBar` / `SliverAppBar` | height: 52, padding: `AMNTDesignTokens.spacingM` (16) |
| `.app-header-title` | `Text` | `textTheme.titleMedium` + w700, fontSize 1.15rem→18 |
| `.app-header-sub` | `Text` | `textTheme.bodySmall` + `colorScheme.onSurfaceVariant` |
| `.content` | `Padding` + `SingleChildScrollView` | `EdgeInsets.symmetric(horizontal: 16)` |
| `.g2` | `GridView(crossAxisCount: 2)` | gap: 10 (`AMNTDesignTokens.spacing10`) |
| `.g3` | `GridView(crossAxisCount: 3)` | gap: 10 |
| `.sec` | `Row(mainAxisAlignment: spaceBetween)` | margin: `EdgeInsets.only(top: 20, bottom: 10)` |
| `.sec-title` | `Text` | `textTheme.titleSmall` + w700, fontSize 0.95rem→15 |
| `.sec-link` | `TextButton` / `GestureDetector` | `colorScheme.primary` 文字色 |
| `.kanban-scroll` | `SingleChildScrollView(scrollDirection: Axis.horizontal)` | gap: 12 |
| `.kanban-col` | `SizedBox(width: 260)` | tablet: flex:1, minWidth: 300 |

---

## 卡片组件

| CSS Class | Flutter Widget | AMNT Token / 说明 |
|-----------|---------------|-------------------|
| `.card` | `Container(decoration: BoxDecoration(...))` | bg: `colorScheme.surfaceContainer`, border: `colorScheme.outline.withOpacity(0.5)`, borderRadius: 14 (`AMNTTokens.radiusM`+2 或自定义 14), padding: 16 |
| `.card:hover` | `MouseRegion` / `InkWell` | hover border: `colorScheme.primary.withOpacity(0.2)` |
| `.card-flat` | `Container` lighter bg | bg: `colorScheme.surfaceContainerLow`, border: `colorScheme.outline.withOpacity(0.3)`, borderRadius: 12 |
| `.card-action` | `InkWell` wrapping card | `onTap` handler, `InkWell` splash, active: `transform scale(0.98)` |

---

## 按钮组件

| CSS Class | Flutter Widget | AMNT Token / 说明 |
|-----------|---------------|-------------------|
| `.btn` | Base button style | padding: `EdgeInsets.symmetric(horizontal: 20, vertical: 10)`, borderRadius: 10, fontSize: 14 (`AMNTTokens.fontSizeS`), w600 |
| `.btn-primary` | `FilledButton` | bg: `colorScheme.primary`, text: white |
| `.btn-primary:hover` | FilledButton hover | bg: darker primary |
| `.btn-ghost` | `OutlinedButton` | text: `colorScheme.onSurfaceVariant`, border: `colorScheme.outline` |
| `.btn-sm` | Button + `minimumSize` | padding: `EdgeInsets.symmetric(horizontal: 14, vertical: 6)`, fontSize: 13, borderRadius: 8, height: `AMNTTokens.buttonHeightS` (32) |

---

## 标签组件 (语义色 — 允许硬编码为常量)

| CSS Class | Flutter Widget | 色值定义 |
|-----------|---------------|---------|
| `.tag` | `Container` 胶囊形 | padding: `EdgeInsets.symmetric(horizontal: 8, vertical: 2)`, borderRadius: 999, fontSize: 10.5, w600 |
| `.tag-green` | Container | bg: `Color(0xFF22C55E).withOpacity(0.12)`, text: `Color(0xFF22C55E)` |
| `.tag-blue` | Container | bg: `Color(0xFF3B82F6).withOpacity(0.12)`, text: `Color(0xFF3B82F6)` |
| `.tag-purple` | Container | bg: `Color(0xFFA855F7).withOpacity(0.12)`, text: `Color(0xFFA855F7)` |
| `.tag-amber` | Container | bg: `Color(0xFFF59E0B).withOpacity(0.12)`, text: `Color(0xFFF59E0B)` |
| `.tag-rose` | Container | bg: `Color(0xFFF43F5E).withOpacity(0.12)`, text: `Color(0xFFF43F5E)` |
| `.tag-cyan` | Container | bg: `Color(0xFF06B6D4).withOpacity(0.12)`, text: `Color(0xFF06B6D4)` |

---

## 头像组件

| CSS Class | Flutter Widget | 尺寸 |
|-----------|---------------|------|
| `.avatar` | `Container` 方形圆角 | 40x40, borderRadius: 12, centered text w700 fontSize 14 |
| `.avatar-sm` | `Container` | 32x32, borderRadius: 10, fontSize 12 |
| `.avatar-lg` | `Container` | 52x52, borderRadius: 14, fontSize 16 |

---

## 文字样式

| CSS Class | Flutter TextTheme | 补充样式 |
|-----------|------------------|---------|
| `.h1` | `textTheme.headlineLarge` | fontWeight: w800, height: 1.2 |
| `.h2` | `textTheme.titleLarge` | fontWeight: w700 |
| `.h3` | `textTheme.titleSmall` | fontWeight: w600 |
| `.body` | `textTheme.bodyMedium` | color: `colorScheme.onSurface.withOpacity(0.85)` |
| `.caption` | `textTheme.bodySmall` | color: `colorScheme.onSurfaceVariant` |
| `.overline` | `textTheme.labelSmall` | uppercase, letterSpacing: 0.06em, `colorScheme.onSurfaceVariant` |

---

## 统计数据

| CSS Class | Flutter Widget | 说明 |
|-----------|---------------|------|
| `.stat-val` | `Text` | fontSize: 24, fontWeight: w800, fontFeatures: [FontFeature.tabularFigures()], height: 1.0 |
| `.stat-label` | `Text` | fontSize: 11, color: `colorScheme.onSurfaceVariant`, marginTop: 2 |

---

## 输入组件

| CSS Class | Flutter Widget | AMNT Token / 说明 |
|-----------|---------------|-------------------|
| `.input` | `TextField` + `InputDecoration` | borderRadius: 10, bg: `colorScheme.surfaceContainerLow`, border: `colorScheme.outline`, focusBorder: `colorScheme.primary`, fontSize: 14 |
| `.search-bar` | `TextField` with prefix icon | borderRadius: 12 (`AMNTTokens.radiusM`), bg: `colorScheme.surfaceContainerLow`, icon: `Icons.search` color `colorScheme.onSurfaceVariant` |

---

## 对话气泡

| CSS Class | Flutter Widget | 说明 |
|-----------|---------------|------|
| `.chat-bubble` | `Container` | maxWidth: 85% of screen, padding: `EdgeInsets.symmetric(horizontal: 14, vertical: 10)`, fontSize: 14, height: 1.5 |
| `.chat-bubble-ai` | `Container` left-aligned | bg: `colorScheme.surfaceContainer`, border: `colorScheme.outline.withOpacity(0.4)`, borderRadius: `BorderRadius.only(topLeft: 16, topRight: 16, bottomRight: 16, bottomLeft: 4)` |
| `.chat-bubble-user` | `Container` right-aligned | bg: `colorScheme.primary`, text: white, borderRadius: `BorderRadius.only(topLeft: 16, topRight: 16, bottomRight: 4, bottomLeft: 16)` |

---

## 进度条

| CSS Class | Flutter Widget | 说明 |
|-----------|---------------|------|
| `.progress-bar` | `Container` | height: 4, bg: `colorScheme.surfaceContainerHighest`, borderRadius: 2 |
| `.progress-fill` | inner `Container` | borderRadius: 2, animated width, duration: 300ms |

---

## 列表项

| CSS Class | Flutter Widget | 说明 |
|-----------|---------------|------|
| `.list-item` | `Container` + `Row` | padding: `EdgeInsets.symmetric(vertical: 12)`, gap: 12, bottom border: `colorScheme.outline.withOpacity(0.25)`, last child no border |

---

## 工具类

| CSS Class | Flutter | 说明 |
|-----------|---------|------|
| `.mt-1` ~ `.mt-6` | `SizedBox(height: x)` | 4/8/12/16/24 |
| `.mb-2` ~ `.mb-4` | `SizedBox(height: x)` | 8/12/16 |
| `.gap-2` / `.gap-3` | `SizedBox(width/height: x)` 或 Wrap/Column gap | 8/12 |
| `.flex` | `Row` / `Flex` | — |
| `.flex-col` | `Column` | — |
| `.items-center` | `CrossAxisAlignment.center` | — |
| `.justify-between` | `MainAxisAlignment.spaceBetween` | — |
| `.truncate` | `Text(overflow: TextOverflow.ellipsis, maxLines: 1)` | — |
| `.text-center` | `TextAlign.center` | — |

---

## 响应式断点

| CSS Media Query | Flutter 判断 | 说明 |
|----------------|-------------|------|
| `@media (min-width:768px)` | `MediaQuery.sizeOf(context).width >= 768` | Tablet 断点 (`AMNTTokens.breakpointTablet`) |
| Tab bar → sidebar | `PlatformNavigationScaffold` | 自动处理 mobile/tablet/desktop |
