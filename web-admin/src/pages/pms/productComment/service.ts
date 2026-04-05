import { request } from 'umi';
import type {
  AuditCommentParams,
  BaseResp,
  CommentDetailParams,
  CommentDetailResponse,
  CommentListParams,
  HandleCommentAppealParams,
  QueryCommentAuditLogResp,
  QueryCommentListResp,
  UpdateCommentParams,
} from './data.d';

type GovernancePayload = {
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
};

export async function queryCommentList(params: CommentListParams, scope?: GovernancePayload) {
  return request<QueryCommentListResp>('/api/pms/comment/queryCommentList', {
    method: 'GET',
    params: {
      ...params,
      ...scope,
    },
  });
}

export async function queryCommentDetail(params: CommentDetailParams, scope?: GovernancePayload) {
  return request<CommentDetailResponse>('/api/pms/comment/queryCommentDetail', {
    method: 'GET',
    params: {
      ...params,
      ...scope,
    },
  });
}

export async function auditComment(params: AuditCommentParams, scope?: GovernancePayload) {
  return request<BaseResp>(`/api/pms/comment/${params.id}/audit`, {
    method: 'POST',
    data: {
      auditStatus: params.auditStatus,
      auditRemark: params.auditRemark,
      ...scope,
    },
  });
}

export async function restoreComment(id: string, scope?: GovernancePayload) {
  return request<BaseResp>(`/api/pms/comment/${id}/restore`, {
    method: 'POST',
    data: {
      ...scope,
    },
  });
}

export async function handleCommentAppeal(
  params: HandleCommentAppealParams,
  scope?: GovernancePayload,
) {
  return request<BaseResp>(`/api/pms/comment/${params.id}/appeal`, {
    method: 'POST',
    data: {
      appealStatus: params.appealStatus,
      appealReply: params.appealReply,
      ...scope,
    },
  });
}

export async function queryCommentAuditLog(
  params: { id: string; current?: number; pageSize?: number },
  scope?: GovernancePayload,
) {
  return request<QueryCommentAuditLogResp>('/api/pms/comment/audit-log', {
    method: 'GET',
    params: {
      ...params,
      ...scope,
    },
  });
}

export async function updateComment(params: UpdateCommentParams, scope?: GovernancePayload) {
  return request<BaseResp>('/api/pms/comment/updateComment', {
    method: 'POST',
    data: {
      ...params,
      ...scope,
    },
  });
}

export async function batchUpdateComment(ids: string[], showStatus: number, scope?: GovernancePayload) {
  return updateComment(
    {
      ids: ids.join(','),
      showStatus,
    },
    scope,
  );
}
