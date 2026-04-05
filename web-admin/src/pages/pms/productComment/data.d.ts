export interface BaseResp {
  code: string;
  message: string;
}

export interface CommentAuditLogItem {
  id: number;
  commentId: string;
  action: string;
  fromStatus: number;
  toStatus: number;
  operatorId: number;
  operatorName: string;
  remark: string;
  createdAt: string;
}

export interface CommentReplayItem {
  id: string;
  commentId: string;
  type: number;
  memberNickName: string;
  memberIcon: string;
  content: string;
  createTime: string;
}

export interface CommentListItem {
  id: string;
  productId: number;
  productName: string;
  memberNickName: string;
  memberId: number;
  star: number;
  content: string;
  pics: string;
  memberIcon: string;
  showStatus: number;
  auditStatus: number;
  hidden: number;
  auditRemark: string;
  auditorId: number;
  auditorName: string;
  auditedAt: string;
  appealStatus: number;
  appealReason: string;
  appealReply: string;
  appealedAt: string;
  appealHandledAt: string;
  productAttribute: string;
  replayCount: number;
  memberIp: string;
  createTime: string;
}

export interface CommentDetailData extends CommentListItem {
  collectCount: number;
  readCount: number;
}

export interface CommentDetailResponse {
  code: string;
  message: string;
  data: CommentDetailData;
  replays?: CommentReplayItem[];
  auditLogs?: CommentAuditLogItem[];
}

export interface QueryCommentListResp {
  code: string;
  message: string;
  current: number;
  data: CommentListItem[];
  pageSize: number;
  success: boolean;
  total: number;
}

export interface QueryCommentAuditLogResp {
  code: string;
  message: string;
  current: number;
  data: CommentAuditLogItem[];
  pageSize: number;
  success: boolean;
  total: number;
}

export interface CommentListParams {
  productId?: number;
  productName?: string;
  memberName?: string;
  showStatus?: number;
  auditStatus?: number;
  hidden?: number;
  startTime?: string;
  endTime?: string;
  pageSize?: number;
  current?: number;
}

export interface CommentDetailParams {
  id: string;
}

export interface AuditCommentParams {
  id: string;
  auditStatus: number;
  auditRemark?: string;
}

export interface HandleCommentAppealParams {
  id: string;
  appealStatus: number;
  appealReply: string;
}

export interface UpdateCommentParams {
  id?: string;
  ids?: string;
  showStatus: number;
}
