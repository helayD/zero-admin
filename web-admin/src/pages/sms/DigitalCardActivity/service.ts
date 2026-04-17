import { request } from 'umi';
import type {
  DrawActivityFormValues,
  DrawActivityListParams,
  PreviewDrawActivityPublishReadinessData,
} from './data.d';
import { serializeDrawActivityPayload } from './helper';

export async function queryDrawActivityList(params: DrawActivityListParams) {
  return request('/api/sms/drawActivity/queryDrawActivityList', {
    method: 'GET',
    params,
  });
}

export async function queryDrawActivityDetail(id: number, params: Partial<DrawActivityListParams>) {
  return request('/api/sms/drawActivity/queryDrawActivityDetail', {
    method: 'GET',
    params: {
      id,
      ...params,
    },
  });
}

export async function addDrawActivity(
  params: DrawActivityFormValues & Partial<DrawActivityListParams>,
) {
  return request('/api/sms/drawActivity/addDrawActivity', {
    method: 'POST',
    data: {
      ...serializeDrawActivityPayload(params),
      scopeType: params.scopeType,
      platformId: params.platformId,
      tenantId: params.tenantId,
      merchantId: params.merchantId,
    },
  });
}

export async function updateDrawActivity(
  params: DrawActivityFormValues & Partial<DrawActivityListParams>,
) {
  return request('/api/sms/drawActivity/updateDrawActivity', {
    method: 'POST',
    data: {
      ...serializeDrawActivityPayload(params),
      scopeType: params.scopeType,
      platformId: params.platformId,
      tenantId: params.tenantId,
      merchantId: params.merchantId,
    },
  });
}

export async function removeDrawActivity(
  params: { ids: number[] } & Partial<DrawActivityListParams>,
) {
  return request('/api/sms/drawActivity/deleteDrawActivity', {
    method: 'GET',
    params,
  });
}

export async function updateDrawActivityStatus(
  params: { ids: number[]; status: number } & Partial<DrawActivityListParams>,
) {
  return request('/api/sms/drawActivity/updateDrawActivityStatus', {
    method: 'POST',
    data: params,
  });
}

export async function previewDrawActivityPublishReadiness(
  id: number,
  params: Partial<DrawActivityListParams>,
) {
  return request<{
    data: PreviewDrawActivityPublishReadinessData;
  }>('/api/sms/drawActivity/previewDrawActivityPublishReadiness', {
    method: 'GET',
    params: {
      id,
      ...params,
    },
  });
}
