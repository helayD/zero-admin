import type { GovernanceScopeValue } from '@/pages/system/components/governance';

export interface ScopePayload extends GovernanceScopeValue {
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
}

export const readCatalogErrorMessage = (error: any, fallback: string) => {
  return error?.data?.message || error?.message || fallback;
};
