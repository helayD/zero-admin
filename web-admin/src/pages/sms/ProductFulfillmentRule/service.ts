import { request } from 'umi';
import type {
  AddProductFulfillmentRuleParams,
  CheckProductFulfillmentRuleBindingResponse,
  ProductFulfillmentRuleDetailResponse,
  ProductFulfillmentRuleListItem,
  ProductFulfillmentRuleListParams,
  UpdateProductFulfillmentRuleParams,
  UpdateProductFulfillmentRuleStatusParams,
} from './data.d';

type GovernancePayload = {
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
};

type QueryProductFulfillmentRuleListResp = {
  code?: string;
  message?: string;
  data?: {
    list?: ProductFulfillmentRuleListItem[];
    total?: number;
  };
};

export async function queryProductFulfillmentRuleList(params: ProductFulfillmentRuleListParams) {
  const response = await request<QueryProductFulfillmentRuleListResp>(
    '/api/sms/productFulfillmentRule/queryProductFulfillmentRuleList',
    {
      method: 'GET',
      params,
    },
  );

  const list = Array.isArray(response?.data?.list) ? response.data.list : [];
  const total = Number(response?.data?.total ?? 0);

  return {
    data: list,
    total,
    success: true,
  };
}

export async function queryProductFulfillmentRuleDetail(id: number, scope?: GovernancePayload) {
  return request<ProductFulfillmentRuleDetailResponse>(
    '/api/sms/productFulfillmentRule/queryProductFulfillmentRuleDetail',
    {
      method: 'GET',
      params: { id, ...scope },
    },
  );
}

export async function addProductFulfillmentRule(params: AddProductFulfillmentRuleParams) {
  return request('/api/sms/productFulfillmentRule/addProductFulfillmentRule', {
    method: 'POST',
    data: params,
  });
}

export async function updateProductFulfillmentRule(params: UpdateProductFulfillmentRuleParams) {
  return request('/api/sms/productFulfillmentRule/updateProductFulfillmentRule', {
    method: 'POST',
    data: params,
  });
}

export async function updateProductFulfillmentRuleStatus(
  params: UpdateProductFulfillmentRuleStatusParams,
) {
  return request('/api/sms/productFulfillmentRule/updateProductFulfillmentRuleStatus', {
    method: 'POST',
    data: params,
  });
}

export async function deleteProductFulfillmentRule(id: number, scope?: GovernancePayload) {
  return request('/api/sms/productFulfillmentRule/deleteProductFulfillmentRule', {
    method: 'POST',
    data: { id, ...scope },
  });
}

export async function checkProductFulfillmentRuleBinding(
  id: number,
  scope?: GovernancePayload,
) {
  return request<CheckProductFulfillmentRuleBindingResponse>(
    '/api/sms/productFulfillmentRule/checkProductFulfillmentRuleBinding',
    {
      method: 'GET',
      params: { id, ...scope },
    },
  );
}
