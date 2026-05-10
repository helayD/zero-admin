-- Story 10.7 Fix: 回填存量已支付订单的 pay_time
--
-- 问题：UpdateOrder 历史版本不会写 pay_time，导致已支付订单 pay_time 仍为 NULL，
-- 订单详情页时间线缺失「支付成功」节点（QueryOrderDetailLogic.buildTimeline 依赖
-- pay_time != ""）。
--
-- 修复：对所有 order_status >= 2 且 pay_time IS NULL 的订单，把 pay_time 回填为
-- update_time（最近一次状态变更时间，最接近真实支付时刻），update_time 也为 NULL
-- 时退化到 create_time。
--
-- 幂等：仅回填 NULL 行，不会覆盖已写入的 pay_time。

UPDATE oms_order_main
SET pay_time = COALESCE(update_time, create_time)
WHERE order_status IN (2, 3, 4, 7)
  AND pay_time IS NULL
  AND is_deleted = 0;
