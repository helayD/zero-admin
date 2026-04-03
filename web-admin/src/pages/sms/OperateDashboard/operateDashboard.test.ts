import {
  buildDashboardNotice,
  buildOperateTableRows,
  buildOperateTrendOption,
  formatPercent,
  getDashboardViewState,
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
});
