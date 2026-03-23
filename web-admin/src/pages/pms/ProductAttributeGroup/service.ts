import { request } from 'umi';
import type { ProductAttributeGroupListParams, ProductAttributeGroupListItem } from './data.d';
import type { GovernanceScopeValue } from '@/pages/system/components/governance';

interface ScopePayload extends GovernanceScopeValue {
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
}

export async function addProductAttributeGroup(params: ProductAttributeGroupListItem & ScopePayload) {
  return request('/api/pms/attributeGroup/addAttributeGroup', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function removeProductAttributeGroup(ids: number[], scope?: ScopePayload) {
  return request('/api/pms/attributeGroup/deleteAttributeGroup', {
    method: 'GET',
    params: {
      ids: ids.join(','),
      ...scope,
    },
  });
}

export async function updateProductAttributeGroup(params: ProductAttributeGroupListItem & ScopePayload) {
  return request('/api/pms/attributeGroup/updateAttributeGroup', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function updateProductAttributeGroupStatus(params: { ids: number[]; status: number } & ScopePayload) {
  return request('/api/pms/attributeGroup/updateAttributeGroupStatus', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function queryProductAttributeGroupDetail(id: number, scope?: ScopePayload) {
  return request('/api/pms/attributeGroup/queryAttributeGroupDetail', {
    method: 'GET',
    params: {
      id,
      ...scope,
    },
  });
}

export async function queryProductAttributeGroupList(params: ProductAttributeGroupListParams & ScopePayload) {
  return request('/api/pms/attributeGroup/queryAttributeGroupList', {
    method: 'GET',
    params: {
      ...params,
    },
  });
}
