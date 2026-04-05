import { request } from 'umi';
import type {
  ExportRepeatPurchaseAnalysisParams,
  QueryRepeatPurchaseAnalysisData,
  QueryRepeatPurchaseAnalysisParams,
  QueryRepeatPurchaseAnalysisResp,
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

const emptyRepeatPurchaseOverview = {
  paidBuyerCount: 0,
  repeatBuyerCount: 0,
  repeatRate: 0,
  repeatOrderCount: 0,
  repeatGmv: 0,
  avgDaysToRepeat: 0,
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

const normalizeRepeatPurchaseData = (
  input?: Partial<QueryRepeatPurchaseAnalysisData>,
): QueryRepeatPurchaseAnalysisData => ({
  overview: {
    ...emptyRepeatPurchaseOverview,
    ...(input?.overview || {}),
  },
  trends: Array.isArray(input?.trends) ? input?.trends || [] : [],
  details: Array.isArray(input?.details) ? input?.details || [] : [],
  total: Number(input?.total || 0),
  pageNum: Number(input?.pageNum || 1),
  pageSize: Number(input?.pageSize || 20),
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

export async function queryRepeatPurchaseAnalysis(
  params: QueryRepeatPurchaseAnalysisParams,
): Promise<QueryRepeatPurchaseAnalysisResp> {
  const res = await request<QueryRepeatPurchaseAnalysisResp>(
    '/api/sms/operateDashboard/queryRepeatPurchaseAnalysis',
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
    data: normalizeRepeatPurchaseData(res?.data),
  };
}

export async function exportRepeatPurchaseAnalysis(params: ExportRepeatPurchaseAnalysisParams) {
  const res = await request('/api/sms/operateDashboard/exportRepeatPurchaseAnalysis', {
    method: 'GET',
    params: {
      ...params,
      activityId: params.activityId || undefined,
    },
    responseType: 'blob',
  });

  return res;
}
