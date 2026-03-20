# API Domain Mapping

37 个 OPC-OS 页面 → API 域映射。每个域对应一组 API endpoint，共享同一组数据库表和 Go domain 包。

> **HTTP Method 规则**: 查询用 GET，创建/修改/删除用 POST。以下 endpoint 中的 POST 仅为默认标注，执行时根据实际语义选择。

---

## 域总览

| API Domain | Go Package | 涉及 Issue | 核心实体 |
|------------|-----------|-----------|---------|
| `opc-home` | `domain/opc_home` | #001 | 首页聚合数据 |
| `opc-worker` | `domain/opc_worker` | #002, #003, #004, #023, #024 | AI 员工配置/状态/协作/记忆 |
| `opc-briefing` | `domain/opc_briefing` | #005 | 团队简报 |
| `opc-dashboard` | `domain/opc_dashboard` | #006 | 数据看板聚合 |
| `opc-onboarding` | `domain/opc_onboarding` | #007, #008 | 新手引导/赛道评估 |
| `opc-notification` | `domain/opc_notification` | #009 | 通知中心 |
| `opc-content` | `domain/opc_content` | #010, #011 | 内容管理+审核 |
| `opc-crm` | `domain/opc_crm` | #012 | 客户管理 |
| `opc-project` | `domain/opc_project` | #013 | 项目管理 |
| `opc-profile` | `domain/opc_profile` | #014, #015, #031 | 用户资料/作品集/品牌 |
| `opc-finance` | `domain/opc_finance` | #016, #017, #021, #022 | 积分/订阅/财务/托管 |
| `opc-marketplace` | `domain/opc_marketplace` | #018, #019, #020 | 需求大厅/投递/企业端 |
| `opc-automation` | `domain/opc_automation` | #025, #026, #027, #028 | 自动化工作流/MCP |
| `opc-knowledge` | `domain/opc_knowledge` | #029 | 知识库 |
| `opc-messaging` | `domain/opc_messaging` | #030 | 消息渠道 |
| `opc-phone` | `domain/opc_phone` | #032, #033, #034 | 手机OS设备管理 |
| `opc-quality` | `domain/opc_quality` | #035, #036, #037 | 信用评分/AI质量/合规 |

---

## 详细 API 设计

### 1. opc-home (#001 首页)

**聚合 API** — 首页数据由多个子服务聚合，一次请求返回。

```
POST /api/opc/v1/home/page-data
```

**Request**:
```json
{ "space_id": "string", "user_id": "string" }
```

**Response** (聚合):
```json
{
  "code": 0,
  "data": {
    "user_name": "string",
    "briefing_date": "string",
    "briefing_items": [{ "dot_color": "string", "content": "string" }],
    "quick_stats": [{ "value": "string", "label": "string", "value_color": "string" }],
    "credits_balance": 0,
    "credits_used": 0,
    "credits_total": 0,
    "income_growth": "string",
    "income_bars": [0.0],
    "phone_status": {
      "device_name": "string", "model_name": "string", "is_online": true,
      "worker_name": "string", "task_description": "string",
      "current_step": 0, "total_steps": 0
    },
    "workers": [{ "id": "string", "name": "string", "title": "string",
      "avatar_char": "string", "avatar_color": "string",
      "status": "string", "current_task": "string" }],
    "recent_tasks": [{ "id": "string", "title": "string", "worker_name": "string",
      "status": "string", "time": "string" }],
    "reminders": [{ "id": "string", "content": "string", "priority": "string" }],
    "quick_actions": [{ "icon": "string", "label": "string", "route": "string", "color": "string" }]
  }
}
```

**数据来源**:
- `briefing_items` → opc-briefing 服务
- `quick_stats` → opc-dashboard 服务聚合
- `credits_*` → opc-finance 服务
- `phone_status` → opc-phone 服务
- `workers` → opc-worker 服务
- `recent_tasks` → opc-worker 任务服务

---

### 2. opc-worker (#002-#004, #023-#024)

#### 2.1 员工列表 (#002)
```
POST /api/opc/v1/worker/list
```
Request: `{ "space_id", "keyword?", "status?", "page", "page_size" }`
Response: `{ "items": [WorkerBrief], "total": int, "engines": [Engine], "templates": [Template] }`

#### 2.2 员工详情 (#003)
```
POST /api/opc/v1/worker/detail
```
Request: `{ "space_id", "worker_id" }`
Response: `{ "profile": WorkerProfile, "config": WorkerConfig, "skills": WorkerSkills, "memory_stats": MemoryStats, "history": [TaskHistory] }`

#### 2.3 员工聊天 (#004)
```
POST /api/opc/v1/worker/chat/history
POST /api/opc/v1/worker/chat/send
```
- history: `{ "worker_id", "page", "page_size" }` → `{ "messages": [ChatMessage] }`
- send: `{ "worker_id", "content", "attachments?" }` → `{ "message_id" }`

**复用**: coze-studio `conversation` domain 的消息收发能力

#### 2.4 员工协作 (#023)
```
POST /api/opc/v1/worker/collaboration/list
```
Request: `{ "space_id" }`
Response: `{ "active": [Collaboration], "history": [CollaborationHistory] }`

#### 2.5 员工记忆 (#024)
```
POST /api/opc/v1/worker/memory/list
POST /api/opc/v1/worker/memory/update
POST /api/opc/v1/worker/memory/resolve-conflict
```

**复用**: coze-studio `memory` domain

---

### 3. opc-briefing (#005)

```
POST /api/opc/v1/briefing/today
```
Request: `{ "space_id", "user_id" }`
Response: `{ "date_text", "performance_stats": [], "worker_reports": [], "collaboration_records": [] }`

**数据来源**: 聚合 opc-worker 的当日任务统计

---

### 4. opc-dashboard (#006)

```
POST /api/opc/v1/dashboard/overview
```
Request: `{ "space_id", "period": "day|week|month" }`
Response: `{ "revenue_kpi", "key_metrics": [], "revenue_trend": [], "worker_efficiency": [], "content_metrics": [], "funnel_levels": [], "ai_insights": [], "track_ranking" }`

**数据来源**: 聚合多个子服务的统计数据

---

### 5. opc-onboarding (#007, #008)

#### 5.1 引导状态
```
POST /api/opc/v1/onboarding/status
```
Response: `{ "current_step": int, "completed_steps": [], "skills": [], "experience": {} }`

#### 5.2 保存引导进度
```
POST /api/opc/v1/onboarding/save-progress
```

#### 5.3 赛道评估 (#008)
```
POST /api/opc/v1/onboarding/track-assessment
```
Request: `{ "selected_skills": [], "experience_level": "string" }`
Response: `{ "ai_result": {}, "tracks": [TrackInfo] }`

---

### 6. opc-notification (#009)

```
POST /api/opc/v1/notification/list
POST /api/opc/v1/notification/mark-read
POST /api/opc/v1/notification/mark-all-read
```

---

### 7. opc-content (#010, #011)

#### 7.1 内容管理 (#010)
```
POST /api/opc/v1/content/overview
POST /api/opc/v1/content/list
POST /api/opc/v1/content/schedule
```

#### 7.2 内容审核 (#011)
```
POST /api/opc/v1/content/review/list
POST /api/opc/v1/content/review/approve
POST /api/opc/v1/content/review/reject
```

---

### 8. opc-crm (#012)

```
POST /api/opc/v1/crm/overview
POST /api/opc/v1/crm/customer/list
POST /api/opc/v1/crm/pipeline/list
POST /api/opc/v1/crm/follow-up/list
```

---

### 9. opc-project (#013)

```
POST /api/opc/v1/project/list
POST /api/opc/v1/project/kanban
POST /api/opc/v1/project/revenue-stats
```

---

### 10. opc-profile (#014, #015, #031)

#### 10.1 个人资料 (#014)
```
POST /api/opc/v1/profile/info
POST /api/opc/v1/profile/update
```

#### 10.2 作品集 (#015)
```
POST /api/opc/v1/profile/portfolio
```

#### 10.3 品牌形象 (#031)
```
POST /api/opc/v1/profile/brand
POST /api/opc/v1/profile/brand/update
```

---

### 11. opc-finance (#016, #017, #021, #022)

#### 11.1 积分管理 (#016)
```
POST /api/opc/v1/finance/credits/overview
POST /api/opc/v1/finance/credits/transactions
```

#### 11.2 会员订阅 (#017)
```
POST /api/opc/v1/finance/subscription/current
POST /api/opc/v1/finance/subscription/plans
POST /api/opc/v1/finance/subscription/upgrade
```

#### 11.3 财务概览 (#021)
```
POST /api/opc/v1/finance/overview
POST /api/opc/v1/finance/revenue-chart
POST /api/opc/v1/finance/income-sources
POST /api/opc/v1/finance/expenses
```

#### 11.4 资金托管 (#022)
```
POST /api/opc/v1/finance/escrow/list
POST /api/opc/v1/finance/escrow/withdraw
```

---

### 12. opc-marketplace (#018, #019, #020)

#### 12.1 需求大厅 (#018)
```
POST /api/opc/v1/marketplace/demands
```

#### 12.2 市场投递 (#019)
```
POST /api/opc/v1/marketplace/applications
POST /api/opc/v1/marketplace/apply
```

#### 12.3 企业端 (#020)
```
POST /api/opc/v1/marketplace/enterprise/overview
POST /api/opc/v1/marketplace/enterprise/demands
```

---

### 13. opc-automation (#025-#028)

#### 13.1 自动化总览 (#025)
```
POST /api/opc/v1/automation/overview
POST /api/opc/v1/automation/workflow/list
POST /api/opc/v1/automation/workflow/toggle
```

#### 13.2 Coze 工作流 (#026)
```
POST /api/opc/v1/automation/coze/overview
POST /api/opc/v1/automation/coze/agents
```
**复用**: coze-studio `workflow` domain + `agent` domain

#### 13.3 ComfyUI 工作流 (#027)
```
POST /api/opc/v1/automation/comfyui/templates
POST /api/opc/v1/automation/comfyui/generate
POST /api/opc/v1/automation/comfyui/history
```

#### 13.4 MCP 工具 (#028)
```
POST /api/opc/v1/automation/mcp/connected
POST /api/opc/v1/automation/mcp/recommended
POST /api/opc/v1/automation/mcp/connect
POST /api/opc/v1/automation/mcp/disconnect
```
**复用**: coze-studio `plugin` domain

---

### 14. opc-knowledge (#029)

```
POST /api/opc/v1/knowledge/overview
POST /api/opc/v1/knowledge/methodologies
POST /api/opc/v1/knowledge/cases
POST /api/opc/v1/knowledge/templates
```
**复用**: coze-studio `knowledge` domain

---

### 15. opc-messaging (#030)

```
POST /api/opc/v1/messaging/channels
POST /api/opc/v1/messaging/channel/toggle
POST /api/opc/v1/messaging/push-rules
POST /api/opc/v1/messaging/quiet-hours
```

---

### 16. opc-phone (#032-#034)

#### 16.1 手机OS (#032)
```
POST /api/opc/v1/phone/device-status
POST /api/opc/v1/phone/current-task
POST /api/opc/v1/phone/capabilities
POST /api/opc/v1/phone/metrics
POST /api/opc/v1/phone/recent-tasks
POST /api/opc/v1/phone/subscription
```

#### 16.2 手机OS编辑 (#033)
```
POST /api/opc/v1/phone/device/settings
POST /api/opc/v1/phone/device/update-settings
```

#### 16.3 手机任务 (#034)
```
POST /api/opc/v1/phone/tasks/active
POST /api/opc/v1/phone/tasks/queued
POST /api/opc/v1/phone/tasks/completed
POST /api/opc/v1/phone/tasks/pause
POST /api/opc/v1/phone/tasks/abort
```

---

### 17. opc-quality (#035-#037)

#### 17.1 信用评分 (#035)
```
POST /api/opc/v1/quality/credit-score
```

#### 17.2 AI 质量监控 (#036)
```
POST /api/opc/v1/quality/ai-overview
POST /api/opc/v1/quality/model-usage
POST /api/opc/v1/quality/agent-quality
POST /api/opc/v1/quality/trace-list
POST /api/opc/v1/quality/cost-summary
```

#### 17.3 合规管理 (#037)
```
POST /api/opc/v1/quality/compliance/status
POST /api/opc/v1/quality/contracts
POST /api/opc/v1/quality/ip-items
```

---

## coze-studio 已有服务复用清单

| OPC API | 复用 coze-studio 服务 | 说明 |
|---------|---------------------|------|
| opc-worker chat | `conversation` domain | 消息收发 |
| opc-worker memory | `memory` domain | 记忆存取 |
| opc-automation coze | `workflow` domain | Coze 工作流 CRUD |
| opc-automation mcp | `plugin` domain | 插件管理 |
| opc-knowledge | `knowledge` domain | 知识库 CRUD |
| opc-worker agent | `agent` domain | Agent 配置 |
