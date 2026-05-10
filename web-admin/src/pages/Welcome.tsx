import {
  Alert,
  Button,
  Card,
  Col,
  DatePicker,
  Empty,
  Form,
  Radio,
  Row,
  Select,
  Skeleton,
  Space,
  Statistic,
  Tooltip,
  Typography,
} from 'antd';
import { InfoCircleOutlined, ReloadOutlined } from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-layout';
import ReactEcharts from 'echarts-for-react';
import type { EChartsOption } from 'echarts';
import moment from 'moment';
import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useModel } from 'umi';

import GovernanceScopeBar from '@/pages/system/components/GovernanceScopeBar';
import {
  defaultGovernanceScope,
  toGovernancePayload,
  type GovernanceScopeValue,
} from '@/pages/system/components/governance';
import type {
  OperateFunnelMetricCard,
  QueryOperateFunnelDashboardData,
} from '@/pages/sms/OperateDashboard/data.d';
import { queryOperateFunnelDashboard } from '@/pages/sms/OperateDashboard/service';
import {
  ACTIVITY_TYPE_OPTIONS,
  CHANNEL_OPTIONS,
  buildActivityOptions,
  buildDashboardNotice,
  buildOperateTrendOption,
  formatPercent,
  getDashboardViewState,
} from '@/pages/sms/OperateDashboard/helper';

const { RangePicker } = DatePicker;
const { Title, Text } = Typography;

type WelcomeFormValues = {
  timeRange: [moment.Moment, moment.Moment];
  channel?: string;
  activityType?: string;
  activityId?: number;
  bucket: 'day' | 'hour';
};

const defaultTimeRange: [moment.Moment, moment.Moment] = [
  moment().subtract(6, 'days').startOf('day'),
  moment().endOf('day'),
];

// 六阶段漏斗配色：遵循 ui-ux-pro-max chart skill 推荐 —— 主色单色渐变 + 支付阶段强调色
// 主色来自项目 antd token (#1677ff)，透明度递减展示漏斗衰减；支付/核销用绿色/紫色强调转化终点
const funnelStageColorMap: Record<string, string> = {
  exposure: 'rgba(22, 119, 255, 0.95)',
  click: 'rgba(22, 119, 255, 0.82)',
  add_cart: 'rgba(22, 119, 255, 0.68)',
  order_created: 'rgba(22, 119, 255, 0.55)',
  pay_success: '#52c41a',
  coupon_redeem: '#722ed1',
};

// 指标卡文字主色（KPI 数值），与漏斗渐变解耦，保持可读性
const metricCardColorMap: Record<string, string> = {
  exposure: '#1677ff',
  click: '#1677ff',
  add_cart: '#d48806',
  order_created: '#d46b08',
  pay_success: '#52c41a',
  coupon_redeem: '#722ed1',
};

const greetingByHour = (hour: number) => {
  if (hour < 6) return '凌晨好';
  if (hour < 11) return '早上好';
  if (hour < 13) return '中午好';
  if (hour < 18) return '下午好';
  return '晚上好';
};

const formatNumber = (value?: number) => new Intl.NumberFormat('zh-CN').format(Number(value || 0));

// 检测用户是否偏好减少动画（ui-ux-pro-max ux skill HIGH: prefers-reduced-motion）
// echarts 动画在满足该偏好时完全关闭，避免诱发晕动
const usePrefersReducedMotion = (): boolean => {
  const getInitial = () => {
    if (typeof window === 'undefined' || !window.matchMedia) return false;
    return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  };
  const [prefers, setPrefers] = useState<boolean>(getInitial);
  useEffect(() => {
    if (typeof window === 'undefined' || !window.matchMedia) return undefined;
    const mql = window.matchMedia('(prefers-reduced-motion: reduce)');
    const handler = (event: MediaQueryListEvent) => setPrefers(event.matches);
    // Safari <14 用 addListener，其余用 addEventListener
    if (mql.addEventListener) {
      mql.addEventListener('change', handler);
      return () => mql.removeEventListener('change', handler);
    }
    mql.addListener(handler);
    return () => mql.removeListener(handler);
  }, []);
  return prefers;
};

const buildFunnelOption = (
  cards: OperateFunnelMetricCard[],
  animationDuration: number,
): EChartsOption => ({
  tooltip: {
    trigger: 'item',
    formatter: (params: any) => {
      const card = cards[params.dataIndex];
      if (!card) return '';
      const rate = card.rate ? `（${card.rateLabel}：${formatPercent(card.rate)}）` : '';
      return `${card.label}：${formatNumber(card.value)}${rate}`;
    },
  },
  series: [
    {
      name: '经营漏斗',
      type: 'funnel',
      left: '10%',
      right: '10%',
      top: 10,
      bottom: 10,
      minSize: '16%',
      sort: 'descending',
      gap: 2,
      label: {
        show: true,
        position: 'inside',
        color: '#fff',
        // chart skill 推荐：每阶段显式展示转化率文本，提升信息密度和专业度
        formatter: (params: any) => {
          const card = cards[params.dataIndex];
          if (!card) return '';
          const rateLine = card.rate
            ? `\n${card.rateLabel} ${formatPercent(card.rate)}`
            : '';
          return `${card.label} ${formatNumber(card.value)}${rateLine}`;
        },
      },
      data: (cards || []).map((card) => ({
        name: card.label,
        value: Math.max(card.value || 0, 0),
        itemStyle: {
          color: funnelStageColorMap[card.key] || '#1677ff',
        },
      })),
    },
  ],
  animation: animationDuration > 0,
  animationDuration,
});

const buildStageBarOption = (
  cards: OperateFunnelMetricCard[],
  animationDuration: number,
): EChartsOption => ({
  tooltip: {
    trigger: 'axis',
    axisPointer: { type: 'shadow' },
    valueFormatter: (value: any) => formatNumber(Number(value)),
  },
  grid: { left: 48, right: 24, top: 16, bottom: 32 },
  xAxis: {
    type: 'category',
    data: (cards || []).map((card) => card.label),
  },
  yAxis: {
    type: 'value',
    name: '规模',
  },
  series: [
    {
      name: '规模',
      type: 'bar',
      barMaxWidth: 48,
      data: (cards || []).map((card) => ({
        value: Math.max(card.value || 0, 0),
        itemStyle: { color: funnelStageColorMap[card.key] || '#1677ff' },
      })),
      label: {
        show: true,
        position: 'top',
        formatter: (params: any) => formatNumber(params.value),
      },
    },
  ],
  animation: animationDuration > 0,
  animationDuration,
});

const Welcome: React.FC = () => {
  const { initialState } = useModel('@@initialState');
  const displayName = initialState?.currentUser?.data?.name || 'Operator';

  const [form] = Form.useForm<WelcomeFormValues>();
  const [scope, setScope] = useState<GovernanceScopeValue>(defaultGovernanceScope);
  const [loading, setLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string>('');
  const [dashboardData, setDashboardData] = useState<QueryOperateFunnelDashboardData>();
  const [selectedActivityType, setSelectedActivityType] = useState<string>('');
  // a11y: 尊重 prefers-reduced-motion，关闭 echarts 入场动画
  const prefersReducedMotion = usePrefersReducedMotion();
  const chartAnimationDuration = prefersReducedMotion ? 0 : 300;

  const loadDashboard = useCallback(
    async (values: WelcomeFormValues | undefined, currentScope: GovernanceScopeValue) => {
      const currentValues = values ?? form.getFieldsValue();
      const rangeValue = currentValues?.timeRange ?? defaultTimeRange;
      const bucket = currentValues?.bucket ?? 'day';

      setLoading(true);
      setErrorMessage('');
      try {
        const response = await queryOperateFunnelDashboard({
          ...toGovernancePayload(currentScope),
          startTime: rangeValue?.[0]?.format('YYYY-MM-DD HH:mm:ss'),
          endTime: rangeValue?.[1]?.format('YYYY-MM-DD HH:mm:ss'),
          channel: currentValues?.channel || undefined,
          activityType: currentValues?.activityType || undefined,
          activityId: currentValues?.activityId || undefined,
          bucket,
        });
        setDashboardData(response?.data);
      } catch (error: any) {
        const message =
          error?.data?.message || error?.message || '查询经营概览失败，请稍后重试';
        setErrorMessage(message);
      } finally {
        setLoading(false);
      }
    },
    [form],
  );

  useEffect(() => {
    form.setFieldsValue({ timeRange: defaultTimeRange, bucket: 'day' });
    loadDashboard(form.getFieldsValue(), scope);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const activityOptions = useMemo(
    () => buildActivityOptions(dashboardData?.activityOptions || [], selectedActivityType),
    [dashboardData?.activityOptions, selectedActivityType],
  );

  const notice = useMemo(() => buildDashboardNotice(dashboardData), [dashboardData]);
  const viewState = useMemo(() => getDashboardViewState(dashboardData), [dashboardData]);

  const overviewCards: OperateFunnelMetricCard[] = useMemo(
    () => dashboardData?.overview?.cards || [],
    [dashboardData?.overview?.cards],
  );

  const funnelOption = useMemo(
    () => buildFunnelOption(overviewCards, chartAnimationDuration),
    [overviewCards, chartAnimationDuration],
  );
  const stageBarOption = useMemo(
    () => buildStageBarOption(overviewCards, chartAnimationDuration),
    [overviewCards, chartAnimationDuration],
  );
  const trendOption = useMemo(
    () => buildOperateTrendOption(dashboardData?.series || [], dashboardData?.bucket || 'day'),
    [dashboardData?.series, dashboardData?.bucket],
  );

  const handleScopeChange = (next: GovernanceScopeValue) => {
    setScope(next);
    loadDashboard(form.getFieldsValue(), next);
  };

  const handleActivityTypeChange = (nextType?: string) => {
    setSelectedActivityType(nextType || '');
    form.setFieldsValue({ activityType: nextType, activityId: undefined });
  };

  const handleReset = () => {
    form.resetFields();
    form.setFieldsValue({ timeRange: defaultTimeRange, bucket: 'day' });
    setSelectedActivityType('');
    loadDashboard(form.getFieldsValue(), scope);
  };

  const renderMetricCard = (card: OperateFunnelMetricCard) => {
    // 指标卡数值用 metricCardColorMap（高对比度），与漏斗的主色渐变解耦保证可读性
    const color = metricCardColorMap[card.key] || '#1677ff';
    const hasRate = Boolean(card.rateLabel);
    return (
      <Col key={card.key} xs={24} sm={12} md={8} xl={4}>
        <Card bordered hoverable bodyStyle={{ padding: 20 }}>
          <Statistic
            title={
              <Space size={4}>
                <Text style={{ fontSize: 14, color: 'rgba(0, 0, 0, 0.65)' }}>{card.label}</Text>
                <Tooltip
                  title={
                    hasRate
                      ? `${card.rateLabel}：${formatPercent(card.rate)}（上一阶段的转化比例）`
                      : `${card.label}：来自经营埋点的真实聚合数据`
                  }
                >
                  <InfoCircleOutlined
                    style={{ color: 'rgba(0, 0, 0, 0.45)' }}
                    aria-label={`${card.label}指标说明`}
                  />
                </Tooltip>
              </Space>
            }
            value={card.value}
            valueStyle={{ color, fontSize: 28, fontWeight: 600 }}
            formatter={(value) => formatNumber(Number(value))}
          />
          <div style={{ marginTop: 8, minHeight: 22 }}>
            {hasRate ? (
              <Text type="secondary" style={{ fontSize: 12 }}>
                {card.rateLabel}：<Text strong>{formatPercent(card.rate)}</Text>
              </Text>
            ) : (
              <Text type="secondary" style={{ fontSize: 12 }}>
                漏斗首阶段（基线）
              </Text>
            )}
          </div>
        </Card>
      </Col>
    );
  };

  const cardPlaceholders = Array.from({ length: 6 }, (_, index) => (
    <Col key={`ph-${index}`} xs={24} sm={12} md={8} xl={4}>
      <Card bordered>
        <Skeleton active paragraph={{ rows: 2 }} />
      </Card>
    </Col>
  ));

  const isCharEmpty = viewState === 'empty' && !loading;

  return (
    <PageContainer
      header={{
        title: `${greetingByHour(moment().hour())}，${displayName}`,
        subTitle: '九克城 · 经营概览（实时经营漏斗）',
        breadcrumb: {},
      }}
    >
      <Card bordered={false} bodyStyle={{ padding: '16px 24px' }} style={{ marginBottom: 16 }}>
        <Title level={5} style={{ marginBottom: 4 }}>
          欢迎回到九克城管理中心
        </Title>
        <Text type="secondary">
          下面的指标来自平台真实埋点（曝光 / 点击 / 加购 / 下单 / 支付 / 核销），按治理作用域、时间范围、渠道与活动联合筛选。
          {dashboardData?.trackingStartedAt && (
            <>
              {' '}可信数据起点：
              <Text code>{dashboardData.trackingStartedAt}</Text>
            </>
          )}
        </Text>
      </Card>

      <GovernanceScopeBar
        value={scope}
        onChange={handleScopeChange}
        entityLabel="经营数据"
        style={{ marginBottom: 16 }}
      />

      <Card bordered={false} style={{ marginBottom: 16 }}>
        <Form<WelcomeFormValues>
          form={form}
          layout="vertical"
          onFinish={(values) => loadDashboard(values, scope)}
          initialValues={{ timeRange: defaultTimeRange, bucket: 'day' }}
        >
          <Row gutter={[16, 12]}>
            <Col xs={24} md={10} lg={8}>
              <Form.Item label="时间范围" name="timeRange" style={{ marginBottom: 0 }}>
                <RangePicker
                  showTime
                  style={{ width: '100%' }}
                  allowClear={false}
                  ranges={{
                    今日: [moment().startOf('day'), moment().endOf('day')],
                    近7天: [moment().subtract(6, 'days').startOf('day'), moment().endOf('day')],
                    近30天: [moment().subtract(29, 'days').startOf('day'), moment().endOf('day')],
                  }}
                />
              </Form.Item>
            </Col>
            <Col xs={12} md={4} lg={4}>
              <Form.Item label="渠道" name="channel" style={{ marginBottom: 0 }}>
                <Select options={CHANNEL_OPTIONS} allowClear placeholder="全部渠道" />
              </Form.Item>
            </Col>
            <Col xs={12} md={4} lg={4}>
              <Form.Item label="活动类型" name="activityType" style={{ marginBottom: 0 }}>
                <Select
                  options={ACTIVITY_TYPE_OPTIONS}
                  allowClear
                  placeholder="全部活动"
                  onChange={handleActivityTypeChange}
                />
              </Form.Item>
            </Col>
            <Col xs={12} md={3} lg={4}>
              <Form.Item label="具体活动" name="activityId" style={{ marginBottom: 0 }}>
                <Select
                  options={activityOptions}
                  allowClear
                  disabled={!selectedActivityType || activityOptions.length === 0}
                  placeholder={selectedActivityType ? '全部活动实例' : '先选活动类型'}
                  showSearch
                  optionFilterProp="label"
                />
              </Form.Item>
            </Col>
            <Col xs={12} md={3} lg={4}>
              <Form.Item label="聚合粒度" name="bucket" style={{ marginBottom: 0 }}>
                <Radio.Group buttonStyle="solid">
                  <Radio.Button value="day">按日</Radio.Button>
                  <Radio.Button value="hour">按小时</Radio.Button>
                </Radio.Group>
              </Form.Item>
            </Col>
          </Row>
          <Row style={{ marginTop: 16 }}>
            <Col span={24}>
              <Space>
                <Button type="primary" htmlType="submit" loading={loading}>
                  查询
                </Button>
                <Button onClick={handleReset} icon={<ReloadOutlined />}>
                  重置
                </Button>
              </Space>
            </Col>
          </Row>
        </Form>
      </Card>

      {errorMessage && (
        // ux skill HIGH: Error Recovery — 错误必须给出恢复路径（此处提供"重试"）
        // Alert 内建 role="alert"，屏幕阅读器可识别
        <Alert
          type="error"
          showIcon
          closable
          message="查询经营概览失败"
          description={errorMessage}
          style={{ marginBottom: 16 }}
          onClose={() => setErrorMessage('')}
          action={
            <Button
              size="small"
              type="primary"
              icon={<ReloadOutlined />}
              loading={loading}
              onClick={() => loadDashboard(form.getFieldsValue(), scope)}
            >
              重试
            </Button>
          }
        />
      )}

      <Alert
        type={notice.type}
        showIcon
        message={notice.message}
        description={notice.description}
        style={{ marginBottom: 16 }}
      />

      <Row gutter={[16, 16]}>
        {loading && !dashboardData ? cardPlaceholders : overviewCards.map(renderMetricCard)}
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} lg={12}>
          <Card title="经营漏斗（按筛选条件聚合）" bordered hoverable>
            {loading ? (
              <Skeleton active paragraph={{ rows: 6 }} />
            ) : isCharEmpty ? (
              <Empty
                description="当前筛选条件下暂无漏斗数据"
                style={{ margin: '40px 0' }}
              >
                <Button type="primary" onClick={handleReset}>
                  恢复默认筛选并重试
                </Button>
              </Empty>
            ) : (
              <>
                <ReactEcharts
                  option={funnelOption}
                  style={{ height: 320 }}
                  notMerge
                  lazyUpdate
                  opts={{ renderer: 'canvas' }}
                />
                {/* chart skill a11y fallback：屏幕阅读器可读的线性漏斗摘要 */}
                <ul
                  role="list"
                  aria-label="经营漏斗阶段摘要"
                  style={{
                    marginTop: 12,
                    paddingLeft: 0,
                    listStyle: 'none',
                    display: 'grid',
                    gridTemplateColumns: 'repeat(auto-fit, minmax(120px, 1fr))',
                    gap: 8,
                    fontSize: 12,
                  }}
                >
                  {overviewCards.map((card) => (
                    <li
                      key={card.key}
                      style={{
                        padding: '6px 10px',
                        background: 'rgba(0, 0, 0, 0.02)',
                        borderLeft: `3px solid ${funnelStageColorMap[card.key] || '#1677ff'}`,
                      }}
                    >
                      <Text strong>{card.label}</Text>
                      <div>
                        <Text type="secondary">值 </Text>
                        <Text>{formatNumber(card.value)}</Text>
                      </div>
                      {card.rateLabel && (
                        <div>
                          <Text type="secondary">{card.rateLabel} </Text>
                          <Text>{formatPercent(card.rate)}</Text>
                        </div>
                      )}
                    </li>
                  ))}
                </ul>
              </>
            )}
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card title="各阶段规模对比" bordered hoverable>
            {loading ? (
              <Skeleton active paragraph={{ rows: 6 }} />
            ) : isCharEmpty ? (
              <Empty
                description="当前筛选条件下暂无规模数据"
                style={{ margin: '40px 0' }}
              >
                <Button type="primary" onClick={handleReset}>
                  恢复默认筛选并重试
                </Button>
              </Empty>
            ) : (
              <ReactEcharts
                option={stageBarOption}
                style={{ height: 320 }}
                notMerge
                lazyUpdate
                opts={{ renderer: 'canvas' }}
              />
            )}
          </Card>
        </Col>
      </Row>

      <Row style={{ marginTop: 16 }}>
        <Col span={24}>
          <Card
            title={`经营漏斗趋势（${dashboardData?.bucket === 'hour' ? '按小时' : '按天'}）`}
            bordered
            hoverable
          >
            {loading ? (
              <Skeleton active paragraph={{ rows: 8 }} />
            ) : isCharEmpty || (dashboardData?.series || []).length === 0 ? (
              <Empty
                description="当前筛选条件下暂无趋势数据"
                style={{ margin: '40px 0' }}
              >
                <Button type="primary" onClick={handleReset}>
                  恢复默认筛选并重试
                </Button>
              </Empty>
            ) : (
              <ReactEcharts
                option={trendOption}
                style={{ height: 360 }}
                notMerge
                lazyUpdate
                opts={{ renderer: 'canvas' }}
              />
            )}
          </Card>
        </Col>
      </Row>
    </PageContainer>
  );
};

export default Welcome;
