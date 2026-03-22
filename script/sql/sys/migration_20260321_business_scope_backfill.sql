-- Story 1.6A business scope truth-source migration.
-- Safe to execute multiple times on an existing environment.

DROP PROCEDURE IF EXISTS add_column_if_missing;
DROP PROCEDURE IF EXISTS add_index_if_missing;

DELIMITER $$

CREATE PROCEDURE add_column_if_missing(
    IN p_table VARCHAR(64),
    IN p_column VARCHAR(64),
    IN p_definition TEXT
)
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.COLUMNS
        WHERE TABLE_SCHEMA = DATABASE()
          AND TABLE_NAME = p_table
          AND COLUMN_NAME = p_column
    ) THEN
        SET @sql = CONCAT('ALTER TABLE `', p_table, '` ADD COLUMN ', p_definition);
        PREPARE stmt FROM @sql;
        EXECUTE stmt;
        DEALLOCATE PREPARE stmt;
    END IF;
END $$

CREATE PROCEDURE add_index_if_missing(
    IN p_table VARCHAR(64),
    IN p_index VARCHAR(64),
    IN p_statement TEXT
)
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.STATISTICS
        WHERE TABLE_SCHEMA = DATABASE()
          AND TABLE_NAME = p_table
          AND INDEX_NAME = p_index
    ) THEN
        SET @sql = p_statement;
        PREPARE stmt FROM @sql;
        EXECUTE stmt;
        DEALLOCATE PREPARE stmt;
    END IF;
END $$

DELIMITER ;

CREATE TABLE IF NOT EXISTS sys_scope_backfill_issue
(
    id           bigint auto_increment comment '主键'
        primary key,
    source_table varchar(64)                         not null comment '来源表',
    source_id    bigint                              not null comment '来源主键',
    issue_code   varchar(64)                         not null comment '问题编码',
    issue_detail varchar(255) default ''             not null comment '问题说明',
    created_at   timestamp    default CURRENT_TIMESTAMP not null comment '创建时间',
    constraint uk_scope_backfill_issue
        unique (source_table, source_id, issue_code)
) comment '主体范围历史回填异常记录';

CALL add_column_if_missing('pms_product_spu', 'platform_id', '`platform_id` bigint DEFAULT 1 NOT NULL COMMENT ''平台ID'' AFTER `id`');
CALL add_column_if_missing('pms_product_spu', 'tenant_id', '`tenant_id` bigint DEFAULT 0 NOT NULL COMMENT ''租户ID'' AFTER `platform_id`');
CALL add_column_if_missing('pms_product_spu', 'merchant_id', '`merchant_id` bigint DEFAULT 0 NOT NULL COMMENT ''商户ID'' AFTER `tenant_id`');
CALL add_index_if_missing('pms_product_spu', 'idx_scope_publish_verify', 'CREATE INDEX `idx_scope_publish_verify` ON `pms_product_spu` (`platform_id`, `tenant_id`, `merchant_id`, `publish_status`, `verify_status`, `id`)');

CALL add_column_if_missing('pms_product_sku', 'platform_id', '`platform_id` bigint DEFAULT 1 NOT NULL COMMENT ''平台ID'' AFTER `id`');
CALL add_column_if_missing('pms_product_sku', 'tenant_id', '`tenant_id` bigint DEFAULT 0 NOT NULL COMMENT ''租户ID'' AFTER `platform_id`');
CALL add_column_if_missing('pms_product_sku', 'merchant_id', '`merchant_id` bigint DEFAULT 0 NOT NULL COMMENT ''商户ID'' AFTER `tenant_id`');
CALL add_index_if_missing('pms_product_sku', 'idx_scope_spu_status', 'CREATE INDEX `idx_scope_spu_status` ON `pms_product_sku` (`platform_id`, `tenant_id`, `merchant_id`, `spu_id`, `publish_status`, `verify_status`, `id`)');

CALL add_column_if_missing('oms_order_main', 'platform_id', '`platform_id` bigint DEFAULT 1 NOT NULL COMMENT ''平台ID'' AFTER `id`');
CALL add_column_if_missing('oms_order_main', 'tenant_id', '`tenant_id` bigint DEFAULT 0 NOT NULL COMMENT ''租户ID'' AFTER `platform_id`');
CALL add_column_if_missing('oms_order_main', 'merchant_id', '`merchant_id` bigint DEFAULT 0 NOT NULL COMMENT ''商户ID'' AFTER `tenant_id`');
CALL add_index_if_missing('oms_order_main', 'idx_scope_order_status', 'CREATE INDEX `idx_scope_order_status` ON `oms_order_main` (`platform_id`, `tenant_id`, `merchant_id`, `order_status`, `user_id`, `id`)');

CALL add_column_if_missing('oms_order_item', 'platform_id', '`platform_id` bigint DEFAULT 1 NOT NULL COMMENT ''平台ID'' AFTER `id`');
CALL add_column_if_missing('oms_order_item', 'tenant_id', '`tenant_id` bigint DEFAULT 0 NOT NULL COMMENT ''租户ID'' AFTER `platform_id`');
CALL add_column_if_missing('oms_order_item', 'merchant_id', '`merchant_id` bigint DEFAULT 0 NOT NULL COMMENT ''商户ID'' AFTER `tenant_id`');
CALL add_index_if_missing('oms_order_item', 'idx_scope_order', 'CREATE INDEX `idx_scope_order` ON `oms_order_item` (`platform_id`, `tenant_id`, `merchant_id`, `order_id`, `sku_id`)');

CALL add_column_if_missing('oms_order_delivery', 'platform_id', '`platform_id` bigint DEFAULT 1 NOT NULL COMMENT ''平台ID'' AFTER `id`');
CALL add_column_if_missing('oms_order_delivery', 'tenant_id', '`tenant_id` bigint DEFAULT 0 NOT NULL COMMENT ''租户ID'' AFTER `platform_id`');
CALL add_column_if_missing('oms_order_delivery', 'merchant_id', '`merchant_id` bigint DEFAULT 0 NOT NULL COMMENT ''商户ID'' AFTER `tenant_id`');
CALL add_index_if_missing('oms_order_delivery', 'idx_scope_delivery_order', 'CREATE INDEX `idx_scope_delivery_order` ON `oms_order_delivery` (`platform_id`, `tenant_id`, `merchant_id`, `order_id`)');

CALL add_column_if_missing('oms_order_payment', 'platform_id', '`platform_id` bigint DEFAULT 1 NOT NULL COMMENT ''平台ID'' AFTER `id`');
CALL add_column_if_missing('oms_order_payment', 'tenant_id', '`tenant_id` bigint DEFAULT 0 NOT NULL COMMENT ''租户ID'' AFTER `platform_id`');
CALL add_column_if_missing('oms_order_payment', 'merchant_id', '`merchant_id` bigint DEFAULT 0 NOT NULL COMMENT ''商户ID'' AFTER `tenant_id`');
CALL add_index_if_missing('oms_order_payment', 'idx_scope_payment_order', 'CREATE INDEX `idx_scope_payment_order` ON `oms_order_payment` (`platform_id`, `tenant_id`, `merchant_id`, `order_id`, `pay_status`)');

CALL add_column_if_missing('oms_order_promotion', 'platform_id', '`platform_id` bigint DEFAULT 1 NOT NULL COMMENT ''平台ID'' AFTER `id`');
CALL add_column_if_missing('oms_order_promotion', 'tenant_id', '`tenant_id` bigint DEFAULT 0 NOT NULL COMMENT ''租户ID'' AFTER `platform_id`');
CALL add_column_if_missing('oms_order_promotion', 'merchant_id', '`merchant_id` bigint DEFAULT 0 NOT NULL COMMENT ''商户ID'' AFTER `tenant_id`');
CALL add_index_if_missing('oms_order_promotion', 'idx_scope_promotion_order', 'CREATE INDEX `idx_scope_promotion_order` ON `oms_order_promotion` (`platform_id`, `tenant_id`, `merchant_id`, `order_id`, `promotion_type`)');

CALL add_column_if_missing('oms_order_operation_log', 'platform_id', '`platform_id` bigint DEFAULT 1 NOT NULL COMMENT ''平台ID'' AFTER `id`');
CALL add_column_if_missing('oms_order_operation_log', 'tenant_id', '`tenant_id` bigint DEFAULT 0 NOT NULL COMMENT ''租户ID'' AFTER `platform_id`');
CALL add_column_if_missing('oms_order_operation_log', 'merchant_id', '`merchant_id` bigint DEFAULT 0 NOT NULL COMMENT ''商户ID'' AFTER `tenant_id`');
CALL add_index_if_missing('oms_order_operation_log', 'idx_scope_order_operation', 'CREATE INDEX `idx_scope_order_operation` ON `oms_order_operation_log` (`platform_id`, `tenant_id`, `merchant_id`, `order_id`, `operation_type`)');

CALL add_column_if_missing('sms_coupon', 'platform_id', '`platform_id` bigint DEFAULT 1 NOT NULL COMMENT ''平台ID'' AFTER `id`');
CALL add_column_if_missing('sms_coupon', 'tenant_id', '`tenant_id` bigint DEFAULT 0 NOT NULL COMMENT ''租户ID'' AFTER `platform_id`');
CALL add_column_if_missing('sms_coupon', 'merchant_id', '`merchant_id` bigint DEFAULT 0 NOT NULL COMMENT ''商户ID'' AFTER `tenant_id`');
CALL add_index_if_missing('sms_coupon', 'idx_coupon_scope_status', 'CREATE INDEX `idx_coupon_scope_status` ON `sms_coupon` (`platform_id`, `tenant_id`, `merchant_id`, `status`, `is_enabled`, `id`)');

CALL add_column_if_missing('cms_subject', 'platform_id', '`platform_id` bigint DEFAULT 1 NOT NULL COMMENT ''平台ID'' AFTER `id`');
CALL add_column_if_missing('cms_subject', 'tenant_id', '`tenant_id` bigint DEFAULT 0 NOT NULL COMMENT ''租户ID'' AFTER `platform_id`');
CALL add_column_if_missing('cms_subject', 'merchant_id', '`merchant_id` bigint DEFAULT 0 NOT NULL COMMENT ''商户ID'' AFTER `tenant_id`');
CALL add_index_if_missing('cms_subject', 'idx_subject_scope_status', 'CREATE INDEX `idx_subject_scope_status` ON `cms_subject` (`platform_id`, `tenant_id`, `merchant_id`, `show_status`, `recommend_status`, `id`)');

CALL add_column_if_missing('cms_subject_product_relation', 'platform_id', '`platform_id` bigint DEFAULT 1 NOT NULL COMMENT ''平台ID'' AFTER `id`');
CALL add_column_if_missing('cms_subject_product_relation', 'tenant_id', '`tenant_id` bigint DEFAULT 0 NOT NULL COMMENT ''租户ID'' AFTER `platform_id`');
CALL add_column_if_missing('cms_subject_product_relation', 'merchant_id', '`merchant_id` bigint DEFAULT 0 NOT NULL COMMENT ''商户ID'' AFTER `tenant_id`');
CALL add_index_if_missing('cms_subject_product_relation', 'idx_subject_relation_scope_subject', 'CREATE INDEX `idx_subject_relation_scope_subject` ON `cms_subject_product_relation` (`platform_id`, `tenant_id`, `merchant_id`, `subject_id`)');
CALL add_index_if_missing('cms_subject_product_relation', 'idx_subject_relation_scope_product', 'CREATE INDEX `idx_subject_relation_scope_product` ON `cms_subject_product_relation` (`platform_id`, `tenant_id`, `merchant_id`, `product_id`)');

UPDATE pms_product_spu SET platform_id = 0, tenant_id = 0, merchant_id = 0;
UPDATE pms_product_sku SET platform_id = 0, tenant_id = 0, merchant_id = 0;
UPDATE oms_order_item SET platform_id = 0, tenant_id = 0, merchant_id = 0;
UPDATE oms_order_main SET platform_id = 0, tenant_id = 0, merchant_id = 0;
UPDATE oms_order_delivery SET platform_id = 0, tenant_id = 0, merchant_id = 0;
UPDATE oms_order_payment SET platform_id = 0, tenant_id = 0, merchant_id = 0;
UPDATE oms_order_promotion SET platform_id = 0, tenant_id = 0, merchant_id = 0;
UPDATE oms_order_operation_log SET platform_id = 0, tenant_id = 0, merchant_id = 0;
UPDATE sms_coupon SET platform_id = 0, tenant_id = 0, merchant_id = 0;
UPDATE cms_subject SET platform_id = 0, tenant_id = 0, merchant_id = 0;
UPDATE cms_subject_product_relation SET platform_id = 0, tenant_id = 0, merchant_id = 0;

UPDATE pms_product_spu spu
LEFT JOIN sys_user u ON u.id = spu.create_by
SET spu.platform_id = COALESCE(u.platform_id, 0),
    spu.tenant_id = COALESCE(u.tenant_id, 0),
    spu.merchant_id = COALESCE(u.merchant_id, 0);

UPDATE pms_product_sku sku
LEFT JOIN pms_product_spu spu ON spu.id = sku.spu_id
LEFT JOIN sys_user u ON u.id = sku.create_by
SET sku.platform_id = CASE WHEN spu.platform_id > 0 THEN spu.platform_id ELSE COALESCE(u.platform_id, 0) END,
    sku.tenant_id = CASE WHEN spu.platform_id > 0 THEN spu.tenant_id ELSE COALESCE(u.tenant_id, 0) END,
    sku.merchant_id = CASE WHEN spu.platform_id > 0 THEN spu.merchant_id ELSE COALESCE(u.merchant_id, 0) END;

UPDATE sms_coupon coupon
LEFT JOIN sys_user u ON u.id = coupon.create_by
SET coupon.platform_id = COALESCE(u.platform_id, 0),
    coupon.tenant_id = COALESCE(u.tenant_id, 0),
    coupon.merchant_id = COALESCE(u.merchant_id, 0);

UPDATE cms_subject subject
LEFT JOIN sys_user u ON u.user_name = subject.create_by
SET subject.platform_id = COALESCE(u.platform_id, 0),
    subject.tenant_id = COALESCE(u.tenant_id, 0),
    subject.merchant_id = COALESCE(u.merchant_id, 0);

UPDATE cms_subject_product_relation rel
LEFT JOIN cms_subject subject ON subject.id = rel.subject_id
LEFT JOIN pms_product_spu spu ON spu.id = rel.product_id
SET rel.platform_id = CASE WHEN subject.platform_id > 0 THEN subject.platform_id ELSE COALESCE(spu.platform_id, 0) END,
    rel.tenant_id = CASE WHEN subject.platform_id > 0 THEN subject.tenant_id ELSE COALESCE(spu.tenant_id, 0) END,
    rel.merchant_id = CASE WHEN subject.platform_id > 0 THEN subject.merchant_id ELSE COALESCE(spu.merchant_id, 0) END;

UPDATE oms_order_item item
LEFT JOIN pms_product_sku sku ON sku.id = item.sku_id
SET item.platform_id = COALESCE(sku.platform_id, 0),
    item.tenant_id = COALESCE(sku.tenant_id, 0),
    item.merchant_id = COALESCE(sku.merchant_id, 0);

UPDATE oms_order_main main_order
JOIN (
    SELECT order_id,
           MAX(platform_id) AS platform_id,
           MAX(tenant_id) AS tenant_id,
           MAX(merchant_id) AS merchant_id
    FROM oms_order_item
    WHERE platform_id > 0
    GROUP BY order_id
    HAVING COUNT(DISTINCT CONCAT_WS(':', platform_id, tenant_id, merchant_id)) = 1
) scoped_item ON scoped_item.order_id = main_order.id
SET main_order.platform_id = scoped_item.platform_id,
    main_order.tenant_id = scoped_item.tenant_id,
    main_order.merchant_id = scoped_item.merchant_id;

UPDATE oms_order_delivery delivery
LEFT JOIN oms_order_main main_order ON main_order.id = delivery.order_id
SET delivery.platform_id = COALESCE(main_order.platform_id, 0),
    delivery.tenant_id = COALESCE(main_order.tenant_id, 0),
    delivery.merchant_id = COALESCE(main_order.merchant_id, 0);

UPDATE oms_order_payment payment
LEFT JOIN oms_order_main main_order ON main_order.id = payment.order_id
SET payment.platform_id = COALESCE(main_order.platform_id, 0),
    payment.tenant_id = COALESCE(main_order.tenant_id, 0),
    payment.merchant_id = COALESCE(main_order.merchant_id, 0);

UPDATE oms_order_promotion promotion
LEFT JOIN oms_order_main main_order ON main_order.id = promotion.order_id
SET promotion.platform_id = COALESCE(main_order.platform_id, 0),
    promotion.tenant_id = COALESCE(main_order.tenant_id, 0),
    promotion.merchant_id = COALESCE(main_order.merchant_id, 0);

UPDATE oms_order_operation_log operation_log
LEFT JOIN oms_order_main main_order ON main_order.id = operation_log.order_id
SET operation_log.platform_id = COALESCE(main_order.platform_id, 0),
    operation_log.tenant_id = COALESCE(main_order.tenant_id, 0),
    operation_log.merchant_id = COALESCE(main_order.merchant_id, 0);

INSERT IGNORE INTO sys_scope_backfill_issue (source_table, source_id, issue_code, issue_detail)
SELECT 'pms_product_spu', id, 'missing_creator_scope', '无法通过 create_by -> sys_user 回填商品主体'
FROM pms_product_spu
WHERE platform_id = 0;

INSERT IGNORE INTO sys_scope_backfill_issue (source_table, source_id, issue_code, issue_detail)
SELECT 'pms_product_sku', id, 'missing_spu_scope', '无法通过 spu/create_by 回填 SKU 主体'
FROM pms_product_sku
WHERE platform_id = 0;

INSERT IGNORE INTO sys_scope_backfill_issue (source_table, source_id, issue_code, issue_detail)
SELECT 'sms_coupon', id, 'missing_creator_scope', '无法通过 create_by -> sys_user 回填优惠券主体'
FROM sms_coupon
WHERE platform_id = 0;

INSERT IGNORE INTO sys_scope_backfill_issue (source_table, source_id, issue_code, issue_detail)
SELECT 'cms_subject', id, 'missing_creator_scope', '无法通过 create_by -> sys_user.user_name 回填专题主体'
FROM cms_subject
WHERE platform_id = 0;

INSERT IGNORE INTO sys_scope_backfill_issue (source_table, source_id, issue_code, issue_detail)
SELECT 'cms_subject_product_relation', rel.id, 'subject_product_scope_mismatch', '专题与商品主体不一致，关系记录需要人工校准'
FROM cms_subject_product_relation rel
JOIN cms_subject subject ON subject.id = rel.subject_id
JOIN pms_product_spu spu ON spu.id = rel.product_id
WHERE subject.platform_id > 0
  AND spu.platform_id > 0
  AND (subject.platform_id <> spu.platform_id OR subject.tenant_id <> spu.tenant_id OR subject.merchant_id <> spu.merchant_id);

INSERT IGNORE INTO sys_scope_backfill_issue (source_table, source_id, issue_code, issue_detail)
SELECT 'oms_order_item', id, 'missing_sku_scope', '无法通过 sku_id -> pms_product_sku 回填订单项主体'
FROM oms_order_item
WHERE platform_id = 0;

INSERT IGNORE INTO sys_scope_backfill_issue (source_table, source_id, issue_code, issue_detail)
SELECT 'oms_order_main', main_order.id, 'mixed_or_missing_item_scope', '订单项主体缺失或存在跨主体混单，需要人工拆分/补数后再启用隔离'
FROM oms_order_main main_order
LEFT JOIN (
    SELECT order_id,
           COUNT(*) AS item_count,
           SUM(CASE WHEN platform_id > 0 THEN 1 ELSE 0 END) AS scoped_count,
           COUNT(DISTINCT CONCAT_WS(':', platform_id, tenant_id, merchant_id)) AS scope_versions
    FROM oms_order_item
    GROUP BY order_id
) agg ON agg.order_id = main_order.id
WHERE main_order.platform_id = 0
   OR agg.item_count IS NULL
   OR agg.scoped_count <> agg.item_count
   OR agg.scope_versions <> 1;

INSERT IGNORE INTO sys_scope_backfill_issue (source_table, source_id, issue_code, issue_detail)
SELECT 'oms_order_delivery', id, 'missing_order_scope', '无法通过 order_id -> oms_order_main 回填收货地址主体'
FROM oms_order_delivery
WHERE platform_id = 0;

INSERT IGNORE INTO sys_scope_backfill_issue (source_table, source_id, issue_code, issue_detail)
SELECT 'oms_order_payment', id, 'missing_order_scope', '无法通过 order_id -> oms_order_main 回填支付主体'
FROM oms_order_payment
WHERE platform_id = 0;

INSERT IGNORE INTO sys_scope_backfill_issue (source_table, source_id, issue_code, issue_detail)
SELECT 'oms_order_promotion', id, 'missing_order_scope', '无法通过 order_id -> oms_order_main 回填优惠主体'
FROM oms_order_promotion
WHERE platform_id = 0;

INSERT IGNORE INTO sys_scope_backfill_issue (source_table, source_id, issue_code, issue_detail)
SELECT 'oms_order_operation_log', id, 'missing_order_scope', '无法通过 order_id -> oms_order_main 回填操作日志主体'
FROM oms_order_operation_log
WHERE platform_id = 0;

DROP PROCEDURE IF EXISTS add_column_if_missing;
DROP PROCEDURE IF EXISTS add_index_if_missing;
