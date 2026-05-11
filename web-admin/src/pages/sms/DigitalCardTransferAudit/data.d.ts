// Story 10.7 Task 9.3 — 后台转赠审计记录列表项
export interface DigitalCardTransferLogListItem {
  id: number;
  assetInstanceId: number;
  assetNo: string;
  templateName: string;
  operationType: string;
  operatorType: string;
  fromStatus: string;
  toStatus: string;
  reasonCode: string;
  reasonText: string;
  traceId: string;
  payloadJson: string;
  createTime: string;
}

export interface QueryDigitalCardTransferLogListParams {
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
  current?: number;
  pageSize?: number;
  assetInstanceId?: number;
  assetNo?: string;
  operationType?: string;
  traceId?: string;
  dateFrom?: string;
  dateTo?: string;
}

export interface QueryDigitalCardTransferLogListResp {
  code: string;
  message: string;
  total: number;
  current: number;
  pageSize: number;
  data: DigitalCardTransferLogListItem[];
  success: boolean;
}

// Story 10.7 Task 9.5 — operation_type 中文映射
export const OPERATION_TYPE_LABELS: Record<string, string> = {
  holder_transferred: '持有人变更',
  claim_token_generated: '生成分享凭证',
  claim_token_revoked: '吊销分享凭证',
  claim_attempt_failed: '领取尝试失败',
  asset_transferred: '越权直转(已下线)',
  asset_transferred_rolled_back: '越权回滚',
  holder_transferred_backfilled: '事后追认',
};
