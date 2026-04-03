import { request } from 'umi';
import type {
  QueryOperateFunnelDashboardData,
  QueryOperateFunnelDashboardParams,
  QueryOperateFunnelDashboardResp,
} from './data';

const emptyOverview = {
  exposure: 0,
  click: 0,
  addCart: 0,
  orderCreated: 0,
  paySuccess: 0,
  couponRedeem: 0,
  clickRate: 0,
  addCartRate: 0,
  orderRate: 0,
  payRate: 0,
  couponRedeemRate: 0,
  cards: [],
};

const normalizeDashboardData = (input?: Partial<QueryOperateFunnelDashboardData>): QueryOperateFunnelDashboardData => ({
  overview: {
    ...emptyOverview,
    ...(input?.overview || {}),
    cards: Array.isArray(input?.overview?.cards) ? input?.overview?.cards || [] : [],
  },
  series: Array.isArray(input?.series) ? input?.series || [] : [],
  activityOptions: Array.isArray(input?.activityOptions) ? input?.activityOptions || [] : [],
  trackingStartedAt: input?.trackingStartedAt || '',
  partialMetrics: Array.isArray(input?.partialMetrics) ? input?.partialMetrics || [] : [],
  bucket: input?.bucket || 'day',
});

export async function queryOperateFunnelDashboard(
  params: QueryOperateFunnelDashboardParams,
): Promise<QueryOperateFunnelDashboardResp> {
  const res = await request<QueryOperateFunnelDashboardResp>(
    '/api/sms/operateDashboard/queryOperateFunnelDashboard',
    {
      method: 'GET',
      params: {
        ...params,
        activityId: params.activityId || undefined,
      },
    },
  );

  return {
    ...res,
    data: normalizeDashboardData(res?.data),
  };
}
