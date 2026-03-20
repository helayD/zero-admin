---
name: html-to-flutter
description: "将 HTML 原型页面迁移为 Flutter 页面。使用 AMNT 设计系统规范，通过 ui-ux-pro-max 技能审核 UI 质量，并更新迁移进度。"
allowed-tools: Bash, Read, Write, Grep, Glob, Edit, LS
argument-hint: "[page_name] (e.g. home, workers, briefing, dashboard)"
disable-model-invocation: true
---

# HTML → Flutter 页面迁移

将 `wib/ModelAi/app-prototype/` 下的 HTML 原型页面迁移为 Flutter Widget 页面，遵循 AMNT 设计系统，使用 `ui-ux-pro-max` 技能审核 UI 质量。

## When to Apply

当需要：
- 将 HTML 原型页面转换为 Flutter 页面
- 新增 OPC-OS 功能模块的 UI 页面
- 用户说"迁移页面"、"migrate page"、"html to flutter"

## Scripts

| Script | Purpose |
|--------|---------|
| `preflight-check.py` | 迁移预检：验证 HTML 文件存在、目标模块路径、依赖完整性 |
| `update-progress.py` | 更新 MIGRATION-PROGRESS.md 中的页面状态和统计 |

## Templates

| Template | Purpose |
|----------|---------|
| `page-template.dart` | Flutter 页面骨架模板（ConsumerWidget + AMNT theme） |
| `widget-template.dart` | Flutter 子组件模板（可复用卡片/列表/统计等） |

## Reference Files

| File | Purpose |
|------|---------|
| `reference/css-flutter-mapping.md` | HTML/CSS class → Flutter Widget + AMNT Token 映射表 |
| `reference/color-mapping.md` | 原型 CSS 变量 → AMNT colorScheme 映射表 |
| `reference/page-module-mapping.md` | 37 个 HTML 页面 → Flutter feature 模块 + 优先级映射 |

## Additional Resources

- 集成方案：`flutter_client/docs/OPC-OS-INTEGRATION-PLAN.md`
- 迁移进度：`flutter_client/docs/MIGRATION-PROGRESS.md`
- AMNT 设计系统：`flutter_client/lib/src/core/theme/`
- 原型样式：`wib/ModelAi/app-prototype/app.css`
- 原型导航：`wib/ModelAi/app-prototype/app-nav.js`

---

## ⚠️ CRITICAL REQUIREMENTS

### 🚨 绝对禁止硬编码配色 (ZERO HARDCODED COLORS)

Flutter 客户端同时支持 **Light Mode** 和 **Dark Mode**，硬编码色值会导致某个模式下 UI 破损。

```dart
// ✅ CORRECT — 通过 Theme 获取，Light/Dark 自动适配
final colorScheme = Theme.of(context).colorScheme;
Container(color: colorScheme.surfaceContainer)     // 卡片背景
Text('text', style: textTheme.bodyMedium)           // 文字样式
Icon(Icons.check, color: colorScheme.primary)       // 图标颜色
Divider(color: colorScheme.outline)                 // 分割线

// ❌ WRONG — 硬编码色值，Light 模式下会看不见/丢失对比度
Container(color: Color(0xFF161618))                 // Dark 专用色
Text('text', style: TextStyle(color: Colors.white)) // Dark 专用
Icon(Icons.check, color: Color(0xFF00E676))         // Dark 专用绿
Divider(color: Color(0xFF27272A))                   // Dark 专用边框

// ❌ WRONG — 使用 Colors.xxx 常量而不是 colorScheme
Container(color: Colors.black)                      // 不跟随主题
Text('text', style: TextStyle(color: Colors.grey))   // 不跟随主题
```

**唯一允许硬编码的例外**: 6 种语义标签色（tag-green/blue/purple/amber/rose/cyan），必须定义为常量。

**颜色 API 优先级**（详见 `reference/color-mapping.md`）:
1. `colorScheme.xxx` — Material 3 标准属性
2. `colorScheme.success/warning/info` — ColorSchemeExtension 扩展
3. `Theme.of(context).extension<AMNTCustomColors>()!.xxx` — 交互态/容器色
4. `AMNTBrandColors.kelingGreenScale[n]` — 品牌色阶（固定值，慎用）

### 其他强制要求

- **中文输出**: 所有日志和进度更新使用中文
- **ui-ux-pro-max 审核**: 每个页面完成后必须通过 UI 质量检查
- **进度追踪**: 每完成一个页面必须更新 MIGRATION-PROGRESS.md
- **No Partial Implementation**: 不做半成品页面，每个页面必须包含完整的加载态/空态/错误态

---

## 📐 页面完成度定义 (100% 还原标准)

> **迁移不是只搭框架**。每个页面必须 100% 还原原型的所有视觉元素、数据字段、交互和状态。
> 只有通过以下全部 7 项检查，才能标记为 `✅ completed`。

### 完成度 7 项检查

| # | 检查维度 | 权重 | 具体要求 |
|---|---------|------|---------|
| 1 | **区块完整性** | 20% | 原型中的每个 Section/Card/List 都已实现为对应 Widget，无遗漏 |
| 2 | **数据字段** | 15% | 所有文字、数值、标签、状态指示都有对应 Widget 渲染（使用 Mock 数据） |
| 3 | **视觉还原** | 20% | 布局结构、间距、圆角、字体层级、颜色语义与原型一致（通过 colorScheme） |
| 4 | **交互行为** | 15% | 所有 onclick/href 跳转已用 GoRouter 实现，按钮/卡片有 InkWell/splash |
| 5 | **三态完整** | 10% | 加载态（Shimmer）、空状态（引导页）、错误态（重试）全部实现 |
| 6 | **双模式兼容** | 10% | Light Mode + Dark Mode 下视觉均正确，无硬编码色值 |
| 7 | **响应式** | 10% | 375px(手机) / 768px(平板) / 1024px(桌面) 断点正常 |

### 每个页面的验收清单模板

在 Step 1 分析完原型后，生成该页面的**具体验收清单**：

```markdown
## home.html 验收清单

### 区块完整性 (所有区块必须实现)
- [ ] Header: 问候语 + 通知铃铛(红点) + 用户头像
- [ ] 今日团队简报卡片: 渐变背景 + 4条简报条目(彩色圆点)
- [ ] 快速统计: 3列网格(本月收入/活跃客户/AI任务完成)
- [ ] 积分&收入: 2列(积分余额+进度条 / 收入趋势迷你柱状图)
- [ ] OPC-Phone 状态卡片: 手机图标 + 在线状态 + 任务进度
- [ ] AI团队状态: 水平滚动5张员工卡(头像+姓名+职位+状态灯+当前任务)
- [ ] 最近任务列表: 4条(图标+标题+描述+状态标签)
- [ ] 待办提醒列表: 2条(图标+标题+描述+紧急度标签)
- [ ] 快捷操作网格: 2x2(手机)/4列(平板)

### 数据字段 (全部使用 Mock 数据渲染)
- [ ] 用户名、时间问候语
- [ ] 统计数值(¥8.2万/23/156)及数值颜色语义
- [ ] 积分(3,400/10,000)及进度条比例
- [ ] 迷你柱状图(6根柱子高度比例)
- [ ] 员工状态(工作中/待审核/在线/待命/执行中)
- [ ] 任务进度(步骤6/10)
- [ ] 时间标签(2分钟前/15分钟前/今天14:00)

### 交互行为
- [ ] 简报卡片 → briefing.html (GoRouter)
- [ ] 积分卡片 → credits.html
- [ ] 收入卡片 → finance.html
- [ ] Phone卡片 → phone-os.html
- [ ] 员工卡片 → chat.html
- [ ] "管理 →" → workers.html
- [ ] 通知铃铛 → notifications.html
- [ ] 审核内容 → content-review.html
- [ ] 快捷操作4按钮 → 各自目标页
- [ ] 所有可点击元素有 InkWell/splash 反馈

### 三态
- [ ] 加载态: 骨架屏(统计卡/员工卡/任务列表)
- [ ] 空状态: 无AI员工时的引导页
- [ ] 错误态: 数据加载失败+重试

### 双模式 + 响应式
- [ ] Light Mode 下所有元素可见且美观
- [ ] Dark Mode 下所有元素可见且美观
- [ ] 375px 单列正常
- [ ] 768px+ 快捷操作变4列
```

### 完成度计算

- **0%** = 未开始
- **20%** = 文件已创建，基本骨架（仅 Scaffold + AppBar）
- **40%** = 主要区块 Widget 已实现，但缺少数据/交互
- **60%** = 所有区块 + 数据字段已实现，交互部分完成
- **80%** = 全部区块 + 数据 + 交互 + 三态，但未通过双模式/审核
- **100%** = 通过全部 7 项检查 + Step 5 UI 审核清单

**只有 100% 才能标记为 completed。80% 以下标记为 in_progress。**

---

## Required Rules

执行前必须读取：
1. `.Codex/rules/datetime.md` — ISO 8601 时间戳（如存在）
2. 参考文件 `reference/css-flutter-mapping.md` — 组件映射参考
3. 参考文件 `reference/color-mapping.md` — 颜色映射参考

---

## 7-Step Workflow

### Step 0: Preflight Check

```bash
python3 .Codex/skills/html-to-flutter/scripts/preflight-check.py $ARGUMENTS
```

脚本验证：
- HTML 原型文件是否存在：`wib/ModelAi/app-prototype/$ARGUMENTS.html`
- 目标 Flutter 模块路径（从 `reference/page-module-mapping.md` 查找）
- AMNT 设计系统文件是否完整
- 依赖包是否已安装

如果脚本失败，报告错误并终止。

---

### Step 1: 分析 HTML 原型

读取 HTML 原型文件，提取以下信息并记录：

**必须提取**:
1. **页面标题和用途** — `<title>` 和 header 内容
2. **页面结构** — 所有 section/card/list/grid 的层级关系
3. **CSS class 列表** — 使用了哪些 `app.css` 中的 class
4. **交互行为** — 跳转链接(`href`)、按钮操作、Tab 切换、搜索
5. **数据模型** — 页面展示的所有数据字段（名称、类型、来源）
6. **颜色语义** — 使用的 tag 颜色、状态颜色、强调色

**同时读取共享资源**:
- `wib/ModelAi/app-prototype/app.css` — 样式定义
- `wib/ModelAi/app-prototype/app-nav.js` — 导航结构

**输出格式**（包含分析结果 + 该页面的验收清单）:
```
## HTML 分析: $ARGUMENTS.html
### 页面用途: ...
### 结构树:
  - Header: ...
  - Section 1: ...
    - Card: ...
  - Section 2: ...
### CSS Classes: [card, g3, tag-green, stat-val, ...]
### 交互: [跳转→briefing.html, 按钮→添加员工, ...]
### 数据字段: [name:string, status:enum, score:number, ...]
### 颜色语义: [green→成功/在线, blue→信息/数据, amber→警告/积分, ...]

## $ARGUMENTS.html 验收清单 (必须全部通过才能标记 completed)

### 区块完整性 (X 个区块)
- [ ] [列出原型中每个 Section/Card/List]
- [ ] ...

### 数据字段 (X 个字段)
- [ ] [列出所有需要渲染的数据字段和 Mock 值]
- [ ] ...

### 交互行为 (X 个交互)
- [ ] [列出所有 onclick/href 跳转目标]
- [ ] [列出所有按钮/Tab/搜索操作]
- [ ] 所有可点击元素有 InkWell/splash 反馈

### 三态
- [ ] 加载态: [描述骨架屏覆盖哪些区块]
- [ ] 空状态: [描述空态场景和引导文案]
- [ ] 错误态: 错误信息 + 重试按钮

### 双模式 + 响应式
- [ ] Light Mode 所有元素可见且美观
- [ ] Dark Mode 所有元素可见且美观
- [ ] 零硬编码色值（仅 OPCTagColors 常量除外）
- [ ] 375px / 768px / 1024px 断点正常
```

**⚠️ 此验收清单将在 Step 7 用于逐项确认完成度。不通过则不能标记 completed。**

---

### Step 2: CSS→Flutter 组件映射

参考 `reference/css-flutter-mapping.md` 和 `reference/color-mapping.md`，将 Step 1 提取的每个 CSS class 映射为具体的 Flutter Widget + AMNT Token。

**核心映射规则**（完整列表见 reference/css-flutter-mapping.md）:

| 类别 | HTML/CSS | Flutter | AMNT |
|------|----------|---------|------|
| 布局 | `.content` | `Padding(horizontal: 16)` | `AMNTDesignTokens.spacingM` |
| 布局 | `.g2` / `.g3` | `GridView(crossAxisCount: 2/3)` | gap: `AMNTDesignTokens.spacing10` |
| 卡片 | `.card` | `Container(decoration: ...)` | `AMNTTokens.cardBorderRadius`, `colorScheme.surfaceContainer` |
| 按钮 | `.btn-primary` | `FilledButton` | `colorScheme.primary` |
| 按钮 | `.btn-ghost` | `OutlinedButton` | `colorScheme.outline` |
| 标签 | `.tag.tag-green` | `Container` 圆角胶囊 | bg: `Color(0xFF22C55E).withOpacity(0.12)` |
| 头像 | `.avatar` / `.avatar-lg` | `Container` 方形圆角 | md:40x40 r12, lg:52x52 r14 |
| 文字 | `.h1` | `textTheme.headlineLarge` + w800 | — |
| 文字 | `.h3` | `textTheme.titleSmall` + w600 | — |
| 文字 | `.caption` | `textTheme.bodySmall` | `colorScheme.onSurfaceVariant` |
| 统计 | `.stat-val` | `Text(fontSize: 24, w800)` | tabular figures |
| 搜索 | `.search-bar` | `TextField` | `AMNTTokens.radiusM`, `colorScheme.surfaceContainerLow` |
| 气泡 | `.chat-bubble-ai` | `Container` | borderRadius: `16,16,16,4` |

**颜色映射**（完整列表见 reference/color-mapping.md）:

| CSS Variable | Flutter AMNT |
|-------------|-------------|
| `--brand-400/500` | `colorScheme.primary` |
| `--s900` | `colorScheme.surface` |
| `--s800`/`--s700` | `colorScheme.surfaceContainerLow` / `surfaceContainer` |
| `--s600` | `colorScheme.outline` |
| `--s400` | `colorScheme.onSurfaceVariant` |
| `--s50` | `colorScheme.onSurface` |
| `--blue` | `colorScheme.secondary` |
| `--purple` | `colorScheme.tertiary` |
| `--amber` | `colorScheme.warning` |
| `--rose` | `colorScheme.error` |

---

### Step 3: 使用 ui-ux-pro-max 获取规范

根据页面类型调用 ui-ux-pro-max：

**所有页面必须执行**:
```bash
python3 .Codex/skills/ui-ux-pro-max/scripts/search.py "dark mode professional dashboard" --stack flutter
```

**数据展示页**（dashboard/briefing/finance/credits）补充：
```bash
python3 .Codex/skills/ui-ux-pro-max/scripts/search.py "chart data visualization" --domain chart
```

**列表管理页**（workers/crm/content/notifications）补充：
```bash
python3 .Codex/skills/ui-ux-pro-max/scripts/search.py "list card interaction touch target" --domain ux
```

**表单配置页**（onboarding/brand/phone-os-edit）补充：
```bash
python3 .Codex/skills/ui-ux-pro-max/scripts/search.py "form input validation accessibility" --domain ux
```

**对话页**（chat）补充：
```bash
python3 .Codex/skills/ui-ux-pro-max/scripts/search.py "chat messaging real-time" --domain ux
```

将获得的规范建议记录，在 Step 4 实现时遵守。

---

### Step 4: 创建 Flutter 页面

按以下结构创建文件：

```
flutter_client/lib/src/features/<module>/
├── data/
│   ├── models/                    # 数据模型
│   │   └── <model>.dart
│   └── <module>_repository.dart   # 数据仓库（先 Mock）
├── presentation/
│   ├── <page>_page.dart           # 主页面
│   └── widgets/                   # 子组件
│       ├── <widget_1>.dart
│       └── <widget_2>.dart
└── state/
    └── <module>_providers.dart    # Riverpod providers
```

**编码规范（强制）**:

1. **🚨 绝对禁止硬编码配色** — 所有颜色必须来自 Theme：
   ```dart
   // ✅ CORRECT — 通过 colorScheme 获取颜色
   final colorScheme = Theme.of(context).colorScheme;
   final textTheme = Theme.of(context).textTheme;
   Container(color: colorScheme.surfaceContainer)   // 卡片背景
   Text('标题', style: textTheme.titleSmall?.copyWith(fontWeight: FontWeight.w600))
   Icon(Icons.star, color: colorScheme.warning)      // 语义色用 extension
   Divider(color: colorScheme.outline)               // 边框用 outline
   Container(color: colorScheme.primaryContainer)    // 品牌色容器

   // ❌ WRONG — 这些写法会导致 Light/Dark 切换时 UI 破损
   Container(color: Color(0xFF161618))               // 硬编码 Dark 背景色
   Container(color: Color(0xFF1E293B))               // 硬编码原型 CSS 色值
   Text('标题', style: TextStyle(color: Colors.white))  // 硬编码白色
   Text('标题', style: TextStyle(fontSize: 14))         // 硬编码字号
   Icon(Icons.star, color: Color(0xFFF59E0B))        // 硬编码警告色
   Divider(color: Colors.grey)                       // 用 Colors.xxx 而非 colorScheme
   ```
   
   **检查方法**: 搜索代码中的 `Color(0x` 和 `Colors.`，除了标签色常量外不应出现。

2. **状态管理 — Riverpod**:
   ```dart
   // 页面使用 ConsumerWidget
   class XxxPage extends ConsumerWidget { ... }
   // 异步数据使用 AsyncValue
   ref.watch(xxxProvider).when(data: ..., loading: ..., error: ...)
   ```

3. **数据状态三态**:
   - **加载态**: Shimmer/Skeleton 骨架屏
   - **空状态**: 引导性空页面（图标 + 文字 + CTA）
   - **错误态**: 错误信息 + 重试按钮

4. **响应式布局**:
   ```dart
   // 使用 AMNT 断点
   final isTablet = MediaQuery.sizeOf(context).width >= 768;
   // 手机: 单列, 平板: 双列, 桌面: 三列+侧边栏
   ```

5. **无障碍**:
   - 触摸目标 ≥ 44x44
   - 图标提供 `semanticLabel`
   - 色彩对比度 ≥ 4.5:1

6. **动画**:
   - 微交互 150-300ms (`AMNTTokens.animationDurationFast`)
   - 尊重 `MediaQuery.disableAnimations`

7. **语义标签色**（唯一允许硬编码的 6 种色值，必须定义为常量）:
   ```dart
   // 在共享常量文件中定义（见 templates/widget-template.dart 中的 OPCTagColors）
   // 这是整个项目中唯一允许 Color(0xFFxxxxxx) 硬编码的地方
   class OPCTagColors {
     OPCTagColors._();
     static const Color green  = Color(0xFF22C55E); // 成功/获客/在线
     static const Color blue   = Color(0xFF3B82F6); // 数据/信息
     static const Color purple = Color(0xFFA855F7); // AI/内容/创意
     static const Color amber  = Color(0xFFF59E0B); // 运营/积分/警告
     static const Color rose   = Color(0xFFF43F5E); // 紧急/错误
     static const Color cyan   = Color(0xFF06B6D4); // 链接/辅助
   }
   // 其他任何场景都必须使用 colorScheme.xxx
   ```

**使用模板**:
- 页面骨架: `templates/page-template.dart`
- 子组件: `templates/widget-template.dart`

---

### Step 5: UI 质量审核

页面完成后执行审核：

```bash
python3 .Codex/skills/ui-ux-pro-max/scripts/search.py "accessibility touch interaction loading" --domain ux
```

**自检清单（全部通过才算完成）**:

**🚨 硬编码配色检查（最高优先级）**:
- [ ] 搜索 `Color(0x` — 除 OPCTagColors 常量定义外，不应出现任何硬编码色值
- [ ] 搜索 `Colors.` — 不应使用 `Colors.white`/`Colors.black`/`Colors.grey` 等，改用 `colorScheme.onSurface`/`colorScheme.surface`/`colorScheme.onSurfaceVariant`
- [ ] 搜索 `TextStyle(` — 不应包含硬编码 `color:` 或 `fontSize:`，改用 `textTheme.xxx.copyWith()`
- [ ] 切换到 Light Mode 检查 — 所有卡片/文字/边框/图标在白色背景下可读且美观
- [ ] 切换到 Dark Mode 检查 — 所有卡片/文字/边框/图标在深色背景下可读且美观

**视觉质量**:
- [ ] 不使用 emoji 作为图标（使用 `Icons.xxx` 或自定义 SVG）
- [ ] hover/press 状态不导致布局抖动

**交互**:
- [ ] 所有可点击元素有视觉反馈（InkWell / splash）
- [ ] 过渡动画 150-300ms
- [ ] 触摸目标 ≥ 44x44
- [ ] 长列表使用 `ListView.builder` / `SliverList`

**双模式兼容**:
- [ ] Light Mode: 文字对比度 ≥ 4.5:1（onSurface vs surface）
- [ ] Dark Mode: 文字对比度 ≥ 4.5:1
- [ ] 卡片层级在两种模式下都可辨识（surface → surfaceContainer → surfaceContainerHigh）
- [ ] 边框在两种模式下都可见（outline / outlineVariant）

**布局**:
- [ ] 无内容被 AppBar/BottomNav 遮挡（SafeArea）
- [ ] 375px / 768px / 1024px 三断点正常
- [ ] 无水平溢出

**数据状态**:
- [ ] 加载态（Shimmer/Skeleton）
- [ ] 空状态（引导性空页面）
- [ ] 错误态（重试按钮）

---

### Step 6: 注册路由

在 `flutter_client/lib/src/core/navigation/app_routes.dart` 中：

1. 添加路由常量（如 `static const String aiWorkers = '/ai-workers';`）
2. 在 GoRouter 配置中添加 `GoRoute` 注册
3. 如果是 Tab 级页面，更新 `MainNavigationDestinations`

---

### Step 7: 更新迁移进度

**只有通过 Step 5 全部审核 + 页面完成度 7 项检查后，才能标记 completed。**

```bash
# 开始迁移时标记为 in_progress
python3 .Codex/skills/html-to-flutter/scripts/update-progress.py $ARGUMENTS in_progress

# 通过全部检查后标记为 completed
python3 .Codex/skills/html-to-flutter/scripts/update-progress.py $ARGUMENTS completed
```

脚本自动更新 `flutter_client/docs/MIGRATION-PROGRESS.md`：
- 将对应页面状态图标改为 `🔄` 或 `✅`
- 更新总览统计数字
- 追加变更日志条目

**完成时输出验收报告**:
```
✅ 已完成: $ARGUMENTS.html → features/<module>/<page>_page.dart

验收结果:
  [✅] 区块完整性: X/X 区块全部实现
  [✅] 数据字段: X 个字段全部渲染
  [✅] 视觉还原: colorScheme 语义映射正确
  [✅] 交互行为: X 个跳转/点击全部实现
  [✅] 三态完整: 加载/空/错误态全部实现
  [✅] 双模式兼容: Light + Dark 均正确
  [✅] 响应式: 375/768/1024 断点正常

创建的文件:
  - flutter_client/lib/src/features/<module>/presentation/<page>_page.dart
  - flutter_client/lib/src/features/<module>/presentation/widgets/...
  - flutter_client/lib/src/features/<module>/state/<module>_providers.dart
  - flutter_client/lib/src/features/<module>/data/models/...
```

---

### Step 8: 更新 Epic 文件

每完成一个页面迁移，更新 Epic 文件记录：

**Epic 文件路径**: `.Codex/epics/OPC_OS_HTML_to_Flutter_Migration/epic.md`

**更新内容**:
1. 将对应 Task 的 `状态` 改为 `已完成`
2. 补充 `创建文件` 列表
3. 更新 `## 5. 统计` 中的已完成数量和完成率
4. 更新 frontmatter 中的 `updated` 时间和 `progress` 百分比

**进度计算**: `progress = (已完成页面数 / 37) * 100`

---

## Error Recovery

| 错误 | 处理 |
|------|------|
| HTML 文件不存在 | 列出可用的 HTML 文件，让用户选择 |
| 目标模块已有同名文件 | 询问用户是否覆盖或合并 |
| AMNT 主题文件缺失 | 终止，提示检查 `core/theme/` 目录 |
| ui-ux-pro-max 脚本错误 | 跳过审核，使用内置检查清单 |
| 路由已存在 | 跳过路由注册，提示用户检查 |

---

## Quick Reference: 原型→模块映射

完整映射见 `reference/page-module-mapping.md`，核心映射：

| HTML 原型 | Flutter 模块 | 优先级 | 导航层级 |
|----------|-------------|--------|---------|
| home.html | home/ | P0 | Tab: 首页 |
| workers.html | ai_workers/ | P0 | Tab: AI团队 |
| worker-detail.html | ai_workers/ | P0 | 二级页面 |
| chat.html | ai_workers/ | P0 | 三级页面 |
| briefing.html | briefing/ | P0 | 首页→简报 |
| dashboard.html | dashboard/ | P0 | Tab: 数据 |
| onboarding.html | onboarding/ | P0 | 独立流程 |
| projects.html | projects/ | P1 | Tab: 项目 |
| market.html | marketplace/ | P1 | Tab: 市场 |
| profile.html | profile/ | P1 | Tab: 我的 |
