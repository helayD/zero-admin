import {request} from 'umi';
import type { ProductBrandListParams, ProductBrandListItem } from './data.d';
import type { GovernanceScopeValue } from '@/pages/system/components/governance';

interface ScopePayload extends GovernanceScopeValue {
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
}

// 添加商品品牌
export async function addProductBrand(params: ProductBrandListItem & ScopePayload) {
  return request('/api/pms/brand/addProductBrand', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

// 删除商品品牌
export async function removeProductBrand(ids: number[], scope?: ScopePayload) {
  return request('/api/pms/brand/deleteProductBrand', {
    method: 'GET',
    params: {
      ids: ids.join(','),
      ...scope,
    },
  });
}


// 更新商品品牌
export async function updateProductBrand(params: ProductBrandListItem & ScopePayload) {
  return request('/api/pms/brand/updateProductBrand', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

// 批量更新商品品牌状态
export async function updateProductBrandStatus(params: { ids: number[], status: number } & ScopePayload) {
  return request('/api/pms/brand/updateProductBrandStatus', {
    method: 'POST',
    data: {
      ...params,
    },

  });
}

// 批量更新商品品牌状态
export async function updateProductBrandRecommendStatus(params: { ids: number[], status: number } & ScopePayload) {
  return request('/api/pms/brand/updateProductBrandRecommendStatus', {
    method: 'POST',
    data: {
      ...params,
    },

  });
}

// 查询商品品牌详情
export async function queryProductBrandDetail(id: number, scope?: ScopePayload) {
  return request('/api/pms/brand/queryProductBrandDetail', {
    method: 'GET',
    params: {
      id,
      ...scope,
    },
  });
}

// 分页查询商品品牌列表
export async function queryProductBrandList(params: ProductBrandListParams) {

  return request('/api/pms/brand/queryProductBrandList', {
    method: 'GET',
    params: {
      ...params,
    },
  });
}
