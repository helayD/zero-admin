# 原型 CSS 变量 → AMNT 颜色映射表

> 将 HTML 原型中的 CSS 自定义属性映射为 Flutter AMNT 设计系统颜色。
>
> **源码依据**:
> - `flutter_client/lib/src/core/theme/color_schemes.dart` — ColorScheme 定义 + ColorSchemeExtension
> - `flutter_client/lib/src/core/theme/theme_extensions.dart` — AMNTCustomColors / AMNTCustomShadows / AMNTCustomSpacing
> - `flutter_client/lib/src/core/theme/brand_colors.dart` — AMNTBrandColors 品牌色常量
> - `flutter_client/lib/src/core/theme/app_theme.dart` — 主题组装（确认使用 lightScheme/darkScheme）
> - `wib/ModelAi/app-prototype/app.css` — 原型 CSS 变量定义

---

## ⚠️ 核心原则

### 1. 永远使用 colorScheme，不硬编码色值

Flutter 客户端同时支持 **Light Mode** 和 **Dark Mode**（`app_theme.dart` line 41-44 确认使用 `AMNTColorSchemes.lightScheme` / `AMNTColorSchemes.darkScheme`）。

```dart
// ✅ CORRECT — 自动适配 Light/Dark
final colorScheme = Theme.of(context).colorScheme;
Container(color: colorScheme.surfaceContainer)

// ❌ WRONG — 只在 Dark 模式下正确
Container(color: Color(0xFF161618))
```

### 2. 按语义映射，不按色值匹配

原型的 `--s700` (#1e293b) 是"卡片背景"语义 → 映射到 `colorScheme.surfaceContainer`。
不要因为 hex 值接近就映射到错误的角色。

### 3. 三层颜色 API 优先级

```dart
// 层1: ColorScheme (Material 3 标准) — 最优先
colorScheme.primary
colorScheme.surfaceContainer
colorScheme.onSurfaceVariant

// 层2: ColorSchemeExtension (color_schemes.dart 底部定义) — 语义扩展
colorScheme.success   // 通过 extension on ColorScheme 实现
colorScheme.warning
colorScheme.info

// 层3: AMNTCustomColors (theme_extensions.dart) — 交互态 + 容器色
final customColors = Theme.of(context).extension<AMNTCustomColors>()!;
customColors.hover
customColors.successContainer
customColors.warningContainer
```

---

## 完整 ColorScheme 对照表

> 数据直接从 `color_schemes.dart` 源码提取，每个值都标注了源码行号。

### 主色调

| 角色 | Light (lightScheme) | Dark (darkScheme) | 说明 |
|------|--------------------|--------------------|------|
| `primary` | `#00B36B` (L54) | `#00E676` (L112) | 品牌绿：薄荷绿/霓虹绿 |
| `onPrimary` | `#FFFFFF` (L55) | `#000000` (L113) | primary 上的文字色 |
| `primaryContainer` | `#E8FAF0` (L56) | `#00331A` (L114) | 品牌绿容器背景 |
| `onPrimaryContainer` | `#003822` (L57) | `#69F0AE` (L115) | 容器上的文字 |

### 次要色调

| 角色 | Light | Dark | 说明 |
|------|-------|------|------|
| `secondary` | `#448AFF` (L60) | `#448AFF` (L118) | 科技蓝（两模式相同） |
| `onSecondary` | `#FFFFFF` (L61) | `#FFFFFF` (L119) | — |
| `secondaryContainer` | `#448AFF` (L62) | `#448AFF` (L120) | 用户指定：科技蓝背景 |
| `onSecondaryContainer` | `#FFFFFF` (L63) | `#FFFFFF` (L121) | — |

### 第三色调

| 角色 | Light | Dark | 说明 |
|------|-------|------|------|
| `tertiary` | `#3949AB` (L66) | `#B388FF` (L124) | 靛蓝/亮紫（AI/Tech） |
| `onTertiary` | `#FFFFFF` (L67) | `#000000` (L125) | — |
| `tertiaryContainer` | `#DEE0FF` (L68) | `#451192` (L126) | — |
| `onTertiaryContainer` | `#0F1540` (L69) | `#EBDCFF` (L127) | — |

### 错误色

| 角色 | Light | Dark | 说明 |
|------|-------|------|------|
| `error` | `#BA1A1A` (L72) | `#FF897D` (L130) | 深红/柔红 |
| `onError` | `#FFFFFF` (L73) | `#000000` (L131) | — |
| `errorContainer` | `#FFDAD6` (L74) | `#93000A` (L132) | — |
| `onErrorContainer` | `#410002` (L75) | `#FFDAD6` (L133) | — |

### 表面色 (Surface 层级)

| 角色 | Light | Dark | 语义 |
|------|-------|------|------|
| `surface` | `#FFFFFF` (L78) | `#050505` (L136) | 页面主背景 |
| `onSurface` | `#1D1D1F` (L79) | `#F5F5F7` (L137) | 主文字色 |
| `surfaceContainerLowest` | `#FFFFFF` (L82) | `#000000` (L140) | 最底层 |
| `surfaceContainerLow` | `#FAFAFA` (L83) | `#101010` (L141) | 次级背景 |
| `surfaceContainer` | `#F5F7F9` (L84) | `#161618` (L142) | **卡片/容器背景** |
| `surfaceContainerHigh` | `#F0F2F5` (L85) | `#222224` (L143) | 悬浮层/弹窗 |
| `surfaceContainerHighest` | `#E8EAED` (L86) | `#2C2C2E` (L144) | 最高层级 |
| `onSurfaceVariant` | `#6B7280` (L88) | `#A1A1AA` (L146) | 辅助文字/caption |

### 轮廓色

| 角色 | Light | Dark | 语义 |
|------|-------|------|------|
| `outline` | `#E5E7EB` (L91) | `#27272A` (L149) | 主边框/分割线 |
| `outlineVariant` | `#F3F4F6` (L92) | `#3F3F46` (L150) | 辅助/弱边框 |

### 反转色

| 角色 | Light | Dark | 语义 |
|------|-------|------|------|
| `inverseSurface` | `#2C2C2E` (L99) | `#F1F3F4` (L157) | 反转背景(SnackBar等) |
| `onInverseSurface` | `#F1F3F4` (L100) | `#000000` (L158) | — |
| `inversePrimary` | `#69F0AE` (L101) | `#00A854` (L159) | — |

---

## ColorSchemeExtension 扩展色

> 定义在 `color_schemes.dart` 底部 (L306-332)，是 `extension on ColorScheme` 的扩展方法。
> 根据 `brightness` 自动选择 Light/Dark 值。

| 扩展属性 | Light 值 | Dark 值 | 来源常量 |
|---------|---------|---------|---------|
| `colorScheme.success` | `#1B8043` 谷歌绿 | `#4ADE80` 荧光绿 | `AMNTColorSchemes.successLight/Dark` (L209-210) |
| `colorScheme.warning` | `#D97706` Amber 600 | `#FCD34D` Amber 300 | `AMNTColorSchemes.warningLight/Dark` (L215-216) |
| `colorScheme.info` | `#0284C7` Sky 600 | `#38BDF8` Sky 400 | `AMNTColorSchemes.infoLight/Dark` (L221-222) |
| `colorScheme.onSuccess` | `#FFFFFF` | `#003915` | L211-212 |
| `colorScheme.onWarning` | `#FFFFFF` | `#451A03` | L217-218 |
| `colorScheme.onInfo` | `#FFFFFF` | `#082F49` | L223-224 |
| `colorScheme.primaryGradient` | `[#1976D2, #42A5F5]` | `[#90CAF9, #1976D2]` | L265-275 |
| `colorScheme.secondaryGradient` | `[#03DAC6, #4DD0E1]` | `[#80CBC4, #03DAC6]` | L278-288 |

---

## AMNTCustomColors (ThemeExtension)

> 定义在 `theme_extensions.dart`，在 `app_theme.dart` L147-151 注册为 ThemeExtension。
> 访问: `Theme.of(context).extension<AMNTCustomColors>()!`

### ⚠️ 注意：success 值与 ColorSchemeExtension 不同

| 属性 | AMNTCustomColors.light | AMNTCustomColors.dark | ColorSchemeExtension |
|------|----------------------|---------------------|---------------------|
| `success` | `#4CAF50` | `#81C784` | `#1B8043` / `#4ADE80` |
| `warning` | = warningLight `#D97706` | = warningDark `#FCD34D` | 相同 |
| `info` | = infoLight `#0284C7` | = infoDark `#38BDF8` | 相同 |

**建议**: 优先使用 `colorScheme.success/warning/info`（ColorSchemeExtension），因为更简洁且与 colorScheme 一致。
仅在需要 `Container` 版本色（如 `successContainer`、`warningContainer`）时使用 AMNTCustomColors。

### 交互态色值

| 属性 | Light | Dark |
|------|-------|------|
| `hover` | `0x14000000` (8% black) | `0x14FFFFFF` (8% white) |
| `focus` | `0x1F000000` (12% black) | `0x1FFFFFFF` (12% white) |
| `press` | `0x1F000000` (12% black) | `0x1FFFFFFF` (12% white) |
| `drag` | `0x29000000` (16% black) | `0x29FFFFFF` (16% white) |
| `selected` | `0x1F000000` (12% black) | `0x1FFFFFFF` (12% white) |

### 扩展表面色

| 属性 | Light | Dark |
|------|-------|------|
| `surfaceDim` | `#DDD8E1` | `#101014` |
| `surfaceBright` | `#FFFBFE` | `#38343A` |
| `gradientStart` | = lightScheme.primary | = darkScheme.primary |
| `gradientEnd` | = lightScheme.secondary | = darkScheme.secondary |

---

## 其他 ThemeExtension

`app_theme.dart` L147-165 还注册了:
- **`AMNTCustomSpacing.standard`** — 响应式间距
- **`AMNTCustomShadows.light/dark`** — 阴影系统
- **`AMNTChartTheme.light/dark`** — 图表配色

---

## Keling 品牌常量 (color_schemes.dart L13-44)

> 这些是**静态常量**，不随主题切换。主要用于特殊场景或工具方法。
> 日常开发应使用 `colorScheme.xxx`，而非这些常量。

| 常量 | 色值 | 用途 |
|------|------|------|
| `kelingPrimaryGreen` | `#00D26A` | 品牌柔和绿 |
| `kelingSecondaryGreen` | `#00B85C` | 品牌次要绿 |
| `kelingPureBlack` | `#000000` | 纯黑 |
| `kelingDeepBlack` | `#0D0D0D` | 侧边栏背景 |
| `kelingCardBackground` | `#1C1C1E` | 卡片背景（比 darkScheme.surfaceContainer 略亮） |
| `kelingContainerBackground` | `#2C2C2E` | 容器/hover 背景 |
| `kelingDialogBackground` | `#3A3A3C` | 弹窗背景 |
| `kelingBorder` | `#48484A` | 边框 |
| `kelingTextPrimary` | `#FFFFFF` | 主文字 |
| `kelingTextSecondary` | `#AEAEB2` | 次要文字 |
| `kelingTextTertiary` | `#8E8E93` | 第三文字 |

---

## AMNTBrandColors 核心品牌色 (brand_colors.dart)

| 常量 | 色值 | 说明 |
|------|------|------|
| `kelingMint` | `#00B36B` | = lightScheme.primary |
| `kelingNeon` | `#00E676` | = darkScheme.primary |
| `kelingBlue` | `#448AFF` | = secondary（两模式） |
| `kelingPurple` | `#B388FF` | = darkScheme.tertiary |
| `warningOrange` | `#D97706` | = warningLight |
| `errorRed` | `#BA1A1A` | = lightScheme.error |
| `accentGreen` | = kelingNeon | = darkScheme.primary |

### kelingGreenScale 色阶 (brand_colors.dart L92-104)

| 阶度 | 色值 | 对应 |
|------|------|------|
| 50 | `#E8FAF0` | = lightScheme.primaryContainer |
| 500 | `#00B36B` | = lightScheme.primary |
| 600 | `#00A05D` | 按钮按下态 |
| 950 | `#00381B` | 最深 |

---

## 原型 CSS → Flutter 映射

### 品牌色

| CSS Variable | 原型 Hex | Flutter 写法 | Light 值 | Dark 值 |
|-------------|---------|-------------|---------|---------|
| `--brand-400/500` | #4ade80/#22c55e | `colorScheme.primary` | `#00B36B` | `#00E676` |
| `--brand-50` | #f0fdf4 | `colorScheme.primaryContainer` | `#E8FAF0` | `#00331A` |

### 灰色系（按语义映射）

| CSS Variable | 原型 Hex | 语义 | Flutter 写法 |
|-------------|---------|------|-------------|
| `--s900` | #020617 | 页面主背景 | `colorScheme.surface` |
| `--s800` | #0f172a | 次级/沉降背景 | `colorScheme.surfaceContainerLow` |
| `--s700` | #1e293b | 卡片/容器 | `colorScheme.surfaceContainer` |
| — | — | 悬浮层 | `colorScheme.surfaceContainerHigh` |
| — | — | 最高层级 | `colorScheme.surfaceContainerHighest` |
| `--s600` | #334155 | 边框/分割线 | `colorScheme.outline` |
| `--s500`/`--s400` | #475569/#64748b | 辅助文字/caption | `colorScheme.onSurfaceVariant` |
| `--s300` | #94a3b8 | body描述 | `colorScheme.onSurface` + `opacity(0.7)` |
| `--s200` | #cbd5e1 | 较亮文字 | `colorScheme.onSurface` + `opacity(0.85)` |
| `--s50` | #f8fafc | 主标题文字 | `colorScheme.onSurface` |

### 强调色

| CSS Variable | 原型 Hex | 语义 | Flutter 写法 |
|-------------|---------|------|-------------|
| `--blue` | #3b82f6 | 信息/科技 | `colorScheme.secondary` |
| `--purple` | #a855f7 | AI/Tech | `colorScheme.tertiary` |
| `--amber` | #f59e0b | 警告/积分 | `colorScheme.warning` |
| `--rose` | #f43f5e | 错误/负面 | `colorScheme.error` |
| `--cyan` | #06b6d4 | 辅助链接 | 自定义常量 `Color(0xFF06B6D4)` |
| — | — | 成功/正面 | `colorScheme.success` |

### 语义标签色（唯一允许硬编码的 6 色，定义为常量）

| CSS Class | 色值 | 用途 |
|-----------|------|------|
| `.tag-green` | `#22C55E` | 成功/获客/在线 |
| `.tag-blue` | `#3B82F6` | 数据/信息 |
| `.tag-purple` | `#A855F7` | AI/内容/创意 |
| `.tag-amber` | `#F59E0B` | 运营/积分/警告 |
| `.tag-rose` | `#F43F5E` | 紧急/错误 |
| `.tag-cyan` | `#06B6D4` | 链接/辅助 |

用法:
```dart
Container(
  decoration: BoxDecoration(
    color: tagColor.withOpacity(0.12),
    borderRadius: BorderRadius.circular(999),
  ),
  child: Text(label, style: TextStyle(color: tagColor)),
)
```

### 特殊场景

| 场景 | Flutter 写法 |
|------|-------------|
| 卡片背景 | `colorScheme.surfaceContainer` |
| 卡片边框 | `colorScheme.outline.withOpacity(0.5)` |
| hover 边框 | `colorScheme.primary.withOpacity(0.2)` |
| 渐变卡片 | `LinearGradient(colors: [colorScheme.primary.withOpacity(0.06), colorScheme.tertiary.withOpacity(0.04)])` |
| 在线圆点 | `color: colorScheme.primary` |
| 离线圆点 | `color: colorScheme.onSurfaceVariant` |
| hover 交互 | `customColors.hover` |
| press 交互 | `customColors.press` |
| 反转背景(SnackBar) | `colorScheme.inverseSurface` |

---

## 使用优先级

1. **`colorScheme.xxx`** — Material 3 标准属性（自动适配 Light/Dark）
2. **`colorScheme.success/warning/info`** — ColorSchemeExtension 扩展（自动适配）
3. **`customColors.xxx`** — AMNTCustomColors（交互态 hover/press/focus + Container 色）
4. **`AMNTBrandColors.kelingGreenScale[n]`** — 品牌色阶（固定值，不随主题切换）
5. **硬编码 `Color(0xFFxxxxxx)`** — **仅用于语义标签色常量**（6 色）
