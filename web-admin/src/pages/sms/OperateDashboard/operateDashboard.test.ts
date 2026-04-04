import {
  buildDashboardNotice,
  buildOperateTableRows,
  buildOperateTrendOption,
  buildRepeatPurchaseDetailRows,
  buildRepeatPurchaseNotice,
  buildRepeatPurchaseTrendOption,
  formatPercent,
  getDashboardViewState,
  getRepeatPurchaseViewState,
} from './helper';

const buildDashboardData = (overrides: Record<string, any> = {}) => ({
  overview: {
    exposure: 120,
    click: 30,
    addCart: 12,
    orderCreated: 6,
    paySuccess: 3,
    couponRedeem: 2,
    clickRate: 0.25,
    addCartRate: 0.4,
    orderRate: 0.5,
    payRate: 0.5,
    couponRedeemRate: 2 / 3,
    cards: [],
  },
  series: [
    {
      bucketLabel: '2026-04-03',
      bucketStart: '2026-04-03 00:00:00',
      bucketEnd: '2026-04-04 00:00:00',
      exposure: 120,
      click: 30,
      addCart: 12,
      orderCreated: 6,
      paySuccess: 3,
      couponRedeem: 2,
      clickRate: 0.25,
      addCartRate: 0.4,
      orderRate: 0.5,
      payRate: 0.5,
      couponRedeemRate: 2 / 3,
    },
  ],
  activityOptions: [],
  trackingStartedAt: '',
  partialMetrics: [],
  bucket: 'day',
  ...overrides,
});

const buildRepeatPurchaseData = (overrides: Record<string, any> = {}) => ({
  overview: {
    paidBuyerCount: 20,
    repeatBuyerCount: 5,
    repeatRate: 0.25,
    repeatOrderCount: 8,
    repeatGmv: 256.8,
    avgDaysToRepeat: 7.2,
  },
  trends: [
    {
      bucketLabel: '2026-04-03',
      bucketStart: '2026-04-03 00:00:00',
      bucketEnd: '2026-04-04 00:00:00',
      paidBuyerCount: 6,
      repeatBuyerCount: 2,
      repeatRate: 1 / 3,
      repeatOrderCount: 3,
      repeatGmv: 88.6,
      avgDaysToRepeat: 4.5,
    },
  ],
  details: [
    {
      memberId: 3001,
      nicknameMasked: '会***员',
      mobileMasked: '138****0001',
      firstValidPayTime: '2026-03-20 10:00:00',
      latestRepeatPayTime: '2026-04-03 09:00:00',
      repeatOrderCount: 2,
      repeatGmv: 88.6,
      latestChannel: 'app',
      latestActivityType: 'coupon',
      latestActivityId: 11,
      platformId: 1,
      tenantId: 10,
      merchantId: 88,
    },
  ],
  total: 1,
  pageNum: 1,
  pageSize: 20,
  activityOptions: [],
  trackingStartedAt: '',
  partialMetrics: [],
  bucket: 'day',
  ...overrides,
});

describe('operateDashboard helper', () => {
  it('formats rate and chart series consistently', () => {
    expect(formatPercent(0.2567)).toBe('25.67%');

    const option = buildOperateTrendOption(buildDashboardData().series, 'day');
    const series = Array.isArray(option.series) ? option.series : [];
    expect(series).toHaveLength(6);
    expect(option.xAxis).toMatchObject({
      data: ['2026-04-03'],
    });
  });

  it('keeps table rows stable for downstream rendering', () => {
    const rows = buildOperateTableRows(buildDashboardData().series);
    expect(rows).toEqual([
      expect.objectContaining({
        key: '2026-04-03 00:00:00',
        bucketLabel: '2026-04-03',
        couponRedeem: 2,
      }),
    ]);
  });

  it('distinguishes empty, partial and full dashboard states', () => {
    const emptyData = buildDashboardData({
      overview: {
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
      },
      series: [],
    });
    expect(getDashboardViewState(emptyData)).toBe('empty');
    expect(buildDashboardNotice(emptyData)?.message).toContain('暂无经营漏斗数据');

    const partialData = buildDashboardData({
      partialMetrics: ['exposure', 'click'],
      trackingStartedAt: '2026-04-03 12:00:00',
    });
    expect(getDashboardViewState(partialData)).toBe('partial');
    expect(buildDashboardNotice(partialData)?.description).toContain('2026-04-03 12:00:00');

    const fullData = buildDashboardData();
    expect(getDashboardViewState(fullData)).toBe('full');
    expect(buildDashboardNotice(fullData)?.message).toContain('经营漏斗已按');
  });

  it('builds repeat purchase chart and table rows consistently', () => {
    const repeatData = buildRepeatPurchaseData();

    const option = buildRepeatPurchaseTrendOption(repeatData.trends);
    const series = Array.isArray(option.series) ? option.series : [];
    expect(series).toHaveLength(4);
    expect(option.xAxis).toMatchObject({
      data: ['2026-04-03'],
    });

    const rows = buildRepeatPurchaseDetailRows(repeatData.details);
    expect(rows).toEqual([
      expect.objectContaining({
        key: '3001-2026-04-03 09:00:00-0',
        memberId: 3001,
        latestActivityId: 11,
      }),
    ]);
  });

  it('distinguishes empty, partial and full repeat purchase states', () => {
    const emptyData = buildRepeatPurchaseData({
      overview: {
        paidBuyerCount: 0,
        repeatBuyerCount: 0,
        repeatRate: 0,
        repeatOrderCount: 0,
        repeatGmv: 0,
        avgDaysToRepeat: 0,
      },
      trends: [],
      details: [],
    });
    expect(getRepeatPurchaseViewState(emptyData)).toBe('empty');
    expect(buildRepeatPurchaseNotice(emptyData)?.message).toContain('暂无复购分析数据');

    const partialData = buildRepeatPurchaseData({
      partialMetrics: ['repeatBuyerCount', 'repeatRate'],
      trackingStartedAt: '2026-04-03 12:00:00',
    });
    expect(getRepeatPurchaseViewState(partialData)).toBe('partial');
    expect(buildRepeatPurchaseNotice(partialData)?.description).toContain('2026-04-03 12:00:00');

    const fullData = buildRepeatPurchaseData();
    expect(getRepeatPurchaseViewState(fullData)).toBe('full');
    expect(buildRepeatPurchaseNotice(fullData)?.message).toContain('复购分析已按天聚合');
  });
});
