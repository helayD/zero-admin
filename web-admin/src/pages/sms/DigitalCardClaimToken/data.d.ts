/**
 * Story 10.7 Task 8.8 / S5+S6 — 后台分享凭证管理类型
 *
 * 监管约束：后台只能看到脱敏后的 token（首 8 位 + ***），原始 token 不暴露
 */
import type { GovernancePayload } from '@/pages/system/components/governance';

export interface DigitalCardClaimTokenListItem {
  id: number;
  tokenMasked: string;
  cardInstanceId: number;
  assetNo: string;
  templateName: string;
  issuerId: number;
  issuerType: string;
  expireAt: string;
  maxClaims: number;
  claimedCount: number;
  status: string;
  claimedBy: number;
  claimedAt: string;
  platformId: number;
  tenantId: number;
  merchantId: number;
  createTime: string;
  updateTime: string;
}

export interface QueryDigitalCardClaimTokenListParams extends GovernancePayload {
  current?: number;
  pageSize?: number;
  tokenId?: number;
  cardInstanceId?: number;
  assetNo?: string;
  issuerId?: number;
  status?: string;
  dateFrom?: string;
  dateTo?: string;
}

export interface QueryDigitalCardClaimTokenListResp {
  code: string;
  message: string;
  total: number;
  current: number;
  pageSize: number;
  data: DigitalCardClaimTokenListItem[];
  success: boolean;
}

export interface AdminRevokeDigitalCardClaimTokenParams extends GovernancePayload {
  tokenId: number;
  reason: string;
  traceId?: string;
}

export interface AdminRevokeDigitalCardClaimTokenResp {
  code: string;
  message: string;
  tokenId: number;
  fromStatus: string;
  toStatus: string;
  success: boolean;
}

/** sms_card_claim_token.status */
export const CLAIM_TOKEN_STATUS_LABELS: Record<string, string> = {
  active: '活跃',
  consumed: '已被领取',
  revoked: '已吊销',
  expired: '已过期',
};

export const CLAIM_TOKEN_STATUS_TAG_COLORS: Record<string, string> = {
  active: 'success',
  consumed: 'blue',
  revoked: 'warning',
  expired: 'default',
};
