import React, { useMemo, useRef, useState } from 'react';
import { Alert, Button, Input, Modal, Space, Tag, Typography, message } from 'antd';
import { ExportOutlined, StopOutlined } from '@ant-design/icons';
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
import {
  adminRevokeDigitalCardClaimToken,
  queryDigitalCardClaimTokenList,
} from './service';
import {
  CLAIM_TOKEN_STATUS_LABELS,
  CLAIM_TOKEN_STATUS_TAG_COLORS,
  type DigitalCardClaimTokenListItem,
  type QueryDigitalCardClaimTokenListParams,
} from './data.d';

const { Paragraph, Text } = Typography;
const { TextArea } = Input;

const STATUS_OPTIONS = Object.entries(CLAIM_TOKEN_STATUS_LABELS).map(([value, label]) => ({
  value,
  label,
}));

/**
 * Story 10.7 Task 8.8 / S5+S6 — 后台分享凭证管理（独立菜单）
 *
 * 监管约束：
 * - 后台只展示脱敏 token（首 8 位 + ***），原始 token 永不暴露
 * - 手动吊销必须填写原因，写入审计日志 operator_type=admin
 */
const DigitalCardClaimTokenPage: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const [scope, setScope] = useState<GovernanceScopeValue>(defaultGovernanceScope);
  const [detailItem, setDetailItem] = useState<DigitalCardClaimTokenListItem | null>(null);
  const [revokeItem, setRevokeItem] = useState<DigitalCardClaimTokenListItem | null>(null);
  const [revokeReason, setRevokeReason] = useState('');
  const [revoking, setRevoking] = useState(false);

  const scopeLabel = buildGovernanceScopeLabel(scope);

  const handleConfirmRevoke = async () => {
    if (!revokeItem) return;
    const reason = revokeReason.trim();
    if (!reason) {
      message.warning('请输入吊销原因');
      return;
    }
    setRevoking(true);
    try {
      const resp = await adminRevokeDigitalCardClaimToken({
        ...toGovernancePayload(scope),
        tokenId: revokeItem.id,
        reason,
      });
      if (resp?.success) {
        message.success(resp.message || '凭证已吊销');
        setRevokeItem(null);
        setRevokeReason('');
        actionRef.current?.reload();
      } else {
        message.error(resp?.message || '吊销失败');
      }
    } catch (err: any) {
      message.error(err?.message || '吊销失败，请稍后重试');
    } finally {
      setRevoking(false);
    }
  };

  const handleExportCsv = async () => {
    try {
      const resp = await queryDigitalCardClaimTokenList({
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
        'Token ID',
        'Token 脱敏',
        '资产编号',
        '模板',
        '签发人ID',
        '签发人类型',
        '过期时间',
        '最大领取次数',
        '已领取次数',
        '状态',
        '领取人ID',
        '领取时间',
        '创建时间',
      ];
      const rows = list.map((item) => [
        item.id,
        item.tokenMasked,
        item.assetNo,
        item.templateName,
        item.issuerId,
        item.issuerType,
        item.expireAt,
        item.maxClaims,
        item.claimedCount,
        CLAIM_TOKEN_STATUS_LABELS[item.status] || item.status,
        item.claimedBy || '',
        item.claimedAt,
        item.createTime,
      ]);
      const csv = [headers, ...rows]
        .map((r) => r.map((cell) => `"${String(cell ?? '').replace(/"/g, '""')}"`).join(','))
        .join('\n');
      const blob = new Blob(['\uFEFF' + csv], { type: 'text/csv;charset=utf-8;' });
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `digital-card-claim-token-${Date.now()}.csv`;
      link.click();
      URL.revokeObjectURL(url);
      message.success(`已导出 ${list.length} 条记录`);
    } catch (err) {
      message.error('导出失败，请稍后重试');
    }
  };

  const columns: ProColumns<DigitalCardClaimTokenListItem>[] = useMemo(
    () => [
      {
        title: 'Token ID',
        dataIndex: 'tokenId',
        width: 100,
        valueType: 'digit',
        render: (_: any, entity: DigitalCardClaimTokenListItem) => entity.id,
      },
      {
        title: 'Token (脱敏)',
        dataIndex: 'tokenMasked',
        width: 140,
        hideInSearch: true,
        render: (_: any, entity: DigitalCardClaimTokenListItem) => (
          <Text code>{entity.tokenMasked}</Text>
        ),
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
        title: '签发人ID',
        dataIndex: 'issuerId',
        width: 100,
        valueType: 'digit',
      },
      {
        title: '状态',
        dataIndex: 'status',
        width: 100,
        valueType: 'select',
        fieldProps: { options: STATUS_OPTIONS, placeholder: '全部' },
        render: (_: any, entity: DigitalCardClaimTokenListItem) => (
          <Tag color={CLAIM_TOKEN_STATUS_TAG_COLORS[entity.status] || 'default'}>
            {CLAIM_TOKEN_STATUS_LABELS[entity.status] || entity.status}
          </Tag>
        ),
      },
      {
        title: '领取进度',
        width: 100,
        hideInSearch: true,
        render: (_: any, entity: DigitalCardClaimTokenListItem) => (
          <Text type={entity.claimedCount >= entity.maxClaims ? 'warning' : undefined}>
            {entity.claimedCount} / {entity.maxClaims}
          </Text>
        ),
      },
      {
        title: '过期时间',
        dataIndex: 'expireAt',
        width: 170,
        hideInSearch: true,
        render: (_: any, entity: DigitalCardClaimTokenListItem) =>
          entity.expireAt || <Text type="secondary">永不过期</Text>,
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
        width: 150,
        fixed: 'right',
        render: (_: any, entity: DigitalCardClaimTokenListItem) => (
          <Space size="small">
            <Button type="link" size="small" onClick={() => setDetailItem(entity)}>
              详情
            </Button>
            {entity.status === 'active' && (
              <Button
                type="link"
                danger
                size="small"
                icon={<StopOutlined />}
                onClick={() => {
                  setRevokeItem(entity);
                  setRevokeReason('');
                }}
              >
                吊销
              </Button>
            )}
          </Space>
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
        type="warning"
        style={{ marginBottom: 16 }}
        message={`当前查询范围：${scopeLabel}`}
        description="分享凭证 token 仅展示前 8 位脱敏值，原始值不予暴露。吊销操作将写入审计日志（operator_type=admin），且不可逆。"
      />
      <ProTable<DigitalCardClaimTokenListItem, QueryDigitalCardClaimTokenListParams>
        headerTitle="分享凭证管理"
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
          const payload: QueryDigitalCardClaimTokenListParams = {
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
          const resp = await queryDigitalCardClaimTokenList(payload);
          return {
            data: resp?.data ?? [],
            total: resp?.total ?? 0,
            success: resp?.success ?? false,
          };
        }}
      />

      <Modal
        title="分享凭证详情"
        open={!!detailItem}
        footer={null}
        onCancel={() => setDetailItem(null)}
        width={640}
      >
        {detailItem && (
          <Space direction="vertical" style={{ width: '100%' }}>
            <Paragraph>
              <Text strong>Token：</Text>
              <Text code>{detailItem.tokenMasked}</Text>
              <Text type="secondary" style={{ marginLeft: 8 }}>
                (#{detailItem.id})
              </Text>
            </Paragraph>
            <Paragraph>
              <Text strong>状态：</Text>
              <Tag color={CLAIM_TOKEN_STATUS_TAG_COLORS[detailItem.status] || 'default'}>
                {CLAIM_TOKEN_STATUS_LABELS[detailItem.status] || detailItem.status}
              </Tag>
            </Paragraph>
            <Paragraph>
              <Text strong>资产：</Text>
              {detailItem.assetNo} (#{detailItem.cardInstanceId}) ·{' '}
              {detailItem.templateName || '-'}
            </Paragraph>
            <Paragraph>
              <Text strong>签发人：</Text>#{detailItem.issuerId} ({detailItem.issuerType})
            </Paragraph>
            <Paragraph>
              <Text strong>领取进度：</Text>
              {detailItem.claimedCount} / {detailItem.maxClaims}
            </Paragraph>
            <Paragraph>
              <Text strong>过期时间：</Text>
              {detailItem.expireAt || '永不过期'}
            </Paragraph>
            {detailItem.claimedAt && (
              <Paragraph>
                <Text strong>最近领取：</Text>
                #{detailItem.claimedBy} @ {detailItem.claimedAt}
              </Paragraph>
            )}
            <Paragraph>
              <Text strong>创建：</Text>
              {detailItem.createTime}
              <Text type="secondary" style={{ marginLeft: 16 }}>
                更新：{detailItem.updateTime}
              </Text>
            </Paragraph>
          </Space>
        )}
      </Modal>

      <Modal
        title={`吊销分享凭证 #${revokeItem?.id ?? ''}`}
        open={!!revokeItem}
        onOk={handleConfirmRevoke}
        confirmLoading={revoking}
        okText="确认吊销"
        okButtonProps={{ danger: true }}
        cancelText="取消"
        onCancel={() => {
          setRevokeItem(null);
          setRevokeReason('');
        }}
      >
        {revokeItem && (
          <Space direction="vertical" style={{ width: '100%' }}>
            <Alert
              type="warning"
              showIcon
              message="此操作将立即吊销分享凭证，且不可恢复。已成功领取的资产不会受影响。"
            />
            <div>
              <Text strong>Token：</Text>
              <Text code>{revokeItem.tokenMasked}</Text>
            </div>
            <div>
              <Text strong>资产：</Text>
              {revokeItem.assetNo} · {revokeItem.templateName || '-'}
            </div>
            <div>
              <Text strong>签发人：</Text>#{revokeItem.issuerId}
            </div>
            <div>
              <Text strong type="danger">
                吊销原因（必填）：
              </Text>
              <TextArea
                value={revokeReason}
                onChange={(e) => setRevokeReason(e.target.value)}
                rows={3}
                placeholder="例：用户投诉、风控异常、分享链路违规等"
                maxLength={500}
                showCount
              />
            </div>
          </Space>
        )}
      </Modal>
    </PageContainer>
  );
};

export default DigitalCardClaimTokenPage;
