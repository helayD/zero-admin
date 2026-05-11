import React, { useMemo, useRef, useState } from 'react';
import { Alert, Button, Modal, Space, Tag, Typography, message } from 'antd';
import { ExportOutlined } from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-layout';
import type { ActionType, ProColumns } from '@ant-design/pro-table';
import ProTable from '@ant-design/pro-table';
import GovernanceScopeBar from '@/pages/system/components/GovernanceScopeBar';
import {
  buildGovernanceScopeLabel,
  defaultGovernanceScope,
  toGovernancePayload,
  type GovernanceScopeValue,
} from '@/pages/system/components/governance';
import { queryDigitalCardTransferLogList } from './service';
import {
  OPERATION_TYPE_LABELS,
  type DigitalCardTransferLogListItem,
  type QueryDigitalCardTransferLogListParams,
} from './data.d';

const { Paragraph, Text } = Typography;

const OPERATION_TYPE_OPTIONS = Object.entries(OPERATION_TYPE_LABELS).map(([value, label]) => ({
  value,
  label,
}));

const OPERATION_TYPE_TAG_COLORS: Record<string, string> = {
  holder_transferred: 'blue',
  claim_token_generated: 'cyan',
  claim_token_revoked: 'orange',
  claim_attempt_failed: 'red',
  asset_transferred: 'magenta',
  asset_transferred_rolled_back: 'volcano',
  holder_transferred_backfilled: 'gold',
};

function safeFormatJson(raw: string): string {
  if (!raw || raw.trim() === '') return '';
  try {
    const parsed = JSON.parse(raw);
    return JSON.stringify(parsed, null, 2);
  } catch {
    return raw;
  }
}

function summarizePayload(raw: string): string {
  if (!raw) return '';
  try {
    const parsed = JSON.parse(raw);
    if (parsed && typeof parsed === 'object') {
      const keys = Object.keys(parsed);
      if (keys.length === 0) return '';
      const preview = keys.slice(0, 3).map((k) => `${k}=${JSON.stringify(parsed[k])}`);
      const suffix = keys.length > 3 ? ', ...' : '';
      return preview.join(', ') + suffix;
    }
  } catch {
    // fallthrough
  }
  return raw.length > 50 ? raw.slice(0, 50) + '...' : raw;
}

/**
 * Story 10.7 Task 9.3 — 后台合规审计：转赠/分享/领取事件跨资产检索
 *
 * 与 DigitalCardAssetWorkbench 的"资产维度"视图互补：
 * - DigitalCardAssetWorkbench：以单张提货卡为锚点查看其完整生命周期
 * - DigitalCardTransferAudit（本页）：以"事件"为锚点，横向追查异常模式（越权转赠、爆破领取等）
 */
const DigitalCardTransferAuditPage: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const [scope, setScope] = useState<GovernanceScopeValue>(defaultGovernanceScope);
  const [detailItem, setDetailItem] = useState<DigitalCardTransferLogListItem | null>(null);

  const scopeLabel = buildGovernanceScopeLabel(scope);

  const handleExportCsv = async (filters: Record<string, any>) => {
    try {
      const resp = await queryDigitalCardTransferLogList({
        ...toGovernancePayload(scope),
        ...filters,
        current: 1,
        pageSize: 10000, // Task 9.5: CSV 导出上限 10000 行
      } as QueryDigitalCardTransferLogListParams);
      const list = resp?.data ?? [];
      if (list.length === 0) {
        message.info('当前筛选条件下没有可导出的数据');
        return;
      }
      const headers = [
        '操作时间',
        '资产ID',
        '资产编号',
        '模板名',
        '操作类型',
        '操作者',
        '来源状态',
        '目标状态',
        '原因码',
        '原因',
        'TraceId',
        'Payload',
      ];
      const rows = list.map((item) => [
        item.createTime,
        item.assetInstanceId,
        item.assetNo,
        item.templateName,
        OPERATION_TYPE_LABELS[item.operationType] || item.operationType,
        item.operatorType,
        item.fromStatus,
        item.toStatus,
        item.reasonCode,
        item.reasonText,
        item.traceId,
        item.payloadJson,
      ]);
      const csv = [headers, ...rows]
        .map((r) => r.map((cell) => `"${String(cell ?? '').replace(/"/g, '""')}"`).join(','))
        .join('\n');
      // Excel 兼容 BOM
      const blob = new Blob(['\uFEFF' + csv], { type: 'text/csv;charset=utf-8;' });
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `digital-card-transfer-audit-${Date.now()}.csv`;
      link.click();
      URL.revokeObjectURL(url);
      message.success(`已导出 ${list.length} 条记录`);
    } catch (err) {
      message.error('导出失败，请稍后重试');
    }
  };

  const columns: ProColumns<DigitalCardTransferLogListItem>[] = useMemo(
    () => [
      {
        title: '操作时间',
        dataIndex: 'createTime',
        width: 170,
        hideInSearch: true,
        fixed: 'left',
      },
      {
        title: '资产编号',
        dataIndex: 'assetNo',
        width: 160,
        fieldProps: { placeholder: '请输入资产编号' },
      },
      {
        title: '模板',
        dataIndex: 'templateName',
        width: 160,
        hideInSearch: true,
        ellipsis: true,
      },
      {
        title: '操作类型',
        dataIndex: 'operationType',
        width: 140,
        valueType: 'select',
        fieldProps: { options: OPERATION_TYPE_OPTIONS, placeholder: '全部' },
        render: (_: any, entity: DigitalCardTransferLogListItem) => (
          <Tag color={OPERATION_TYPE_TAG_COLORS[entity.operationType] || 'default'}>
            {OPERATION_TYPE_LABELS[entity.operationType] || entity.operationType}
          </Tag>
        ),
      },
      {
        title: '操作者',
        dataIndex: 'operatorType',
        width: 100,
        hideInSearch: true,
      },
      {
        title: '状态变更',
        width: 180,
        hideInSearch: true,
        render: (_: any, entity: DigitalCardTransferLogListItem) => {
          if (!entity.fromStatus && !entity.toStatus) return '-';
          return (
            <Text type="secondary">
              {entity.fromStatus || '-'} → {entity.toStatus || '-'}
            </Text>
          );
        },
      },
      {
        title: '原因',
        dataIndex: 'reasonText',
        ellipsis: true,
        hideInSearch: true,
      },
      {
        title: 'TraceId',
        dataIndex: 'traceId',
        width: 200,
        ellipsis: true,
        copyable: true,
      },
      {
        title: '操作时间起',
        dataIndex: 'dateFrom',
        valueType: 'dateTime',
        hideInTable: true,
      },
      {
        title: '操作时间止',
        dataIndex: 'dateTo',
        valueType: 'dateTime',
        hideInTable: true,
      },
      {
        title: 'Payload 摘要',
        width: 240,
        hideInSearch: true,
        render: (_: any, entity: DigitalCardTransferLogListItem) => {
          const summary = summarizePayload(entity.payloadJson);
          return summary ? <Text ellipsis>{summary}</Text> : '-';
        },
      },
      {
        title: '操作',
        valueType: 'option',
        width: 80,
        fixed: 'right',
        render: (_: any, entity: DigitalCardTransferLogListItem) => (
          <Button type="link" size="small" onClick={() => setDetailItem(entity)}>
            详情
          </Button>
        ),
      },
    ],
    [],
  );

  return (
    <PageContainer>
      <GovernanceScopeBar
        value={scope}
        onChange={(next) => {
          setScope(next);
          actionRef.current?.reload();
        }}
      />
      <Alert
        type="info"
        style={{ marginBottom: 16 }}
        message={`当前查询范围：${scopeLabel}`}
        description="本页跨资产展示提货卡的转赠、分享凭证生成/吊销、领取尝试等审计事件。Story 10.7 Task 9 实现，与「提货卡资产」抽屉的资产维度审计互补。"
      />
      <ProTable<DigitalCardTransferLogListItem, QueryDigitalCardTransferLogListParams>
        headerTitle="转赠审计记录"
        actionRef={actionRef}
        rowKey="id"
        scroll={{ x: 1500 }}
        columns={columns}
        toolBarRender={() => [
          <Button
            key="export"
            icon={<ExportOutlined />}
            onClick={() => handleExportCsv({})}
          >
            导出 CSV
          </Button>,
        ]}
        request={async (params: any) => {
          const { current, pageSize, dateFrom, dateTo, ...rest } = params as any;
          const payload: QueryDigitalCardTransferLogListParams = {
            ...toGovernancePayload(scope),
            ...rest,
            current,
            pageSize,
            dateFrom: dateFrom ? new Date(dateFrom).toISOString().slice(0, 19).replace('T', ' ') : undefined,
            dateTo: dateTo ? new Date(dateTo).toISOString().slice(0, 19).replace('T', ' ') : undefined,
          };
          const resp = await queryDigitalCardTransferLogList(payload);
          return {
            data: resp?.data ?? [],
            total: resp?.total ?? 0,
            success: resp?.success ?? false,
          };
        }}
      />

      <Modal
        title="审计事件详情"
        open={!!detailItem}
        footer={null}
        onCancel={() => setDetailItem(null)}
        width={720}
      >
        {detailItem && (
          <Space direction="vertical" style={{ width: '100%' }}>
            <Paragraph>
              <Text strong>操作时间：</Text>
              {detailItem.createTime}
            </Paragraph>
            <Paragraph>
              <Text strong>资产编号：</Text>
              {detailItem.assetNo} (#{detailItem.assetInstanceId})
            </Paragraph>
            <Paragraph>
              <Text strong>模板：</Text>
              {detailItem.templateName || '-'}
            </Paragraph>
            <Paragraph>
              <Text strong>操作类型：</Text>
              <Tag color={OPERATION_TYPE_TAG_COLORS[detailItem.operationType] || 'default'}>
                {OPERATION_TYPE_LABELS[detailItem.operationType] || detailItem.operationType}
              </Tag>
              （{detailItem.operatorType}）
            </Paragraph>
            <Paragraph>
              <Text strong>状态变更：</Text>
              {detailItem.fromStatus || '-'} → {detailItem.toStatus || '-'}
            </Paragraph>
            <Paragraph>
              <Text strong>原因：</Text>
              {detailItem.reasonCode ? <Tag>{detailItem.reasonCode}</Tag> : null}
              {detailItem.reasonText || '-'}
            </Paragraph>
            <Paragraph>
              <Text strong>TraceId：</Text>
              <Text code copyable>
                {detailItem.traceId || '-'}
              </Text>
            </Paragraph>
            <div>
              <Text strong>Payload：</Text>
              <pre
                style={{
                  background: '#f5f5f5',
                  padding: 12,
                  borderRadius: 4,
                  maxHeight: 320,
                  overflow: 'auto',
                  marginTop: 8,
                }}
              >
                {safeFormatJson(detailItem.payloadJson) || '(空)'}
              </pre>
            </div>
          </Space>
        )}
      </Modal>
    </PageContainer>
  );
};

export default DigitalCardTransferAuditPage;
