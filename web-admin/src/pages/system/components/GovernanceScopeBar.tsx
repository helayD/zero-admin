import { Alert, Col, InputNumber, Row, Select, Space, Tag } from 'antd';
import React, { useState, useEffect } from 'react';
import type { GovernanceScopeValue, GovernanceScopeType } from './governance';
import {
  governanceImpactText,
  governanceScopeColor,
  normalizeGovernanceScope,
  buildGovernanceScopeLabel,
} from './governance';

interface GovernanceScopeBarProps {
  value: GovernanceScopeValue;
  onChange: (next: GovernanceScopeValue) => void;
  entityLabel?: string;
  style?: React.CSSProperties;
}

const scopeOptions: { label: string; value: GovernanceScopeType }[] = [
  { label: '平台级', value: 'platform' },
  { label: '租户级', value: 'tenant' },
  { label: '商户级', value: 'merchant' },
];

/** 当前 draft 是否满足后端最低要求（tenant 需 tenantId>0，merchant 需 merchantId>0） */
const isScopeReady = (s: GovernanceScopeValue): boolean => {
  if (s.scopeType === 'tenant') return (s.tenantId ?? 0) > 0;
  if (s.scopeType === 'merchant') return (s.merchantId ?? 0) > 0;
  return true; // platform 无额外要求
};

const GovernanceScopeBar: React.FC<GovernanceScopeBarProps> = ({
  value,
  onChange,
  entityLabel = '治理元数据',
  style,
}) => {
  // draft: 内部编辑态，允许不完整；只有 ready 时才通知父组件
  const [draft, setDraft] = useState<GovernanceScopeValue>(() => normalizeGovernanceScope(value));

  // 外部 value 变化时同步 draft
  useEffect(() => {
    setDraft(normalizeGovernanceScope(value));
  }, [value.scopeType, value.platformId, value.tenantId, value.merchantId]);

  const updateDraft = (patch: Partial<GovernanceScopeValue>) => {
    const next = normalizeGovernanceScope({ ...draft, ...patch });
    setDraft(next);
    if (isScopeReady(next)) {
      onChange(next);
    }
  };

  const pending = !isScopeReady(draft);

  return (
    <Alert
      showIcon
      type={pending ? 'warning' : 'info'}
      message="Scope Context Bar"
      style={style}
      description={
        <Space direction="vertical" style={{ width: '100%' }} size={12}>
          <Row gutter={12}>
            <Col xs={24} sm={8}>
              <Select
                style={{ width: '100%' }}
                value={draft.scopeType}
                options={scopeOptions}
                onChange={(nextScopeType) => {
                  if (nextScopeType === 'platform') {
                    updateDraft({ scopeType: nextScopeType, tenantId: 0, merchantId: 0 });
                    return;
                  }
                  if (nextScopeType === 'tenant') {
                    updateDraft({ scopeType: nextScopeType, merchantId: 0 });
                    return;
                  }
                  updateDraft({ scopeType: nextScopeType });
                }}
              />
            </Col>
            <Col xs={24} sm={draft.scopeType === 'merchant' ? 8 : 16}>
              <InputNumber
                style={{ width: '100%' }}
                min={1}
                disabled={draft.scopeType === 'platform'}
                placeholder={
                  draft.scopeType === 'merchant' ? '请输入租户 ID（可选）' : '请输入租户 ID'
                }
                status={draft.scopeType === 'tenant' && !draft.tenantId ? 'warning' : undefined}
                value={draft.scopeType === 'platform' ? undefined : draft.tenantId || undefined}
                onChange={(nextTenantId) => updateDraft({ tenantId: Number(nextTenantId || 0) })}
              />
            </Col>
            {draft.scopeType === 'merchant' && (
              <Col xs={24} sm={8}>
                <InputNumber
                  style={{ width: '100%' }}
                  min={1}
                  placeholder="请输入商户 ID"
                  status={!draft.merchantId ? 'warning' : undefined}
                  value={draft.merchantId || undefined}
                  onChange={(nextMerchantId) =>
                    updateDraft({ merchantId: Number(nextMerchantId || 0) })
                  }
                />
              </Col>
            )}
          </Row>

          <Space wrap size={[8, 8]}>
            <Tag color={governanceScopeColor(draft.scopeType)}>
              {buildGovernanceScopeLabel(draft)}
            </Tag>
            {pending && (
              <Tag color="orange">
                请先填写{draft.scopeType === 'tenant' ? '租户' : '商户'} ID 后生效
              </Tag>
            )}
            <span>{governanceImpactText(draft, entityLabel)}</span>
          </Space>
        </Space>
      }
    />
  );
};

export default GovernanceScopeBar;
