import { request } from 'umi';
import type { ProductAttributeListParams, ProductAttributeListItem } from './data.d';
import type { GovernanceScopeValue } from '@/pages/system/components/governance';

interface ScopePayload extends GovernanceScopeValue {
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
}

export async function addProductAttribute(params: ProductAttributeListItem & ScopePayload) {
  return request('/api/pms/attribute/addAttribute', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function removeProductAttribute(ids: number[], scope?: ScopePayload) {
  return request('/api/pms/attribute/deleteAttribute', {
    method: 'GET',
    params: {
      ids: ids.join(','),
      ...scope,
    },
  });
}

export async function updateProductAttribute(params: ProductAttributeListItem & ScopePayload) {
  return request('/api/pms/attribute/updateAttribute', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function updateProductAttributeStatus(params: { ids: number[]; status: number } & ScopePayload) {
  return request('/api/pms/attribute/updateAttributeStatus', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function queryProductAttributeDetail(id: number, scope?: ScopePayload) {
  return request('/api/pms/attribute/queryAttributeDetail', {
    method: 'GET',
    params: {
      id,
      ...scope,
    },
  });
}

export async function queryProductAttributeList(params: ProductAttributeListParams & ScopePayload) {
  return request('/api/pms/attribute/queryAttributeList', {
    method: 'GET',
    params: {
      ...params,
    },
  });
}
