export type GovernanceScopeType = 'platform' | 'tenant' | 'merchant';

export interface DigitalCardAssetItem {
  assetInstanceId: number;
  assetNo: string;
  memberId: number;
  activityId: number;
  activityName: string;
  templateId: number;
  templateName: string;
  cardFaceImage: string;
  rarity: string;
  tokenId: string;
  mintTaskId: number;
  mintStatus: string;
  mintStatusText: string;
  chainStatus: string;
  chainStatusText: string;
  displayStatus: string;
  displayStatusText: string;
  complianceStatus: string;
  complianceStatusText: string;
  tokenStatusText: string;
  complianceRuleSummary: string;
  lastReceiptSummary: string;
  obtainedAt: string;
  disposedAt: string;
  latestReasonSummary: string;
}

export interface DigitalCardAssetParticipationSummary {
  participationRecordId: number;
  requestId: string;
  resultType: string;
  resultStatus: string;
  resultStatusText: string;
  failureReason: string;
  createTime: string;
}

export interface DigitalCardAssetMintTaskSummary {
  taskId: number;
  requestId: string;
  traceId: string;
  taskStatus: string;
  taskStatusText: string;
  mintStatus: string;
  mintStatusText: string;
  chainStatus: string;
  chainStatusText: string;
  chainTxId: string;
  lastReceiptSummary: string;
  availableActions: string[];
}

export interface DigitalCardAssetLogItem {
  operationType: string;
  operatorType: string;
  fromStatus: string;
  toStatus: string;
  reasonText: string;
  traceId: string;
  payloadJson: string;
  createTime: string;
}

export interface DigitalCardAssetDetailData {
  item: DigitalCardAssetItem;
  participationSummary: DigitalCardAssetParticipationSummary;
  mintTaskSummary: DigitalCardAssetMintTaskSummary;
  traceId: string;
  requestId: string;
  ruleSnapshotJson: string;
  logs: DigitalCardAssetLogItem[];
  availableAssetActions: string[];
}

export interface QueryDigitalCardAssetListResp {
  code: string;
  message: string;
  data: {
    list: DigitalCardAssetItem[];
    total: number;
  };
  current: number;
  pageSize: number;
  total: number;
  success: boolean;
}

export interface QueryDigitalCardAssetDetailResp {
  code: string;
  message: string;
  data: DigitalCardAssetDetailData;
  success: boolean;
}

export interface DigitalCardAssetActionResp {
  code: string;
  message: string;
  assetInstanceId: number;
  displayStatus: string;
  displayStatusText: string;
  complianceStatus: string;
  complianceStatusText: string;
  success: boolean;
}

export interface DigitalCardAssetListParams {
  current?: number;
  pageSize?: number;
  activityId?: number;
  activityName?: string;
  memberId?: number;
  templateId?: number;
  templateName?: string;
  assetNo?: string;
  tokenId?: string;
  mintStatus?: string;
  chainStatus?: string;
  displayStatus?: string;
  complianceStatus?: string;
  startTime?: string;
  endTime?: string;
  scopeType?: GovernanceScopeType;
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
  filter?: Record<string, (string | number)[]>;
  sorter?: Record<string, string>;
}

export const MINT_STATUS_OPTIONS = [
  { label: '全部发放状态', value: '' },
  { label: '待发放', value: 'mint_pending' },
  { label: '链上处理中', value: 'mint_processing' },
  { label: '已到账', value: 'mint_success' },
  { label: '发放失败', value: 'mint_failed' },
  { label: '补偿中', value: 'mint_compensating' },
  { label: '人工复核中', value: 'mint_manual_review' },
  { label: '已冻结', value: 'mint_frozen' },
];

export const CHAIN_STATUS_OPTIONS = [
  { label: '全部链上状态', value: '' },
  { label: '未知', value: 'unknown' },
  { label: '处理中', value: 'processing' },
  { label: '成功', value: 'success' },
  { label: '失败', value: 'failed' },
  { label: '冻结', value: 'frozen' },
];

export const DISPLAY_STATUS_OPTIONS = [
  { label: '全部展示状态', value: '' },
  { label: '正常展示', value: 'display_visible' },
  { label: '受限展示', value: 'display_hidden' },
  { label: '已下线展示', value: 'display_offlined' },
  { label: '已回收', value: 'display_recycled' },
];

export const COMPLIANCE_STATUS_OPTIONS = [
  { label: '全部合规状态', value: '' },
  { label: '合规正常', value: 'compliance_clear' },
  { label: '人工复核中', value: 'compliance_review' },
  { label: '已限制展示', value: 'compliance_restricted' },
  { label: '回收处理中', value: 'compliance_recycle_requested' },
  { label: '已回收', value: 'compliance_recycled' },
];
