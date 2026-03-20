# HTML 原型 → Flutter 模块映射表

> 37 个 HTML 原型页面与 Flutter feature 模块的完整对应关系。
> 包含优先级、导航层级、模块状态（新建/重构/扩展）。

---

## P0: MVP 核心体验

| HTML 原型 | Flutter 模块 | 页面文件 | 导航层级 | 模块状态 |
|----------|-------------|---------|---------|---------|
| home.html | features/home/ | home_page.dart | Tab: 首页 | 🔄重构 |
| workers.html | features/ai_workers/ | workers_page.dart | Tab: AI团队 | 🔄重构channels/ |
| worker-detail.html | features/ai_workers/ | worker_detail_page.dart | 二级: AI团队→详情 | 🔄重构channels/ |
| chat.html | features/ai_workers/ | worker_chat_page.dart | 三级: 详情→对话 | 🔄复用messaging_tabs/ |
| briefing.html | features/briefing/ | briefing_page.dart | 二级: 首页→简报 | ❌新建 |
| dashboard.html | features/dashboard/ | dashboard_page.dart | Tab: 数据 | 🔄重构 |
| onboarding.html | features/onboarding/ | onboarding_page.dart | 独立流程 | ❌新建 |
| track-assessment.html | features/onboarding/ | track_assessment_page.dart | 二级: 引导→评估 | ❌新建 |

---

## P1: 业务闭环

| HTML 原型 | Flutter 模块 | 页面文件 | 导航层级 | 模块状态 |
|----------|-------------|---------|---------|---------|
| notifications.html | features/notifications/ | notifications_page.dart | 全局入口(AppBar) | ❌新建 |
| content.html | features/content/ | content_page.dart | 二级: 项目→内容 | ❌新建 |
| content-review.html | features/content/ | content_review_page.dart | 三级: 内容→审核 | ❌新建 |
| crm.html | features/crm/ | crm_page.dart | 二级: 项目→CRM | 🔄重构contacts/ |
| projects.html | features/projects/ | projects_page.dart | Tab: 项目 | ❌新建 |
| profile.html | features/profile/ | profile_page.dart | Tab: 我的 | 🔄扩展 |
| portfolio.html | features/profile/ | portfolio_page.dart | 二级: 我的→作品集 | ❌新建 |
| credits.html | features/finance/ | credits_page.dart | 二级: 我的→积分 | ❌新建 |
| subscription.html | features/subscription/ | subscription_page.dart | 二级: 我的→订阅 | 🔄扩展 |
| demands.html | features/marketplace/ | demands_page.dart | 二级: 市场→需求 | ❌新建 |
| market.html | features/marketplace/ | marketplace_page.dart | Tab: 市场 | ❌新建 |
| enterprise.html | features/marketplace/ | enterprise_page.dart | 二级: 市场→企业 | ❌新建 |

---

## P2: 高级功能

| HTML 原型 | Flutter 模块 | 页面文件 | 导航层级 | 模块状态 |
|----------|-------------|---------|---------|---------|
| finance.html | features/finance/ | finance_page.dart | 二级: 我的→财务 | ❌新建 |
| escrow.html | features/finance/ | escrow_page.dart | 三级: 财务→Escrow | ❌新建 |
| worker-collaboration.html | features/ai_workers/ | worker_collaboration_page.dart | 二级: AI团队→协作 | ❌新建 |
| worker-memory.html | features/ai_workers/ | worker_memory_page.dart | 三级: 详情→记忆 | ❌新建 |
| automation.html | features/automation/ | automation_page.dart | 侧边栏(Desktop) | ❌新建 |
| coze-workflows.html | features/automation/ | coze_workflows_page.dart | 二级: 自动化→Coze | ❌新建 |
| comfyui-workflows.html | features/automation/ | comfyui_workflows_page.dart | 二级: 自动化→ComfyUI | ❌新建 |
| mcp-tools.html | features/automation/ | mcp_tools_page.dart | 二级: 自动化→MCP | ❌新建 |
| knowledge.html | features/knowledge/ | knowledge_page.dart | 侧边栏(Desktop) | 🔄改造notebook/ |
| messaging-channels.html | features/messaging_channels/ | messaging_channels_page.dart | 二级: 我的→消息渠道 | ❌新建 |
| brand.html | features/profile/ | brand_page.dart | 二级: 我的→品牌 | ❌新建 |

---

## P3: 全功能

| HTML 原型 | Flutter 模块 | 页面文件 | 导航层级 | 模块状态 |
|----------|-------------|---------|---------|---------|
| phone-os.html | features/phone_os/ | phone_os_page.dart | 二级: 我的→Phone | 🔄扩展device/ |
| phone-os-edit.html | features/phone_os/ | phone_os_edit_page.dart | 三级: Phone→编辑 | ❌新建 |
| phone-tasks.html | features/phone_os/ | phone_tasks_page.dart | 二级: Phone→任务 | ❌新建 |
| credit-score.html | features/credit_score/ | credit_score_page.dart | 二级: 我的→信用 | ❌新建 |
| ai-quality.html | features/ai_quality/ | ai_quality_page.dart | 侧边栏(Desktop) | ❌新建 |
| comply.html | features/ai_quality/ | comply_page.dart | 二级: 质量→合规 | ❌新建 |

---

## 统计

| 模块状态 | 数量 |
|---------|------|
| ❌ 新建 | 27 |
| 🔄 重构 | 6 |
| 🔄 扩展 | 3 |
| 🔄 改造 | 1 |
| **总计** | **37** |
