CREATE TABLE IF NOT EXISTS sys_channel_integration_template
(
    id                     bigint auto_increment
        PRIMARY KEY,
    template_code          varchar(64)                            NOT NULL,
    template_name          varchar(100)                           NOT NULL,
    template_type          enum ('channel', 'integration')        NOT NULL,
    target_code            varchar(64)                            NOT NULL,
    scope_type             enum ('platform', 'tenant', 'merchant') DEFAULT 'platform' NOT NULL,
    platform_id            bigint                                 DEFAULT 1 NOT NULL,
    tenant_id              bigint                                 DEFAULT 0 NOT NULL,
    merchant_id            bigint                                 DEFAULT 0 NOT NULL,
    status                 enum ('draft', 'enabled', 'disabled', 'archived') DEFAULT 'draft' NOT NULL,
    metadata_config        json                                   NULL,
    secret_ref_config      json                                   NULL,
    intent_contract_config json                                   NULL,
    impact_scope_config    json                                   NULL,
    remark                 varchar(255)                           DEFAULT '' NOT NULL,
    create_by              varchar(50)                            DEFAULT 'admin' NOT NULL,
    create_time            timestamp                              DEFAULT CURRENT_TIMESTAMP NOT NULL,
    update_by              varchar(50)                            DEFAULT '' NOT NULL,
    update_time            datetime                               NULL ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT uk_channel_integration_template_code_scope
        UNIQUE (template_code, scope_type, platform_id, tenant_id, merchant_id),
    KEY idx_channel_integration_template_target_status (target_code, status),
    KEY idx_channel_integration_template_scope (scope_type, platform_id, tenant_id, merchant_id)
) COMMENT '渠道与第三方集成模板';

CREATE TABLE IF NOT EXISTS sys_channel_integration_binding
(
    id                    bigint auto_increment
        PRIMARY KEY,
    template_id           bigint                                   NOT NULL,
    subject_type          enum ('tenant', 'merchant')              NOT NULL,
    platform_id           bigint                                   DEFAULT 1 NOT NULL,
    tenant_id             bigint                                   DEFAULT 0 NOT NULL,
    merchant_id           bigint                                   DEFAULT 0 NOT NULL,
    target_code           varchar(64)                              NOT NULL,
    binding_status        enum ('draft', 'enabled', 'disabled', 'archived') DEFAULT 'draft' NOT NULL,
    binding_source        varchar(32)                              DEFAULT 'manual' NOT NULL,
    effect_scope_snapshot json                                     NULL,
    published_by          varchar(50)                              DEFAULT '' NOT NULL,
    published_at          datetime                                 NULL,
    remark                varchar(255)                             DEFAULT '' NOT NULL,
    create_by             varchar(50)                              DEFAULT 'admin' NOT NULL,
    create_time           timestamp                                DEFAULT CURRENT_TIMESTAMP NOT NULL,
    update_by             varchar(50)                              DEFAULT '' NOT NULL,
    update_time           datetime                                 NULL ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT uk_channel_integration_binding_subject
        UNIQUE (template_id, subject_type, platform_id, tenant_id, merchant_id, target_code),
    KEY idx_channel_integration_binding_subject_status (subject_type, platform_id, tenant_id, merchant_id, binding_status),
    KEY idx_channel_integration_binding_template (template_id, binding_status)
) COMMENT '渠道与第三方集成模板绑定真相表';

INSERT INTO sys_channel_integration_template (
    id,
    template_code,
    template_name,
    template_type,
    target_code,
    scope_type,
    platform_id,
    tenant_id,
    merchant_id,
    status,
    metadata_config,
    secret_ref_config,
    intent_contract_config,
    impact_scope_config,
    remark,
    create_by,
    create_time,
    update_by,
    update_time
)
VALUES
    (
        1,
        'h5_default_template',
        'H5 渠道模板',
        'channel',
        'h5',
        'platform',
        1,
        0,
        0,
        'enabled',
        JSON_OBJECT('channelCode', 'h5', 'entryMode', 'intent_route', 'schemaVersion', 'v1'),
        JSON_OBJECT(),
        JSON_OBJECT('supportedIntents', JSON_ARRAY('home', 'product_detail', 'coupon_center', 'member_message'), 'routeMode', 'intent'),
        JSON_OBJECT('subjectTypes', JSON_ARRAY('tenant', 'merchant'), 'requiresBinding', true),
        '平台统一 H5 模板',
        'codex',
        NOW(),
        'codex',
        NOW()
    ),
    (
        2,
        'mini_program_default_template',
        '小程序渠道模板',
        'channel',
        'mini_program',
        'platform',
        1,
        0,
        0,
        'enabled',
        JSON_OBJECT('channelCode', 'mini_program', 'entryMode', 'intent_route', 'schemaVersion', 'v1'),
        JSON_OBJECT('credentialRefFields', JSON_ARRAY('appSecretRef', 'privateKeyRef')),
        JSON_OBJECT('supportedIntents', JSON_ARRAY('home', 'product_detail', 'coupon_center', 'member_message'), 'routeMode', 'intent'),
        JSON_OBJECT('subjectTypes', JSON_ARRAY('tenant', 'merchant'), 'requiresBinding', true),
        '平台统一小程序模板',
        'codex',
        NOW(),
        'codex',
        NOW()
    ),
    (
        3,
        'logistics_tracking_default_template',
        '物流轨迹集成模板',
        'integration',
        'logistics_tracking',
        'platform',
        1,
        0,
        0,
        'enabled',
        JSON_OBJECT('providerMode', 'tracking', 'schemaVersion', 'v1'),
        JSON_OBJECT('credentialRefFields', JSON_ARRAY('accessKeyIdRef', 'accessKeySecretRef', 'apiTokenRef')),
        JSON_OBJECT(),
        JSON_OBJECT('subjectTypes', JSON_ARRAY('tenant', 'merchant'), 'requiresBinding', true),
        '平台统一物流轨迹模板',
        'codex',
        NOW(),
        'codex',
        NOW()
    ),
    (
        4,
        'sms_provider_default_template',
        '短信服务商集成模板',
        'integration',
        'sms_provider',
        'platform',
        1,
        0,
        0,
        'enabled',
        JSON_OBJECT('providerMode', 'sms', 'schemaVersion', 'v1'),
        JSON_OBJECT('credentialRefFields', JSON_ARRAY('accessKeyIdRef', 'accessKeySecretRef', 'signatureSecretRef')),
        JSON_OBJECT(),
        JSON_OBJECT('subjectTypes', JSON_ARRAY('tenant', 'merchant'), 'requiresBinding', true),
        '平台统一短信服务商模板',
        'codex',
        NOW(),
        'codex',
        NOW()
    ),
    (
        5,
        'member_message_default_template',
        '站内消息集成模板',
        'integration',
        'member_message',
        'platform',
        1,
        0,
        0,
        'enabled',
        JSON_OBJECT('providerMode', 'ums_member_message', 'schemaVersion', 'v1'),
        JSON_OBJECT(),
        JSON_OBJECT('supportedIntents', JSON_ARRAY('order_status', 'coupon_notice', 'system_notice'), 'routeMode', 'intent'),
        JSON_OBJECT('subjectTypes', JSON_ARRAY('tenant', 'merchant'), 'requiresBinding', true),
        '平台统一站内消息模板',
        'codex',
        NOW(),
        'codex',
        NOW()
    )
ON DUPLICATE KEY UPDATE
    template_name = VALUES(template_name),
    template_type = VALUES(template_type),
    target_code = VALUES(target_code),
    status = VALUES(status),
    metadata_config = VALUES(metadata_config),
    secret_ref_config = VALUES(secret_ref_config),
    intent_contract_config = VALUES(intent_contract_config),
    impact_scope_config = VALUES(impact_scope_config),
    remark = VALUES(remark),
    update_by = VALUES(update_by),
    update_time = VALUES(update_time);

UPDATE sys_menu
SET menu_name = '渠道与集成模板',
    parent_id = 2,
    menu_path = '/system/channelIntegrationTemplate/list',
    menu_type = 1,
    menu_sort = 11,
    menu_status = 1,
    is_deleted = 1,
    is_visible = 1,
    remark = '渠道与第三方集成模板治理资源',
    vue_path = 'channelIntegrationTemplate',
    vue_component = 'system/channel_integration_template/index',
    vue_icon = 'el-icon-connection',
    vue_redirect = '',
    background_url = '/api/sys/channelIntegrationTemplate/queryChannelIntegrationTemplateList',
    update_by = 'codex',
    update_time = NOW()
WHERE background_url = '/api/sys/channelIntegrationTemplate/queryChannelIntegrationTemplateList';

INSERT INTO sys_menu (
    id,
    menu_name,
    parent_id,
    menu_path,
    menu_perms,
    menu_type,
    menu_icon,
    menu_sort,
    create_by,
    create_time,
    update_by,
    update_time,
    menu_status,
    is_deleted,
    is_visible,
    remark,
    vue_path,
    vue_component,
    vue_icon,
    vue_redirect,
    background_url
)
SELECT
    314,
    '渠道与集成模板',
    2,
    '/system/channelIntegrationTemplate/list',
    '',
    1,
    '',
    11,
    'codex',
    NOW(),
    'codex',
    NOW(),
    1,
    1,
    1,
    '渠道与第三方集成模板治理资源',
    'channelIntegrationTemplate',
    'system/channel_integration_template/index',
    'el-icon-connection',
    '',
    '/api/sys/channelIntegrationTemplate/queryChannelIntegrationTemplateList'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1
    FROM sys_menu
    WHERE id = 314
       OR background_url = '/api/sys/channelIntegrationTemplate/queryChannelIntegrationTemplateList'
);

INSERT INTO sys_menu (
    id,
    menu_name,
    parent_id,
    menu_path,
    menu_perms,
    menu_type,
    menu_icon,
    menu_sort,
    create_by,
    create_time,
    update_by,
    update_time,
    menu_status,
    is_deleted,
    is_visible,
    remark,
    vue_path,
    vue_component,
    vue_icon,
    vue_redirect,
    background_url
)
SELECT
    315,
    '查询渠道与集成模板详情',
    314,
    '',
    '',
    2,
    '',
    1,
    'codex',
    NOW(),
    'codex',
    NOW(),
    1,
    1,
    1,
    '渠道与第三方集成模板详情权限',
    '',
    '',
    '',
    '',
    '/api/sys/channelIntegrationTemplate/queryChannelIntegrationTemplateDetail'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1
    FROM sys_menu
    WHERE id = 315
       OR background_url = '/api/sys/channelIntegrationTemplate/queryChannelIntegrationTemplateDetail'
);

INSERT INTO sys_menu (
    id,
    menu_name,
    parent_id,
    menu_path,
    menu_perms,
    menu_type,
    menu_icon,
    menu_sort,
    create_by,
    create_time,
    update_by,
    update_time,
    menu_status,
    is_deleted,
    is_visible,
    remark,
    vue_path,
    vue_component,
    vue_icon,
    vue_redirect,
    background_url
)
SELECT
    316,
    '新建渠道与集成模板',
    314,
    '',
    '',
    2,
    '',
    2,
    'codex',
    NOW(),
    'codex',
    NOW(),
    1,
    1,
    1,
    '渠道与第三方集成模板新建权限',
    '',
    '',
    '',
    '',
    '/api/sys/channelIntegrationTemplate/createChannelIntegrationTemplate'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1
    FROM sys_menu
    WHERE id = 316
       OR background_url = '/api/sys/channelIntegrationTemplate/createChannelIntegrationTemplate'
);

INSERT INTO sys_menu (
    id,
    menu_name,
    parent_id,
    menu_path,
    menu_perms,
    menu_type,
    menu_icon,
    menu_sort,
    create_by,
    create_time,
    update_by,
    update_time,
    menu_status,
    is_deleted,
    is_visible,
    remark,
    vue_path,
    vue_component,
    vue_icon,
    vue_redirect,
    background_url
)
SELECT
    317,
    '更新渠道与集成模板',
    314,
    '',
    '',
    2,
    '',
    3,
    'codex',
    NOW(),
    'codex',
    NOW(),
    1,
    1,
    1,
    '渠道与第三方集成模板更新权限',
    '',
    '',
    '',
    '',
    '/api/sys/channelIntegrationTemplate/updateChannelIntegrationTemplate'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1
    FROM sys_menu
    WHERE id = 317
       OR background_url = '/api/sys/channelIntegrationTemplate/updateChannelIntegrationTemplate'
);

INSERT INTO sys_menu (
    id,
    menu_name,
    parent_id,
    menu_path,
    menu_perms,
    menu_type,
    menu_icon,
    menu_sort,
    create_by,
    create_time,
    update_by,
    update_time,
    menu_status,
    is_deleted,
    is_visible,
    remark,
    vue_path,
    vue_component,
    vue_icon,
    vue_redirect,
    background_url
)
SELECT
    318,
    '更新渠道与集成模板状态',
    314,
    '',
    '',
    2,
    '',
    4,
    'codex',
    NOW(),
    'codex',
    NOW(),
    1,
    1,
    1,
    '渠道与第三方集成模板状态权限',
    '',
    '',
    '',
    '',
    '/api/sys/channelIntegrationTemplate/updateChannelIntegrationTemplateStatus'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1
    FROM sys_menu
    WHERE id = 318
       OR background_url = '/api/sys/channelIntegrationTemplate/updateChannelIntegrationTemplateStatus'
);

INSERT INTO sys_menu_template_item (template_id, menu_id)
VALUES
    (1, 314),
    (2, 314)
ON DUPLICATE KEY UPDATE
    menu_id = VALUES(menu_id);
