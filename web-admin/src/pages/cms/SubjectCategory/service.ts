import { request } from 'umi';
import type { SubjectCategoryListParams, SubjectCategoryListItem } from './data.d';

type GovernancePayload = {
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
};

export async function addSubjectCategory(params: SubjectCategoryListItem & GovernancePayload) {
  return request('/api/cms/subjectCategory/addSubjectCategory', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function updateSubjectCategory(params: SubjectCategoryListItem & GovernancePayload) {
  return request('/api/cms/subjectCategory/updateSubjectCategory', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function updateSubjectCategoryStatus(
  params: GovernancePayload & {
    ids: number[];
    showStatus: number;
  },
) {
  return request('/api/cms/subjectCategory/updateSubjectCategoryStatus', {
    method: 'POST',
    data: {
      ...params,
    },
  });
}

export async function removeSubjectCategory(ids: number[], scope?: GovernancePayload) {
  return request('/api/cms/subjectCategory/deleteSubjectCategory', {
    method: 'GET',
    params: {
      ids: ids.join(','),
      ...scope,
    },
  });
}

export async function querySubjectCategoryDetail(id: number, scope?: GovernancePayload) {
  return request('/api/cms/subjectCategory/querySubjectCategoryDetail', {
    method: 'GET',
    params: {
      id,
      ...scope,
    },
  });
}

export async function querySubjectCategoryList(params: SubjectCategoryListParams) {
  return request('/api/cms/subjectCategory/querySubjectCategoryList', {
    method: 'GET',
    params: {
      ...params,
    },
  });
}
