# DDL Schema Design

OPC-OS 平台数据库表设计。遵循 coze-studio 风格：MySQL 8.4, BIGINT 时间戳, 软删除, JSON 扩展字段。

---

## 设计原则

1. **表前缀**: `opc_` 区分 OPC 业务表与 coze-studio 原有表
2. **主键**: `BIGINT UNSIGNED AUTO_INCREMENT`
3. **租户隔离**: `space_id` + `user_id` 组合
4. **时间戳**: `BIGINT` Unix 毫秒时间戳（非 DATETIME）
5. **软删除**: `deleted_at BIGINT DEFAULT 0`（0=未删除）
6. **枚举值**: `TINYINT` + COMMENT 说明含义
7. **JSON 字段**: 复杂嵌套/可变结构用 JSON 类型
8. **索引**: 查询字段必建索引，避免全表扫描

---

## 1. 用户与工作空间

### opc_user_profile（用户 OPC 资料）
```sql
CREATE TABLE IF NOT EXISTS `opc_user_profile` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `space_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '工作空间ID',
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
  `display_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '显示名',
  `avatar_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '头像URL',
  `bio` TEXT COMMENT '个人简介',
  `phone` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '手机号',
  `email` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '邮箱',
  `track_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '当前赛道ID',
  `onboarding_step` TINYINT NOT NULL DEFAULT 0 COMMENT '引导进度: 0-6',
  `onboarding_completed` TINYINT NOT NULL DEFAULT 0 COMMENT '0=未完成, 1=已完成',
  `extra` JSON DEFAULT NULL COMMENT '扩展字段',
  `created_at` BIGINT NOT NULL DEFAULT 0,
  `updated_at` BIGINT NOT NULL DEFAULT 0,
  `deleted_at` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_space_user` (`space_id`, `user_id`, `deleted_at`),
  KEY `idx_track` (`track_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='OPC用户资料';
```
**涉及 Issue**: #001, #007, #008, #014

---

### opc_user_portfolio（作品集）
```sql
CREATE TABLE IF NOT EXISTS `opc_user_portfolio` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `space_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `about_me` TEXT COMMENT '关于我',
  `skills` JSON COMMENT '[{"name":"AI Agent","level":"expert","tag_color":"green"}]',
  `service_packages` JSON COMMENT '[{"name":"基础版","price":999,"features":[]}]',
  `works` JSON COMMENT '[{"title":"","cover_url":"","description":"","tags":[]}]',
  `client_reviews` JSON COMMENT '[{"client_name":"","rating":5,"content":""}]',
  `created_at` BIGINT NOT NULL DEFAULT 0,
  `updated_at` BIGINT NOT NULL DEFAULT 0,
  `deleted_at` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_space_user` (`space_id`, `user_id`, `deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='OPC作品集';
```
**涉及 Issue**: #015

---

### opc_user_brand（品牌形象）
```sql
CREATE TABLE IF NOT EXISTS `opc_user_brand` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `space_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `brand_name` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '品牌名',
  `slogan` VARCHAR(256) NOT NULL DEFAULT '' COMMENT '品牌标语',
  `logo_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT 'Logo URL',
  `brand_color` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '品牌色 hex',
  `brand_quote` TEXT COMMENT '品牌引用语',
  `brand_story` TEXT COMMENT '品牌故事',
  `tags` JSON COMMENT '品牌标签 ["AI","创新"]',
  `assets` JSON COMMENT '品牌资产 [{"name":"","preview_url":"","type":""}]',
  `platforms` JSON COMMENT '平台形象 [{"name":"小红书","icon":"","sync_status":""}]',
  `materials` JSON COMMENT '品牌物料 [{"name":"","preview_url":"","type":""}]',
  `created_at` BIGINT NOT NULL DEFAULT 0,
  `updated_at` BIGINT NOT NULL DEFAULT 0,
  `deleted_at` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_space_user` (`space_id`, `user_id`, `deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='OPC品牌形象';
```
**涉及 Issue**: #031

---

## 2. AI 员工

### opc_worker（AI 员工配置）
```sql
CREATE TABLE IF NOT EXISTS `opc_worker` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `space_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `worker_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '员工唯一标识',
  `name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '员工名',
  `title` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '职位',
  `avatar_char` VARCHAR(4) NOT NULL DEFAULT '' COMMENT '头像字符',
  `avatar_color` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '头像颜色 enum',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '0=offline,1=online,2=working,3=standby,4=pending_review,5=executing',
  `agent_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '绑定的 coze agent ID',
  `autonomy_level` TINYINT NOT NULL DEFAULT 1 COMMENT '自主性等级 1-5',
  `notification_enabled` TINYINT NOT NULL DEFAULT 1 COMMENT '通知开关',
  `current_task` VARCHAR(256) NOT NULL DEFAULT '' COMMENT '当前任务描述',
  `skills_config` JSON COMMENT '技能配置 {"coze_workflows":[],"comfyui":[],"mcp":[],"phone_skills":[]}',
  `extra` JSON DEFAULT NULL,
  `created_at` BIGINT NOT NULL DEFAULT 0,
  `updated_at` BIGINT NOT NULL DEFAULT 0,
  `deleted_at` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_space_worker` (`space_id`, `worker_id`, `deleted_at`),
  KEY `idx_space_user` (`space_id`, `user_id`),
  KEY `idx_agent` (`agent_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='OPC AI员工';
```
**涉及 Issue**: #002, #003, #004, #023

---

### opc_worker_task（员工任务记录）
```sql
CREATE TABLE IF NOT EXISTS `opc_worker_task` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `space_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `worker_id` VARCHAR(64) NOT NULL DEFAULT '',
  `task_type` TINYINT NOT NULL DEFAULT 0 COMMENT '0=normal,1=collaboration,2=phone',
  `title` VARCHAR(256) NOT NULL DEFAULT '',
  `description` TEXT,
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '0=pending,1=running,2=completed,3=failed,4=cancelled',
  `credits_used` INT NOT NULL DEFAULT 0 COMMENT '消耗积分',
  `result_summary` TEXT COMMENT '结果摘要',
  `tags` JSON COMMENT '标签 [{"text":"","color":""}]',
  `started_at` BIGINT NOT NULL DEFAULT 0,
  `completed_at` BIGINT NOT NULL DEFAULT 0,
  `created_at` BIGINT NOT NULL DEFAULT 0,
  `updated_at` BIGINT NOT NULL DEFAULT 0,
  `deleted_at` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_space_worker` (`space_id`, `worker_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='OPC员工任务';
```
**涉及 Issue**: #001, #003, #005, #034

---

### opc_worker_collaboration（员工协作记录）
```sql
CREATE TABLE IF NOT EXISTS `opc_worker_collaboration` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `space_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `mode` TINYINT NOT NULL DEFAULT 0 COMMENT '0=sequential,1=parallel,2=assist,3=event',
  `initiator_worker_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '发起方',
  `target_worker_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '目标方',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '0=active,1=completed',
  `description` TEXT,
  `flow_nodes` JSON COMMENT '流程节点 [{"worker_id":"","status":"","step":""}]',
  `messages` JSON COMMENT '协助消息 [{"role":"","content":"","time":""}]',
  `created_at` BIGINT NOT NULL DEFAULT 0,
  `updated_at` BIGINT NOT NULL DEFAULT 0,
  `deleted_at` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_space` (`space_id`),
  KEY `idx_initiator` (`initiator_worker_id`),
  KEY `idx_target` (`target_worker_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='OPC员工协作';
```
**涉及 Issue**: #005, #023

---

## 3. 通知与消息

### opc_notification（通知）
```sql
CREATE TABLE IF NOT EXISTS `opc_notification` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `space_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `type` TINYINT NOT NULL DEFAULT 0 COMMENT '0=system,1=worker,2=task,3=finance,4=content',
  `title` VARCHAR(256) NOT NULL DEFAULT '',
  `content` TEXT,
  `is_read` TINYINT NOT NULL DEFAULT 0 COMMENT '0=unread,1=read',
  `action_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '跳转链接',
  `source_worker_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '来源员工',
  `extra` JSON DEFAULT NULL,
  `created_at` BIGINT NOT NULL DEFAULT 0,
  `updated_at` BIGINT NOT NULL DEFAULT 0,
  `deleted_at` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_space_user` (`space_id`, `user_id`),
  KEY `idx_is_read` (`is_read`),
  KEY `idx_type` (`type`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='OPC通知';
```
**涉及 Issue**: #009

---

### opc_messaging_channel（消息渠道）
```sql
CREATE TABLE IF NOT EXISTS `opc_messaging_channel` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `space_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `channel_type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT 'wechat/feishu/dingtalk/email/sms/webhook',
  `channel_name` VARCHAR(128) NOT NULL DEFAULT '',
  `icon_url` VARCHAR(512) NOT NULL DEFAULT '',
  `is_connected` TINYINT NOT NULL DEFAULT 0,
  `is_enabled` TINYINT NOT NULL DEFAULT 1,
  `config` JSON COMMENT '渠道配置 {"webhook_url":"","token":""}',
  `created_at` BIGINT NOT NULL DEFAULT 0,
  `updated_at` BIGINT NOT NULL DEFAULT 0,
  `deleted_at` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_space_user` (`space_id`, `user_id`),
  KEY `idx_channel_type` (`channel_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='OPC消息渠道';
```
**涉及 Issue**: #030

---

### opc_push_rule（推送规则）
```sql
CREATE TABLE IF NOT EXISTS `opc_push_rule` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `space_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `event_type` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '事件类型',
  `channel_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '推送渠道',
  `is_enabled` TINYINT NOT NULL DEFAULT 1,
  `schedule` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'cron 表达式或时间段',
  `quiet_start` VARCHAR(8) NOT NULL DEFAULT '' COMMENT '免打扰开始 HH:mm',
  `quiet_end` VARCHAR(8) NOT NULL DEFAULT '' COMMENT '免打扰结束 HH:mm',
  `created_at` BIGINT NOT NULL DEFAULT 0,
  `updated_at` BIGINT NOT NULL DEFAULT 0,
  `deleted_at` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_space_user` (`space_id`, `user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='OPC推送规则';
```
**涉及 Issue**: #030

---

## 4. 内容管理

### opc_content（内容）
```sql
CREATE TABLE IF NOT EXISTS `opc_content` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `space_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `worker_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '创作员工',
  `title` VARCHAR(256) NOT NULL DEFAULT '',
  `content_body` MEDIUMTEXT COMMENT '内容正文',
  `platform` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '目标平台: xiaohongshu/wechat/jike/douyin',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '0=draft,1=pending_review,2=approved,3=rejected,4=scheduled,5=published',
  `scheduled_at` BIGINT NOT NULL DEFAULT 0 COMMENT '排期时间',
  `published_at` BIGINT NOT NULL DEFAULT 0 COMMENT '发布时间',
  `review_note` TEXT COMMENT '审核备注',
  `metrics` JSON COMMENT '发布效果 {"views":0,"likes":0,"comments":0,"shares":0,"favorites":0}',
  `tags` JSON COMMENT '标签',
  `created_at` BIGINT NOT NULL DEFAULT 0,
  `updated_at` BIGINT NOT NULL DEFAULT 0,
  `deleted_at` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_space_user` (`space_id`, `user_id`),
  KEY `idx_worker` (`worker_id`),
  KEY `idx_status` (`status`),
  KEY `idx_platform` (`platform`),
  KEY `idx_scheduled_at` (`scheduled_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='OPC内容';
```
**涉及 Issue**: #010, #011

---

## 5. CRM

### opc_customer（客户）
```sql
CREATE TABLE IF NOT EXISTS `opc_customer` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `space_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `name` VARCHAR(128) NOT NULL DEFAULT '',
  `company` VARCHAR(256) NOT NULL DEFAULT '',
  `contact_info` JSON COMMENT '{"phone":"","email":"","wechat":""}',
  `pipeline_stage` TINYINT NOT NULL DEFAULT 0 COMMENT '0=new,1=contacted,2=replied,3=deep_talk,4=closed_won,5=closed_lost',
  `source` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '来源渠道',
  `match_score` INT NOT NULL DEFAULT 0 COMMENT '匹配度 0-100',
  `last_contact_at` BIGINT NOT NULL DEFAULT 0,
  `tags` JSON,
  `notes` TEXT,
  `assigned_worker_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '负责员工',
  `created_at` BIGINT NOT NULL DEFAULT 0,
  `updated_at` BIGINT NOT NULL DEFAULT 0,
  `deleted_at` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_space_user` (`space_id`, `user_id`),
  KEY `idx_pipeline` (`pipeline_stage`),
  KEY `idx_worker` (`assigned_worker_id`),
  KEY `idx_last_contact` (`last_contact_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='OPC客户';
```
**涉及 Issue**: #012

---

### opc_follow_up（跟进提醒）
```sql
CREATE TABLE IF NOT EXISTS `opc_follow_up` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `space_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `customer_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `worker_id` VARCHAR(64) NOT NULL DEFAULT '',
  `content` TEXT,
  `due_at` BIGINT NOT NULL DEFAULT 0,
  `is_completed` TINYINT NOT NULL DEFAULT 0,
  `completed_at` BIGINT NOT NULL DEFAULT 0,
  `created_at` BIGINT NOT NULL DEFAULT 0,
  `updated_at` BIGINT NOT NULL DEFAULT 0,
  `deleted_at` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_space` (`space_id`),
  KEY `idx_customer` (`customer_id`),
  KEY `idx_due_at` (`due_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='OPC跟进提醒';
```
**涉及 Issue**: #012

---

## 6. 项目管理

### opc_project（项目）
```sql
CREATE TABLE IF NOT EXISTS `opc_project` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `space_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `name` VARCHAR(256) NOT NULL DEFAULT '',
  `description` TEXT,
  `client_name` VARCHAR(128) NOT NULL DEFAULT '',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '0=pending,1=in_progress,2=completed,3=cancelled',
  `kanban_stage` VARCHAR(32) NOT NULL DEFAULT 'backlog' COMMENT 'backlog/todo/in_progress/review/done',
  `budget` BIGINT NOT NULL DEFAULT 0 COMMENT '预算(分)',
  `revenue` BIGINT NOT NULL DEFAULT 0 COMMENT '收入(分)',
  `progress` TINYINT NOT NULL DEFAULT 0 COMMENT '进度 0-100',
  `deadline_at` BIGINT NOT NULL DEFAULT 0,
  `escrow_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联托管ID',
  `assigned_workers` JSON COMMENT '["worker_id_1","worker_id_2"]',
  `created_at` BIGINT NOT NULL DEFAULT 0,
  `updated_at` BIGINT NOT NULL DEFAULT 0,
  `deleted_at` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_space_user` (`space_id`, `user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_deadline` (`deadline_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='OPC项目';
```
**涉及 Issue**: #013

---

## 7. 财务

### opc_credits_account（积分账户）
```sql
CREATE TABLE IF NOT EXISTS `opc_credits_account` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `space_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `balance` INT NOT NULL DEFAULT 0 COMMENT '当前余额',
  `total_purchased` INT NOT NULL DEFAULT 0 COMMENT '总购买量',
  `total_used` INT NOT NULL DEFAULT 0 COMMENT '总消耗量',
  `total_earned` INT NOT NULL DEFAULT 0 COMMENT '总赚取量',
  `created_at` BIGINT NOT NULL DEFAULT 0,
  `updated_at` BIGINT NOT NULL DEFAULT 0,
  `deleted_at` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_space_user` (`space_id`, `user_id`, `deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='OPC积分账户';
```
**涉及 Issue**: #001, #016

---

### opc_credits_transaction（积分流水）
```sql
CREATE TABLE IF NOT EXISTS `opc_credits_transaction` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `space_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `type` TINYINT NOT NULL DEFAULT 0 COMMENT '0=consume,1=purchase,2=earn,3=refund',
  `amount` INT NOT NULL DEFAULT 0 COMMENT '数量(正数)',
  `balance_after` INT NOT NULL DEFAULT 0 COMMENT '交易后余额',
  `worker_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '关联员工',
  `task_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联任务',
  `description` VARCHAR(256) NOT NULL DEFAULT '',
  `created_at` BIGINT NOT NULL DEFAULT 0,
  `deleted_at` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_space_user` (`space_id`, `user_id`),
  KEY `idx_worker` (`worker_id`),
  KEY `idx_type` (`type`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='OPC积分流水';
```
**涉及 Issue**: #016

---

### opc_subscription（会员订阅）
```sql
CREATE TABLE IF NOT EXISTS `opc_subscription` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `space_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `plan_id` VARCHAR(32) NOT NULL DEFAULT '' COMMENT 'free/starter/pro/enterprise',
  `billing_cycle` TINYINT NOT NULL DEFAULT 0 COMMENT '0=monthly,1=yearly',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '0=inactive,1=active,2=expired,3=cancelled',
  `started_at` BIGINT NOT NULL DEFAULT 0,
  `expires_at` BIGINT NOT NULL DEFAULT 0,
  `payment_method` VARCHAR(32) NOT NULL DEFAULT '' COMMENT 'alipay/wechat/card',
  `auto_renew` TINYINT NOT NULL DEFAULT 0,
  `created_at` BIGINT NOT NULL DEFAULT 0,
  `updated_at` BIGINT NOT NULL DEFAULT 0,
  `deleted_at` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_space_user` (`space_id`, `user_id`, `deleted_at`),
  KEY `idx_plan` (`plan_id`),
  KEY `idx_expires` (`expires_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='OPC会员订阅';
```
**涉及 Issue**: #017

---

### opc_finance_record（财务记录）
```sql
CREATE TABLE IF NOT EXISTS `opc_finance_record` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `space_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `type` TINYINT NOT NULL DEFAULT 0 COMMENT '0=income,1=expense',
  `category` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '分类: project/subscription/credits/tax',
  `amount` BIGINT NOT NULL DEFAULT 0 COMMENT '金额(分)',
  `description` VARCHAR(256) NOT NULL DEFAULT '',
  `project_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `reference_id` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '外部单号',
  `recorded_at` BIGINT NOT NULL DEFAULT 0 COMMENT '记账日期',
  `created_at` BIGINT NOT NULL DEFAULT 0,
  `updated_at` BIGINT NOT NULL DEFAULT 0,
  `deleted_at` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_space_user` (`space_id`, `user_id`),
  KEY `idx_type_category` (`type`, `category`),
  KEY `idx_recorded_at` (`recorded_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='OPC财务记录';
```
**涉及 Issue**: #021

---

### opc_escrow（资金托管）
```sql
CREATE TABLE IF NOT EXISTS `opc_escrow` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `space_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `project_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `client_name` VARCHAR(128) NOT NULL DEFAULT '',
  `total_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '总金额(分)',
  `released_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '已释放(分)',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '0=pending,1=active,2=completed,3=disputed',
  `milestones` JSON COMMENT '[{"name":"","amount":0,"status":"","completed_at":0}]',
  `commission_rate` INT NOT NULL DEFAULT 0 COMMENT '佣金比例(万分比)',
  `created_at` BIGINT NOT NULL DEFAULT 0,
  `updated_at` BIGINT NOT NULL DEFAULT 0,
  `deleted_at` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_space_user` (`space_id`, `user_id`),
  KEY `idx_project` (`project_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='OPC资金托管';
```
**涉及 Issue**: #022

---

## 8. 市场

### opc_demand（需求）
```sql
CREATE TABLE IF NOT EXISTS `opc_demand` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `space_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `publisher_user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '发布者(企业)',
  `title` VARCHAR(256) NOT NULL DEFAULT '',
  `description` TEXT,
  `category` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '分类',
  `budget_min` BIGINT NOT NULL DEFAULT 0 COMMENT '预算下限(分)',
  `budget_max` BIGINT NOT NULL DEFAULT 0 COMMENT '预算上限(分)',
  `deadline_at` BIGINT NOT NULL DEFAULT 0,
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '0=open,1=in_progress,2=completed,3=cancelled',
  `required_skills` JSON COMMENT '["AI Agent","内容创作"]',
  `tags` JSON,
  `created_at` BIGINT NOT NULL DEFAULT 0,
  `updated_at` BIGINT NOT NULL DEFAULT 0,
  `deleted_at` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_space` (`space_id`),
  KEY `idx_publisher` (`publisher_user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_category` (`category`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='OPC需求';
```
**涉及 Issue**: #018, #019, #020

---

### opc_application（投递申请）
```sql
CREATE TABLE IF NOT EXISTS `opc_application` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `space_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `demand_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `applicant_user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `cover_letter` TEXT,
  `proposed_price` BIGINT NOT NULL DEFAULT 0 COMMENT '报价(分)',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '0=pending,1=accepted,2=rejected,3=withdrawn',
  `match_score` INT NOT NULL DEFAULT 0 COMMENT '匹配度 0-100',
  `created_at` BIGINT NOT NULL DEFAULT 0,
  `updated_at` BIGINT NOT NULL DEFAULT 0,
  `deleted_at` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_demand` (`demand_id`),
  KEY `idx_applicant` (`applicant_user_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='OPC投递申请';
```
**涉及 Issue**: #019

---

## 9. 手机 OS

### opc_phone_device（手机设备）
```sql
CREATE TABLE IF NOT EXISTS `opc_phone_device` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `space_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `device_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '设备唯一标识',
  `device_name` VARCHAR(128) NOT NULL DEFAULT '',
  `model_name` VARCHAR(128) NOT NULL DEFAULT '',
  `is_online` TINYINT NOT NULL DEFAULT 0,
  `battery_level` TINYINT NOT NULL DEFAULT 0 COMMENT '电量 0-100',
  `storage_used_gb` DECIMAL(5,2) NOT NULL DEFAULT 0,
  `storage_total_gb` DECIMAL(5,2) NOT NULL DEFAULT 0,
  `connection_settings` JSON COMMENT '{"ws_endpoint":"","heartbeat_interval":30,"network_mode":""}',
  `ai_settings` JSON COMMENT '{"gui_agent":"ui-tars-7b","screenshot_quality":"medium"}',
  `security_settings` JSON COMMENT '{"auto_lock":true,"screenshot_mask":false}',
  `power_settings` JSON COMMENT '{"auto_sleep":true,"low_power_mode":false}',
  `blacklisted_apps` JSON COMMENT '["app1","app2"]',
  `subscription_plan` VARCHAR(32) NOT NULL DEFAULT 'free' COMMENT 'free/basic/pro',
  `last_heartbeat_at` BIGINT NOT NULL DEFAULT 0,
  `created_at` BIGINT NOT NULL DEFAULT 0,
  `updated_at` BIGINT NOT NULL DEFAULT 0,
  `deleted_at` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_device` (`device_id`, `deleted_at`),
  KEY `idx_space_user` (`space_id`, `user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='OPC手机设备';
```
**涉及 Issue**: #032, #033

---

### opc_phone_task（手机任务）
```sql
CREATE TABLE IF NOT EXISTS `opc_phone_task` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `space_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `device_id` VARCHAR(64) NOT NULL DEFAULT '',
  `worker_id` VARCHAR(64) NOT NULL DEFAULT '',
  `title` VARCHAR(256) NOT NULL DEFAULT '',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '0=queued,1=running,2=completed,3=failed,4=paused,5=aborted',
  `total_steps` INT NOT NULL DEFAULT 0,
  `current_step` INT NOT NULL DEFAULT 0,
  `steps` JSON COMMENT '[{"step":1,"action":"","status":"","screenshot_url":""}]',
  `screenshot_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '当前截屏',
  `priority` TINYINT NOT NULL DEFAULT 0 COMMENT '0=normal,1=high',
  `started_at` BIGINT NOT NULL DEFAULT 0,
  `completed_at` BIGINT NOT NULL DEFAULT 0,
  `created_at` BIGINT NOT NULL DEFAULT 0,
  `updated_at` BIGINT NOT NULL DEFAULT 0,
  `deleted_at` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_device` (`device_id`),
  KEY `idx_worker` (`worker_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='OPC手机任务';
```
**涉及 Issue**: #034

---

## 10. 质量与合规

### opc_credit_score（信用评分）
```sql
CREATE TABLE IF NOT EXISTS `opc_credit_score` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `space_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `total_score` INT NOT NULL DEFAULT 0 COMMENT '总分 0-100',
  `level` VARCHAR(16) NOT NULL DEFAULT '' COMMENT 'bronze/silver/gold/diamond',
  `rank_percentile` INT NOT NULL DEFAULT 0 COMMENT '超过百分比',
  `dimensions` JSON COMMENT '[{"name":"交付质量","score":92,"max":100}]',
  `improvement_tips` JSON COMMENT '[{"title":"","content":"","priority":""}]',
  `monthly_scores` JSON COMMENT '[{"month":"2026-01","score":88}]',
  `calculated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '最近计算时间',
  `created_at` BIGINT NOT NULL DEFAULT 0,
  `updated_at` BIGINT NOT NULL DEFAULT 0,
  `deleted_at` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_space_user` (`space_id`, `user_id`, `deleted_at`),
  KEY `idx_level` (`level`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='OPC信用评分';
```
**涉及 Issue**: #035

---

### opc_ai_quality_snapshot（AI 质量快照）
```sql
CREATE TABLE IF NOT EXISTS `opc_ai_quality_snapshot` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `space_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `health_score` INT NOT NULL DEFAULT 0 COMMENT '健康度 0-100',
  `model_usage` JSON COMMENT '[{"model":"gpt-4","calls":120,"tokens":50000,"cost_cents":800}]',
  `agent_quality` JSON COMMENT '[{"agent_id":"","name":"","success_rate":95,"avg_latency_ms":200}]',
  `trace_items` JSON COMMENT '[{"trace_id":"","status":"ok|degraded","latency_ms":0}]',
  `cost_today_cents` BIGINT NOT NULL DEFAULT 0,
  `cost_month_cents` BIGINT NOT NULL DEFAULT 0,
  `budget_month_cents` BIGINT NOT NULL DEFAULT 0,
  `snapshot_at` BIGINT NOT NULL DEFAULT 0,
  `created_at` BIGINT NOT NULL DEFAULT 0,
  `deleted_at` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_space_user` (`space_id`, `user_id`),
  KEY `idx_snapshot_at` (`snapshot_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='OPC AI质量快照';
```
**涉及 Issue**: #036

---

### opc_compliance（合规记录）
```sql
CREATE TABLE IF NOT EXISTS `opc_compliance` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `space_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `type` TINYINT NOT NULL DEFAULT 0 COMMENT '0=contract,1=ip,2=tax',
  `title` VARCHAR(256) NOT NULL DEFAULT '',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '0=pending,1=active,2=expired',
  `detail` JSON COMMENT '类型相关详情',
  `expires_at` BIGINT NOT NULL DEFAULT 0,
  `created_at` BIGINT NOT NULL DEFAULT 0,
  `updated_at` BIGINT NOT NULL DEFAULT 0,
  `deleted_at` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_space_user` (`space_id`, `user_id`),
  KEY `idx_type_status` (`type`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='OPC合规记录';
```
**涉及 Issue**: #037

---

## 表统计

| 分类 | 表数量 | 表名 |
|------|--------|------|
| 用户 | 3 | opc_user_profile, opc_user_portfolio, opc_user_brand |
| AI 员工 | 3 | opc_worker, opc_worker_task, opc_worker_collaboration |
| 通知消息 | 3 | opc_notification, opc_messaging_channel, opc_push_rule |
| 内容 | 1 | opc_content |
| CRM | 2 | opc_customer, opc_follow_up |
| 项目 | 1 | opc_project |
| 财务 | 4 | opc_credits_account, opc_credits_transaction, opc_subscription, opc_finance_record, opc_escrow |
| 市场 | 2 | opc_demand, opc_application |
| 手机 | 2 | opc_phone_device, opc_phone_task |
| 质量 | 3 | opc_credit_score, opc_ai_quality_snapshot, opc_compliance |
| **合计** | **24** | |

> **注意**: #007/#008 引导数据存储在 `opc_user_profile.onboarding_step` + Redis 缓存中，不单独建表。
> #006 数据看板、#005 简报 为聚合查询，不单独建表。
> #025-#028 自动化、#029 知识库 复用 coze-studio 已有表。
