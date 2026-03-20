import { Alert, Col, InputNumber, Row, Select, Space, Tag } from 'antd';
import React from 'react';
import type { GovernanceScopeValue, GovernanceScopeType } from './governance';
import { governanceImpactText, governanceScopeColor, normalizeGovernanceScope } from './governance';

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

const GovernanceScopeBar: React.FC<GovernanceScopeBarProps> = ({
  value,
  onChange,
  entityLabel = '治理元数据',
  style,
}) => {
  const scope = normalizeGovernanceScope(value);

  const updateScope = (patch: Partial<GovernanceScopeValue>) => {
    onChange(normalizeGovernanceScope({ ...scope, ...patch }));
  };

  return (
    <Alert
      showIcon
      type="info"
      message="Scope Context Bar"
      style={style}
      description={
        <Space direction="vertical" style={{ width: '100%' }} size={12}>
          <Row gutter={12}>
            <Col xs={24} sm={8}>
              <Select
                style={{ width: '100%' }}
                value={scope.scopeType}
                options={scopeOptions}
                onChange={(nextScopeType) => {
                  if (nextScopeType === 'platform') {
                    updateScope({ scopeType: nextScopeType, tenantId: 0, merchantId: 0 });
                    return;
                  }
                  if (nextScopeType === 'tenant') {
                    updateScope({ scopeType: nextScopeType, merchantId: 0 });
                    return;
                  }
                  updateScope({ scopeType: nextScopeType });
                }}
              />
            </Col>
            <Col xs={24} sm={scope.scopeType === 'merchant' ? 8 : 16}>
              <InputNumber
                style={{ width: '100%' }}
                min={1}
                disabled={scope.scopeType === 'platform'}
                placeholder={
                  scope.scopeType === 'merchant' ? '请输入租户 ID（可选）' : '请输入租户 ID'
                }
                value={scope.scopeType === 'platform' ? undefined : scope.tenantId || undefined}
                onChange={(nextTenantId) => updateScope({ tenantId: Number(nextTenantId || 0) })}
              />
            </Col>
            {scope.scopeType === 'merchant' && (
              <Col xs={24} sm={8}>
                <InputNumber
                  style={{ width: '100%' }}
                  min={1}
                  placeholder="请输入商户 ID"
                  value={scope.merchantId || undefined}
                  onChange={(nextMerchantId) =>
                    updateScope({ merchantId: Number(nextMerchantId || 0) })
                  }
                />
              </Col>
            )}
          </Row>

          <Space wrap size={[8, 8]}>
            <Tag color={governanceScopeColor(scope.scopeType)}>{scope.scopeLabel}</Tag>
            <span>{governanceImpactText(scope, entityLabel)}</span>
          </Space>
        </Space>
      }
    />
  );
};

export default GovernanceScopeBar;
