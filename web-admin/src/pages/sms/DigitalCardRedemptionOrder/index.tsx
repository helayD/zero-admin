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
import { queryDigitalCardRedemptionOrderList } from './service';
import {
  REDEMPTION_ORDER_STATUS_LABELS,
  REDEMPTION_ORDER_STATUS_TAG_COLORS,
  type DigitalCardRedemptionOrderListItem,
  type QueryDigitalCardRedemptionOrderListParams,
} from './data.d';

const { Paragraph, Text } = Typography;

const STATUS_OPTIONS = Object.entries(REDEMPTION_ORDER_STATUS_LABELS).map(([value, label]) => ({
  value,
  label,
}));

/**
 * Story 10.7 Task 2.8 / S3+S4 — 后台提货单管理（独立菜单）
 * 与「提货卡资产」抽屉视图互补：以"订单"为锚点跨资产检索 + CSV 导出
 */
const DigitalCardRedemptionOrderPage: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const [scope, setScope] = useState<GovernanceScopeValue>(defaultGovernanceScope);
  const [detailItem, setDetailItem] = useState<DigitalCardRedemptionOrderListItem | null>(null);

  const scopeLabel = buildGovernanceScopeLabel(scope);

  const handleExportCsv = async () => {
    try {
      const resp = await queryDigitalCardRedemptionOrderList({
        ...toGovernancePayload(scope),
        current: 1,
        pageSize: 10000,
      });
      const list = resp?.data ?? [];
      if (list.length === 0) {
        message.info('当前筛选条件下没有可导出的数据');
        return;
      }
      const headers = [
        '提货单号',
        '资产编号',
        '模板',
        '持有人ID',
        '收货人',
        '收货电话',
        '收货地址',
        '状态',
        '发货时间',
        '签收时间',
        '取消原因',
        'OMS订单ID',
        '创建时间',
      ];
      const rows = list.map((item) => [
        item.orderNo,
        item.assetNo,
        item.templateName,
        item.holderId,
        item.receiverName,
        item.receiverPhone,
        item.receiverAddress,
        REDEMPTION_ORDER_STATUS_LABELS[item.status] || item.status,
        item.shippedAt,
        item.deliveredAt,
        item.cancelReason,
        item.omsOrderId,
        item.createTime,
      ]);
      const csv = [headers, ...rows]
        .map((r) => r.map((cell) => `"${String(cell ?? '').replace(/"/g, '""')}"`).join(','))
        .join('\n');
      const blob = new Blob(['\uFEFF' + csv], { type: 'text/csv;charset=utf-8;' });
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `digital-card-redemption-order-${Date.now()}.csv`;
      link.click();
      URL.revokeObjectURL(url);
      message.success(`已导出 ${list.length} 条记录`);
    } catch (err) {
      message.error('导出失败，请稍后重试');
    }
  };

  const columns: ProColumns<DigitalCardRedemptionOrderListItem>[] = useMemo(
    () => [
      {
        title: '提货单号',
        dataIndex: 'orderNo',
        width: 200,
        fixed: 'left',
        copyable: true,
        fieldProps: { placeholder: '支持模糊查询' },
      },
      {
        title: '资产编号',
        dataIndex: 'assetNo',
        width: 160,
        fieldProps: { placeholder: '支持模糊查询' },
      },
      {
        title: '模板',
        dataIndex: 'templateName',
        width: 160,
        hideInSearch: true,
        ellipsis: true,
      },
      {
        title: '持有人ID',
        dataIndex: 'holderId',
        width: 100,
        valueType: 'digit',
      },
      {
        title: '收货人',
        dataIndex: 'receiverName',
        width: 120,
        hideInSearch: true,
      },
      {
        title: '状态',
        dataIndex: 'status',
        width: 100,
        valueType: 'select',
        fieldProps: { options: STATUS_OPTIONS, placeholder: '全部' },
        render: (_: any, entity: DigitalCardRedemptionOrderListItem) => (
          <Tag color={REDEMPTION_ORDER_STATUS_TAG_COLORS[entity.status] || 'default'}>
            {REDEMPTION_ORDER_STATUS_LABELS[entity.status] || entity.status}
          </Tag>
        ),
      },
      {
        title: 'OMS订单ID',
        dataIndex: 'omsOrderId',
        width: 120,
        valueType: 'digit',
        render: (_: any, entity: DigitalCardRedemptionOrderListItem) =>
          entity.omsOrderId > 0 ? entity.omsOrderId : <Text type="secondary">-</Text>,
      },
      {
        title: '创建时间',
        dataIndex: 'createTime',
        width: 170,
        hideInSearch: true,
      },
      {
        title: '创建时间起',
        dataIndex: 'dateFrom',
        valueType: 'dateTime',
        hideInTable: true,
      },
      {
        title: '创建时间止',
        dataIndex: 'dateTo',
        valueType: 'dateTime',
        hideInTable: true,
      },
      {
        title: '操作',
        valueType: 'option',
        width: 80,
        fixed: 'right',
        render: (_: any, entity: DigitalCardRedemptionOrderListItem) => (
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
        description="本页跨资产管理所有数字卡提货单。Story 10.7 Task 2.8 / S3+S4 实现：以订单为锚点搜索 + CSV 导出。"
      />
      <ProTable<DigitalCardRedemptionOrderListItem, QueryDigitalCardRedemptionOrderListParams>
        headerTitle="提货单管理"
        actionRef={actionRef}
        rowKey="id"
        scroll={{ x: 1500 }}
        columns={columns}
        toolBarRender={() => [
          <Button key="export" icon={<ExportOutlined />} onClick={handleExportCsv}>
            导出 CSV
          </Button>,
        ]}
        request={async (params: any) => {
          const { current, pageSize, dateFrom, dateTo, ...rest } = params;
          const payload: QueryDigitalCardRedemptionOrderListParams = {
            ...toGovernancePayload(scope),
            ...rest,
            current,
            pageSize,
            dateFrom: dateFrom
              ? new Date(dateFrom).toISOString().slice(0, 19).replace('T', ' ')
              : undefined,
            dateTo: dateTo
              ? new Date(dateTo).toISOString().slice(0, 19).replace('T', ' ')
              : undefined,
          };
          const resp = await queryDigitalCardRedemptionOrderList(payload);
          return {
            data: resp?.data ?? [],
            total: resp?.total ?? 0,
            success: resp?.success ?? false,
          };
        }}
      />

      <Modal
        title="提货单详情"
        open={!!detailItem}
        footer={null}
        onCancel={() => setDetailItem(null)}
        width={720}
      >
        {detailItem && (
          <Space direction="vertical" style={{ width: '100%' }}>
            <Paragraph>
              <Text strong>提货单号：</Text>
              <Text code copyable>
                {detailItem.orderNo}
              </Text>
            </Paragraph>
            <Paragraph>
              <Text strong>状态：</Text>
              <Tag color={REDEMPTION_ORDER_STATUS_TAG_COLORS[detailItem.status] || 'default'}>
                {REDEMPTION_ORDER_STATUS_LABELS[detailItem.status] || detailItem.status}
              </Tag>
              {detailItem.cancelReason && (
                <Text type="danger" style={{ marginLeft: 8 }}>
                  ({detailItem.cancelReason})
                </Text>
              )}
            </Paragraph>
            <Paragraph>
              <Text strong>资产：</Text>
              {detailItem.assetNo} (#{detailItem.cardInstanceId}) · {detailItem.templateName || '-'}
            </Paragraph>
            <Paragraph>
              <Text strong>持有人：</Text>#{detailItem.holderId}
            </Paragraph>
            <Paragraph>
              <Text strong>收货人：</Text>
              {detailItem.receiverName} / {detailItem.receiverPhone}
            </Paragraph>
            <Paragraph>
              <Text strong>收货地址：</Text>
              {detailItem.receiverAddress}
            </Paragraph>
            <Paragraph>
              <Text strong>OMS 订单：</Text>
              {detailItem.omsOrderId > 0 ? `#${detailItem.omsOrderId}` : '未生成'}
            </Paragraph>
            <Paragraph>
              <Text strong>发货时间：</Text>
              {detailItem.shippedAt || '-'}
            </Paragraph>
            <Paragraph>
              <Text strong>签收时间：</Text>
              {detailItem.deliveredAt || '-'}
            </Paragraph>
            <Paragraph>
              <Text strong>创建时间：</Text>
              {detailItem.createTime}
              <Text type="secondary" style={{ marginLeft: 16 }}>
                更新：{detailItem.updateTime}
              </Text>
            </Paragraph>
          </Space>
        )}
      </Modal>
    </PageContainer>
  );
};

export default DigitalCardRedemptionOrderPage;
