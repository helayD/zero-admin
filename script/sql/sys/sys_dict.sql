create table sys_dict_type
(
    id          bigint                                 not null auto_increment comment '字典id',
    platform_id bigint       default 1                 not null comment '平台ID',
    tenant_id   bigint       default 0                 not null comment '租户ID',
    merchant_id bigint       default 0                 not null comment '商户ID',
    dict_name   varchar(100) default ''                not null comment '字典名称',
    dict_type   varchar(100) default ''                not null comment '字典类型',
    status      tinyint      default 0                 not null comment '状态（0：停用，1:正常）',
    remark      varchar(500) default ''                not null comment '备注',
    create_by   varchar(50)  default 'admin'           not null comment '创建者',
    create_time timestamp    default CURRENT_TIMESTAMP not null comment '创建时间',
    update_by   varchar(50)  default ''                not null comment '更新者',
    update_time datetime                               null on update CURRENT_TIMESTAMP comment '更新时间',
    primary key (id),
    constraint uk_sys_dict_type_scope_type
        unique (platform_id, tenant_id, merchant_id, dict_type)
) comment = '字典类型表';

create index idx_sys_dict_type_scope_status
    on sys_dict_type (platform_id, tenant_id, merchant_id, status);

INSERT INTO sys_dict_type (dict_name, dict_type, status, remark) VALUES ('用户性别', 'sys_user_sex', 1, '用户性别列表');
INSERT INTO sys_dict_type (dict_name, dict_type, status, remark) VALUES ('通知类型', 'sys_notice_type', 1, '通知类型列表');


create table sys_dict_item
(
    id           bigint                                 not null auto_increment    comment '字典数据id',
    platform_id  bigint       default 1                 not null comment '平台ID',
    tenant_id    bigint       default 0                 not null comment '租户ID',
    merchant_id  bigint       default 0                 not null comment '商户ID',
    dict_type_id bigint       default 0                 not null comment '字典类型ID',
    dict_sort   int          default 0                 not null comment '字典排序',
    dict_label  varchar(100) default ''                not null comment '字典标签',
    dict_value  varchar(100) default ''                not null comment '字典键值',
    dict_type   varchar(100) default ''                not null comment '字典类型',
    css_class   varchar(100) default ''                not null comment '样式属性（其他样式扩展）',
    list_class  varchar(100) default ''                not null comment '表格回显样式',
    is_default  char(1)      default 'N'               not null comment '是否默认（Y是 N否）',
    status      tinyint      default 0                 not null comment '状态（0：停用，1:正常）',
    remark      varchar(500) default ''                not null comment '备注',
    create_by   varchar(50)  default 'admin'           not null comment '创建者',
    create_time timestamp    default CURRENT_TIMESTAMP not null comment '创建时间',
    update_by   varchar(50)  default ''                not null comment '更新者',
    update_time datetime                               null on update CURRENT_TIMESTAMP comment '更新时间',
    primary key (id),
    constraint uk_sys_dict_item_scope_value
        unique (platform_id, tenant_id, merchant_id, dict_type_id, dict_value)
)comment = '字典数据表';

create index idx_sys_dict_item_scope_type
    on sys_dict_item (platform_id, tenant_id, merchant_id, dict_type_id, status);


INSERT INTO sys_dict_item (dict_type_id, dict_sort, dict_label, dict_value, dict_type, css_class, list_class, is_default, status, remark) VALUES (1, 1, '男', '0', 'sys_user_sex', '1', '1', 'N', 1, '性别男');
INSERT INTO sys_dict_item (dict_type_id, dict_sort, dict_label, dict_value, dict_type, css_class, list_class, is_default, status, remark) VALUES (1, 2, '女', '1', 'sys_user_sex', '1', '1', 'N', 1, '性别女');
INSERT INTO sys_dict_item (dict_type_id, dict_sort, dict_label, dict_value, dict_type, css_class, list_class, is_default, status, remark) VALUES (1, 3, '未知', '2', 'sys_user_sex', '1', '1', 'N', 1, '性别未知');
INSERT INTO sys_dict_item (dict_type_id, dict_sort, dict_label, dict_value, dict_type, css_class, list_class, is_default, status, remark) VALUES (2, 1, '通知', '1', 'sys_notice_type', '1', '1', 'N', 1, '通知');
INSERT INTO sys_dict_item (dict_type_id, dict_sort, dict_label, dict_value, dict_type, css_class, list_class, is_default, status, remark) VALUES (2, 2, '公告', '2', 'sys_notice_type', '1', '1', 'N', 1, '公告');

