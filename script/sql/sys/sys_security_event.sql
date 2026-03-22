drop table if exists sys_security_event;
create table sys_security_event
(
    id              bigint auto_increment comment '安全事件ID'
        primary key,
    trace_id        varchar(64)   default ''                not null comment '链路追踪ID',
    event_type      varchar(64)   default ''                not null comment '事件类型',
    action          varchar(128)  default ''                not null comment '操作动作',
    resource_type   varchar(64)   default ''                not null comment '资源类型',
    resource_id     bigint         default 0                 not null comment '资源ID',
    scope_type      varchar(32)   default ''                not null comment '治理范围类型',
    platform_id     bigint         default 1                 not null comment '平台ID',
    tenant_id       bigint         default 0                 not null comment '租户ID',
    merchant_id     bigint         default 0                 not null comment '商户ID',
    operator_id     bigint         default 0                 not null comment '操作者ID',
    operator_name   varchar(64)   default ''                not null comment '操作者名称',
    request_summary varchar(500)  default ''                not null comment '请求摘要',
    result          varchar(32)   default ''                not null comment '处理结果',
    payload         text                                    not null comment '结构化事件载荷',
    created_at      datetime      default CURRENT_TIMESTAMP not null comment '创建时间',
    key idx_security_event_trace (trace_id),
    key idx_security_event_scope (scope_type, platform_id, tenant_id, merchant_id),
    key idx_security_event_resource (resource_type, resource_id, created_at),
    key idx_security_event_type (event_type, created_at)
) comment = '治理安全事件表';
