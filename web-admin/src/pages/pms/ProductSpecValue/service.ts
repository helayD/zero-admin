import { request } from 'umi';
import type { ProductSpecValueListParams, ProductSpecValueListItem } from './data.d';
import type { GovernanceScopeValue } from '@/pages/system/components/governance';

interface ScopePayload extends GovernanceScopeValue {
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
}

export async function addProductSpecValue(params: ProductSpecValueListItem & ScopePayload) {
  return request('/api/pms/productSpecValue/addSpecValue', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function removeProductSpecValue(ids: number[], scope?: ScopePayload) {
  return request('/api/pms/productSpecValue/deleteSpecValue', {
    method: 'GET',
    params: {
      ids: ids.join(','),
      ...scope,
    },
  });
}

export async function updateProductSpecValue(params: ProductSpecValueListItem & ScopePayload) {
  return request('/api/pms/productSpecValue/updateSpecValue', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function updateProductSpecValueStatus(params: { ids: number[]; status: number } & ScopePayload) {
  return request('/api/pms/productSpecValue/updateSpecValueStatus', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function queryProductSpecValueDetail(id: number, scope?: ScopePayload) {
  return request('/api/pms/productSpecValue/querySpecValueDetail', {
    method: 'GET',
    params: {
      id,
      ...scope,
    },
  });
}

export async function queryProductSpecValueList(params: ProductSpecValueListParams & ScopePayload) {
  return request('/api/pms/productSpecValue/querySpecValueList', {
    method: 'GET',
    params: {
      ...params,
    },
  });
}
