import { request } from 'umi';
import type { ProductSkuListItem, ProductSkuListParams } from './data.d';

type GovernancePayload = {
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
};

export async function addProductSku(params: ProductSkuListItem & GovernancePayload) {
  return request('/api/pms/product/addProductSku', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function removeProductSku(ids: number[], scope?: GovernancePayload) {
  return request('/api/pms/product/deleteProductSku', {
    method: 'GET',
    params: {
      ids: ids.join(','),
      ...scope,
    },
  });
}

export async function updateProductSku(
  params: GovernancePayload & { data: ProductSkuListItem[] },
) {
  return request('/api/pms/product/updateProductSku', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function queryProductSkuDetail(id: number, scope?: GovernancePayload) {
  return request('/api/pms/product/queryProductSkuDetail', {
    method: 'GET',
    params: {
      id,
      ...scope,
    },
  });
}

export async function queryProductSkuList(params: ProductSkuListParams) {
  return request('/api/pms/product/queryProductSkuList', {
    method: 'GET',
    params: {
      ...params,
    },
  });
}
