-- Story 10.10 卡片模板新增种子数据。
-- 现有种子 (920001-920003) 都属于 platform_id=1 / tenant_id=10 / merchant_id=88 或租户级，
-- 缺一条「平台级公版」(tenant_id=0 merchant_id=0) 让所有租户/商户都能引用。
--
-- 幂等：使用 UNIQUE KEY (platform_id, tenant_id, merchant_id, template_code, is_deleted) 守卫。

INSERT INTO sms_card_template (
    platform_id, tenant_id, merchant_id,
    template_code, template_name, card_face_image,
    copyright_owner, copyright_proof_summary,
    rarity, issue_limit,
    display_copy, circulation_limit_summary,
    display_status, content_audit_status,
    provider_code, credential_ref,
    status, audit_status,
    create_by, create_time, update_by, update_time, is_deleted
)
SELECT
    1, 0, 0,
    'TPL-PLATFORM-COMMEMORATIVE', '平台公版-周年纪念卡', '',
    '平台官方', '官方版权·内部留存',
    'SSR', 5000,
    '平台周年纪念限量发行，凭借此卡可参与年度专属抽奖与会员特权活动。',
    '本卡仅供个人收藏与限定活动使用，不支持二次转赠。',
    1, 2,
    'PLATFORM_OFFICIAL', '',
    1, 2,
    1, NOW(), 1, NOW(), 0
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1 FROM sms_card_template
    WHERE platform_id = 1
      AND tenant_id = 0
      AND merchant_id = 0
      AND template_code = 'TPL-PLATFORM-COMMEMORATIVE'
      AND is_deleted = 0
);
