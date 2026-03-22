import { request } from 'umi';
import type { SubjectListParams } from './data.d';

export async function querySubjectList(params: SubjectListParams) {
  return request('/api/cms/subject/querySubjectList', {
    method: 'GET',
    params: {
      ...params,
    },
  });
}
