-- Story 10.11 / Task 11.1：重置三张遗留 manual_review 卡片，让 sms-rpc 自循环扫描重派、走真链 mint。
--
-- 背景：
--   - asset_instance_id IN (970012, 970016, 970017) 三张卡在阶段一免费链 mock 时代
--     卡死在 manual_review，错误码包含「免费链能力未启用」/「mint_prerequisite_rejected」。
--   - Story 10.11 真实 FISCO BCOS 3.x 上线后，需要把这三张卡解锁为 pending_dispatch，
--     等 sms-rpc startMintTaskSelfScan(每 30s 跑一次) 自动捡起，走真链 mint。
--
-- 设计原则：
--   - 不删除任何历史日志（sms_card_asset_log 全部保留），只重置任务表 + 实例表的状态字段。
--   - 不动 idempotency_key（保持原值，让链上 _mintedByKey 仍然能去重）。
--   - 显式清空 chain_tx_id / token_id（避免 transitionSuccess 写库时撞上历史脏值）。
--
-- 验证：
--   1) 执行后 30 秒内观察 sms-rpc 日志：
--        "sms-rpc 自循环扫描提货卡任务: dispatched=3 ..."
--   2) 等 5 分钟，查询：
--        SELECT id, asset_instance_id, task_status, mint_status, chain_tx_id, token_id
--          FROM sms_card_mint_task
--         WHERE asset_instance_id IN (970012, 970016, 970017);
--      预期：task_status=succeeded, mint_status=mint_success, chain_tx_id 64 hex, token_id 非空。
--   3) 手机端打开三张卡详情页，时间线只显示「已到账」，不再有「补偿中 / asset_created」字样。
--
-- 回滚：本迁移属于「数据修复」性质，无需回滚。如果回滚需要，从备份恢复。

START TRANSACTION;

-- 1. 重置任务表
UPDATE sms_card_mint_task
   SET task_status      = 'pending_dispatch',
       mint_status      = 'mint_compensating',
       chain_status     = 'unknown',
       manual_required  = 0,
       frozen           = 0,
       freeze_reason    = '',
       retry_count      = 0,
       last_error_code  = '',
       last_error_reason= '',
       last_execute_at  = NULL,
       next_retry_at    = NULL,
       token_id         = '',
       chain_tx_id      = '',
       last_receipt_summary = '',
       last_receipt_json    = '',
       update_time      = NOW()
 WHERE asset_instance_id IN (970012, 970016, 970017)
   AND is_deleted = 0;

-- 2. 重置卡片实例表（解锁 manual_review，让 UI 不再锁死）
UPDATE sms_card_instance
   SET mint_status  = 'mint_compensating',
       chain_status = 'unknown',
       update_time  = NOW()
 WHERE id IN (970012, 970016, 970017)
   AND is_deleted = 0;

-- 3. 追加一条审计日志（便于 BI / 审计回溯）
INSERT INTO sms_card_asset_log
    (asset_instance_id, participation_record_id, from_status, to_status,
     operation_type, operator_type, trace_id, reason_code, reason_text,
     payload_json, create_time)
SELECT id,
       participation_record_id,
       mint_status,
       'mint_compensating',
       'mint_retry_requested',
       'manual',
       CONCAT('story-10-11-reset-', UNIX_TIMESTAMP()),
       'story_10_11_legacy_reset',
       'Story 10.11: 重置 mock 时代遗留 manual_review 状态，由 sms-rpc 自循环扫描走真链 FISCO BCOS 3.x mint',
       '{"story":"10.11","sourceErrorPattern":"free_chain_mock_legacy"}',
       NOW()
  FROM sms_card_instance
 WHERE id IN (970012, 970016, 970017)
   AND is_deleted = 0;

COMMIT;
