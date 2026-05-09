import { request } from 'umi';
import type {
  AddCardTemplateParams,
  CardTemplateDetailResponse,
  CardTemplateListItem,
  CardTemplateListParams,
  CheckCardTemplateUsageResponse,
  UpdateCardTemplateParams,
  UpdateCardTemplateStatusParams,
} from './data.d';

type GovernancePayload = {
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
};

type QueryCardTemplateListResp = {
  code?: string;
  message?: string;
  data?: {
    list?: CardTemplateListItem[];
    total?: number;
  };
};

export async function queryCardTemplateList(params: CardTemplateListParams) {
  const response = await request<QueryCardTemplateListResp>(
    '/api/sms/cardTemplate/queryCardTemplateList',
    {
      method: 'GET',
      params,
    },
  );

  const list = Array.isArray(response?.data?.list) ? response.data!.list : [];
  const total = Number(response?.data?.total ?? 0);

  return {
    data: list,
    total,
    success: true,
  };
}

export async function queryCardTemplateDetail(id: number, scope?: GovernancePayload) {
  return request<CardTemplateDetailResponse>(
    '/api/sms/cardTemplate/queryCardTemplateDetail',
    {
      method: 'GET',
      params: { id, ...scope },
    },
  );
}

export async function addCardTemplate(params: AddCardTemplateParams) {
  return request('/api/sms/cardTemplate/addCardTemplate', {
    method: 'POST',
    data: params,
  });
}

export async function updateCardTemplate(params: UpdateCardTemplateParams) {
  return request('/api/sms/cardTemplate/updateCardTemplate', {
    method: 'POST',
    data: params,
  });
}

export async function updateCardTemplateStatus(params: UpdateCardTemplateStatusParams) {
  return request('/api/sms/cardTemplate/updateCardTemplateStatus', {
    method: 'POST',
    data: params,
  });
}

export async function deleteCardTemplate(id: number, scope?: GovernancePayload) {
  return request('/api/sms/cardTemplate/deleteCardTemplate', {
    method: 'GET',
    params: { id, ...scope },
  });
}

export async function checkCardTemplateUsage(id: number, scope?: GovernancePayload) {
  return request<CheckCardTemplateUsageResponse>(
    '/api/sms/cardTemplate/checkCardTemplateUsage',
    {
      method: 'GET',
      params: { id, ...scope },
    },
  );
}
