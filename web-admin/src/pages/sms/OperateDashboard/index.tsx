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
  Tag,
  Typography,
} from 'antd';
import { ReloadOutlined, SearchOutlined } from '@ant-design/icons';
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
} from './data.d';
import {
  ACTIVITY_TYPE_OPTIONS,
  buildActivityOptions,
  buildDashboardNotice,
  buildOperateTableRows,
  buildOperateTrendOption,
  CHANNEL_OPTIONS,
  formatPercent,
  getDashboardViewState,
} from './helper';
import { queryOperateFunnelDashboard } from './service';

const { RangePicker } = DatePicker;
const { Text } = Typography;

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
  const [dashboardData, setDashboardData] = useState<QueryOperateFunnelDashboardData>();
  const [selectedActivityType, setSelectedActivityType] = useState<string>('');

  const loadDashboard = useCallback(async (
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
    void loadDashboard(
      {
        timeRange: defaultTimeRange,
        channel: '',
        activityType: '',
        activityId: undefined,
        bucket: 'day',
      },
      defaultGovernanceScope,
    );
  }, [form, loadDashboard]);

  const scopeLabel = buildGovernanceScopeLabel(scope);
  const dashboardNotice = buildDashboardNotice(dashboardData);
  const activityOptions = buildActivityOptions(
    dashboardData?.activityOptions || [],
    selectedActivityType,
  );
  const tableData = buildOperateTableRows(dashboardData?.series || []);
  const viewState = getDashboardViewState(dashboardData);

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

  return (
    <PageContainer>
      <Space direction="vertical" size={16} style={{ width: '100%' }}>
        <GovernanceScopeBar
          value={scope}
          onChange={(nextScope) => {
            setScope(nextScope);
            void loadDashboard(undefined, nextScope);
          }}
          entityLabel="经营漏斗看板"
        />
        <Alert
          showIcon
          type="info"
          message={`当前查询范围：${scopeLabel}`}
          description="漏斗数据会先按治理范围收敛，再按渠道、活动和时间桶聚合，前端不自行计算权限或转化率。"
        />

        <Card>
          <Form<OperateDashboardFormValues>
            form={form}
            layout="vertical"
            onFinish={(values) => {
              void loadDashboard(values, scope);
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
                    options={[
                      { label: '按天', value: 'day' },
                      { label: '按小时', value: 'hour' },
                    ]}
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
                    options={activityOptions}
                  />
                </Form.Item>
              </Col>
            </Row>

            <Space>
              <Button type="primary" htmlType="submit" icon={<SearchOutlined />}>
                查询看板
              </Button>
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
                  void loadDashboard(
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
                实际聚合：{dashboardData?.bucket === 'hour' ? '按小时' : '按天'}
              </Tag>
              {dashboardData?.trackingStartedAt ? (
                <Tag color="gold">可信起点：{dashboardData.trackingStartedAt}</Tag>
              ) : null}
            </Space>
          </Form>
        </Card>

        {dashboardNotice ? (
          <Alert
            showIcon
            type={dashboardNotice.type}
            message={dashboardNotice.message}
            description={dashboardNotice.description}
          />
        ) : null}

        <Spin spinning={loading}>
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

          <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
            <Col xs={24} xl={16}>
              <Card title="阶段趋势">
                {viewState === 'empty' ? (
                  <Empty description="当前筛选条件下暂无可展示的趋势数据" />
                ) : (
                  <ReactEcharts
                    option={buildOperateTrendOption(
                      dashboardData?.series || [],
                      dashboardData?.bucket || 'day',
                    )}
                    style={{ height: 360 }}
                  />
                )}
              </Card>
            </Col>
            <Col xs={24} xl={8}>
              <Card title="总览说明">
                <Space direction="vertical" size={12} style={{ width: '100%' }}>
                  <Text strong>指标口径</Text>
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
                </Space>
              </Card>
            </Col>
          </Row>

          <Card title="时间桶明细" style={{ marginTop: 16 }}>
            {viewState === 'empty' ? (
              <Empty description="当前筛选条件下没有可展示的明细表格" />
            ) : (
              <Table<OperateDashboardTableRow>
                rowKey="key"
                columns={columns}
                dataSource={tableData}
                pagination={{ pageSize: 10, showSizeChanger: false }}
                scroll={{ x: 1380 }}
              />
            )}
          </Card>
        </Spin>
      </Space>
    </PageContainer>
  );
};

export default OperateDashboard;
