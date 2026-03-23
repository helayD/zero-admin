import {request} from 'umi';
import type { ProductCategoryListParams, ProductCategoryListItem } from './data.d';
import type { GovernanceScopeValue } from '@/pages/system/components/governance';

interface ScopePayload extends GovernanceScopeValue {
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
}

// 添加产品分类
export async function addProductCategory(params: ProductCategoryListItem & ScopePayload) {
  return request('/api/pms/category/addProductCategory', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

// 删除产品分类
export async function removeProductCategory(ids: number[], scope?: ScopePayload) {
  return request('/api/pms/category/deleteProductCategory', {
    method: 'GET',
    params: {
      ids: ids.join(','),
      ...scope,
    },
  });
}


// 更新产品分类
export async function updateProductCategory(params: ProductCategoryListItem & ScopePayload) {
  return request('/api/pms/category/updateProductCategory', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

// 更新产品分类状态
export async function updateProductCategoryStatus(params: { ids: number[], status: number } & ScopePayload) {
  return request('/api/pms/category/updateProductCategoryStatus', {
    method: 'POST',
    data: {
      ...params,
    },

  });
}

// 更新是否显示在导航栏
export async function updateProductCategoryNavStatus(params: { ids: number[], status: number } & ScopePayload) {
  return request('/api/pms/category/updateProductCategoryNavStatus', {
    method: 'POST',
    data: {
      ...params,
    },

  });
}


// 查询产品分类详情
export async function queryProductCategoryDetail(id: number, scope?: ScopePayload) {
  return request('/api/pms/category/queryProductCategoryDetail', {
    method: 'GET',
    params: {
      id,
      ...scope,
    },
  });
}

// 分页查询产品分类列表
export async function queryProductCategoryList(params: ProductCategoryListParams) {

  return request('/api/pms/category/queryProductCategoryList', {
    method: 'GET',
    params: {
      ...params,
    },
  });
}
