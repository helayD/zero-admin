import type { EChartsOption } from 'echarts';
import type {
  OperateDashboardTableRow,
  OperateFunnelActivityOption,
  OperateFunnelSeriesPoint,
  QueryRepeatPurchaseAnalysisData,
  QueryOperateFunnelDashboardData,
  RepeatPurchaseDetailTableRow,
  RepeatPurchaseTrendPoint,
} from './data.d';

export const CHANNEL_OPTIONS = [
  { label: '全部渠道', value: '' },
  { label: 'App', value: 'app' },
  { label: 'PC', value: 'pc' },
  { label: 'H5', value: 'h5' },
  { label: '小程序', value: 'mini_program' },
  { label: '未知', value: 'unknown' },
];

export const ACTIVITY_TYPE_OPTIONS = [
  { label: '全部活动', value: '' },
  { label: '首页广告', value: 'home_advertise' },
  { label: '优惠券', value: 'coupon' },
  { label: '秒杀活动', value: 'seckill_activity' },
  { label: '无活动归因', value: 'none' },
];

const metricLabelMap: Record<string, string> = {
  exposure: '曝光',
  click: '点击',
  add_cart: '加购',
  order_created: '下单',
  pay_success: '支付',
  coupon_redeem: '核销',
  paidBuyerCount: '支付买家数',
  repeatBuyerCount: '复购买家数',
  repeatRate: '复购率',
  repeatOrderCount: '复购订单数',
  repeatGmv: '复购GMV',
  avgDaysToRepeat: '平均复购天数',
};

export const formatPercent = (value?: number) => `${((value || 0) * 100).toFixed(2)}%`;

export const resolvePartialMetricLabels = (metrics: string[]) =>
  metrics.map((metric) => metricLabelMap[metric] || metric);

export const hasAnyMetricValue = (data?: QueryOperateFunnelDashboardData) => {
  if (!data) {
    return false;
  }

  const overview = data.overview;
  return [
    overview.exposure,
    overview.click,
    overview.addCart,
    overview.orderCreated,
    overview.paySuccess,
    overview.couponRedeem,
  ].some((value) => Number(value || 0) > 0);
};

export const getDashboardViewState = (data?: QueryOperateFunnelDashboardData) => {
  if (!data) {
    return 'empty';
  }
  if ((data.partialMetrics || []).length > 0) {
    return 'partial';
  }
  if (hasAnyMetricValue(data)) {
    return 'full';
  }
  return 'empty';
};

export const buildDashboardNotice = (data?: QueryOperateFunnelDashboardData) => {
  const viewState = getDashboardViewState(data);
  if (!data) {
    return {
      type: 'info' as const,
      message: '请选择筛选条件后查询经营漏斗',
      description: '支持按时间、渠道、活动和治理范围联合筛选，并自动回显实际聚合时间桶。',
    };
  }

  if (viewState === 'partial') {
    const labels = resolvePartialMetricLabels(data.partialMetrics || []).join('、');
    return {
      type: 'warning' as const,
      message: '部分指标仍处于“可信起点之后”状态',
      description: `${labels} 仅从 ${data.trackingStartedAt || '当前埋点可信起点'} 开始可信，图表与表格会保留完整时间桶，方便识别历史缺口。`,
    };
  }

  if (viewState === 'empty') {
    return {
      type: 'info' as const,
      message: '当前筛选条件下暂无经营漏斗数据',
      description: data.trackingStartedAt
        ? `当前没有命中有效指标，但系统已记录可信起点：${data.trackingStartedAt}。`
        : '可以尝试放宽时间范围、切换活动实例或调整治理范围后重新查询。',
    };
  }

  return {
    type: 'success' as const,
    message: `经营漏斗已按${data.bucket === 'hour' ? '小时' : '天'}聚合`,
    description: data.trackingStartedAt
      ? `当前结果从 ${data.trackingStartedAt} 起具备稳定埋点基础，可继续结合活动筛选观察转化走势。`
      : '当前时间范围内已返回完整总览、趋势与明细表，可直接用于活动复盘。',
  };
};

export const hasAnyRepeatPurchaseMetricValue = (data?: QueryRepeatPurchaseAnalysisData) => {
  if (!data) {
    return false;
  }

  const overview = data.overview;
  return [
    overview.paidBuyerCount,
    overview.repeatBuyerCount,
    overview.repeatRate,
    overview.repeatOrderCount,
    overview.repeatGmv,
    overview.avgDaysToRepeat,
  ].some((value) => Number(value || 0) > 0);
};

export const getRepeatPurchaseViewState = (data?: QueryRepeatPurchaseAnalysisData) => {
  if (!data) {
    return 'empty';
  }
  if ((data.partialMetrics || []).length > 0) {
    return 'partial';
  }
  if (hasAnyRepeatPurchaseMetricValue(data) || (data.details || []).length > 0) {
    return 'full';
  }
  return 'empty';
};

export const buildRepeatPurchaseNotice = (data?: QueryRepeatPurchaseAnalysisData) => {
  const viewState = getRepeatPurchaseViewState(data);
  if (!data) {
    return {
      type: 'info' as const,
      message: '请选择筛选条件后查询复购分析',
      description: '支持按治理范围、时间、渠道和活动筛选复购结果，并与导出结果保持同一口径。',
    };
  }

  if (viewState === 'partial') {
    const labels = resolvePartialMetricLabels(data.partialMetrics || []).join('、');
    return {
      type: 'warning' as const,
      message: '部分复购指标仍处于“可信起点之后”状态',
      description: `${labels} 仅从 ${data.trackingStartedAt || '当前埋点可信起点'} 开始可信，页面与导出都会保留该提示。`,
    };
  }

  if (viewState === 'empty') {
    return {
      type: 'info' as const,
      message: '当前筛选条件下暂无复购分析数据',
      description: data.trackingStartedAt
        ? `当前没有命中有效复购指标，但系统已记录可信起点：${data.trackingStartedAt}。`
        : '可以尝试放宽时间范围、切换活动实例或调整治理范围后重新查询。',
    };
  }

  return {
    type: 'success' as const,
    message: '复购分析已按天聚合',
    description: data.trackingStartedAt
      ? `当前结果从 ${data.trackingStartedAt} 起具备稳定归因基础，可直接用于留存复盘与导出。`
      : '当前时间范围内已返回复购总览、趋势与详情，可直接导出分析结果。',
  };
};

export const buildActivityOptions = (
  options: OperateFunnelActivityOption[],
  activityType?: string,
) => {
  const normalized = activityType || '';
  return (options || [])
    .filter((item) => !normalized || item.activityType === normalized)
    .map((item) => ({
      label: item.activityLabel,
      value: item.activityId,
    }));
};

export const buildOperateTrendOption = (
  series: OperateFunnelSeriesPoint[],
  bucketLabel: string,
): EChartsOption => ({
  tooltip: {
    trigger: 'axis',
  },
  legend: {
    top: 0,
    data: ['曝光', '点击', '加购', '下单', '支付', '核销'],
  },
  grid: {
    left: 36,
    right: 24,
    top: 56,
    bottom: 28,
  },
  xAxis: {
    type: 'category',
    boundaryGap: false,
    data: (series || []).map((item) => item.bucketLabel),
  },
  yAxis: {
    type: 'value',
    name: bucketLabel === 'hour' ? '小时指标值' : '日指标值',
  },
  series: [
    { name: '曝光', type: 'line', smooth: true, data: (series || []).map((item) => item.exposure) },
    { name: '点击', type: 'line', smooth: true, data: (series || []).map((item) => item.click) },
    { name: '加购', type: 'line', smooth: true, data: (series || []).map((item) => item.addCart) },
    { name: '下单', type: 'line', smooth: true, data: (series || []).map((item) => item.orderCreated) },
    { name: '支付', type: 'line', smooth: true, data: (series || []).map((item) => item.paySuccess) },
    { name: '核销', type: 'line', smooth: true, data: (series || []).map((item) => item.couponRedeem) },
  ],
});

export const buildOperateTableRows = (series: OperateFunnelSeriesPoint[]): OperateDashboardTableRow[] =>
  (series || []).map((item) => ({
    ...item,
    key: item.bucketStart,
  }));

export const buildRepeatPurchaseTrendOption = (
  series: RepeatPurchaseTrendPoint[],
): EChartsOption => ({
  tooltip: {
    trigger: 'axis',
  },
  legend: {
    top: 0,
    data: ['支付买家数', '复购买家数', '复购订单数', '复购GMV', '复购率'],
  },
  grid: {
    left: 36,
    right: 60,
    top: 56,
    bottom: 28,
  },
  xAxis: {
    type: 'category',
    boundaryGap: false,
    data: (series || []).map((item) => item.bucketLabel),
  },
  yAxis: [
    {
      type: 'value',
      name: '人数 / 单量',
    },
    {
      type: 'value',
      name: 'GMV / 复购率',
      axisLabel: {
        formatter: (val: number) => val > 1 ? `${val}` : `${(val * 100).toFixed(0)}%`,
      },
    },
  ],
  series: [
    { name: '支付买家数', type: 'line', smooth: true, data: (series || []).map((item) => item.paidBuyerCount) },
    { name: '复购买家数', type: 'line', smooth: true, data: (series || []).map((item) => item.repeatBuyerCount) },
    { name: '复购订单数', type: 'line', smooth: true, data: (series || []).map((item) => item.repeatOrderCount) },
    { name: '复购GMV', type: 'line', smooth: true, yAxisIndex: 1, data: (series || []).map((item) => item.repeatGmv) },
    { name: '复购率', type: 'line', smooth: true, yAxisIndex: 1, lineStyle: { type: 'dashed' }, data: (series || []).map((item) => item.repeatRate) },
  ],
});

export const buildRepeatPurchaseDetailRows = (
  details: QueryRepeatPurchaseAnalysisData['details'],
): RepeatPurchaseDetailTableRow[] =>
  (details || []).map((item, index) => ({
    ...item,
    key: `${item.memberId}-${item.latestRepeatPayTime}-${index}`,
  }));
