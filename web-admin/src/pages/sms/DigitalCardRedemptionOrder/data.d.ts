/**
 * Story 10.7 Task 2.8 / S3+S4 — 后台提货单管理列表类型
 */
import type { GovernancePayload } from '@/pages/system/components/governance';

export interface DigitalCardRedemptionOrderListItem {
  id: number;
  orderNo: string;
  cardInstanceId: number;
  assetNo: string;
  templateName: string;
  holderId: number;
  receiverName: string;
  receiverPhone: string;
  receiverAddress: string;
  status: string;
  shippedAt: string;
  deliveredAt: string;
  cancelReason: string;
  omsOrderId: number;
  platformId: number;
  tenantId: number;
  merchantId: number;
  createTime: string;
  updateTime: string;
}

export interface QueryDigitalCardRedemptionOrderListParams extends GovernancePayload {
  current?: number;
  pageSize?: number;
  orderId?: number;
  orderNo?: string;
  cardInstanceId?: number;
  assetNo?: string;
  holderId?: number;
  status?: string;
  omsOrderId?: number;
  dateFrom?: string;
  dateTo?: string;
}

export interface QueryDigitalCardRedemptionOrderListResp {
  code: string;
  message: string;
  total: number;
  current: number;
  pageSize: number;
  data: DigitalCardRedemptionOrderListItem[];
  success: boolean;
}

/** 提货单状态：sms_card_redemption_order.status */
export const REDEMPTION_ORDER_STATUS_LABELS: Record<string, string> = {
  pending: '待处理',
  processing: '处理中',
  shipped: '已发货',
  delivered: '已签收',
  cancelled: '已取消',
  failed: '已失败',
};

export const REDEMPTION_ORDER_STATUS_TAG_COLORS: Record<string, string> = {
  pending: 'default',
  processing: 'processing',
  shipped: 'blue',
  delivered: 'success',
  cancelled: 'warning',
  failed: 'error',
};
