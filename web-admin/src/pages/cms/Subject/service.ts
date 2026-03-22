import { request } from 'umi';
import type { SubjectListItem, SubjectListParams } from './data.d';

type GovernancePayload = {
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
};

export async function addSubject(params: SubjectListItem & GovernancePayload) {
  return request('/api/cms/subject/addSubject', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function updateSubject(params: SubjectListItem & GovernancePayload) {
  return request('/api/cms/subject/updateSubject', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function updateSubjectStatus(
  params: GovernancePayload & {
    ids: number[];
    showStatus: number;
    recommendStatus: number;
  },
) {
  return request('/api/cms/subject/updateSubjectStatus', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function removeSubject(ids: number[], scope?: GovernancePayload) {
  return request('/api/cms/subject/deleteSubject', {
    method: 'GET',
    params: {
      ids: ids.join(','),
      ...scope,
    },
  });
}

export async function querySubjectDetail(id: number, scope?: GovernancePayload) {
  return request('/api/cms/subject/querySubjectDetail', {
    method: 'GET',
    params: {
      id,
      ...scope,
    },
  });
}

export async function querySubjectList(params: SubjectListParams) {
  return request('/api/cms/subject/querySubjectList', {
    method: 'GET',
    params: {
      ...params,
    },
  });
}
