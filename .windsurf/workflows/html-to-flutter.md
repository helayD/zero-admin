---
description: 将 HTML 原型页面迁移为 Flutter 页面。使用 AMNT 设计系统规范，通过 ui-ux-pro-max 技能审核 UI 质量，并更新迁移进度。
---

# HTML → Flutter 页面迁移工作流

> **Claude Skill**: `.claude/skills/html-to-flutter/SKILL.md` — 包含完整的 7 步工作流、映射表、脚本和模板
> **调用方式**: 使用 `@html-to-flutter` skill 或按以下步骤手动执行

将 `wib/ModelAi/app-prototype/` 下的 HTML 原型页面迁移为 Flutter Widget 页面，遵循现有 AMNT 设计系统，并使用 `@ui-ux-pro-max` 技能审核 UI 质量。

---

## 前置条件

- 确认目标页面的 HTML 原型文件路径：`wib/ModelAi/app-prototype/<page>.html`
- 确认目标 Flutter feature 模块路径（参考集成方案）：`flutter_client/docs/OPC-OS-INTEGRATION-PLAN.md`
- 确认迁移进度文件已存在：`flutter_client/docs/MIGRATION-PROGRESS.md`
- 参考映射表：`.claude/skills/html-to-flutter/reference/`

---

## 步骤 1：分析 HTML 原型

读取目标 HTML 原型文件，提取以下信息：

```
需要提取的信息：
  1. 页面标题和用途
  2. 页面结构（Header / Content Sections / Cards / Lists / Actions）
  3. 使用的 CSS class（对应 app.css 中的组件）
  4. 交互行为（跳转链接、按钮操作、Tab切换等）
  5. 数据模型（页面展示了哪些数据字段）
  6. 图标和颜色语义（tag-green/tag-blue/tag-purple 等）
```

读取原型文件：
```bash
cat wib/ModelAi/app-prototype/<page>.html
```

同时参考共享样式和导航：
- 样式规范：`wib/ModelAi/app-prototype/app.css`
- 导航定义：`wib/ModelAi/app-prototype/app-nav.js`

---

## 步骤 2：使用 ui-ux-pro-max 获取 Flutter 规范

// turbo
```bash
python3 .claude/skills/ui-ux-pro-max/scripts/search.py "dark mode dashboard business OS professional" --stack flutter
```

根据页面类型补充搜索：

- 如果是**数据展示页**（dashboard/briefing/finance）：
  // turbo
  ```bash
  python3 .claude/skills/ui-ux-pro-max/scripts/search.py "chart data visualization dashboard" --domain chart
  ```

- 如果是**列表管理页**（workers/crm/content/notifications）：
  // turbo
  ```bash
  python3 .claude/skills/ui-ux-pro-max/scripts/search.py "list card interaction touch" --domain ux
  ```

- 如果是**表单/配置页**（onboarding/brand/phone-os-edit）：
  // turbo
  ```bash
  python3 .claude/skills/ui-ux-pro-max/scripts/search.py "form input accessibility" --domain ux
  ```

---

## 步骤 3：映射 HTML→Flutter 组件

使用以下 **CSS→Flutter 映射表** 将原型组件转换为 AMNT 设计系统 Widget：

### 3.1 布局映射

| HTML/CSS | Flutter Widget | AMNT Token |
|----------|---------------|------------|
| `.app-shell` | `Scaffold` | — |
| `.app-header` | `SliverAppBar` / AppBar | `AMNTDesignTokens.spacing16` padding |
| `.content` | `SingleChildScrollView` + `Padding` | `EdgeInsets.symmetric(horizontal: 16)` |
| `.g2` | `GridView` 2列 | `crossAxisCount: 2, crossAxisSpacing: 10, mainAxisSpacing: 10` |
| `.g3` | `GridView` 3列 | `crossAxisCount: 3` |
| `.sec` | `Row(mainAxisAlignment: spaceBetween)` | `AMNTDesignTokens.space5` vertical margin |
| `.kanban-scroll` | `SingleChildScrollView(scrollDirection: Axis.horizontal)` | — |

### 3.2 组件映射

| HTML/CSS | Flutter Widget | AMNT 规范 |
|----------|---------------|-----------|
| `.card` | `Container` with decoration | `AMNTTokens.cardBorderRadius` (12px), `colorScheme.surfaceContainer` bg, `colorScheme.outline` border |
| `.card-flat` | `Container` lighter bg | `colorScheme.surfaceContainerLow` |
| `.card-action` | `InkWell` + card | 添加 `onTap`, splash effect |
| `.btn.btn-primary` | `FilledButton` | 使用 `colorScheme.primary` |
| `.btn.btn-ghost` | `OutlinedButton` | 使用 `colorScheme.outline` border |
| `.btn.btn-sm` | Button + `minimumSize` | `AMNTTokens.buttonHeightS` (32px) |
| `.tag.tag-green` | `Chip` / Container | bg: `Color(0xFF22C55E).withOpacity(0.12)`, text: `Color(0xFF22C55E)` |
| `.tag.tag-blue` | `Chip` / Container | bg: `Color(0xFF3B82F6).withOpacity(0.12)`, text: `Color(0xFF3B82F6)` |
| `.tag.tag-purple` | `Chip` / Container | bg: `Color(0xFFA855F7).withOpacity(0.12)`, text: `Color(0xFFA855F7)` |
| `.tag.tag-amber` | `Chip` / Container | bg: `Color(0xFFF59E0B).withOpacity(0.12)`, text: `Color(0xFFF59E0B)` |
| `.tag.tag-rose` | `Chip` / Container | bg: `Color(0xFFF43F5E).withOpacity(0.12)`, text: `Color(0xFFF43F5E)` |
| `.tag.tag-cyan` | `Chip` / Container | bg: `Color(0xFF06B6D4).withOpacity(0.12)`, text: `Color(0xFF06B6D4)` |
| `.avatar` / `.avatar-sm` / `.avatar-lg` | `Container` 方形圆角 | sm: 32x32 r10, md: 40x40 r12, lg: 52x52 r14 |
| `.search-bar` | `TextField` with decoration | `AMNTTokens.radiusM` border, `colorScheme.surfaceContainerLow` fill |
| `.stat-val` | `Text` | `fontSize: 24, fontWeight: w800, fontVariantNumeric: tabular` |
| `.stat-label` | `Text` | `fontSize: 11, color: colorScheme.onSurfaceVariant` |
| `.list-item` | `ListTile` / custom Row | Bottom border: `colorScheme.outline.withOpacity(0.25)` |
| `.progress-bar` | `LinearProgressIndicator` / Container | height: 4, `colorScheme.surfaceContainerHighest` track |
| `.chat-bubble-ai` | `Container` left-aligned | `colorScheme.surfaceContainer` bg, borderRadius: `16,16,16,4` |
| `.chat-bubble-user` | `Container` right-aligned | `colorScheme.primary` bg, borderRadius: `16,16,4,16` |
| `.h1` | `Text` | `Theme.of(context).textTheme.headlineLarge` + w800 |
| `.h2` | `Text` | `Theme.of(context).textTheme.titleLarge` + w700 |
| `.h3` | `Text` | `Theme.of(context).textTheme.titleSmall` + w600 |
| `.body` | `Text` | `Theme.of(context).textTheme.bodyMedium` |
| `.caption` | `Text` | `Theme.of(context).textTheme.bodySmall` + `colorScheme.onSurfaceVariant` |
| `.overline` | `Text` | `Theme.of(context).textTheme.labelSmall` + uppercase + letterSpacing 0.06em |

### 3.3 颜色语义映射（原型 CSS var → AMNT）

| CSS Variable | 用途 | Flutter AMNT |
|-------------|------|-------------|
| `--brand-400` (#4ade80) | 品牌强调/正面 | `colorScheme.primary` |
| `--brand-500` (#22c55e) | 品牌主色 | `colorScheme.primary` |
| `--s900` (#020617) | 主背景 | `colorScheme.surface` |
| `--s800` (#0f172a) | 次级背景 | `colorScheme.surfaceContainer` |
| `--s700` (#1e293b) | 卡片背景 | `colorScheme.surfaceContainerHigh` |
| `--s600` (#334155) | 边框 | `colorScheme.outline` |
| `--s500` (#475569) | 次要文字/inactive tab | `colorScheme.onSurfaceVariant` |
| `--s400` (#64748b) | caption文字 | `colorScheme.onSurfaceVariant` |
| `--s300` (#94a3b8) | body文字 | `colorScheme.onSurface.withOpacity(0.7)` |
| `--s200` (#cbd5e1) | 亮文字 | `colorScheme.onSurface.withOpacity(0.85)` |
| `--s50` (#f8fafc) | 主文字 | `colorScheme.onSurface` |
| `--blue` (#3b82f6) | 信息/数据 | `colorScheme.secondary` 或自定义 |
| `--purple` (#a855f7) | AI/Tech | `colorScheme.tertiary` |
| `--amber` (#f59e0b) | 警告/积分 | `AMNTCustomColors.warning` |
| `--rose` (#f43f5e) | 错误/负面 | `colorScheme.error` |
| `--cyan` (#06b6d4) | 辅助信息 | 自定义 `Color(0xFF06B6D4)` |

---

## 步骤 4：创建 Flutter 页面文件

按照以下目录结构创建文件：

```
lib/src/features/<module>/
├── data/
│   ├── models/           # 数据模型（从HTML数据字段提取）
│   └── <module>_repository.dart  # 数据仓库（先用Mock）
├── presentation/
│   ├── <page>_page.dart  # 主页面（StatelessWidget + ConsumerWidget）
│   └── widgets/          # 页面内子组件
└── state/
    └── <module>_providers.dart  # Riverpod providers
```

### 4.1 页面骨架模板

```dart
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

class XxxPage extends ConsumerWidget {
  const XxxPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final colorScheme = Theme.of(context).colorScheme;
    final textTheme = Theme.of(context).textTheme;

    return Scaffold(
      appBar: AppBar(
        title: Text('页面标题'),
      ),
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.symmetric(horizontal: 16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // === Section 1: 从HTML提取 ===
              // === Section 2: ... ===
            ],
          ),
        ),
      ),
    );
  }
}
```

### 4.2 编码规范要求

1. **🚨 绝对禁止硬编码配色** — Flutter 同时支持 Light/Dark 双模式，硬编码会导致某模式 UI 破损：
   - ✅ `colorScheme.surfaceContainer` / `colorScheme.primary` / `colorScheme.warning`
   - ✅ `textTheme.bodyMedium` / `textTheme.titleSmall`
   - ❌ `Color(0xFF161618)` / `Color(0xFF1E293B)` — 硬编码 Dark 色值
   - ❌ `Colors.white` / `Colors.grey` — 不跟随主题
   - ❌ `TextStyle(fontSize: 14, color: Colors.white)` — 硬编码字号和颜色
   - 唯一例外：6 种语义标签色常量（OPCTagColors）
   - **检查方法**: 搜索代码中 `Color(0x` 和 `Colors.`，除标签色常量外不应出现

2. **颜色 API 优先级**（详见 `.claude/skills/html-to-flutter/reference/color-mapping.md`）：
   - 层1: `colorScheme.xxx` — Material 3 标准属性
   - 层2: `colorScheme.success/warning/info` — ColorSchemeExtension 扩展
   - 层3: `Theme.of(context).extension<AMNTCustomColors>()!.xxx` — 交互态/容器色
   - 层4: `AMNTBrandColors.kelingGreenScale[n]` — 品牌色阶（固定值，慎用）

3. **状态管理**：
   - 使用 Riverpod `ConsumerWidget` / `ConsumerStatefulWidget`
   - Provider 定义在 `state/` 目录
   - 异步数据用 `AsyncValue` + `when()` 模式

4. **响应式**：
   - 使用 `AMNTTokens.breakpointTablet` (768) 判断
   - 手机：单列 / 平板：双列 / 桌面：三列+侧边栏

5. **无障碍**：
   - 所有可点击元素 minimum touch target 44x44
   - 图片/图标提供 `semanticLabel`
   - 色彩对比度 ≥ 4.5:1

6. **动画**：
   - 微交互 150-300ms (`AMNTTokens.animationDurationFast` / `animationDurationNormal`)
   - 页面切换使用 GoRouter 默认过渡
   - 尊重 `MediaQuery.disableAnimations`

---

## 步骤 5：UI 质量审核

页面完成后，使用 `@ui-ux-pro-max` 进行审核检查：

// turbo
```bash
python3 .claude/skills/ui-ux-pro-max/scripts/search.py "accessibility touch interaction" --domain ux
```

### 5.1 自检清单

**视觉质量**：
- [ ] 不使用 emoji 作为图标（使用 `Icons.xxx` 或 SVG）
- [ ] 所有图标来自统一图标集
- [ ] hover/press 状态不导致布局抖动
- [ ] 颜色全部使用 `colorScheme` / `AMNTCustomColors` / 语义标签色

**交互**：
- [ ] 所有可点击元素有视觉反馈（InkWell/splash）
- [ ] 过渡动画 150-300ms
- [ ] 键盘导航可用（Focus可见）
- [ ] 触摸目标 ≥ 44x44

**深色模式**：
- [ ] 文字对比度足够（onSurface vs surface）
- [ ] 卡片/容器在深色下可辨识
- [ ] 边框在深色下可见

**布局**：
- [ ] 无内容被 AppBar/BottomNav 遮挡
- [ ] 375px / 768px / 1024px 三个断点正常
- [ ] 无水平溢出
- [ ] 长列表使用 `ListView.builder` 或 `SliverList`

**数据状态**：
- [ ] 加载态（Shimmer/Skeleton）
- [ ] 空状态（引导性空页面）
- [ ] 错误态（重试按钮）
- [ ] 下拉刷新（如适用）

---

## 步骤 6：注册路由

在 `flutter_client/lib/src/core/navigation/app_routes.dart` 中注册新路由：

1. 添加路由常量定义
2. 在 GoRouter 配置中添加 `GoRoute`
3. 如果是 Tab 级页面，更新 `MainNavigationDestinations`

---

## 步骤 7：更新迁移进度

**只有通过步骤 5 全部审核 + Step 1 生成的验收清单全部勾选后，才能标记 completed。**

```bash
# 开始迁移
python3 .claude/skills/html-to-flutter/scripts/update-progress.py <page> in_progress
# 通过全部检查
python3 .claude/skills/html-to-flutter/scripts/update-progress.py <page> completed
```

完成度标准（100% 还原，不是框架）：
- **区块完整性** (20%): 原型中每个 Section/Card/List 都已实现
- **数据字段** (15%): 所有文字/数值/标签/状态指示都有 Widget 渲染（Mock 数据）
- **视觉还原** (20%): 布局/间距/圆角/字体层级/颜色语义与原型一致
- **交互行为** (15%): 所有 onclick/href 跳转 + InkWell 反馈
- **三态完整** (10%): 加载态/空状态/错误态
- **双模式兼容** (10%): Light + Dark 均正确，零硬编码
- **响应式** (10%): 375/768/1024 断点正常

状态含义：
- `⬜` 未开始 (0%)
- `🔄` 进行中 (20%-80%)
- `✅` 已完成 (100% — 全部 7 项通过)
- `⏸️` 暂停（被阻塞）

---

## 步骤 8：更新 Epic 文件

每完成一个页面迁移，更新 Epic 文件：

**Epic 路径**: `.claude/epics/OPC_OS_HTML_to_Flutter_Migration/epic.md`

更新内容：
1. 对应 Task 的状态改为 `已完成`
2. 补充创建文件列表
3. 更新统计区的已完成数量和完成率
4. 更新 frontmatter 的 `updated` 时间和 `progress` 百分比（`已完成数/37*100`）

---

## 快速参考：原型→模块对照表

| HTML 原型 | Flutter 模块 | 优先级 |
|----------|-------------|--------|
| home.html | features/home/ | P0 |
| workers.html | features/ai_workers/ | P0 |
| worker-detail.html | features/ai_workers/ | P0 |
| chat.html | features/ai_workers/ | P0 |
| briefing.html | features/briefing/ | P0 |
| dashboard.html | features/dashboard/ | P0 |
| onboarding.html | features/onboarding/ | P0 |
| track-assessment.html | features/onboarding/ | P0 |
| notifications.html | features/notifications/ | P1 |
| content.html | features/content/ | P1 |
| content-review.html | features/content/ | P1 |
| crm.html | features/crm/ | P1 |
| projects.html | features/projects/ | P1 |
| profile.html | features/profile/ | P1 |
| portfolio.html | features/profile/ | P1 |
| credits.html | features/finance/ | P1 |
| subscription.html | features/subscription/ | P1 |
| demands.html | features/marketplace/ | P1 |
| market.html | features/marketplace/ | P1 |
| enterprise.html | features/marketplace/ | P1 |
| finance.html | features/finance/ | P2 |
| escrow.html | features/finance/ | P2 |
| worker-collaboration.html | features/ai_workers/ | P2 |
| worker-memory.html | features/ai_workers/ | P2 |
| automation.html | features/automation/ | P2 |
| coze-workflows.html | features/automation/ | P2 |
| comfyui-workflows.html | features/automation/ | P2 |
| mcp-tools.html | features/automation/ | P2 |
| knowledge.html | features/knowledge/ | P2 |
| messaging-channels.html | features/messaging_channels/ | P2 |
| brand.html | features/profile/ | P2 |
| phone-os.html | features/phone_os/ | P3 |
| phone-os-edit.html | features/phone_os/ | P3 |
| phone-tasks.html | features/phone_os/ | P3 |
| credit-score.html | features/credit_score/ | P3 |
| ai-quality.html | features/ai_quality/ | P3 |
| comply.html | features/ai_quality/ | P3 |
