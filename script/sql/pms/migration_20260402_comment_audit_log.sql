CREATE TABLE IF NOT EXISTS `pms_comment_audit_log` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `comment_id` varchar(32) NOT NULL COMMENT '评价ID（MongoDB ObjectID）',
  `platform_id` bigint NOT NULL DEFAULT 0 COMMENT '平台ID',
  `tenant_id` bigint NOT NULL DEFAULT 0 COMMENT '租户ID',
  `merchant_id` bigint NOT NULL DEFAULT 0 COMMENT '商户ID',
  `action` varchar(32) NOT NULL DEFAULT '' COMMENT '操作类型: approve/reject/hide/restore/appeal_submit/appeal_handle',
  `from_status` int NOT NULL DEFAULT 0 COMMENT '操作前审核状态',
  `to_status` int NOT NULL DEFAULT 0 COMMENT '操作后审核状态',
  `operator_id` bigint NOT NULL DEFAULT 0 COMMENT '操作人ID',
  `operator_name` varchar(64) NOT NULL DEFAULT '' COMMENT '操作人姓名',
  `remark` varchar(500) NOT NULL DEFAULT '' COMMENT '操作备注/拒绝原因/申诉回复',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_comment_id` (`comment_id`),
  KEY `idx_tenant_merchant` (`tenant_id`, `merchant_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品评价审核操作日志表';
