ALTER TABLE pms_product_spu
    ADD COLUMN IF NOT EXISTS publish_man VARCHAR(64) NULL COMMENT '最近上下架操作人' AFTER update_time,
    ADD COLUMN IF NOT EXISTS publish_time DATETIME NULL COMMENT '最近上下架时间' AFTER publish_man,
    ADD COLUMN IF NOT EXISTS publish_detail VARCHAR(255) NOT NULL DEFAULT '' COMMENT '最近上下架说明' AFTER publish_time,
    ADD COLUMN IF NOT EXISTS recommend_man VARCHAR(64) NULL COMMENT '最近推荐操作人' AFTER publish_detail,
    ADD COLUMN IF NOT EXISTS recommend_time DATETIME NULL COMMENT '最近推荐时间' AFTER recommend_man,
    ADD COLUMN IF NOT EXISTS recommend_detail VARCHAR(255) NOT NULL DEFAULT '' COMMENT '最近推荐说明' AFTER recommend_time;
