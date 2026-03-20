export type GovernanceScopeType = 'platform' | 'tenant' | 'merchant';

export interface GovernanceScopeValue {
  scopeType?: GovernanceScopeType;
  scopeLabel?: string;
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
}

export const DEFAULT_PLATFORM_ID = 1;

export const defaultGovernanceScope: GovernanceScopeValue = {
  scopeType: 'platform',
  platformId: DEFAULT_PLATFORM_ID,
  tenantId: 0,
  merchantId: 0,
  scopeLabel: '平台级 #1',
};

export function buildGovernanceScopeLabel(input?: Partial<GovernanceScopeValue>): string {
  const scope = {
    scopeType: (input?.scopeType || 'platform') as GovernanceScopeType,
    platformId: input?.platformId || DEFAULT_PLATFORM_ID,
    tenantId: input?.tenantId || 0,
    merchantId: input?.merchantId || 0,
  };

  if (scope.scopeType === 'tenant') {
    return scope.tenantId ? `租户级 #${scope.tenantId}` : '租户级';
  }
  if (scope.scopeType === 'merchant') {
    return scope.merchantId ? `商户级 #${scope.merchantId}` : '商户级';
  }
  return `平台级 #${scope.platformId}`;
}

export const normalizeGovernanceScope = (
  input?: Partial<GovernanceScopeValue>,
): GovernanceScopeValue => {
  const scopeType = (input?.scopeType || 'platform') as GovernanceScopeType;
  const platformId =
    input?.platformId && input.platformId > 0 ? input.platformId : DEFAULT_PLATFORM_ID;
  const tenantId = scopeType === 'platform' ? 0 : Number(input?.tenantId || 0);
  const merchantId = scopeType === 'merchant' ? Number(input?.merchantId || 0) : 0;

  return {
    scopeType,
    platformId,
    tenantId,
    merchantId,
    scopeLabel: buildGovernanceScopeLabel({
      scopeType,
      platformId,
      tenantId,
      merchantId,
    }),
  };
};

export const toGovernancePayload = (input?: Partial<GovernanceScopeValue>) => {
  const scope = normalizeGovernanceScope(input);
  return {
    scopeType: scope.scopeType,
    platformId: scope.platformId,
    tenantId: scope.tenantId,
    merchantId: scope.merchantId,
  };
};

export const governanceScopeColor = (scopeType?: GovernanceScopeType) => {
  switch (scopeType) {
    case 'tenant':
      return 'green';
    case 'merchant':
      return 'orange';
    default:
      return 'blue';
  }
};

export const governanceImpactText = (
  input: Partial<GovernanceScopeValue>,
  entityLabel = '治理元数据',
) => {
  const scope = normalizeGovernanceScope(input);
  if (scope.scopeType === 'tenant') {
    return `当前只会影响该租户下的${entityLabel}、表单选项和通知配置，不会串写平台级默认数据。`;
  }
  if (scope.scopeType === 'merchant') {
    return `当前只会影响目标商户的${entityLabel}。如需平台或租户复用，请切回更高层级后再操作。`;
  }
  return `当前维护的是平台默认${entityLabel}，后续租户与商户后台可直接复用。`;
};
