import { request } from 'umi';
import type {
  CreateMerchantResponse,
  MerchantDetailResponse,
  MerchantListParams,
  MerchantListResponse,
  MerchantReviewPayload,
  MerchantStatusPayload,
} from './data.d';

export async function queryMerchantList(params: MerchantListParams) {
  return request<MerchantListResponse>('/api/sys/merchant/queryMerchantList', {
    method: 'GET',
    params: {
      ...params,
    },
  });
}

export async function queryMerchantDetail(id: number) {
  return request<MerchantDetailResponse>('/api/sys/merchant/queryMerchantDetail', {
    method: 'GET',
    params: {
      id,
    },
  });
}

export async function createMerchant(params: Record<string, any>) {
  return request<CreateMerchantResponse>('/api/sys/merchant/createMerchant', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function approveMerchant(params: MerchantReviewPayload) {
  return request('/api/sys/merchant/approveMerchant', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function rejectMerchant(params: MerchantReviewPayload) {
  return request('/api/sys/merchant/rejectMerchant', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function requestMerchantMaterial(params: MerchantReviewPayload) {
  return request('/api/sys/merchant/requestMerchantMaterial', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function enableMerchant(params: MerchantStatusPayload) {
  return request('/api/sys/merchant/enableMerchant', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function disableMerchant(params: MerchantStatusPayload) {
  return request('/api/sys/merchant/disableMerchant', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function archiveMerchant(params: MerchantStatusPayload) {
  return request('/api/sys/merchant/archiveMerchant', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}
