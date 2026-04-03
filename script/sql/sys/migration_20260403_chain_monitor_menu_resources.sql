-- Story 7.6 chain monitor menu resource backfill.
-- Safe to execute multiple times on an existing environment.

INSERT INTO sys_menu (
    id, menu_name, parent_id, menu_path, menu_perms, menu_type, menu_icon, menu_sort,
    create_by, create_time, update_by, update_time, is_deleted, vue_path, vue_component,
    vue_icon, vue_redirect, background_url
)
VALUES
    (306, '查询链路监控列表', 24, '', '', 2, '', 4, 'liufeihua', CURRENT_TIMESTAMP, 'liufeihua', CURRENT_TIMESTAMP, 1, '', '', '', '', '/api/oms/order/queryChainMonitorList'),
    (307, '导出链路监控列表', 24, '', '', 2, '', 5, 'liufeihua', CURRENT_TIMESTAMP, 'liufeihua', CURRENT_TIMESTAMP, 1, '', '', '', '', '/api/oms/order/exportChainMonitorList'),
    (308, '查询链路可用动作', 24, '', '', 2, '', 6, 'liufeihua', CURRENT_TIMESTAMP, 'liufeihua', CURRENT_TIMESTAMP, 1, '', '', '', '', '/api/oms/order/queryChainActions'),
    (309, '暂停链路', 24, '', '', 2, '', 7, 'liufeihua', CURRENT_TIMESTAMP, 'liufeihua', CURRENT_TIMESTAMP, 1, '', '', '', '', '/api/oms/order/pauseChain'),
    (310, '重试链路', 24, '', '', 2, '', 8, 'liufeihua', CURRENT_TIMESTAMP, 'liufeihua', CURRENT_TIMESTAMP, 1, '', '', '', '', '/api/oms/order/retryChain'),
    (311, '回放链路', 24, '', '', 2, '', 9, 'liufeihua', CURRENT_TIMESTAMP, 'liufeihua', CURRENT_TIMESTAMP, 1, '', '', '', '', '/api/oms/order/replayChain'),
    (312, '升级链路', 24, '', '', 2, '', 10, 'liufeihua', CURRENT_TIMESTAMP, 'liufeihua', CURRENT_TIMESTAMP, 1, '', '', '', '', '/api/oms/order/escalateChain')
ON DUPLICATE KEY UPDATE
    menu_name = VALUES(menu_name),
    parent_id = VALUES(parent_id),
    menu_type = VALUES(menu_type),
    menu_sort = VALUES(menu_sort),
    background_url = VALUES(background_url),
    update_by = VALUES(update_by),
    update_time = CURRENT_TIMESTAMP;
