import { request } from 'umi';
import type { PreferredAreaListParams, PreferredAreaListItem } from './data.d';

type GovernancePayload = {
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
};

export async function addPreferredArea(params: PreferredAreaListItem & GovernancePayload) {
  return request('/api/cms/prefrenceArea/addPrefrenceArea', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function updatePreferredArea(params: PreferredAreaListItem & GovernancePayload) {
  return request('/api/cms/prefrenceArea/updatePrefrenceArea', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function updatePreferredAreaStatus(
  params: GovernancePayload & {
    ids: number[];
    showStatus: number;
  },
) {
  return request('/api/cms/prefrenceArea/updatePrefrenceAreaStatus', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function removePreferredArea(ids: number[], scope?: GovernancePayload) {
  return request('/api/cms/prefrenceArea/deletePrefrenceArea', {
    method: 'GET',
    params: {
      ids: ids.join(','),
      ...scope,
    },
  });
}

export async function queryPreferredAreaDetail(id: number, scope?: GovernancePayload) {
  return request('/api/cms/prefrenceArea/queryPrefrenceAreaDetail', {
    method: 'GET',
    params: {
      id,
      ...scope,
    },
  });
}

export async function queryPreferredAreaList(params: PreferredAreaListParams) {
  return request('/api/cms/prefrenceArea/queryPrefrenceAreaList', {
    method: 'GET',
    params: {
      ...params,
    },
  });
}
