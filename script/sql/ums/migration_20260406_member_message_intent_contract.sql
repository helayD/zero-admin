ALTER TABLE ums_member_message
    ADD COLUMN IF NOT EXISTS intent_contract TEXT NOT NULL DEFAULT '{}' COMMENT '统一意图契约(JSON)' AFTER related_order_id;

UPDATE ums_member_message
SET intent_contract = '{}'
WHERE intent_contract IS NULL
   OR TRIM(intent_contract) = '';
