ALTER TABLE oms_cart_item
    ADD COLUMN IF NOT EXISTS activity_type varchar(32) default 'none' not null comment '活动类型(home_advertise/coupon/seckill_activity/none)' after source,
    ADD COLUMN IF NOT EXISTS activity_id bigint default 0 not null comment '活动ID' after activity_type;

ALTER TABLE oms_order_main
    ADD COLUMN IF NOT EXISTS activity_type varchar(32) default 'none' not null comment '活动类型(home_advertise/coupon/seckill_activity/none)' after source_type,
    ADD COLUMN IF NOT EXISTS activity_id bigint default 0 not null comment '活动ID' after activity_type;

CREATE INDEX IF NOT EXISTS idx_cart_activity_create_time
    ON oms_cart_item (activity_type, activity_id, create_time);

CREATE INDEX IF NOT EXISTS idx_order_activity_create_time
    ON oms_order_main (activity_type, activity_id, create_time);

CREATE INDEX IF NOT EXISTS idx_order_activity_pay_time
    ON oms_order_main (activity_type, activity_id, pay_time);
