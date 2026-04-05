DROP TABLE IF EXISTS sms_operate_funnel_event;
create table sms_operate_funnel_event
(
    id             bigint auto_increment comment '主键ID'
        primary key,
    event_type     varchar(32)                            not null comment '漏斗阶段(exposure/click/add_cart/order_created/pay_success/coupon_redeem)',
    stat_time      datetime                               not null comment '事件发生时间',
    bucket_date    date                                   not null comment '按天聚合桶',
    bucket_hour    datetime                               null comment '按小时聚合桶',
    platform_id    bigint       default 1                 not null comment '平台ID',
    tenant_id      bigint       default 0                 not null comment '租户ID',
    merchant_id    bigint       default 0                 not null comment '商户ID',
    channel        varchar(32)  default 'unknown'         not null comment '统一渠道(app/pc/h5/mini_program/unknown)',
    activity_type  varchar(32)  default 'none'            not null comment '活动类型(home_advertise/coupon/seckill_activity/none)',
    activity_id    bigint       default 0                 not null comment '活动ID',
    member_id      bigint       default 0                 not null comment '会员ID',
    product_id     bigint       default 0                 not null comment '商品ID',
    order_id       bigint       default 0                 not null comment '订单ID',
    coupon_id      bigint       default 0                 not null comment '优惠券ID',
    trace_id       varchar(128)                           not null comment '幂等追踪ID',
    extra_json     json                                   null comment '扩展信息',
    create_time    datetime     default CURRENT_TIMESTAMP not null comment '创建时间',
    constraint uk_operate_funnel_trace
        unique (event_type, trace_id)
)
    comment '经营漏斗事件事实表';

create index idx_operate_funnel_scope_time
    on sms_operate_funnel_event (platform_id, tenant_id, merchant_id, stat_time);

create index idx_operate_funnel_activity_time
    on sms_operate_funnel_event (activity_type, activity_id, stat_time);

create index idx_operate_funnel_channel_time
    on sms_operate_funnel_event (channel, stat_time);

create index idx_operate_funnel_event_time
    on sms_operate_funnel_event (event_type, stat_time);
