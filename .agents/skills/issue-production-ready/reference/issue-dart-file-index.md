# Issue → Dart File Index

每个 issue 对应的 Flutter dart 文件精确路径。执行 skill 时必须读取这些文件。

**路径前缀**: `flutter_client/lib/src/features/`

---

## #001 首页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `home/data/models/opc_home_models.dart` |
| **Provider** | `home/state/opc_home_providers.dart` |
| **Page** | `home/presentation/pages/home_page.dart` |
| **共享Widget** | `home/presentation/widgets/opc_shared_widgets.dart` |

**Mock 函数**: `_getMockHomeData()` in `opc_home_providers.dart`
**核心 Model**: `OpcHomeData`, `BriefingItem`, `QuickStatItem`, `PhoneStatusModel`, `AiWorkerModel`, `RecentTaskModel`, `ReminderModel`, `QuickActionModel`
**共享枚举**: `OpcTagColor`, `WorkerStatus` (定义在 `opc_home_models.dart`)

---

## #002 AI员工列表页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `ai_workers/data/models/workers_models.dart` |
| **Provider** | `ai_workers/state/workers_providers.dart` |
| **Page** | `ai_workers/presentation/pages/workers_page.dart` |
| **Widget** | `ai_workers/presentation/widgets/worker_card.dart` |
| **Widget** | `ai_workers/presentation/widgets/workers_sections.dart` |

**Mock 函数**: `_getMockWorkersData()` in `workers_providers.dart`
**核心 Model**: `WorkerDetailModel`, `EngineModel`, `TemplateWorkerModel`

---

## #003 AI员工详情页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `ai_workers/data/models/worker_detail_models.dart` |
| **Provider** | `ai_workers/state/worker_detail_providers.dart` |
| **Page** | `ai_workers/presentation/pages/worker_detail_page.dart` |
| **Widget** | `ai_workers/presentation/widgets/worker_detail_profile.dart` |
| **Widget** | `ai_workers/presentation/widgets/worker_detail_config_tab.dart` |
| **Widget** | `ai_workers/presentation/widgets/worker_detail_skills_tab.dart` |
| **Widget** | `ai_workers/presentation/widgets/worker_detail_memory_tab.dart` |
| **Widget** | `ai_workers/presentation/widgets/worker_detail_history_tab.dart` |

**Mock 函数**: `_getMockWorkerDetail()` in `worker_detail_providers.dart`
**核心 Model**: `WorkerDetailData`, `AutonomyLevel`, `CozeWorkflow`, `PhoneSkill`, `MemoryItem`, `TaskHistory`

---

## #004 AI员工聊天页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `ai_workers/data/models/worker_chat_models.dart` |
| **Provider** | `ai_workers/state/worker_chat_providers.dart` |
| **Page** | `ai_workers/presentation/pages/worker_chat_page.dart` |
| **Widget** | `ai_workers/presentation/widgets/chat_message_bubble.dart` |
| **Widget** | `ai_workers/presentation/widgets/chat_input_bar.dart` |

**Mock 函数**: `_getMockChatData()` in `worker_chat_providers.dart`
**核心 Model**: `ChatMessage`, `TaskDecomposition`, `CustomerCard`, `PhoneExecutionCard`
**复用 coze-studio**: `conversation` domain

---

## #005 今日简报页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `briefing/data/models/briefing_models.dart` |
| **Provider** | `briefing/state/briefing_providers.dart` |
| **Page** | `briefing/presentation/pages/briefing_page.dart` |
| **Widget** | `briefing/presentation/widgets/briefing_performance_card.dart` |
| **Widget** | `briefing/presentation/widgets/briefing_worker_report_card.dart` |
| **Widget** | `briefing/presentation/widgets/briefing_collaboration_section.dart` |

**Mock 函数**: `_getMockBriefingData()` in `briefing_providers.dart`
**核心 Model**: `BriefingData`, `TeamPerformanceStat`, `WorkerReportModel`, `CollaborationRecord`, `WorkerAlert`, `PhoneDeviceInfo`

---

## #006 数据看板页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `dashboard/data/models/opc_dashboard_models.dart` |
| **Provider** | `dashboard/state/opc_dashboard_providers.dart` |
| **Page** | `dashboard/presentation/pages/opc_dashboard_page.dart` |

**Mock 函数**: `_getMockDashboardData()` in `opc_dashboard_providers.dart`
**核心 Model**: `OpcDashboardData`, `RevenueKpi`, `KeyMetricCard`, `MiniChartBar`, `RevenueTrendBar`, `WorkerEfficiencyItem`, `ContentMetricRow`, `FunnelLevel`, `AiInsight`, `TrackRanking`

---

## #007 新手引导页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `onboarding/data/models/onboarding_models.dart` |
| **Provider** | `onboarding/state/onboarding_providers.dart` |
| **Page** | `onboarding/presentation/pages/onboarding_page.dart` |
| **Widget** | `onboarding/presentation/widgets/onboarding_step_indicator.dart` |
| **Widget** | `onboarding/presentation/widgets/onboarding_steps.dart` |

**Mock 函数**: Mock 数据内置于 `onboarding_providers.dart`
**核心 Model**: `SkillTag`, `RadarSkill`, `TrackRec`, `TeamMember`, `PhoneCapability`, `BindingStep`, `SubscriptionPlan`

---

## #008 赛道评估页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `onboarding/data/models/track_assessment_models.dart` |
| **Provider** | `onboarding/state/track_assessment_providers.dart` |
| **Page** | `onboarding/presentation/pages/track_assessment_page.dart` |

**Mock 函数**: Mock 数据内置于 `track_assessment_providers.dart`
**核心 Model**: `TrackInfo`, `AiAssessmentResult`, `TrackAssessmentData`

---

## #009 通知中心页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `notifications/data/models/notification_models.dart` |
| **Provider** | `notifications/state/notification_providers.dart` |
| **Page** | `notifications/presentation/pages/notifications_page.dart` |

**Mock 函数**: Mock 数据内置于 `notification_providers.dart`
**核心 Model**: `NotificationItem`, `NotificationGroup`, `NotificationsData`

---

## #010 内容管理页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `content/data/models/content_models.dart` |
| **Provider** | `content/state/content_providers.dart` |
| **Page** | `content/presentation/pages/content_page.dart` |

**Mock 函数**: Mock 数据内置于 `content_providers.dart`
**核心 Model**: `WeeklyOverview`, `PendingContent`, `ScheduledContent`, `TopContent`, `ContentPlan`

---

## #011 内容审核页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `content/data/models/content_review_models.dart` |
| **Provider** | `content/state/content_review_providers.dart` |
| **Page** | `content/presentation/pages/content_review_page.dart` |

**Mock 函数**: Mock 数据内置于 `content_review_providers.dart`
**核心 Model**: `ReviewStats`, `ReviewItem`, `ReviewTag`, `PublishedPerformanceItem`, `ContentReviewData`

---

## #012 客户管理页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `crm/data/models/crm_models.dart` |
| **Provider** | `crm/state/crm_providers.dart` |
| **Page** | `crm/presentation/pages/crm_page.dart` |

**Mock 函数**: Mock 数据内置于 `crm_providers.dart`
**核心 Model**: `CrmFilter`, `CrmStatItem`, `PipelineStage`, `CustomerItem`, `FollowUpReminder`, `CrmPageData`

---

## #013 项目管理页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `projects/data/models/projects_models.dart` |
| **Provider** | `projects/state/projects_providers.dart` |
| **Page** | `projects/presentation/pages/projects_page.dart` |

**Mock 函数**: Mock 数据内置于 `projects_providers.dart`
**核心 Model**: `ProjectFilter`, `KanbanColumn`, `ProjectCard`, `RevenueStat`, `EscrowInfo`, `ProjectsPageData`

---

## #014 个人资料页修复

| 类型 | 文件路径 |
|------|---------|
| **Model** | `profile/data/models/opc_profile_models.dart` |
| **Provider** | `profile/state/providers/opc_profile_providers.dart` |
| **Page** | `profile/presentation/pages/opc_profile_page.dart` |

**Mock 函数**: Mock 数据内置于 `opc_profile_providers.dart`
**核心 Model**: OPC Profile 相关模型

---

## #015 作品集页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `profile/data/models/opc_portfolio_models.dart` |
| **Provider** | `profile/state/providers/opc_portfolio_providers.dart` |
| **Page** | `profile/presentation/pages/opc_portfolio_page.dart` |

**Mock 函数**: Mock 数据内置于 `opc_portfolio_providers.dart`
**核心 Model**: `PortfolioProfile`, `SkillTag`, `SkillLevel`, `ServicePackage`, `PortfolioWork`, `ClientReview`

---

## #016 积分管理页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `finance/data/models/opc_credits_models.dart` |
| **Provider** | `finance/state/opc_credits_providers.dart` |
| **Page** | `finance/presentation/pages/opc_credits_page.dart` |

**Mock 函数**: Mock 数据内置于 `opc_credits_providers.dart`
**核心 Model**: `CreditsOverview`, `WorkerUsage`, `AgentTypeUsage`, `CreditTransaction`, `IconDataRef`

---

## #017 会员订阅页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `subscription/data/models/opc_subscription_models.dart` |
| **Provider** | `subscription/state/opc_subscription_providers.dart` |
| **Page** | `subscription/presentation/pages/opc_subscription_page.dart` |

**Mock 函数**: Mock 数据内置于 `opc_subscription_providers.dart`
**核心 Model**: `CurrentPlan`, `MembershipPlan`, `PlanFeature`, `PaymentMethod`, `CreditPackage`, `BillingCycleType`

---

## #018 需求大厅页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `marketplace/data/models/opc_demands_models.dart` |
| **Provider** | `marketplace/state/opc_demands_providers.dart` |
| **Page** | `marketplace/presentation/pages/opc_demands_page.dart` |

**Mock 函数**: Mock 数据内置于 `opc_demands_providers.dart`
**核心 Model**: `DemandsStats`, `DemandItem`, `MatchLevel`, `DemandTag`

---

## #019 市场投递页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `marketplace/data/models/opc_market_models.dart` |
| **Provider** | `marketplace/state/opc_market_providers.dart` |
| **Page** | `marketplace/presentation/pages/opc_market_page.dart` |

**Mock 函数**: Mock 数据内置于 `opc_market_providers.dart`
**核心 Model**: `MarketDemand`, `DemandTagItem`, `ApplicationItem`

---

## #020 企业端页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `marketplace/data/models/opc_enterprise_models.dart` |
| **Provider** | `marketplace/state/opc_enterprise_providers.dart` |
| **Page** | `marketplace/presentation/pages/opc_enterprise_page.dart` |

**Mock 函数**: Mock 数据内置于 `opc_enterprise_providers.dart`
**核心 Model**: `EnterpriseProfile`, `QuickStats`, `EnterpriseDemand`, `RecommendedOpc`, `ActiveProject`, `SpendingOverview`

---

## #021 财务概览页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `finance/data/models/opc_finance_models.dart` |
| **Provider** | `finance/state/opc_finance_providers.dart` |
| **Page** | `finance/presentation/pages/opc_finance_page.dart` |

**Mock 函数**: Mock 数据内置于 `opc_finance_providers.dart`
**核心 Model**: `BalanceOverview`, `QuickStat`, `RevenueChart`, `ChartBar`, `IncomeSource`, `ExpenseItem`, `TaxReminder`, `OpcFinanceIconRef`

---

## #022 资金托管页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `finance/data/models/opc_escrow_models.dart` |
| **Provider** | `finance/state/opc_escrow_providers.dart` |
| **Page** | `finance/presentation/pages/opc_escrow_page.dart` |

**Mock 函数**: Mock 数据内置于 `opc_escrow_providers.dart`
**核心 Model**: `EscrowPageData`, `EscrowOverview`, `ActiveEscrow`, `EscrowStatus`, `CompletedEscrow`, `WithdrawalInfo`

---

## #023 员工协作页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `ai_workers/data/models/worker_collaboration_models.dart` |
| **Provider** | `ai_workers/state/worker_collaboration_providers.dart` |
| **Page** | `ai_workers/presentation/pages/worker_collaboration_page.dart` |

**Mock 函数**: Mock 数据内置于 `worker_collaboration_providers.dart`
**核心 Model**: `CollaborationMode`, `CollabNodeStatus`, `CollabFlowNode`, `CollabParallelWorker`, `CollabAssistMessage`, `Collaboration`, `CollaborationHistory`

---

## #024 员工记忆页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `ai_workers/data/models/worker_memory_models.dart` |
| **Provider** | `ai_workers/state/worker_memory_providers.dart` |
| **Page** | `ai_workers/presentation/pages/worker_memory_page.dart` |

**Mock 函数**: Mock 数据内置于 `worker_memory_providers.dart`
**核心 Model**: `MemoryType`, `MemoryTag`, `MemoryItem`, `MemoryGroup`, `MemoryConflict`, `MemoryStats`, `WorkerMemoryData`
**复用 coze-studio**: `memory` domain

---

## #025 自动化工作流页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `automation/data/models/automation_models.dart` |
| **Provider** | `automation/state/automation_providers.dart` |
| **Page** | `automation/presentation/pages/automation_page.dart` |

**Mock 函数**: Mock 数据内置于 `automation_providers.dart`
**核心 Model**: `WorkflowTriggerType`, `WorkflowStep`, `WorkflowItem`, `IconRef`, `WorkflowStats`, `AutomationPageData`

---

## #026 Coze工作流页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `automation/data/models/coze_workflow_models.dart` |
| **Provider** | `automation/state/coze_workflows_providers.dart` |
| **Page** | `automation/presentation/pages/coze_workflows_page.dart` |

**Mock 函数**: Mock 数据内置于 `coze_workflows_providers.dart`
**核心 Model**: `CozeAgentWorkflow`, `CozeWorkflowNode`, `LLMRouteLevel`, `RAGKnowledgeLevel`, `CozeOverviewStats`, `CozeWorkflowsPageData`
**复用 coze-studio**: `workflow` domain + `agent` domain

---

## #027 ComfyUI工作流页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `automation/data/models/comfyui_workflow_models.dart` |
| **Provider** | `automation/state/comfyui_workflows_providers.dart` |
| **Page** | `automation/presentation/pages/comfyui_workflows_page.dart` |

**Mock 函数**: Mock 数据内置于 `comfyui_workflows_providers.dart`
**核心 Model**: `ComfyUIWorkflowTemplate`, `ComfyUIGeneration`, `ComfyUIOverviewStats`, `ComfyUIWorkflowsPageData`

---

## #028 MCP工具页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `automation/data/models/mcp_tools_models.dart` |
| **Provider** | `automation/state/mcp_tools_providers.dart` |
| **Page** | `automation/presentation/pages/mcp_tools_page.dart` |

**Mock 函数**: Mock 数据内置于 `mcp_tools_providers.dart`
**核心 Model**: `MCPCategory`, `MCPConnectedTool`, `MCPRecommendedTool`, `MCPUsageStats`, `MCPToolsPageData`
**复用 coze-studio**: `plugin` domain

---

## #029 知识库页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `knowledge/data/models/knowledge_models.dart` |
| **Provider** | `knowledge/state/knowledge_providers.dart` |
| **Page** | `knowledge/presentation/pages/knowledge_page.dart` |

**Mock 函数**: Mock 数据内置于 `knowledge_providers.dart`
**核心 Model**: `KnowledgeStats`, `KnowledgeCategory`, `KnowledgeMethodology`, `KnowledgeTag`, `KnowledgeCase`, `KnowledgeTemplate`, `KnowledgePageData`
**复用 coze-studio**: `knowledge` domain

---

## #030 消息渠道页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `messaging_channels/data/models/messaging_channel_models.dart` |
| **Provider** | `messaging_channels/state/messaging_channels_providers.dart` |
| **Page** | `messaging_channels/presentation/pages/messaging_channels_page.dart` |

**Mock 函数**: Mock 数据内置于 `messaging_channels_providers.dart`
**核心 Model**: `ConnectedChannel`, `AvailableChannel`, `PushRule`, `QuietHoursSetting`, `MessagingChannelsPageData`

---

## #031 品牌形象页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `profile/data/models/brand_models.dart` |
| **Provider** | `profile/state/brand_providers.dart` |
| **Page** | `profile/presentation/pages/brand_page.dart` |

**Mock 函数**: Mock 数据内置于 `brand_providers.dart`
**核心 Model**: `BrandInfo`, `BrandTag`, `BrandAsset`, `BrandStory`, `BrandPlatform`, `BrandMaterial`, `BrandPageData`

---

## #032 手机OS页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `phone_os/data/models/phone_os_models.dart` |
| **Provider** | `phone_os/state/phone_os_providers.dart` |
| **Page** | `phone_os/presentation/pages/phone_os_page.dart` |

**Mock 函数**: Mock 数据内置于 `phone_os_providers.dart`
**核心 Model**: `PhoneDeviceStatus`, `DeviceQuota`, `PhoneTaskStep`, `PhoneCurrentTask`, `PhoneCapability`, `PerformanceMetric`, `PhoneRecentTask`, `PhoneSubscription`, `PhoneOSPageData`

---

## #033 手机OS编辑页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `phone_os/data/models/phone_os_edit_models.dart` |
| **Provider** | `phone_os/state/phone_os_providers.dart` (共用，含 `phoneOSEditDataProvider`) |
| **Page** | `phone_os/presentation/pages/phone_os_edit_page.dart` |

**Mock 函数**: `phoneOSEditDataProvider` in `phone_os_providers.dart`
**核心 Model**: `DeviceEditInfo`, `ConnectionSettings`, `GuiAgentOption`, `ScreenshotQualityOption`, `SecuritySettings`, `BlacklistedApp`, `PowerSettings`, `OtaUpdateInfo`, `PhoneOSEditPageData`

---

## #034 手机任务页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `phone_os/data/models/phone_tasks_models.dart` |
| **Provider** | `phone_os/state/phone_os_providers.dart` (共用，含 `phoneTasksDataProvider`) |
| **Page** | `phone_os/presentation/pages/phone_tasks_page.dart` |

**Mock 函数**: `phoneTasksDataProvider` in `phone_os_providers.dart`
**核心 Model**: `ActivePhoneTask`, `PhoneTaskStepItem`, `QueuedPhoneTask`, `CompletedPhoneTask`, `TaskDeviceStatus`, `PhoneTasksPageData`

---

## #035 信用评分页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `credit_score/data/models/credit_score_models.dart` |
| **Provider** | `credit_score/state/credit_score_providers.dart` |
| **Page** | `credit_score/presentation/pages/credit_score_page.dart` |

**Mock 函数**: Mock 数据内置于 `credit_score_providers.dart`
**核心 Model**: `ScoreDimension`, `CreditLevel`, `ImprovementTip`, `MonthScore`, `CreditScorePageData`

---

## #036 AI质量监控页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `ai_quality/data/models/ai_quality_models.dart` |
| **Provider** | `ai_quality/state/ai_quality_providers.dart` |
| **Page** | `ai_quality/presentation/pages/ai_quality_page.dart` |

**Mock 函数**: Mock 数据内置于 `ai_quality_providers.dart`
**核心 Model**: `ModelUsage`, `AgentQuality`, `TraceItem`, `AiCostSummary`, `AiQualityPageData`

---

## #037 合规管理页迁移

| 类型 | 文件路径 |
|------|---------|
| **Model** | `ai_quality/data/models/comply_models.dart` |
| **Provider** | `ai_quality/state/ai_quality_providers.dart` (共用，含合规 Provider) |
| **Page** | `ai_quality/presentation/pages/comply_page.dart` |

**Mock 函数**: Mock 数据内置于 `ai_quality_providers.dart`
**核心 Model**: `ComplianceStatus`, `ContractItem`, `ContractTemplate`, `IpItem`, `ComplyPageData`
