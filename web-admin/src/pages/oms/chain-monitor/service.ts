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

// ============================================================
// 链路干预 API（Story 7.6 新增）
// ============================================================

export async function retryChain(orderId: number, remark?: string) {
  return request('/api/oms/order/retryChain', {
    method: 'POST',
    data: { orderId, remark },
  });
}

export async function replayChain(orderId: number, replayReason: string) {
  return request('/api/oms/order/replayChain', {
    method: 'POST',
    data: { orderId, replayReason },
  });
}

export async function pauseChain(orderId: number, pauseReason: string) {
  return request('/api/oms/order/pauseChain', {
    method: 'POST',
    data: { orderId, pauseReason },
  });
}

export async function escalateChain(orderId: number, escalateReason: string) {
  return request('/api/oms/order/escalateChain', {
    method: 'POST',
    data: { orderId, escalateReason },
  });
}

export async function queryChainActions(orderId: number) {
  return request('/api/oms/order/queryChainActions', {
    method: 'GET',
    params: { orderId },
  });
}
