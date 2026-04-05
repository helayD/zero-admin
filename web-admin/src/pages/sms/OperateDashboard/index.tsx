import {
  Alert,
  Button,
  Card,
  Col,
  DatePicker,
  Empty,
  Form,
  Row,
  Select,
  Space,
  Spin,
  Statistic,
  Table,
  Tabs,
  Tag,
  Typography,
  message,
} from 'antd';
import { DownloadOutlined, ReloadOutlined, SearchOutlined } from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-layout';
import type { ColumnsType } from 'antd/lib/table';
import React, { useCallback, useEffect, useState } from 'react';
import ReactEcharts from 'echarts-for-react';
import moment from 'moment';
import GovernanceScopeBar from '@/pages/system/components/GovernanceScopeBar';
import {
  buildGovernanceScopeLabel,
  defaultGovernanceScope,
  toGovernancePayload,
  type GovernanceScopeValue,
} from '@/pages/system/components/governance';
import type {
  OperateDashboardTableRow,
  QueryOperateFunnelDashboardData,
  QueryRepeatPurchaseAnalysisData,
  RepeatPurchaseDetailTableRow,
} from './data.d';
import {
  ACTIVITY_TYPE_OPTIONS,
  buildActivityOptions,
  buildDashboardNotice,
  buildOperateTableRows,
  buildOperateTrendOption,
  buildRepeatPurchaseDetailRows,
  buildRepeatPurchaseNotice,
  buildRepeatPurchaseTrendOption,
  CHANNEL_OPTIONS,
  formatPercent,
  getDashboardViewState,
  getRepeatPurchaseViewState,
} from './helper';
import {
  exportRepeatPurchaseAnalysis,
  queryOperateFunnelDashboard,
  queryRepeatPurchaseAnalysis,
} from './service';

const { RangePicker } = DatePicker;
const { Text } = Typography;
const { TabPane } = Tabs;

type DashboardViewKey = 'funnel' | 'repeatPurchase';

type OperateDashboardFormValues = {
  timeRange: [moment.Moment, moment.Moment];
  channel?: string;
  activityType?: string;
  activityId?: number;
  bucket?: 'day' | 'hour';
};

const defaultTimeRange: [moment.Moment, moment.Moment] = [
  moment().startOf('day').subtract(6, 'day'),
  moment().endOf('day'),
];

const OperateDashboard: React.FC = () => {
  const [form] = Form.useForm<OperateDashboardFormValues>();
  const [scope, setScope] = useState<GovernanceScopeValue>(defaultGovernanceScope);
  const [loading, setLoading] = useState(false);
  const [activeView, setActiveView] = useState<DashboardViewKey>('funnel');
  const [dashboardData, setDashboardData] = useState<QueryOperateFunnelDashboardData>();
  const [repeatPurchaseData, setRepeatPurchaseData] = useState<QueryRepeatPurchaseAnalysisData>();
  const [selectedActivityType, setSelectedActivityType] = useState<string>('');

  const loadFunnelDashboard = useCallback(async (
    values: OperateDashboardFormValues | undefined,
    currentScope: GovernanceScopeValue,
  ) => {
    const currentValues =
      values ||
      form.getFieldsValue([
        'timeRange',
        'channel',
        'activityType',
        'activityId',
        'bucket',
      ]);
    const rangeValue = currentValues.timeRange || defaultTimeRange;

    setLoading(true);
    try {
      const response = await queryOperateFunnelDashboard({
        ...toGovernancePayload(currentScope),
        startTime: rangeValue?.[0]?.format('YYYY-MM-DD HH:mm:ss'),
        endTime: rangeValue?.[1]?.format('YYYY-MM-DD HH:mm:ss'),
        channel: currentValues.channel || undefined,
        activityType: currentValues.activityType || undefined,
        activityId: currentValues.activityId || undefined,
        bucket: currentValues.bucket || 'day',
      });
      setDashboardData(response.data);
    } catch (e) {
      message.error('查询经营漏斗失败，请稍后重试');
    } finally {
      setLoading(false);
    }
  }, [form]);

  const loadRepeatPurchaseDashboard = useCallback(async (
    values: OperateDashboardFormValues | undefined,
    currentScope: GovernanceScopeValue,
    pageNum = 1,
    pageSize = 20,
  ) => {
    const currentValues =
      values ||
      form.getFieldsValue([
        'timeRange',
        'channel',
        'activityType',
        'activityId',
      ]);
    const rangeValue = currentValues.timeRange || defaultTimeRange;

    setLoading(true);
    try {
      const response = await queryRepeatPurchaseAnalysis({
        ...toGovernancePayload(currentScope),
        startTime: rangeValue?.[0]?.format('YYYY-MM-DD HH:mm:ss'),
        endTime: rangeValue?.[1]?.format('YYYY-MM-DD HH:mm:ss'),
        channel: currentValues.channel || undefined,
        activityType: currentValues.activityType || undefined,
        activityId: currentValues.activityId || undefined,
        bucket: 'day',
        pageNum,
        pageSize,
      });
      setRepeatPurchaseData(response.data);
    } catch (e) {
      message.error('查询复购分析失败，请稍后重试');
    } finally {
      setLoading(false);
    }
  }, [form]);

  useEffect(() => {
    form.setFieldsValue({
      timeRange: defaultTimeRange,
      channel: '',
      activityType: '',
      activityId: undefined,
      bucket: 'day',
    });
    void loadFunnelDashboard(
      {
        timeRange: defaultTimeRange,
        channel: '',
        activityType: '',
        activityId: undefined,
        bucket: 'day',
      },
      defaultGovernanceScope,
    );
  }, [form, loadFunnelDashboard]);

  const currentActivityOptions = buildActivityOptions(
    activeView === 'funnel'
      ? dashboardData?.activityOptions || []
      : repeatPurchaseData?.activityOptions || [],
    selectedActivityType,
  );

  const scopeLabel = buildGovernanceScopeLabel(scope);
  const dashboardNotice = buildDashboardNotice(dashboardData);
  const repeatPurchaseNotice = buildRepeatPurchaseNotice(repeatPurchaseData);
  const tableData = buildOperateTableRows(dashboardData?.series || []);
  const viewState = getDashboardViewState(dashboardData);
  const repeatDetailRows = buildRepeatPurchaseDetailRows(repeatPurchaseData?.details || []);
  const repeatViewState = getRepeatPurchaseViewState(repeatPurchaseData);

  const columns: ColumnsType<OperateDashboardTableRow> = [
    {
      title: '时间桶',
      dataIndex: 'bucketLabel',
      width: 160,
    },
    {
      title: '曝光',
      dataIndex: 'exposure',
      width: 96,
    },
    {
      title: '点击',
      dataIndex: 'click',
      width: 96,
    },
    {
      title: '加购',
      dataIndex: 'addCart',
      width: 96,
    },
    {
      title: '下单',
      dataIndex: 'orderCreated',
      width: 96,
    },
    {
      title: '支付',
      dataIndex: 'paySuccess',
      width: 96,
    },
    {
      title: '核销',
      dataIndex: 'couponRedeem',
      width: 96,
    },
    {
      title: '点击率',
      dataIndex: 'clickRate',
      width: 108,
      render: (_, record) => formatPercent(record.clickRate),
    },
    {
      title: '加购率',
      dataIndex: 'addCartRate',
      width: 108,
      render: (_, record) => formatPercent(record.addCartRate),
    },
    {
      title: '下单率',
      dataIndex: 'orderRate',
      width: 108,
      render: (_, record) => formatPercent(record.orderRate),
    },
    {
      title: '支付率',
      dataIndex: 'payRate',
      width: 108,
      render: (_, record) => formatPercent(record.payRate),
    },
    {
      title: '核销率',
      dataIndex: 'couponRedeemRate',
      width: 108,
      render: (_, record) => formatPercent(record.couponRedeemRate),
    },
  ];

  const repeatColumns: ColumnsType<RepeatPurchaseDetailTableRow> = [
    {
      title: '会员ID',
      dataIndex: 'memberId',
      width: 120,
      fixed: 'left',
    },
    {
      title: '昵称',
      dataIndex: 'nicknameMasked',
      width: 120,
    },
    {
      title: '手机号',
      dataIndex: 'mobileMasked',
      width: 140,
    },
    {
      title: '首次有效支付时间',
      dataIndex: 'firstValidPayTime',
      width: 180,
    },
    {
      title: '最近复购支付时间',
      dataIndex: 'latestRepeatPayTime',
      width: 180,
    },
    {
      title: '复购订单数',
      dataIndex: 'repeatOrderCount',
      width: 120,
    },
    {
      title: '复购GMV',
      dataIndex: 'repeatGmv',
      width: 120,
    },
    {
      title: '最近渠道',
      dataIndex: 'latestChannel',
      width: 120,
    },
    {
      title: '最近活动类型',
      dataIndex: 'latestActivityType',
      width: 140,
    },
    {
      title: '最近活动ID',
      dataIndex: 'latestActivityId',
      width: 120,
    },
    {
      title: '平台ID',
      dataIndex: 'platformId',
      width: 100,
    },
    {
      title: '租户ID',
      dataIndex: 'tenantId',
      width: 100,
    },
    {
      title: '商户ID',
      dataIndex: 'merchantId',
      width: 100,
    },
  ];

  const repeatCards = [
    { key: 'paidBuyerCount', label: '支付买家数', value: repeatPurchaseData?.overview.paidBuyerCount || 0 },
    { key: 'repeatBuyerCount', label: '复购买家数', value: repeatPurchaseData?.overview.repeatBuyerCount || 0 },
    { key: 'repeatRate', label: '复购率', value: formatPercent(repeatPurchaseData?.overview.repeatRate) },
    { key: 'repeatOrderCount', label: '复购订单数', value: repeatPurchaseData?.overview.repeatOrderCount || 0 },
    { key: 'repeatGmv', label: '复购GMV', value: repeatPurchaseData?.overview.repeatGmv || 0 },
    {
      key: 'avgDaysToRepeat',
      label: '平均复购天数',
      value: Number(repeatPurchaseData?.overview.avgDaysToRepeat || 0).toFixed(2),
    },
  ];

  const handleQuery = useCallback(async (
    values: OperateDashboardFormValues | undefined,
    currentScope: GovernanceScopeValue,
  ) => {
    if (activeView === 'repeatPurchase') {
      await loadRepeatPurchaseDashboard(values, currentScope, 1, repeatPurchaseData?.pageSize || 20);
      return;
    }
    await loadFunnelDashboard(values, currentScope);
  }, [activeView, loadFunnelDashboard, loadRepeatPurchaseDashboard, repeatPurchaseData?.pageSize]);

  const handleExport = useCallback(async () => {
    const currentValues = form.getFieldsValue([
      'timeRange',
      'channel',
      'activityType',
      'activityId',
    ]);
    const rangeValue = currentValues.timeRange || defaultTimeRange;
    const hide = message.loading('正在导出...');
    try {
      const response = await exportRepeatPurchaseAnalysis({
        ...toGovernancePayload(scope),
        startTime: rangeValue?.[0]?.format('YYYY-MM-DD HH:mm:ss'),
        endTime: rangeValue?.[1]?.format('YYYY-MM-DD HH:mm:ss'),
        channel: currentValues.channel || undefined,
        activityType: currentValues.activityType || undefined,
        activityId: currentValues.activityId || undefined,
        bucket: 'day',
      });
      const blob = new Blob([response], {
        type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
      });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `复购分析_${new Date().toISOString().slice(0, 10)}.xlsx`;
      link.click();
      window.URL.revokeObjectURL(url);
      hide();
      message.success('导出成功');
    } catch {
      hide();
      message.error('导出失败');
    }
  }, [form, scope]);

  return (
    <PageContainer>
      <Space direction="vertical" size={16} style={{ width: '100%' }}>
        <Tabs
          activeKey={activeView}
          onChange={(key) => {
            const nextView = key as DashboardViewKey;
            setActiveView(nextView);
            if (nextView === 'repeatPurchase') {
              form.setFieldsValue({ bucket: 'day' });
              void loadRepeatPurchaseDashboard(undefined, scope, 1, repeatPurchaseData?.pageSize || 20);
              return;
            }
            void loadFunnelDashboard(undefined, scope);
          }}
        >
          <TabPane tab="经营漏斗" key="funnel" />
          <TabPane tab="复购分析" key="repeatPurchase" />
        </Tabs>
        <GovernanceScopeBar
          value={scope}
          onChange={(nextScope) => {
            setScope(nextScope);
            void handleQuery(undefined, nextScope);
          }}
          entityLabel={activeView === 'repeatPurchase' ? '复购分析' : '经营漏斗看板'}
        />
        <Alert
          showIcon
          type="info"
          message={`当前查询范围：${scopeLabel}`}
          description={
            activeView === 'repeatPurchase'
              ? '复购数据会先按治理范围收敛，再按渠道、活动和 180 天回看口径聚合，页面与导出保持同一查询链路。'
              : '漏斗数据会先按治理范围收敛，再按渠道、活动和时间桶聚合，前端不自行计算权限或转化率。'
          }
        />

        <Card>
          <Form<OperateDashboardFormValues>
            form={form}
            layout="vertical"
            onFinish={(values) => {
              void handleQuery(values, scope);
            }}
          >
            <Row gutter={16}>
              <Col xs={24} md={12} xl={8}>
                <Form.Item
                  name="timeRange"
                  label="时间范围"
                  rules={[{ required: true, message: '请选择时间范围' }]}
                >
                  <RangePicker
                    showTime
                    style={{ width: '100%' }}
                    allowClear={false}
                  />
                </Form.Item>
              </Col>
              <Col xs={24} md={12} xl={4}>
                <Form.Item name="bucket" label="时间桶">
                  <Select
                    disabled={activeView === 'repeatPurchase'}
                    options={
                      activeView === 'repeatPurchase'
                        ? [{ label: '按天', value: 'day' }]
                        : [
                            { label: '按天', value: 'day' },
                            { label: '按小时', value: 'hour' },
                          ]
                    }
                  />
                </Form.Item>
              </Col>
              <Col xs={24} md={12} xl={4}>
                <Form.Item name="channel" label="渠道">
                  <Select options={CHANNEL_OPTIONS} />
                </Form.Item>
              </Col>
              <Col xs={24} md={12} xl={4}>
                <Form.Item name="activityType" label="活动类型">
                  <Select
                    options={ACTIVITY_TYPE_OPTIONS}
                    onChange={(value) => {
                      setSelectedActivityType(value || '');
                      form.setFieldsValue({ activityId: undefined });
                    }}
                  />
                </Form.Item>
              </Col>
              <Col xs={24} md={12} xl={4}>
                <Form.Item name="activityId" label="活动实例">
                  <Select
                    allowClear
                    showSearch
                    optionFilterProp="label"
                    placeholder="全部活动实例"
                    options={currentActivityOptions}
                  />
                </Form.Item>
              </Col>
            </Row>

            <Space>
              <Button type="primary" htmlType="submit" icon={<SearchOutlined />}>
                {activeView === 'repeatPurchase' ? '查询复购分析' : '查询看板'}
              </Button>
              {activeView === 'repeatPurchase' ? (
                <Button icon={<DownloadOutlined />} onClick={() => { void handleExport(); }}>
                  导出结果
                </Button>
              ) : null}
              <Button
                icon={<ReloadOutlined />}
                onClick={() => {
                  form.setFieldsValue({
                    timeRange: defaultTimeRange,
                    channel: '',
                    activityType: '',
                    activityId: undefined,
                    bucket: 'day',
                  });
                  setSelectedActivityType('');
                  void handleQuery(
                    {
                      timeRange: defaultTimeRange,
                      channel: '',
                      activityType: '',
                      activityId: undefined,
                      bucket: 'day',
                    },
                    scope,
                  );
                }}
              >
                重置
              </Button>
              <Tag color="blue">
                实际聚合：{
                  activeView === 'repeatPurchase'
                    ? repeatPurchaseData?.bucket === 'day'
                      ? '按天'
                      : repeatPurchaseData?.bucket || '按天'
                    : dashboardData?.bucket === 'hour'
                      ? '按小时'
                      : '按天'
                }
              </Tag>
              {(activeView === 'repeatPurchase'
                ? repeatPurchaseData?.trackingStartedAt
                : dashboardData?.trackingStartedAt) ? (
                <Tag color="gold">
                  可信起点：{activeView === 'repeatPurchase' ? repeatPurchaseData?.trackingStartedAt : dashboardData?.trackingStartedAt}
                </Tag>
              ) : null}
            </Space>
          </Form>
        </Card>

        {(activeView === 'repeatPurchase' ? repeatPurchaseNotice : dashboardNotice) ? (
          <Alert
            showIcon
            type={(activeView === 'repeatPurchase' ? repeatPurchaseNotice : dashboardNotice)?.type}
            message={(activeView === 'repeatPurchase' ? repeatPurchaseNotice : dashboardNotice)?.message}
            description={(activeView === 'repeatPurchase' ? repeatPurchaseNotice : dashboardNotice)?.description}
          />
        ) : null}

        <Spin spinning={loading}>
          {activeView === 'repeatPurchase' ? (
            <Row gutter={[16, 16]}>
              {repeatCards.map((card) => (
                <Col xs={24} sm={12} xl={4} key={card.key}>
                  <Card>
                    <Statistic title={card.label} value={card.value} />
                  </Card>
                </Col>
              ))}
            </Row>
          ) : (
            <Row gutter={[16, 16]}>
              {(dashboardData?.overview.cards || []).map((card) => (
                <Col xs={24} sm={12} xl={4} key={card.key}>
                  <Card>
                    <Statistic title={card.label} value={card.value || 0} />
                    <Text type="secondary">
                      {card.rateLabel || '转化率'}：{formatPercent(card.rate)}
                    </Text>
                  </Card>
                </Col>
              ))}
            </Row>
          )}

          <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
            <Col xs={24} xl={16}>
              <Card title={activeView === 'repeatPurchase' ? '复购趋势' : '阶段趋势'}>
                {(activeView === 'repeatPurchase' ? repeatViewState : viewState) === 'empty' ? (
                  <Empty description={activeView === 'repeatPurchase' ? '当前筛选条件下暂无可展示的复购趋势' : '当前筛选条件下暂无可展示的趋势数据'} />
                ) : (
                  <ReactEcharts
                    option={
                      activeView === 'repeatPurchase'
                        ? buildRepeatPurchaseTrendOption(repeatPurchaseData?.trends || [])
                        : buildOperateTrendOption(
                            dashboardData?.series || [],
                            dashboardData?.bucket || 'day',
                          )
                    }
                    style={{ height: 360 }}
                  />
                )}
              </Card>
            </Col>
            <Col xs={24} xl={8}>
              <Card title="总览说明">
                <Space direction="vertical" size={12} style={{ width: '100%' }}>
                  <Text strong>指标口径</Text>
                  {activeView === 'repeatPurchase' ? (
                    <>
                      <Text>复购买家数 = 当前结果中至少拥有 1 笔复购订单的去重买家</Text>
                      <Text>复购率 = 复购买家数 / 支付买家数</Text>
                      <Text>复购订单数 = 当前结果中命中 180 天回看窗口的订单数</Text>
                      <Text>复购GMV = 当前结果中复购订单的支付金额汇总</Text>
                      <Text>平均复购天数 = 复购订单相对前序有效支付订单的平均间隔</Text>
                      <Alert
                        showIcon
                        type="info"
                        message="导出一致性"
                        description="导出结果严格复用当前页面筛选条件与治理范围，昵称和手机号默认脱敏展示。"
                      />
                    </>
                  ) : (
                    <>
                      <Text>点击率 = 点击 / 曝光</Text>
                      <Text>加购率 = 加购 / 点击</Text>
                      <Text>下单率 = 下单 / 加购</Text>
                      <Text>支付率 = 支付 / 下单</Text>
                      <Text>核销率 = 核销 / 支付</Text>
                      <Alert
                        showIcon
                        type="info"
                        message="核销率语义"
                        description="当前看板中的核销率表示“支付成功订单中的优惠券核销占比”，不是领券到核销的转化率。"
                      />
                    </>
                  )}
                </Space>
              </Card>
            </Col>
          </Row>

          <Card title={activeView === 'repeatPurchase' ? '复购详情' : '时间桶明细'} style={{ marginTop: 16 }}>
            {(activeView === 'repeatPurchase' ? repeatViewState : viewState) === 'empty' ? (
              <Empty description={activeView === 'repeatPurchase' ? '当前筛选条件下没有可展示的复购详情' : '当前筛选条件下没有可展示的明细表格'} />
            ) : (
              activeView === 'repeatPurchase' ? (
                <Table<RepeatPurchaseDetailTableRow>
                  rowKey="key"
                  columns={repeatColumns}
                  dataSource={repeatDetailRows}
                  pagination={{
                    current: repeatPurchaseData?.pageNum || 1,
                    pageSize: repeatPurchaseData?.pageSize || 20,
                    total: repeatPurchaseData?.total || 0,
                    showSizeChanger: true,
                    onChange: (page, pageSize) => {
                      void loadRepeatPurchaseDashboard(undefined, scope, page, pageSize);
                    },
                  }}
                  scroll={{ x: 1800 }}
                />
              ) : (
                <Table<OperateDashboardTableRow>
                  rowKey="key"
                  columns={columns}
                  dataSource={tableData}
                  pagination={{ pageSize: 10, showSizeChanger: false }}
                  scroll={{ x: 1380 }}
                />
              )
            )}
          </Card>
        </Spin>
      </Space>
    </PageContainer>
  );
};

export default OperateDashboard;
