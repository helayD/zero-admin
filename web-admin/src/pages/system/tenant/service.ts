import { request } from 'umi';
import type {
  CreateTenantResponse,
  TenantDetailResponse,
  TenantListParams,
  TenantListResponse,
  TenantStatusPayload,
} from './data.d';

export async function queryTenantList(params: TenantListParams) {
  return request<TenantListResponse>('/api/sys/tenant/queryTenantList', {
    method: 'GET',
    params: {
      ...params,
    },
  });
}

export async function queryTenantDetail(id: number) {
  return request<TenantDetailResponse>('/api/sys/tenant/queryTenantDetail', {
    method: 'GET',
    params: {
      id,
    },
  });
}

export async function createTenant(params: Record<string, any>) {
  return request<CreateTenantResponse>('/api/sys/tenant/createTenant', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function enableTenant(params: TenantStatusPayload) {
  return request('/api/sys/tenant/enableTenant', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function disableTenant(params: TenantStatusPayload) {
  return request('/api/sys/tenant/disableTenant', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function archiveTenant(params: TenantStatusPayload) {
  return request('/api/sys/tenant/archiveTenant', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}
