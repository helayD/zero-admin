import { request } from 'umi';
import type {
  ProductSpuDetailResponse,
  ProductSpuListParams,
  ProductSpuNestedPayload,
  ProductSpuSubmitPayload,
} from './data.d';

type GovernancePayload = {
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
};

const productDataFieldNames = [
  'id',
  'name',
  'productSn',
  'categoryId',
  'categoryIds',
  'categoryName',
  'brandId',
  'brandName',
  'unit',
  'weight',
  'keywords',
  'albumPics',
  'mainPic',
  'publishStatus',
  'newStatus',
  'recommendStatus',
  'verifyStatus',
  'previewStatus',
  'sort',
  'newStatusSort',
  'recommendStatusSort',
  'stock',
  'lowStock',
  'promotionType',
  'subTitle',
  'detailHtml',
  'detailMobileHtml',
] as const;

const emptyNestedPayload: Required<ProductSpuNestedPayload> = {
  ladderList: [],
  fullList: [],
  memberPriceList: [],
  skuList: [],
  attributeValueList: [],
  subjectIds: [],
  prefrenceAreaIds: [],
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

function buildProductSpuPayload(params: ProductSpuSubmitPayload) {
  const {
    scopeType,
    platformId,
    tenantId,
    merchantId,
    ladderList,
    fullList,
    memberPriceList,
    skuList,
    attributeValueList,
    subjectIds,
    prefrenceAreaIds,
    ...rawProductData
  } = params;

  const productData = productDataFieldNames.reduce<Record<string, any>>((acc, fieldName) => {
    const value =
      fieldName === 'subTitle'
        ? rawProductData.subTitle ?? rawProductData.detailTitle
        : rawProductData[fieldName];
    if (value !== undefined) {
      acc[fieldName] = value;
    }
    return acc;
  }, {});

  return {
    productData,
    ladderList: Array.isArray(ladderList) ? ladderList : emptyNestedPayload.ladderList,
    fullList: Array.isArray(fullList) ? fullList : emptyNestedPayload.fullList,
    memberPriceList: Array.isArray(memberPriceList)
      ? memberPriceList
      : emptyNestedPayload.memberPriceList,
    skuList: Array.isArray(skuList) ? skuList : emptyNestedPayload.skuList,
    attributeValueList: Array.isArray(attributeValueList)
      ? attributeValueList
      : emptyNestedPayload.attributeValueList,
    ...buildRelationPayload({ subjectIds, prefrenceAreaIds }),
    scopeType,
    platformId,
    tenantId,
    merchantId,
  };
}

export async function addProductSpu(params: ProductSpuSubmitPayload) {
  return request('/api/pms/product/addProductSpu', {
    method: 'POST',
    data: buildProductSpuPayload(params),
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

export async function updateProductSpu(params: ProductSpuSubmitPayload) {
  return request('/api/pms/product/updateProductSpu', {
    method: 'POST',
    data: buildProductSpuPayload(params),
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
  return request<ProductSpuDetailResponse>('/api/pms/product/queryProductSpuDetail', {
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
