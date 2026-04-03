import type { EChartsOption } from 'echarts';
import type {
  OperateDashboardTableRow,
  OperateFunnelActivityOption,
  OperateFunnelSeriesPoint,
  QueryOperateFunnelDashboardData,
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
