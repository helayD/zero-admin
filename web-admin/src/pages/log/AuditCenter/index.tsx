import React, { useMemo, useRef, useState } from 'react';
import { CopyOutlined } from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-layout';
import ProTable, { ActionType, ProColumns } from '@ant-design/pro-table';
import ProDescriptions, { ProDescriptionsItemProps } from '@ant-design/pro-descriptions';
import { Alert, Button, Drawer, Empty, message, Space, Tag, Timeline, Typography } from 'antd';
import type { AuditCenterDetailData, AuditCenterItem, AuditTimelineItem } from './data.d';
import { queryAuditCenterDetail, queryAuditCenterList } from './service';
import GovernanceScopeBar from '@/pages/system/components/GovernanceScopeBar';
import {
  defaultGovernanceScope,
  governanceScopeColor,
  normalizeGovernanceScope,
  toGovernancePayload,
  type GovernanceScopeValue,
} from '@/pages/system/components/governance';

const { Paragraph, Text } = Typography;

const sourceTypeOptions = {
  login_log: { text: '登录日志' },
  operate_log: { text: '操作日志' },
  security_event: { text: '安全事件' },
  tenant_audit: { text: '租户审计' },
  merchant_audit: { text: '商户审计' },
};

const scopeTypeOptions = {
  platform: { text: '平台级' },
  tenant: { text: '租户级' },
  merchant: { text: '商户级' },
};

const resultOptions = {
  success: { text: '成功', status: 'Success' as const },
  approved: { text: '通过', status: 'Success' as const },
  active: { text: '生效', status: 'Success' as const },
  denied: { text: '拒绝', status: 'Error' as const },
  blocked: { text: '阻断', status: 'Error' as const },
  failed: { text: '失败', status: 'Error' as const },
  rejected: { text: '驳回', status: 'Error' as const },
  pending: { text: '待处理', status: 'Processing' as const },
};

const renderResultTag = (result?: string) => {
  if (!result) {
    return <Tag>未知</Tag>;
  }
  const normalized = result.toLowerCase();
  if (normalized.includes('success') || normalized.includes('approve') || normalized.includes('active')) {
    return <Tag color="success">{result}</Tag>;
  }
  if (
    normalized.includes('den') ||
    normalized.includes('block') ||
    normalized.includes('fail') ||
    normalized.includes('reject')
  ) {
    return <Tag color="error">{result}</Tag>;
  }
  if (normalized.includes('pending') || normalized.includes('processing')) {
    return <Tag color="processing">{result}</Tag>;
  }
  return <Tag color="default">{result}</Tag>;
};

const prettyPayload = (value?: string) => {
  if (!value) {
    return '-';
  }
  const trimmed = value.trim();
  if (!trimmed) {
    return '-';
  }
  try {
    return JSON.stringify(JSON.parse(trimmed), null, 2);
  } catch (error) {
    return trimmed;
  }
};

const renderSummaryLines = (value?: string) => {
  if (!value) {
    return <Text type="secondary">暂无摘要</Text>;
  }
  return (
    <Space direction="vertical" size={4} style={{ width: '100%' }}>
      {value
        .split(/\n+/)
        .map((item) => item.trim())
        .filter(Boolean)
        .map((item, index) => (
          <Text key={`${item}-${index}`}>{item}</Text>
        ))}
    </Space>
  );
};

const buildTimelineDescription = (item: AuditTimelineItem) => {
  const parts = [
    item.operatorName ? `处理人：${item.operatorName}` : '',
    item.resourceType ? `资源：${item.resourceType}${item.resourceName ? ` / ${item.resourceName}` : ''}` : '',
    item.traceId ? `TraceId：${item.traceId}` : '',
  ].filter(Boolean);
  return parts.join(' ｜ ');
};

const AuditCenterPage: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const [detailOpen, setDetailOpen] = useState(false);
  const [detail, setDetail] = useState<AuditCenterDetailData>();
  const [scope, setScope] = useState<GovernanceScopeValue>(defaultGovernanceScope);

  const normalizedScope = useMemo(() => normalizeGovernanceScope(scope), [scope]);
  const detailPayloadText = useMemo(() => prettyPayload(detail?.detailPayload), [detail?.detailPayload]);

  const openDetail = async (row: AuditCenterItem) => {
    try {
      const res = await queryAuditCenterDetail({
        sourceType: row.sourceType,
        sourceId: row.sourceId,
        ...toGovernancePayload(normalizedScope),
      });
      setDetail(res.data);
      setDetailOpen(true);
    } catch (error) {
      message.error('获取审计详情失败，请稍后重试');
    }
  };

  const columns: ProColumns<AuditCenterItem>[] = [
    {
      title: '来源',
      dataIndex: 'sourceType',
      valueEnum: sourceTypeOptions,
      width: 120,
      render: (_, row) => <Tag>{sourceTypeOptions[row.sourceType]?.text || row.sourceType}</Tag>,
    },
    { title: '事件类型', dataIndex: 'eventType', ellipsis: true },
    { title: '动作', dataIndex: 'action', ellipsis: true },
    {
      title: '结果',
      dataIndex: 'result',
      valueEnum: resultOptions,
      render: (_, row) => renderResultTag(row.result),
      width: 100,
    },
    {
      title: '作用域类型',
      dataIndex: 'scopeType',
      valueEnum: scopeTypeOptions,
      width: 110,
      render: (_, row) => <Tag color={governanceScopeColor(row.scopeType as any)}>{scopeTypeOptions[row.scopeType]?.text || row.scopeType || '-'}</Tag>,
    },
    { title: '治理范围', dataIndex: 'scopeLabel', hideInSearch: true, ellipsis: true, width: 150 },
    { title: '操作者', dataIndex: 'operatorName', width: 120 },
    { title: '资源类型', dataIndex: 'resourceType', width: 120 },
    { title: '资源标识', dataIndex: 'resourceId', width: 100 },
    { title: '资源名称', dataIndex: 'resourceName', hideInSearch: true, ellipsis: true, width: 140 },
    { title: 'TraceId', dataIndex: 'traceId', ellipsis: true, width: 180 },
    { title: '摘要', dataIndex: 'requestSummary', hideInSearch: true, ellipsis: true },
    {
      title: '开始时间',
      dataIndex: 'startTime',
      valueType: 'dateTime',
      hideInTable: true,
      fieldProps: { placeholder: '请输入开始时间' },
    },
    {
      title: '结束时间',
      dataIndex: 'endTime',
      valueType: 'dateTime',
      hideInTable: true,
      fieldProps: { placeholder: '请输入结束时间' },
    },
    { title: '发生时间', dataIndex: 'happenedAt', valueType: 'dateTime', hideInSearch: true, width: 180 },
    {
      title: '操作',
      valueType: 'option',
      width: 100,
      render: (_, row) => [
        <a key="detail" onClick={async () => openDetail(row)}>
          查看
        </a>,
      ],
    },
  ];

  const detailDescriptions: ProDescriptionsItemProps<AuditCenterDetailData>[] = [
    {
      title: '来源',
      dataIndex: 'sourceType',
      renderText: (value) => sourceTypeOptions[value as keyof typeof sourceTypeOptions]?.text || value || '-',
    },
    { title: '事件类型', dataIndex: 'eventType' },
    { title: '动作', dataIndex: 'action' },
    { title: '结果', dataIndex: 'result', render: (_, entity) => renderResultTag(entity.result) },
    {
      title: '作用域',
      dataIndex: 'scopeLabel',
      render: (_, entity) => (
        <Space size={8} wrap>
          <Tag color={governanceScopeColor(entity.scopeType as any)}>
            {scopeTypeOptions[entity.scopeType as keyof typeof scopeTypeOptions]?.text || entity.scopeType}
          </Tag>
          <span>{entity.scopeLabel || '-'}</span>
        </Space>
      ),
    },
    { title: '主体信息', dataIndex: 'subjectInfo' },
    { title: '处理人', dataIndex: 'operatorName' },
    { title: '资源类型', dataIndex: 'resourceType' },
    { title: '资源 ID', dataIndex: 'resourceId' },
    { title: '资源名称', dataIndex: 'resourceName' },
    {
      title: 'TraceId',
      dataIndex: 'traceId',
      render: (_, entity) =>
        entity.traceId ? (
          <Space>
            <Text code>{entity.traceId}</Text>
            <Button
              size="small"
              icon={<CopyOutlined />}
              onClick={async () => {
                await navigator.clipboard.writeText(entity.traceId);
                message.success('TraceId 已复制');
              }}
            >
              复制
            </Button>
          </Space>
        ) : (
          '-'
        ),
    },
    { title: '发生时间', dataIndex: 'happenedAt' },
  ];

  return (
    <PageContainer title={false}>
      <GovernanceScopeBar
        value={normalizedScope}
        onChange={(next) => {
          setScope(next);
          actionRef.current?.reload();
        }}
        entityLabel="审计记录"
        style={{ marginBottom: 16 }}
      />

      <Alert
        showIcon
        type="info"
        style={{ marginBottom: 16 }}
        message="统一审计中心"
        description="支持按主体、来源、结果、时间范围、资源和 TraceId 统一检索登录日志、操作日志、安全事件以及租户/商户治理审计。"
      />

      <ProTable<AuditCenterItem>
        headerTitle="治理审计中心"
        actionRef={actionRef}
        rowKey={(row) => `${row.sourceType}-${row.sourceId}`}
        search={{ labelWidth: 100 }}
        request={async (params) => {
          const { current, pageSize, ...rest } = params as Record<string, any>;
          const res = await queryAuditCenterList({
            current,
            pageSize,
            ...rest,
            ...toGovernancePayload(normalizedScope),
          } as any);
          return { data: res.data || [], success: res.success, total: res.total };
        }}
        columns={columns}
      />

      <Drawer
        width={900}
        open={detailOpen}
        onClose={() => setDetailOpen(false)}
        title="审计详情"
        destroyOnClose
      >
        {detail ? (
          <Space direction="vertical" style={{ width: '100%' }} size={16}>
            <ProDescriptions<AuditCenterDetailData>
              column={2}
              bordered
              dataSource={detail}
              columns={detailDescriptions}
            />

            <div>
              <Text strong>请求摘要</Text>
              <div style={{ marginTop: 8 }}>{renderSummaryLines(detail.requestSummary)}</div>
            </div>

            <div>
              <Space align="center" style={{ marginBottom: 8 }}>
                <Text strong>原始明细</Text>
                {detail.sensitiveMasked ? <Tag color="warning">已按当前角色脱敏</Tag> : <Tag color="success">原始可读</Tag>}
              </Space>
              <Paragraph>
                <pre
                  style={{
                    background: '#fafafa',
                    border: '1px solid #f0f0f0',
                    borderRadius: 6,
                    padding: 12,
                    whiteSpace: 'pre-wrap',
                    wordBreak: 'break-word',
                    marginBottom: 0,
                  }}
                >
                  {detailPayloadText}
                </pre>
              </Paragraph>
            </div>

            <div>
              <Text strong>关联时间线</Text>
              <div style={{ marginTop: 12 }}>
                {detail.timeline?.length ? (
                  <Timeline>
                    {detail.timeline.map((item) => (
                      <Timeline.Item key={`${item.sourceType}-${item.sourceId}-${item.happenedAt}`}>
                        <Space direction="vertical" size={4} style={{ width: '100%' }}>
                          <Space wrap>
                            <Tag>{sourceTypeOptions[item.sourceType]?.text || item.sourceType}</Tag>
                            <Text strong>{item.action || item.eventType || '-'}</Text>
                            {renderResultTag(item.result)}
                            <Text type="secondary">{item.happenedAt}</Text>
                          </Space>
                          {buildTimelineDescription(item) ? (
                            <Text type="secondary">{buildTimelineDescription(item)}</Text>
                          ) : null}
                          <div>{renderSummaryLines(item.requestSummary)}</div>
                        </Space>
                      </Timeline.Item>
                    ))}
                  </Timeline>
                ) : (
                  <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="当前记录暂无可串联时间线" />
                )}
              </div>
            </div>
          </Space>
        ) : (
          <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无详情数据" />
        )}
      </Drawer>
    </PageContainer>
  );
};

export default AuditCenterPage;
