import { request } from 'umi';
import type { ProductSpuListItem, ProductSpuListParams } from './data.d';

type GovernancePayload = {
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
};

const emptyNestedPayload = {
  ladderList: [],
  fullList: [],
  memberPriceList: [],
  skuList: [],
  attributeValueList: [],
};

type ProductSpuRelationPayload = {
  subjectIds?: number[];
  prefrenceAreaIds?: number[];
};

function buildRelationPayload(params: ProductSpuRelationPayload) {
  const relationPayload: ProductSpuRelationPayload = {};

  if (Array.isArray(params.subjectIds)) {
    relationPayload.subjectIds = params.subjectIds;
  }
  if (Array.isArray(params.prefrenceAreaIds)) {
    relationPayload.prefrenceAreaIds = params.prefrenceAreaIds;
  }

  return relationPayload;
}

export async function addProductSpu(
  params: ProductSpuListItem & GovernancePayload & ProductSpuRelationPayload,
) {
  const { scopeType, platformId, tenantId, merchantId, ...productData } = params;
  return request('/api/pms/product/addProductSpu', {
    method: 'POST',
    data: {
      productData,
      ...emptyNestedPayload,
      ...buildRelationPayload(params),
      scopeType,
      platformId,
      tenantId,
      merchantId,
    },
  });
}

export async function removeProductSpu(ids: number[], scope?: GovernancePayload) {
  return request('/api/pms/product/deleteProductSpu', {
    method: 'GET',
    params: {
      ids: ids.join(','),
      ...scope,
    },
  });
}

export async function updateProductSpu(
  params: ProductSpuListItem & GovernancePayload & ProductSpuRelationPayload,
) {
  const { scopeType, platformId, tenantId, merchantId, ...productData } = params;
  return request('/api/pms/product/updateProductSpu', {
    method: 'POST',
    data: {
      productData,
      ...emptyNestedPayload,
      ...buildRelationPayload(params),
      scopeType,
      platformId,
      tenantId,
      merchantId,
    },
  });
}

const productSpuStatusEndpoint: Record<string, string> = {
  publish: '/api/pms/product/updatePublishStatus',
  verify: '/api/pms/product/updateVerifyStatus',
  recommend: '/api/pms/product/updateRecommendStatus',
  new: '/api/pms/product/updateNewStatus',
  delete: '/api/pms/product/updateDeleteStatus',
};

export async function updateProductSpuStatus(
  action: 'publish' | 'verify' | 'recommend' | 'new' | 'delete',
  params: GovernancePayload & { ids: number[]; status: number; detail?: string },
) {
  return request(productSpuStatusEndpoint[action], {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function queryProductSpuDetail(id: number, scope?: GovernancePayload) {
  return request('/api/pms/product/queryProductSpuDetail', {
    method: 'GET',
    params: {
      id,
      ...scope,
    },
  });
}

export async function queryProductSpuList(params: ProductSpuListParams) {
  return request('/api/pms/product/queryProductSpuList', {
    method: 'GET',
    params: {
      ...params,
    },
  });
}
