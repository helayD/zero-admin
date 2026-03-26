-- =============================================================================
-- Migration: 20260326 部署种子数据补齐
-- 涉及 Story: 4-5, 4-6, 5-1, 5-2
-- =============================================================================

-- 1. 补齐 cms_preferred_area 表（远程不存在）
CREATE TABLE IF NOT EXISTS cms_preferred_area
(
    id          bigint auto_increment
        primary key comment '主键ID',
    name        varchar(255)                          not null comment '专区名称',
    sub_title   varchar(255)                          not null comment '子标题',
    pic         varchar(500)                          not null comment '展示图片',
    sort        int                                   not null comment '排序',
    show_status tinyint                               not null comment '显示状态：0->不显示；1->显示',
    create_by   varchar(50)                           not null comment '创建者',
    create_time datetime    default CURRENT_TIMESTAMP not null comment '创建时间',
    update_by   varchar(50) default ''                not null comment '更新者',
    update_time datetime                              null on update CURRENT_TIMESTAMP comment '更新时间'
) comment '优选专区' charset = utf8mb4;

INSERT IGNORE INTO cms_preferred_area (id, name, sub_title, pic, sort, show_status, create_by)
VALUES (1, '让音质更出众', '音质不打折 完美现场感', '', 1, 1, 'admin'),
       (2, '让音质更出众22', '让音质更出众22', '', 2, 1, 'admin'),
       (3, '让音质更出众33', ' ', '', 3, 1, 'admin'),
       (4, '让音质更出众44', ' ', '', 4, 1, 'admin');

-- 2. 补齐 cms_preferred_area_product_relation 表
CREATE TABLE IF NOT EXISTS cms_preferred_area_product_relation
(
    id                bigint auto_increment
        primary key comment '主键ID',
    preferred_area_id bigint not null comment '优选专区ID',
    product_id        bigint not null comment '产品ID'
) comment '优选专区和产品关系表' charset = utf8mb4;

INSERT IGNORE INTO cms_preferred_area_product_relation (id, preferred_area_id, product_id)
VALUES (1, 1, 1), (2, 1, 2), (3, 1, 3), (4, 1, 4), (5, 1, 5),
       (6, 2, 3), (7, 2, 4);

-- 3. 更新秒杀活动时间（原数据已过期 2025 年）
UPDATE sms_seckill_activity SET
    start_time = '2026-01-01 00:00:00',
    end_time   = '2029-12-31 23:59:59',
    status     = 0,
    is_enabled = 1
WHERE id IN (1, 2, 3);

-- 确保至少一个秒杀活动是上线状态
UPDATE sms_seckill_activity SET status = 0, is_enabled = 1 WHERE id = 1;

-- 4. 补齐秒杀场次种子数据（如果不存在）
-- sms_seckill_session: start_time/end_time 是 varchar(12) 格式
INSERT IGNORE INTO sms_seckill_session (id, name, start_time, end_time, status, sort, create_by)
VALUES (1, '上午场', '08:00', '12:00', 1, 1, 1),
       (2, '下午场', '12:00', '18:00', 1, 2, 1),
       (3, '晚间场', '18:00', '23:59', 1, 3, 1);

-- 5. 补齐秒杀商品关联（如果不存在）
-- sms_seckill_product 实际字段: activity_id, session_id, sku_id, sku_name, seckill_price, seckill_stock, stock_locked, per_limit, sort, status, create_by
INSERT IGNORE INTO sms_seckill_product (id, activity_id, session_id, sku_id, sku_name, seckill_price, seckill_stock, stock_locked, per_limit, sort, status, create_by)
VALUES (1, 1, 1, 1, '小米手机 金色 128GB', 5999.00, 100, 0, 1, 1, 1, 1),
       (2, 1, 1, 3, '苹果手机 金色 128GB', 5999.00, 100, 0, 1, 2, 1, 1),
       (3, 1, 2, 5, '华为手机 金色 128GB', 5999.00, 100, 0, 1, 1, 1, 1);

-- 6. 确保购物车测试数据足够（补充更多购物车项）
INSERT IGNORE INTO oms_cart_item (id, member_id, product_id, product_sku_id, quantity, price, selected,
    product_name, product_sub_title, product_pic, product_sku_code, product_sn, product_brand,
    product_category_id, product_attr, member_nickname, source, delete_status, expire_time)
VALUES
    (2, 1001, 2, 3, 2, 7999.00, 1,
     'Apple/苹果 iPhone 17', '支持5G', 'http://example.com/pic2.jpg', 'SKU003', 'SN002', 'Apple',
     1, '[]', '张三', 4, 0, '2029-12-31 23:59:59'),
    (3, 1001, 3, 5, 1, 7999.00, 0,
     '华为 Mate X5', '折叠屏手机', 'http://example.com/pic3.jpg', 'SKU005', 'SN003', '华为',
     1, '[]', '张三', 4, 0, '2029-12-31 23:59:59'),
    (4, 1001, 1, 2, 1, 8999.00, 1,
     '小米 REDMI Note15 Pro 256GB', '256GB版', 'http://example.com/pic1b.jpg', 'SKU002', 'SN001B', '小米',
     1, '[]', '张三', 4, 0, '2029-12-31 23:59:59');

-- 7. 补齐 pms_product_spu 操作元数据字段（operation_metadata.go 需要）
SET @col_exists2 = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='pms_product_spu' AND COLUMN_NAME='publish_man');
SET @sql2 = IF(@col_exists2 = 0, 'ALTER TABLE pms_product_spu ADD COLUMN publish_man varchar(100) NOT NULL DEFAULT '''' COMMENT ''上架操作人'' AFTER publish_status, ADD COLUMN publish_time datetime NULL COMMENT ''上架时间'' AFTER publish_man, ADD COLUMN publish_detail varchar(500) NOT NULL DEFAULT '''' COMMENT ''上架说明'' AFTER publish_time, ADD COLUMN recommend_man varchar(100) NOT NULL DEFAULT '''' COMMENT ''推荐操作人'' AFTER recommend_status, ADD COLUMN recommend_time datetime NULL COMMENT ''推荐时间'' AFTER recommend_man, ADD COLUMN recommend_detail varchar(500) NOT NULL DEFAULT '''' COMMENT ''推荐说明'' AFTER recommend_time', 'SELECT 1');
PREPARE stmt2 FROM @sql2;
EXECUTE stmt2;
DEALLOCATE PREPARE stmt2;

-- 8. 补齐 cms_subject_category 的 scope 字段
-- 如果字段不存在则添加（MySQL 不支持 IF NOT EXISTS for ADD COLUMN，用存储过程替代）
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='cms_subject_category' AND COLUMN_NAME='platform_id');
SET @sql = IF(@col_exists = 0, 'ALTER TABLE cms_subject_category ADD COLUMN platform_id bigint NOT NULL DEFAULT 1 COMMENT ''平台ID'' AFTER id, ADD COLUMN tenant_id bigint NOT NULL DEFAULT 0 COMMENT ''租户ID'' AFTER platform_id, ADD COLUMN merchant_id bigint NOT NULL DEFAULT 0 COMMENT ''商户ID'' AFTER tenant_id', 'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 8. 确保优惠券作用域表种子数据（优惠券适用商品分类）
INSERT IGNORE INTO sms_coupon_scope (id, coupon_id, scope_type, scope_id)
VALUES (1, 1, 0, 0),  -- 满减券：全场通用
       (2, 2, 0, 0),  -- 新用户券：全场通用
       (3, 3, 0, 0);  -- 双十一券：全场通用
