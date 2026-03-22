drop table if exists sys_merchant_audit;
create table sys_merchant_audit
(
    id                     bigint auto_increment comment '商户治理审计主键'
        primary key,
    trace_id               varchar(64)                         default ''                not null comment '链路追踪ID',
    platform_id            bigint                                default 1                 not null comment '平台ID',
    tenant_id              bigint                                                             not null comment '租户ID',
    merchant_id            bigint                                                             not null comment '商户ID',
    merchant_code          varchar(32)                                                        not null comment '商户编码',
    action                 varchar(64)                                                        not null comment '动作类型',
    before_review_status   tinyint                                                          null comment '变更前审核状态',
    after_review_status    tinyint                                                          null comment '变更后审核状态',
    before_business_status tinyint                                                          null comment '变更前经营状态',
    after_business_status  tinyint                                                          null comment '变更后经营状态',
    operator_id            bigint                                default 0                 not null comment '操作者ID',
    operator_name          varchar(50)                         default ''                not null comment '操作者名称',
    request_payload        varchar(2000)                       default '{}'              not null comment '请求快照(JSON)',
    event_payload          varchar(2000)                       default '{}'              not null comment '事件快照(JSON)',
    result                 varchar(16)                         default 'success'         not null comment '处理结果',
    created_by             varchar(50)                         default 'admin'           not null comment '创建者',
    created_at             timestamp                           default CURRENT_TIMESTAMP not null comment '创建时间',
    updated_by             varchar(50)                         default ''                not null comment '更新者',
    updated_at             datetime                                                          null on update CURRENT_TIMESTAMP comment '更新时间'
) comment '商户治理审计表';

create index idx_sys_merchant_audit_merchant_time
    on sys_merchant_audit (merchant_id, created_at);

create index idx_sys_merchant_audit_action
    on sys_merchant_audit (action, created_at);
