import { request } from 'umi';
import type { ProductSpecListParams, ProductSpecListItem } from './data.d';
import type { GovernanceScopeValue } from '@/pages/system/components/governance';

interface ScopePayload extends GovernanceScopeValue {
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
}

export async function addProductSpec(params: ProductSpecListItem & ScopePayload) {
  return request('/api/pms/productSpec/addSpec', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function removeProductSpec(ids: number[], scope?: ScopePayload) {
  return request('/api/pms/productSpec/deleteSpec', {
    method: 'GET',
    params: {
      ids: ids.join(','),
      ...scope,
    },
  });
}

export async function updateProductSpec(params: ProductSpecListItem & ScopePayload) {
  return request('/api/pms/productSpec/updateSpec', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function updateProductSpecStatus(params: { ids: number[]; status: number } & ScopePayload) {
  return request('/api/pms/productSpec/updateSpecStatus', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function queryProductSpecDetail(id: number, scope?: ScopePayload) {
  return request('/api/pms/productSpec/querySpecDetail', {
    method: 'GET',
    params: {
      id,
      ...scope,
    },
  });
}

export async function queryProductSpecList(params: ProductSpecListParams & ScopePayload) {
  return request('/api/pms/productSpec/querySpecList', {
    method: 'GET',
    params: {
      ...params,
    },
  });
}
