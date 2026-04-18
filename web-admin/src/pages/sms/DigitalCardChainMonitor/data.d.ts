export type GovernanceScopeType = 'platform' | 'tenant' | 'merchant';

export interface DigitalCardChainItem {
  taskId: number;
  traceId: string;
  assetInstanceId: number;
  assetNo: string;
  memberId: number;
  activityId: number;
  activityName: string;
  templateId: number;
  templateName: string;
  tokenId: string;
  taskStatus: string;
  taskStatusText: string;
  mintStatus: string;
  mintStatusText: string;
  chainStatus: string;
  chainStatusText: string;
  retryCount: number;
  lastError: string;
  lastExecuteAt: string;
  manualRequired: boolean;
  frozen: boolean;
  freezeReason: string;
  lastReceiptSummary: string;
  assetStatusText: string;
  participationRecordId: number;
}

export interface DigitalCardChainLogItem {
  operationType: string;
  operatorType: string;
  fromStatus: string;
  toStatus: string;
  reasonText: string;
  traceId: string;
  payloadJson: string;
  createTime: string;
}

export interface DigitalCardChainDetailData {
  item: DigitalCardChainItem;
  requestId: string;
  chainTxId: string;
  lastReceiptJson: string;
  logs: DigitalCardChainLogItem[];
}

export interface QueryDigitalCardChainListData {
  list: DigitalCardChainItem[];
  total: number;
}

export interface QueryDigitalCardChainListResp {
  code: string;
  message: string;
  data: QueryDigitalCardChainListData;
  current: number;
  pageSize: number;
  total: number;
  success: boolean;
}

export interface QueryDigitalCardChainDetailResp {
  code: string;
  message: string;
  data: DigitalCardChainDetailData;
  success: boolean;
}

export interface QueryDigitalCardChainActionsResp {
  code: string;
  message: string;
  availableActions: string[];
  success: boolean;
}

export interface DigitalCardChainActionResp {
  code: string;
  message: string;
  taskStatus: string;
  mintStatus: string;
  chainStatus: string;
  retryCount: number;
  manualRequired: boolean;
  frozen: boolean;
  success: boolean;
}

export interface DigitalCardChainListParams {
  current?: number;
  pageSize?: number;
  activityId?: number;
  activityName?: string;
  memberId?: number;
  templateId?: number;
  assetNo?: string;
  tokenId?: string;
  taskStatus?: string;
  mintStatus?: string;
  chainStatus?: string;
  manualRequired?: number;
  startTime?: string;
  endTime?: string;
  scopeType?: GovernanceScopeType;
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
  filter?: Record<string, (string | number)[]>;
  sorter?: Record<string, string>;
}

export const TASK_STATUS_OPTIONS = [
  { label: '全部任务状态', value: '' },
  { label: '待派发', value: 'pending_dispatch' },
  { label: '已派发', value: 'dispatched' },
  { label: '执行中', value: 'running' },
  { label: '成功', value: 'succeeded' },
  { label: '失败', value: 'failed' },
  { label: '人工复核', value: 'manual_review' },
  { label: '已冻结', value: 'frozen' },
];

export const MINT_STATUS_OPTIONS = [
  { label: '全部发放状态', value: '' },
  { label: '待发放', value: 'mint_pending' },
  { label: '发放中', value: 'mint_processing' },
  { label: '发放成功', value: 'mint_success' },
  { label: '发放失败', value: 'mint_failed' },
  { label: '补偿中', value: 'mint_compensating' },
  { label: '人工复核', value: 'mint_manual_review' },
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

export const MANUAL_REQUIRED_OPTIONS = [
  { label: '全部', value: 0 },
  { label: '仅人工复核', value: 1 },
  { label: '仅自动链路', value: 2 },
];
