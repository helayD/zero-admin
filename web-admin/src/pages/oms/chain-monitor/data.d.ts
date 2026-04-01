export interface ChainMonitorItem {
  traceId: string;
  platformId: number;
  tenantId: number;
  merchantId: number;
  chainType: number;
  chainTypeText: string;
  entityId: number;
  entityNo: string;
  entityType: number;
  stage: number;
  stageText: string;
  result: number;
  resultText: string;
  retryCount: number;
  lastError: string;
  lastExecuteAt: string;
  createdAt: string;
  actorId: number;
}

export interface ChainMonitorListPagination {
  total: number;
  pageSize: number;
  current: number;
}

export interface ChainMonitorListData {
  list: ChainMonitorItem[];
  pagination: Partial<ChainMonitorListPagination>;
}

export interface ChainMonitorListParams {
  chainType?: number;
  consistencyStage?: number;
  consistencyResult?: number;
  manualRequired?: number;
  startTime?: string;
  endTime?: string;
  orderNo?: string;
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
  pageSize?: number;
  current?: number;
  filter?: { [key: string]: any[] };
  sorter?: { [key: string]: any };
}

export const CHAIN_TYPE_OPTIONS = [
  { label: '全部', value: 0 },
  { label: '订单超时补偿', value: 1 },
];

export const CHAIN_STAGE_OPTIONS = [
  { label: '不限', value: 0 },
  { label: '正常', value: 1 },
  { label: '待处理', value: 2 },
  { label: '支付中', value: 3 },
  { label: '支付确认中', value: 4 },
  { label: '取消补偿中', value: 5 },
  { label: '已取消/已关闭', value: 6 },
  { label: '已完成', value: 7 },
  { label: '售后处理中', value: 8 },
  { label: '已关闭', value: 9 },
  { label: '需人工介入', value: 10 },
];

export const CHAIN_RESULT_OPTIONS = [
  { label: '不限', value: 0 },
  { label: '无结果', value: 1 },
  { label: '处理中', value: 2 },
  { label: '成功', value: 3 },
  { label: '失败', value: 4 },
  { label: '人工处理', value: 5 },
];

export const MANUAL_REQUIRED_OPTIONS = [
  { label: '不限', value: 0 },
  { label: '仅需人工介入', value: 1 },
  { label: '仅自动处理', value: 2 },
];
