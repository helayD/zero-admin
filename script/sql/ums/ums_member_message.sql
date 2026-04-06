drop table if exists ums_member_message;

create table ums_member_message
(
    id              bigint auto_increment
        primary key comment '消息ID',
    member_id       bigint                                not null comment '会员ID',
    message_type    int       default 1                 not null comment '消息类型（1:订单,2:支付,3:售后,4:活动,5:会员）',
    title           varchar(255)                          not null comment '消息标题',
    content         text                                   null comment '消息内容',
    image_url       varchar(500) default ''                null comment '图片URL（可选）',
    link_type       varchar(50)  default ''                null comment '跳转类型(order/product/coupon/activity)',
    link_id         varchar(100) default ''               null comment '跳转目标ID',
    related_order_id bigint    default 0                  null comment '关联订单ID',
    intent_contract text         default '{}'             not null comment '统一意图契约(JSON)',
    status          int       default 0                 not null comment '状态（0:未读,1:已读）',
    read_time       datetime                              null comment '阅读时间',
    create_time     datetime    default CURRENT_TIMESTAMP not null comment '创建时间',
    platform_id     bigint     default 1                 not null comment '平台ID',
    tenant_id       bigint     default 1                 not null comment '租户ID',
    merchant_id     bigint     default 0                 null comment '商户ID（商户相关消息时使用）'
)
    comment '会员消息表';

create index idx_member_id
    on ums_member_message (member_id);

create index idx_create_time
    on ums_member_message (create_time);

create index idx_status
    on ums_member_message (status);

-- 插入测试数据
insert into ums_member_message (id, member_id, message_type, title, content, link_type, link_id, related_order_id, status, platform_id, tenant_id)
values (1, 1001, 1, '订单已创建', '您的订单已成功创建，订单号：ORD202604010001', 'order', '1001', 1001, 0, 1, 1),
       (2, 1001, 2, '支付成功', '您的订单已支付成功，金额：¥299.00', 'order', '1001', 1001, 0, 1, 1),
       (3, 1001, 4, '活动提醒', '春季大促活动火热进行中，点击查看更多优惠', 'activity', '1', 0, 1, 1, 1),
       (4, 1002, 1, '订单已创建', '您的订单已成功创建，订单号：ORD202604010002', 'order', '1002', 1002, 1, 1, 1),
       (5, 1002, 3, '售后申请已受理', '您的售后申请已受理，请等待处理', 'order', '1002', 1002, 0, 1, 1);
