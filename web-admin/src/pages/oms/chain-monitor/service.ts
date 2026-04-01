import { request } from 'umi';
import type { ChainMonitorListParams } from './data';

type GovernancePayload = {
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
};

export async function queryChainMonitorList(params: ChainMonitorListParams) {
  const res = await request('/api/oms/order/queryChainMonitorList', {
    method: 'GET',
    params: {
      ...params,
      chainType: params.chainType ?? 0,
      consistencyStage: params.consistencyStage ?? 0,
      consistencyResult: params.consistencyResult ?? 0,
      manualRequired: params.manualRequired ?? 0,
      pageSize: params.pageSize ?? 10,
      current: params.current ?? 1,
    },
  });

  return {
    ...res,
    data: Array.isArray(res?.data) ? res.data : [],
  };
}

export async function exportChainMonitorList(params: ChainMonitorListParams & GovernancePayload) {
  const res = await request('/api/oms/order/exportChainMonitorList', {
    method: 'GET',
    params: {
      ...params,
      chainType: params.chainType ?? 0,
      consistencyStage: params.consistencyStage ?? 0,
      consistencyResult: params.consistencyResult ?? 0,
      manualRequired: params.manualRequired ?? 0,
      pageSize: 10000,
      current: 1,
    },
    responseType: 'blob',
  });
  return res;
}
