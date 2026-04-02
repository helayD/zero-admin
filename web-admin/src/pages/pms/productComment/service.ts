import { request } from 'umi';
import type {
  CommentDetailParams,
  CommentDetailResponse,
  CommentListData,
  CommentListParams,
  UpdateCommentParams,
} from './data.d';

type GovernancePayload = {
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
};

export async function queryCommentList(
  params: CommentListParams,
  scope?: GovernancePayload,
) {
  // Review Fix H-4: Admin API 响应结构调整为 { data: { list, pagination } }
  return request<{ data: { list: CommentListData; pagination: { total: number } }; code: number; message: string }>(
    '/api/pms/comment/queryCommentList',
    {
      method: 'GET',
      params: {
        ...params,
        ...scope,
      },
    },
  );
}

export async function queryCommentDetail(
  params: CommentDetailParams,
  scope?: GovernancePayload,
) {
  return request<CommentDetailResponse>('/api/pms/comment/queryCommentDetail', {
    method: 'GET',
    params: {
      ...params,
      ...scope,
    },
  });
}

export async function updateComment(
  params: UpdateCommentParams,
  scope?: GovernancePayload,
) {
  return request('/api/pms/comment/updateComment', {
    method: 'POST',
    data: {
      ...params,
      ...scope,
    },
  });
}

export async function batchUpdateComment(
  ids: string[],
  showStatus: number,
  scope?: GovernancePayload,
) {
  return request('/api/pms/comment/updateComment', {
    method: 'POST',
    data: {
      ids: ids.join(','),
      showStatus,
      ...scope,
    },
  });
}
