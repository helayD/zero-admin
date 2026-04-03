export interface OperateFunnelMetricCard {
  key: string;
  label: string;
  value: number;
  rate: number;
  rateLabel: string;
}

export interface OperateFunnelOverview {
  exposure: number;
  click: number;
  addCart: number;
  orderCreated: number;
  paySuccess: number;
  couponRedeem: number;
  clickRate: number;
  addCartRate: number;
  orderRate: number;
  payRate: number;
  couponRedeemRate: number;
  cards: OperateFunnelMetricCard[];
}

export interface OperateFunnelSeriesPoint {
  bucketLabel: string;
  bucketStart: string;
  bucketEnd: string;
  exposure: number;
  click: number;
  addCart: number;
  orderCreated: number;
  paySuccess: number;
  couponRedeem: number;
  clickRate: number;
  addCartRate: number;
  orderRate: number;
  payRate: number;
  couponRedeemRate: number;
}

export interface OperateFunnelActivityOption {
  activityType: string;
  activityId: number;
  activityName: string;
  activityLabel: string;
}

export interface QueryOperateFunnelDashboardData {
  overview: OperateFunnelOverview;
  series: OperateFunnelSeriesPoint[];
  activityOptions: OperateFunnelActivityOption[];
  trackingStartedAt: string;
  partialMetrics: string[];
  bucket: 'day' | 'hour' | string;
}

export interface QueryOperateFunnelDashboardResp {
  code: string;
  message: string;
  success: boolean;
  data: QueryOperateFunnelDashboardData;
}

export interface QueryOperateFunnelDashboardParams {
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
  startTime?: string;
  endTime?: string;
  channel?: string;
  activityType?: string;
  activityId?: number;
  bucket?: 'day' | 'hour';
}

export interface OperateDashboardTableRow extends OperateFunnelSeriesPoint {
  key: string;
}
