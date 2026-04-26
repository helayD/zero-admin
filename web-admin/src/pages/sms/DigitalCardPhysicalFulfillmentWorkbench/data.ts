export type GovernanceScopeType = 'platform' | 'tenant' | 'merchant';

export interface DigitalCardPhysicalFulfillmentItem {
  fulfillmentId: number;
  fulfillmentNo: string;
  assetInstanceId: number;
  assetNo: string;
  activityId: number;
  activityName: string;
  templateId: number;
  templateName: string;
  memberId: number;
  fulfillmentStatus: string;
  fulfillmentStatusText: string;
  productionStatus: string;
  productionStatusText: string;
  shippingStatus: string;
  shippingStatusText: string;
  productionBatchNo: string;
  carrierName: string;
  trackingNo: string;
  failureCode: string;
  failureReason: string;
  shippedAt: string;
  signedAt: string;
  updateTime: string;
}

export interface DigitalCardPhysicalFulfillmentTimelineItem {
  action: string;
  actionText: string;
  statusText: string;
  reason: string;
  createTime: string;
}

export interface DigitalCardPhysicalFulfillmentDetailData {
  item: DigitalCardPhysicalFulfillmentItem;
  receiverName: string;
  receiverPhone: string;
  addressSummary: string;
  requestId: string;
  traceId: string;
  timeline: DigitalCardPhysicalFulfillmentTimelineItem[];
}

export interface QueryDigitalCardPhysicalFulfillmentListResp {
  code: string;
  message: string;
  data: {
    list: DigitalCardPhysicalFulfillmentItem[];
    total: number;
  };
  current: number;
  pageSize: number;
  total: number;
  success: boolean;
}

export interface QueryDigitalCardPhysicalFulfillmentDetailResp {
  code: string;
  message: string;
  data: DigitalCardPhysicalFulfillmentDetailData;
  success: boolean;
}

export interface DigitalCardPhysicalFulfillmentActionResp {
  code: string;
  message: string;
  fulfillmentId: number;
  assetInstanceId: number;
  fulfillmentNo: string;
  fulfillmentStatus: string;
  fulfillmentStatusText: string;
  productionStatus: string;
  productionStatusText: string;
  shippingStatus: string;
  shippingStatusText: string;
  blockedReason: string;
  blockedReasonText: string;
  success: boolean;
}

export interface DigitalCardPhysicalFulfillmentListParams {
  current?: number;
  pageSize?: number;
  activityId?: number;
  templateId?: number;
  memberId?: number;
  assetNo?: string;
  fulfillmentNo?: string;
  productionBatchNo?: string;
  fulfillmentStatus?: string;
  productionStatus?: string;
  shippingStatus?: string;
  trackingNo?: string;
  failureCode?: string;
  startTime?: string;
  endTime?: string;
  scopeType?: GovernanceScopeType;
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
  filter?: Record<string, (string | number)[]>;
  sorter?: Record<string, string>;
}

export const FULFILLMENT_STATUS_OPTIONS = [
  { label: '全部履约状态', value: '' },
  { label: '待权益确认', value: 'pending_digital_confirmation' },
  { label: '待实名后发放', value: 'pending_real_name' },
  { label: '待确认地址', value: 'pending_address' },
  { label: '待制作', value: 'production_pending' },
  { label: '制作中', value: 'production_in_progress' },
  { label: '质检中', value: 'quality_checking' },
  { label: '待发货', value: 'ready_to_ship' },
  { label: '已发货', value: 'shipped' },
  { label: '运输中', value: 'in_transit' },
  { label: '已签收', value: 'signed' },
  { label: '履约异常', value: 'exception' },
  { label: '补发中', value: 'reissue_pending' },
  { label: '已取消', value: 'cancelled' },
];

export const PRODUCTION_STATUS_OPTIONS = [
  { label: '全部制作状态', value: '' },
  { label: '未开始', value: 'not_started' },
  { label: '待制作', value: 'queued' },
  { label: '制作中', value: 'printing' },
  { label: '质检中', value: 'quality_checking' },
  { label: '制作完成', value: 'completed' },
  { label: '制作失败', value: 'failed' },
];

export const SHIPPING_STATUS_OPTIONS = [
  { label: '全部物流状态', value: '' },
  { label: '待发货', value: 'pending' },
  { label: '已发货', value: 'shipped' },
  { label: '运输中', value: 'in_transit' },
  { label: '已送达', value: 'delivered' },
  { label: '已签收', value: 'signed' },
  { label: '物流异常', value: 'exception' },
];
