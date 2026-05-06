-- 添加 returnApply 接口权限到退货列表菜单
-- Safe to execute multiple times on an existing environment.

UPDATE sys_menu
SET background_url = '/api/oms/orderReturn/queryOrderReturnList,/api/oms/returnApply/queryOrderReturnApplyList',
    update_by = 'liufeihua',
    update_time = CURRENT_TIMESTAMP
WHERE id = 22
  AND menu_name = '退货列表';
