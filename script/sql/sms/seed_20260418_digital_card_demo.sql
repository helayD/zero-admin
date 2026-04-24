-- Demo seed for digital card admin pages.
-- Safe to execute multiple times on an existing environment.

INSERT INTO `sms_draw_activity` (
    `id`, `platform_id`, `tenant_id`, `merchant_id`, `activity_code`, `name`, `rule_summary`,
    `start_time`, `end_time`, `real_name_required`, `participant_condition_summary`,
    `consume_rule_summary`, `probability_rule`, `compliance_rule_summary`,
    `circulation_limit_summary`, `approval_record_ref`, `publish_failure_summary`,
    `publish_readiness`, `copyright_status`, `content_audit_status`, `status`,
    `audit_status`, `is_enabled`, `show_on_home`, `home_entry_title`,
    `home_entry_subtitle`, `home_entry_image`, `home_entry_sort`, `home_entry_enabled`,
    `landing_target_type`, `landing_target_value`, `create_by`, `create_time`,
    `update_by`, `update_time`, `is_deleted`
)
VALUES
    (
        910001, 1, 10, 88, 'DRAW-SPRING-2026', '2026 春季限定数字卡片抽卡',
        '参与抽卡不需要实名认证，每次消耗 1 次抽卡次数；中奖后兑卡和发放数字卡片资产时需要完成实名认证。',
        '2026-03-01 00:00:00', '2026-12-31 23:59:59', 1, '完成新手任务并拥有抽卡次数的会员可参与，中奖兑卡时需完成实名认证',
        '每次抽卡消耗 1 次抽卡次数，不支持现金直购', 'SSR 25%，SR 75%',
        '默认禁止收益承诺，并要求活动文案、卡面资源和处置动作均可追溯',
        '禁止集中竞价；禁止连续挂牌；禁止收益承诺', 'APR-DC-2026-001', '发布预检通过',
        1, 2, 2, 1, 2, 1, 1, '春季限定卡池',
        '抽卡不限实名，中奖兑卡需实名', 'https://img.example.com/digital-card/spring-entry.png', 100, 1,
        'activity', '910001', 1, '2026-03-01 09:00:00', 1, '2026-04-18 18:00:00', 0
    ),
    (
        910002, 1, 10, 0, 'DRAW-NEWBIE-2026', '新会员欢迎抽卡',
        '面向租户级新会员的欢迎抽卡活动，可抽取入门卡片模板。',
        '2026-04-01 00:00:00', '2026-12-31 23:59:59', 0, '注册满 1 天且完成资料填写',
        '每位会员限参与 1 次', '保底必中 1 张欢迎卡',
        '仅做会员权益展示，不承诺收益或交易价值',
        '禁止集中竞价；禁止连续挂牌；禁止收益承诺', 'APR-DC-2026-002', '等待审批流通过',
        2, 2, 2, 0, 1, 1, 0, '',
        '', '', 0, 1,
        '', '', 1, '2026-04-01 10:00:00', 1, '2026-04-18 18:00:00', 0
    )
ON DUPLICATE KEY UPDATE
    `activity_code` = VALUES(`activity_code`),
    `name` = VALUES(`name`),
    `rule_summary` = VALUES(`rule_summary`),
    `start_time` = VALUES(`start_time`),
    `end_time` = VALUES(`end_time`),
    `real_name_required` = VALUES(`real_name_required`),
    `participant_condition_summary` = VALUES(`participant_condition_summary`),
    `consume_rule_summary` = VALUES(`consume_rule_summary`),
    `probability_rule` = VALUES(`probability_rule`),
    `compliance_rule_summary` = VALUES(`compliance_rule_summary`),
    `circulation_limit_summary` = VALUES(`circulation_limit_summary`),
    `approval_record_ref` = VALUES(`approval_record_ref`),
    `publish_failure_summary` = VALUES(`publish_failure_summary`),
    `publish_readiness` = VALUES(`publish_readiness`),
    `copyright_status` = VALUES(`copyright_status`),
    `content_audit_status` = VALUES(`content_audit_status`),
    `status` = VALUES(`status`),
    `audit_status` = VALUES(`audit_status`),
    `is_enabled` = VALUES(`is_enabled`),
    `show_on_home` = VALUES(`show_on_home`),
    `home_entry_title` = VALUES(`home_entry_title`),
    `home_entry_subtitle` = VALUES(`home_entry_subtitle`),
    `home_entry_image` = VALUES(`home_entry_image`),
    `home_entry_sort` = VALUES(`home_entry_sort`),
    `home_entry_enabled` = VALUES(`home_entry_enabled`),
    `landing_target_type` = VALUES(`landing_target_type`),
    `landing_target_value` = VALUES(`landing_target_value`),
    `update_by` = VALUES(`update_by`),
    `update_time` = VALUES(`update_time`),
    `is_deleted` = VALUES(`is_deleted`);

INSERT INTO `sms_draw_pool` (
    `id`, `activity_id`, `platform_id`, `tenant_id`, `merchant_id`, `pool_code`, `pool_name`,
    `probability_rule`, `sort`, `status`, `audit_status`, `create_by`, `create_time`,
    `update_by`, `update_time`, `is_deleted`
)
VALUES
    (930001, 910001, 1, 10, 88, 'SPRING-MAIN', '春季主卡池', 'SSR 25%，SR 75%', 1, 1, 2, 1, '2026-03-01 09:05:00', 1, '2026-04-18 18:00:00', 0),
    (930002, 910002, 1, 10, 0, 'NEWBIE-MAIN', '新手欢迎卡池', '必中欢迎卡 100%', 1, 0, 1, 1, '2026-04-01 10:05:00', 1, '2026-04-18 18:00:00', 0)
ON DUPLICATE KEY UPDATE
    `pool_code` = VALUES(`pool_code`),
    `pool_name` = VALUES(`pool_name`),
    `probability_rule` = VALUES(`probability_rule`),
    `sort` = VALUES(`sort`),
    `status` = VALUES(`status`),
    `audit_status` = VALUES(`audit_status`),
    `update_by` = VALUES(`update_by`),
    `update_time` = VALUES(`update_time`),
    `is_deleted` = VALUES(`is_deleted`);

INSERT INTO `sms_card_template` (
    `id`, `platform_id`, `tenant_id`, `merchant_id`, `template_code`, `template_name`,
    `card_face_image`, `copyright_owner`, `copyright_proof_summary`, `rarity`,
    `issue_limit`, `display_copy`, `circulation_limit_summary`, `display_status`,
    `content_audit_status`, `provider_code`, `credential_ref`, `status`, `audit_status`,
    `create_by`, `create_time`, `update_by`, `update_time`, `is_deleted`
)
VALUES
    (
        920001, 1, 10, 88, 'TPL-SSR-RABBIT', 'SSR 银河兔',
        'https://img.example.com/digital-card/tpl-ssr-rabbit.png', '零号实验室',
        '版权登记号 CP-SSR-RABBIT-2026', 'SSR', 500,
        '春季限定 SSR 卡面，适合活动头部奖励展示', '禁止集中竞价；禁止连续挂牌；禁止收益承诺',
        1, 2, 'antchain-demo', 'cred:ssr-rabbit', 1, 2,
        1, '2026-03-01 09:10:00', 1, '2026-04-18 18:00:00', 0
    ),
    (
        920002, 1, 10, 88, 'TPL-SR-FOX', 'SR 星云狐',
        'https://img.example.com/digital-card/tpl-sr-fox.png', '零号实验室',
        '版权登记号 CP-SR-FOX-2026', 'SR', 1500,
        '春季常规 SR 奖励，面向大部分中奖会员', '禁止集中竞价；禁止连续挂牌；禁止收益承诺',
        1, 2, 'antchain-demo', 'cred:sr-fox', 1, 2,
        1, '2026-03-01 09:12:00', 1, '2026-04-18 18:00:00', 0
    ),
    (
        920003, 1, 10, 0, 'TPL-NEWBIE-BADGE', '新手欢迎徽章',
        'https://img.example.com/digital-card/tpl-newbie-badge.png', '零号实验室',
        '版权登记号 CP-NEWBIE-BADGE-2026', 'N', 9999,
        '新会员欢迎徽章，作为租户级入门权益使用', '禁止集中竞价；禁止连续挂牌；禁止收益承诺',
        1, 2, 'antchain-demo', 'cred:newbie-badge', 1, 2,
        1, '2026-04-01 10:10:00', 1, '2026-04-18 18:00:00', 0
    )
ON DUPLICATE KEY UPDATE
    `template_code` = VALUES(`template_code`),
    `template_name` = VALUES(`template_name`),
    `card_face_image` = VALUES(`card_face_image`),
    `copyright_owner` = VALUES(`copyright_owner`),
    `copyright_proof_summary` = VALUES(`copyright_proof_summary`),
    `rarity` = VALUES(`rarity`),
    `issue_limit` = VALUES(`issue_limit`),
    `display_copy` = VALUES(`display_copy`),
    `circulation_limit_summary` = VALUES(`circulation_limit_summary`),
    `display_status` = VALUES(`display_status`),
    `content_audit_status` = VALUES(`content_audit_status`),
    `provider_code` = VALUES(`provider_code`),
    `credential_ref` = VALUES(`credential_ref`),
    `status` = VALUES(`status`),
    `audit_status` = VALUES(`audit_status`),
    `update_by` = VALUES(`update_by`),
    `update_time` = VALUES(`update_time`),
    `is_deleted` = VALUES(`is_deleted`);

INSERT INTO `sms_draw_pool_template` (
    `id`, `activity_id`, `pool_id`, `template_id`, `platform_id`, `tenant_id`, `merchant_id`,
    `rarity`, `probability`, `sale_limit`, `remaining_limit`, `config_limit`, `status`,
    `audit_status`, `create_by`, `create_time`, `update_by`, `update_time`, `is_deleted`
)
VALUES
    (940001, 910001, 930001, 920001, 1, 10, 88, 'SSR', 0.250000, 500, 497, 500, 1, 2, 1, '2026-03-01 09:20:00', 1, '2026-04-18 18:00:00', 0),
    (940002, 910001, 930001, 920002, 1, 10, 88, 'SR', 0.750000, 1500, 1497, 1500, 1, 2, 1, '2026-03-01 09:21:00', 1, '2026-04-18 18:00:00', 0),
    (940003, 910002, 930002, 920003, 1, 10, 0, 'N', 1.000000, 9999, 9999, 9999, 0, 1, 1, '2026-04-01 10:20:00', 1, '2026-04-18 18:00:00', 0)
ON DUPLICATE KEY UPDATE
    `rarity` = VALUES(`rarity`),
    `probability` = VALUES(`probability`),
    `sale_limit` = VALUES(`sale_limit`),
    `remaining_limit` = VALUES(`remaining_limit`),
    `config_limit` = VALUES(`config_limit`),
    `status` = VALUES(`status`),
    `audit_status` = VALUES(`audit_status`),
    `update_by` = VALUES(`update_by`),
    `update_time` = VALUES(`update_time`),
    `is_deleted` = VALUES(`is_deleted`);

INSERT INTO `sms_draw_activity_audit` (
    `id`, `activity_id`, `platform_id`, `tenant_id`, `merchant_id`, `operation_type`,
    `operator_id`, `operator_name`, `approval_result`, `rule_snapshot`,
    `circulation_limit_snapshot`, `failure_summary`, `trace_id`, `payload_json`,
    `status`, `audit_status`, `create_by`, `create_time`
)
VALUES
    (
        950001, 910001, 1, 10, 88, 'create', 1, 'admin', 'approved',
        '实名 + 完成新手任务 + 每次消耗 1 次抽卡次数',
        '禁止集中竞价；禁止连续挂牌；禁止收益承诺',
        '创建活动并提交合规材料', 'trace-seed-activity-001',
        JSON_OBJECT('activityId', 910001, 'action', 'create', 'operator', 'admin'),
        1, 2, 1, '2026-03-01 09:30:00'
    ),
    (
        950002, 910001, 1, 10, 88, 'publish', 1, 'admin', 'approved',
        '发布前审计通过，模板、卡池和首页入口均已配置完成',
        '禁止集中竞价；禁止连续挂牌；禁止收益承诺',
        '发布预检通过并成功上线', 'trace-seed-activity-002',
        JSON_OBJECT('activityId', 910001, 'action', 'publish', 'operator', 'admin'),
        1, 2, 1, '2026-03-01 09:45:00'
    ),
    (
        950003, 910002, 1, 10, 0, 'create', 1, 'admin', 'pending',
        '新会员欢迎抽卡等待租户侧审批流补齐',
        '禁止集中竞价；禁止连续挂牌；禁止收益承诺',
        '等待审批流通过', 'trace-seed-activity-003',
        JSON_OBJECT('activityId', 910002, 'action', 'create', 'operator', 'admin'),
        0, 1, 1, '2026-04-01 10:30:00'
    )
ON DUPLICATE KEY UPDATE
    `operation_type` = VALUES(`operation_type`),
    `operator_id` = VALUES(`operator_id`),
    `operator_name` = VALUES(`operator_name`),
    `approval_result` = VALUES(`approval_result`),
    `rule_snapshot` = VALUES(`rule_snapshot`),
    `circulation_limit_snapshot` = VALUES(`circulation_limit_snapshot`),
    `failure_summary` = VALUES(`failure_summary`),
    `trace_id` = VALUES(`trace_id`),
    `payload_json` = VALUES(`payload_json`),
    `status` = VALUES(`status`),
    `audit_status` = VALUES(`audit_status`),
    `create_by` = VALUES(`create_by`),
    `create_time` = VALUES(`create_time`);

INSERT INTO `sms_draw_participation_record` (
    `id`, `activity_id`, `member_id`, `request_id`, `scope`, `eligibility_snapshot_json`,
    `consume_type`, `consume_amount`, `lottery_times_before`, `lottery_times_after`,
    `result_type`, `result_status`, `pool_id`, `template_id`, `rarity`, `trace_id`,
    `failure_code`, `failure_reason`, `asset_instance_id`, `asset_no`, `asset_status`,
    `asset_created_at`, `create_time`, `update_time`, `is_deleted`
)
VALUES
    (
        960001, 910001, 700001, 'drawreq-spring-001', 'merchant:1:10:88',
        JSON_OBJECT('realNameStatus', 'verified', 'newbieTaskCompleted', true),
        'lottery_times', 1, 5, 4, 'card_asset', 'won', 930001, 920001, 'SSR',
        'trace-seed-draw-001', '', '', 970001, 'ASSET-SPRING-001', 'asset_created',
        '2026-04-18 09:00:05', '2026-04-18 09:00:00', '2026-04-18 09:00:05', 0
    ),
    (
        960002, 910001, 700002, 'drawreq-spring-002', 'merchant:1:10:88',
        JSON_OBJECT('realNameStatus', 'verified', 'newbieTaskCompleted', true),
        'lottery_times', 1, 3, 2, 'card_asset', 'won', 930001, 920002, 'SR',
        'trace-seed-draw-002', '', '', 970002, 'ASSET-SPRING-002', 'asset_created',
        '2026-04-18 09:15:10', '2026-04-18 09:15:00', '2026-04-18 09:20:00', 0
    ),
    (
        960003, 910001, 700003, 'drawreq-spring-003', 'merchant:1:10:88',
        JSON_OBJECT('realNameStatus', 'verified', 'newbieTaskCompleted', true),
        'lottery_times', 1, 2, 1, 'card_asset', 'won', 930001, 920002, 'SR',
        'trace-seed-draw-003', '', '', 970003, 'ASSET-SPRING-003', 'asset_created',
        '2026-04-18 09:30:10', '2026-04-18 09:30:00', '2026-04-18 09:40:00', 0
    ),
    (
        960004, 910001, 700004, 'drawreq-spring-004', 'merchant:1:10:88',
        JSON_OBJECT('realNameStatus', 'verified', 'newbieTaskCompleted', true),
        'lottery_times', 1, 1, 0, 'card_asset', 'won', 930001, 920001, 'SSR',
        'trace-seed-draw-004', '', '', 970004, 'ASSET-SPRING-004', 'asset_created',
        '2026-04-18 09:45:10', '2026-04-18 09:45:00', '2026-04-18 09:50:00', 0
    )
ON DUPLICATE KEY UPDATE
    `scope` = VALUES(`scope`),
    `eligibility_snapshot_json` = VALUES(`eligibility_snapshot_json`),
    `consume_type` = VALUES(`consume_type`),
    `consume_amount` = VALUES(`consume_amount`),
    `lottery_times_before` = VALUES(`lottery_times_before`),
    `lottery_times_after` = VALUES(`lottery_times_after`),
    `result_type` = VALUES(`result_type`),
    `result_status` = VALUES(`result_status`),
    `pool_id` = VALUES(`pool_id`),
    `template_id` = VALUES(`template_id`),
    `rarity` = VALUES(`rarity`),
    `trace_id` = VALUES(`trace_id`),
    `failure_code` = VALUES(`failure_code`),
    `failure_reason` = VALUES(`failure_reason`),
    `asset_instance_id` = VALUES(`asset_instance_id`),
    `asset_no` = VALUES(`asset_no`),
    `asset_status` = VALUES(`asset_status`),
    `asset_created_at` = VALUES(`asset_created_at`),
    `update_time` = VALUES(`update_time`),
    `is_deleted` = VALUES(`is_deleted`);

INSERT INTO `sms_card_instance` (
    `id`, `platform_id`, `tenant_id`, `merchant_id`, `activity_id`, `member_id`,
    `participation_record_id`, `request_id`, `trace_id`, `scope`, `pool_id`,
    `template_id`, `rarity`, `asset_no`, `asset_status`, `mint_status`, `token_id`,
    `chain_status`, `last_receipt_at`, `mint_task_id`, `issued_at`, `create_by`,
    `create_time`, `update_by`, `update_time`, `display_status`, `compliance_status`,
    `display_reason`, `compliance_reason`, `rule_snapshot_json`, `disposed_at`,
    `disposed_by`, `is_deleted`
)
VALUES
    (
        970001, 1, 10, 88, 910001, 700001, 960001, 'drawreq-spring-001',
        'trace-seed-draw-001', 'merchant:1:10:88', 930001, 920001, 'SSR',
        'ASSET-SPRING-001', 'asset_created', 'mint_success', 'token-seed-0001',
        'success', '2026-04-18 09:00:20', 980001, '2026-04-18 09:00:05', 1,
        '2026-04-18 09:00:05', 1, '2026-04-18 09:00:20', 'display_visible',
        'compliance_clear', '', '', JSON_OBJECT('scenario', 'success', 'activityId', 910001),
        NULL, 0, 0
    ),
    (
        970002, 1, 10, 88, 910001, 700002, 960002, 'drawreq-spring-002',
        'trace-seed-draw-002', 'merchant:1:10:88', 930001, 920002, 'SR',
        'ASSET-SPRING-002', 'asset_created', 'mint_manual_review', '',
        'failed', '2026-04-18 09:20:00', 980002, '2026-04-18 09:15:10', 1,
        '2026-04-18 09:15:10', 1, '2026-04-18 09:20:00', 'display_hidden',
        'compliance_review', '回执信息待人工复核', '回执信息待人工复核',
        JSON_OBJECT('scenario', 'manual_review', 'activityId', 910001, 'lastReason', '回执缺少 tokenId'),
        NULL, 0, 0
    ),
    (
        970003, 1, 10, 88, 910001, 700003, 960003, 'drawreq-spring-003',
        'trace-seed-draw-003', 'merchant:1:10:88', 930001, 920002, 'SR',
        'ASSET-SPRING-003', 'asset_created', 'mint_compensating', '',
        'failed', NULL, 980003, '2026-04-18 09:30:10', 1,
        '2026-04-18 09:30:10', 1, '2026-04-18 09:40:00', 'display_offlined',
        'compliance_restricted', '链路失败，资产已下线展示', '链路失败，资产已下线展示',
        JSON_OBJECT('scenario', 'failed_compensating', 'activityId', 910001, 'lastReason', '演示失败补偿中'),
        NULL, 0, 0
    ),
    (
        970004, 1, 10, 88, 910001, 700004, 960004, 'drawreq-spring-004',
        'trace-seed-draw-004', 'merchant:1:10:88', 930001, 920001, 'SSR',
        'ASSET-SPRING-004', 'asset_created', 'mint_frozen', '',
        'frozen', NULL, 980004, '2026-04-18 09:45:10', 1,
        '2026-04-18 09:45:10', 1, '2026-04-18 09:50:00', 'display_hidden',
        'compliance_restricted', '人工冻结链路演示', '人工冻结链路演示',
        JSON_OBJECT('scenario', 'frozen', 'activityId', 910001, 'lastReason', '风控冻结演示'),
        NULL, 0, 0
    )
ON DUPLICATE KEY UPDATE
    `request_id` = VALUES(`request_id`),
    `trace_id` = VALUES(`trace_id`),
    `scope` = VALUES(`scope`),
    `pool_id` = VALUES(`pool_id`),
    `template_id` = VALUES(`template_id`),
    `rarity` = VALUES(`rarity`),
    `asset_no` = VALUES(`asset_no`),
    `asset_status` = VALUES(`asset_status`),
    `mint_status` = VALUES(`mint_status`),
    `token_id` = VALUES(`token_id`),
    `chain_status` = VALUES(`chain_status`),
    `last_receipt_at` = VALUES(`last_receipt_at`),
    `mint_task_id` = VALUES(`mint_task_id`),
    `issued_at` = VALUES(`issued_at`),
    `update_by` = VALUES(`update_by`),
    `update_time` = VALUES(`update_time`),
    `display_status` = VALUES(`display_status`),
    `compliance_status` = VALUES(`compliance_status`),
    `display_reason` = VALUES(`display_reason`),
    `compliance_reason` = VALUES(`compliance_reason`),
    `rule_snapshot_json` = VALUES(`rule_snapshot_json`),
    `disposed_at` = VALUES(`disposed_at`),
    `disposed_by` = VALUES(`disposed_by`),
    `is_deleted` = VALUES(`is_deleted`);

INSERT INTO `sms_card_mint_task` (
    `id`, `platform_id`, `tenant_id`, `merchant_id`, `asset_instance_id`,
    `participation_record_id`, `activity_id`, `member_id`, `request_id`, `trace_id`,
    `idempotency_key`, `task_status`, `mint_status`, `chain_status`, `token_id`,
    `chain_tx_id`, `retry_count`, `max_retry_count`, `last_error_code`,
    `last_error_reason`, `last_receipt_summary`, `last_receipt_json`, `last_execute_at`,
    `next_retry_at`, `manual_required`, `frozen`, `freeze_reason`, `create_by`,
    `update_by`, `create_time`, `update_time`, `is_deleted`
)
VALUES
    (
        980001, 1, 10, 88, 970001, 960001, 910001, 700001, 'drawreq-spring-001',
        'trace-seed-draw-001', 'mint-970001', 'succeeded', 'mint_success', 'success',
        'token-seed-0001', 'tx-seed-0001', 0, 3, '', '',
        '链上确权成功，token 已落库', JSON_OBJECT('txId', 'tx-seed-0001', 'tokenId', 'token-seed-0001', 'status', 'success'),
        '2026-04-18 09:00:20', NULL, 0, 0, '', 1, 1, '2026-04-18 09:00:05', '2026-04-18 09:00:20', 0
    ),
    (
        980002, 1, 10, 88, 970002, 960002, 910001, 700002, 'drawreq-spring-002',
        'trace-seed-draw-002', 'mint-970002', 'manual_review', 'mint_manual_review', 'failed',
        '', '', 1, 3, 'receipt_reconcile_required', '链上回执缺少 tokenId，待人工复核',
        '回执缺少 tokenId', JSON_OBJECT('status', 'failed', 'reason', 'receipt missing tokenId'),
        '2026-04-18 09:20:00', '2026-04-18 10:00:00', 1, 0, '', 1, 1, '2026-04-18 09:15:10', '2026-04-18 09:20:00', 0
    ),
    (
        980003, 1, 10, 88, 970003, 960003, 910001, 700003, 'drawreq-spring-003',
        'trace-seed-draw-003', 'mint-970003', 'failed', 'mint_compensating', 'failed',
        '', '', 2, 3, 'mint_execute_failed', '链上网关超时，等待人工触发补偿',
        '链上执行失败，等待补偿', JSON_OBJECT('status', 'failed', 'reason', 'gateway timeout'),
        '2026-04-18 09:40:00', '2026-04-18 10:10:00', 0, 0, '', 1, 1, '2026-04-18 09:30:10', '2026-04-18 09:40:00', 0
    ),
    (
        980004, 1, 10, 88, 970004, 960004, 910001, 700004, 'drawreq-spring-004',
        'trace-seed-draw-004', 'mint-970004', 'frozen', 'mint_frozen', 'frozen',
        '', '', 0, 3, '', '',
        '链路已冻结，等待人工解除', JSON_OBJECT('status', 'frozen', 'reason', 'manual freeze'),
        '2026-04-18 09:50:00', NULL, 0, 1, '风控冻结演示', 1, 1, '2026-04-18 09:45:10', '2026-04-18 09:50:00', 0
    )
ON DUPLICATE KEY UPDATE
    `request_id` = VALUES(`request_id`),
    `trace_id` = VALUES(`trace_id`),
    `idempotency_key` = VALUES(`idempotency_key`),
    `task_status` = VALUES(`task_status`),
    `mint_status` = VALUES(`mint_status`),
    `chain_status` = VALUES(`chain_status`),
    `token_id` = VALUES(`token_id`),
    `chain_tx_id` = VALUES(`chain_tx_id`),
    `retry_count` = VALUES(`retry_count`),
    `max_retry_count` = VALUES(`max_retry_count`),
    `last_error_code` = VALUES(`last_error_code`),
    `last_error_reason` = VALUES(`last_error_reason`),
    `last_receipt_summary` = VALUES(`last_receipt_summary`),
    `last_receipt_json` = VALUES(`last_receipt_json`),
    `last_execute_at` = VALUES(`last_execute_at`),
    `next_retry_at` = VALUES(`next_retry_at`),
    `manual_required` = VALUES(`manual_required`),
    `frozen` = VALUES(`frozen`),
    `freeze_reason` = VALUES(`freeze_reason`),
    `update_by` = VALUES(`update_by`),
    `update_time` = VALUES(`update_time`),
    `is_deleted` = VALUES(`is_deleted`);

INSERT INTO `sms_card_asset_log` (
    `id`, `asset_instance_id`, `participation_record_id`, `from_status`, `to_status`,
    `operation_type`, `operator_type`, `trace_id`, `reason_code`, `reason_text`,
    `payload_json`, `create_time`
)
VALUES
    (990001, 970001, 960001, '', 'mint_pending', 'mint_requested', 'system', 'trace-seed-draw-001', '', '已创建链上发放任务', JSON_OBJECT('taskId', 980001), '2026-04-18 09:00:05'),
    (990002, 970001, 960001, 'mint_pending', 'mint_processing', 'mint_dispatching', 'system', 'trace-seed-draw-001', '', '任务已派发到链路执行器', JSON_OBJECT('taskId', 980001), '2026-04-18 09:00:10'),
    (990003, 970001, 960001, 'mint_processing', 'mint_success', 'mint_succeeded', 'system', 'trace-seed-draw-001', '', '链上发放成功', JSON_OBJECT('taskId', 980001, 'tokenId', 'token-seed-0001'), '2026-04-18 09:00:20'),
    (990011, 970002, 960002, '', 'mint_pending', 'mint_requested', 'system', 'trace-seed-draw-002', '', '已创建链上发放任务', JSON_OBJECT('taskId', 980002), '2026-04-18 09:15:10'),
    (990012, 970002, 960002, 'mint_pending', 'mint_manual_review', 'mint_manual_review', 'system', 'trace-seed-draw-002', 'receipt_reconcile_required', '链上回执缺少 tokenId，升级人工复核', JSON_OBJECT('taskId', 980002), '2026-04-18 09:20:00'),
    (990013, 970002, 960002, 'compliance_clear', 'compliance_review', 'asset_compliance_review', 'manual', 'trace-seed-draw-002', '', '运营已标记人工复核', JSON_OBJECT('assetInstanceId', 970002), '2026-04-18 09:25:00'),
    (990021, 970003, 960003, '', 'mint_pending', 'mint_requested', 'system', 'trace-seed-draw-003', '', '已创建链上发放任务', JSON_OBJECT('taskId', 980003), '2026-04-18 09:30:10'),
    (990022, 970003, 960003, 'mint_pending', 'mint_compensating', 'mint_failed', 'system', 'trace-seed-draw-003', 'mint_execute_failed', '链上网关超时，进入补偿状态', JSON_OBJECT('taskId', 980003), '2026-04-18 09:40:00'),
    (990023, 970003, 960003, 'compliance_clear', 'compliance_restricted', 'asset_offlined', 'manual', 'trace-seed-draw-003', '', '运营已将异常资产下线展示', JSON_OBJECT('assetInstanceId', 970003), '2026-04-18 09:42:00'),
    (990031, 970004, 960004, '', 'mint_pending', 'mint_requested', 'system', 'trace-seed-draw-004', '', '已创建链上发放任务', JSON_OBJECT('taskId', 980004), '2026-04-18 09:45:10'),
    (990032, 970004, 960004, 'mint_pending', 'mint_frozen', 'mint_frozen', 'manual', 'trace-seed-draw-004', '', '风控命中，链路已冻结', JSON_OBJECT('taskId', 980004), '2026-04-18 09:50:00')
ON DUPLICATE KEY UPDATE
    `from_status` = VALUES(`from_status`),
    `to_status` = VALUES(`to_status`),
    `operation_type` = VALUES(`operation_type`),
    `operator_type` = VALUES(`operator_type`),
    `trace_id` = VALUES(`trace_id`),
    `reason_code` = VALUES(`reason_code`),
    `reason_text` = VALUES(`reason_text`),
    `payload_json` = VALUES(`payload_json`),
    `create_time` = VALUES(`create_time`);
