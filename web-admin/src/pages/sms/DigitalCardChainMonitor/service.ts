import { request } from 'umi';
import type {
  DigitalCardChainActionResp,
  DigitalCardChainListParams,
  QueryDigitalCardChainActionsResp,
  QueryDigitalCardChainDetailResp,
  QueryDigitalCardChainListResp,
} from './data';
import { normalizeDigitalCardChainListResponse } from './helper';

type GovernancePayload = {
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
};

type ActionPayload = GovernancePayload & {
  taskId: number;
  reason: string;
};

export async function queryDigitalCardChainList(params: DigitalCardChainListParams) {
  const response = await request<QueryDigitalCardChainListResp>(
    '/api/sms/digitalCardChain/queryDigitalCardChainList',
    {
      method: 'GET',
      params: {
        ...params,
        current: params.current ?? 1,
        pageSize: params.pageSize ?? 20,
        manualRequired: params.manualRequired ?? 0,
      },
    },
  );

  return normalizeDigitalCardChainListResponse(response);
}

export async function queryDigitalCardChainDetail(taskId: number) {
  return request<QueryDigitalCardChainDetailResp>(
    '/api/sms/digitalCardChain/queryDigitalCardChainDetail',
    {
      method: 'GET',
      params: { taskId },
    },
  );
}

export async function queryDigitalCardChainActions(taskId: number) {
  return request<QueryDigitalCardChainActionsResp>(
    '/api/sms/digitalCardChain/queryDigitalCardChainActions',
    {
      method: 'GET',
      params: { taskId },
    },
  );
}

export async function retryDigitalCardChain(payload: ActionPayload) {
  return request<DigitalCardChainActionResp>('/api/sms/digitalCardChain/retryDigitalCardChain', {
    method: 'POST',
    data: payload,
  });
}

export async function freezeDigitalCardChain(payload: ActionPayload) {
  return request<DigitalCardChainActionResp>('/api/sms/digitalCardChain/freezeDigitalCardChain', {
    method: 'POST',
    data: payload,
  });
}

export async function escalateDigitalCardChain(payload: ActionPayload) {
  return request<DigitalCardChainActionResp>('/api/sms/digitalCardChain/escalateDigitalCardChain', {
    method: 'POST',
    data: payload,
  });
}
