import React, { useRef, useState } from 'react';
import { Alert, Button, message, Space } from 'antd';
import { DownloadOutlined } from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-layout';
import type { ActionType, ProColumns } from '@ant-design/pro-table';
import ProTable from '@ant-design/pro-table';
import type { ChainMonitorItem } from './data.d';
import { queryChainMonitorList, exportChainMonitorList } from './service';
import ChainDetailDrawer from './components/ChainDetailDrawer';
import GovernanceScopeBar from '@/pages/system/components/GovernanceScopeBar';
import {
  buildGovernanceScopeLabel,
  defaultGovernanceScope,
  toGovernancePayload,
  type GovernanceScopeValue,
} from '@/pages/system/components/governance';

const ChainMonitorList: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const [scope, setScope] = useState<GovernanceScopeValue>(defaultGovernanceScope);
  const [detailVisible, setDetailVisible] = useState<boolean>(false);
  const [currentRow, setCurrentRow] = useState<ChainMonitorItem>();

  const scopeLabel = buildGovernanceScopeLabel(scope);

  const handleExport = async () => {
    const hide = Message.loading('正在导出...');
    try {
      const params = toGovernancePayload(scope);
      const response = await exportChainMonitorList({
        ...params,
        pageSize: 10000,
        current: 1,
      });

      const blob = new Blob([response], { type: 'application/vnd.ms-excel' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `链路监控_${new Date().toISOString().slice(0, 10)}.xlsx`;
      link.click();
      window.URL.revokeObjectURL(url);
      hide();
      Message.success('导出成功');
    } catch (error) {
      hide();
      Message.error('导出失败');
    }
  };

  const columns: ProColumns<ChainMonitorItem>[] = [
    {
      title: '链路追踪ID',
      dataIndex: 'traceId',
      hideInSearch: true,
      width: 180,
      ellipsis: true,
    },
    {
      title: '业务编号',
      dataIndex: 'entityNo',
      hideInSearch: true,
      render: (dom, entity) => (
        <a
          onClick={() => {
            setCurrentRow(entity);
            setDetailVisible(true);
          }}
        >
          {dom}
        </a>
      ),
    },
    {
      title: '链路类型',
      dataIndex: 'chainTypeText',
      hideInSearch: true,
      width: 120,
      render: (_, entity) => (
        <span style={{ color: '#1890ff' }}>{entity.chainTypeText || '-'}</span>
      ),
    },
    {
      title: '阶段',
      dataIndex: 'stageText',
      hideInSearch: true,
      width: 120,
      render: (_, entity) => {
        const stageColor: Record<string, string> = {
          '需人工介入': '#ff4d4f',
          '取消补偿中': '#faad14',
          '待处理': '#faad14',
          '支付中': '#1890ff',
          '支付确认中': '#1890ff',
          '售后处理中': '#faad14',
          '已完成': '#52c41a',
          '正常': '#52c41a',
          '已取消/已关闭': '#8c8c8c',
          '已关闭': '#8c8c8c',
        };
        return (
          <span style={{ color: stageColor[entity.stageText || ''] || '#595959' }}>
            {entity.stageText || '-'}
          </span>
        );
      },
    },
    {
      title: '结果',
      dataIndex: 'resultText',
      hideInSearch: true,
      width: 100,
      render: (_, entity) => {
        const resultColor: Record<string, string> = {
          '成功': '#52c41a',
          '失败': '#ff4d4f',
          '人工处理': '#ff4d4f',
          '处理中': '#1890ff',
        };
        return (
          <span style={{ color: resultColor[entity.resultText || ''] || '#595959' }}>
            {entity.resultText || '-'}
          </span>
        );
      },
    },
    {
      title: '重试次数',
      dataIndex: 'retryCount',
      hideInSearch: true,
      width: 80,
      render: (_, entity) => {
        if (entity.retryCount >= 3) {
          return <span style={{ color: '#ff4d4f', fontWeight: 600 }}>{entity.retryCount}</span>;
        }
        return entity.retryCount;
      },
    },
    {
      title: '最近错误',
      dataIndex: 'lastError',
      hideInSearch: true,
      ellipsis: true,
      render: (_, entity) => {
        if (!entity.lastError) return '-';
        return <span style={{ color: '#ff4d4f' }}>{entity.lastError}</span>;
      },
    },
    {
      title: '最近执行时间',
      dataIndex: 'lastExecuteAt',
      hideInSearch: true,
      width: 170,
    },
    {
      title: '链路创建时间',
      dataIndex: 'createdAt',
      hideInSearch: true,
      width: 170,
    },
    {
      title: '业务编号',
      dataIndex: 'entityNo',
      hideInTable: true,
    },
  ];

  return (
    <PageContainer>
      <GovernanceScopeBar
        value={scope}
        onChange={(nextScope) => {
          setScope(nextScope);
          actionRef.current?.reload?.();
        }}
        entityLabel="链路"
        style={{ marginBottom: 16 }}
      />
      <Alert
        showIcon
        type="info"
        style={{ marginBottom: 16 }}
        message={`当前查询范围：${scopeLabel}`}
        description="按租户/商户筛选仅返回该主体下的链路数据；点击业务编号可查看链路详情。"
      />
      <ProTable<ChainMonitorItem>
        headerTitle="链路监控列表"
        actionRef={actionRef}
        rowKey="traceId"
        search={{
          labelWidth: 120,
        }}
        toolBarRender={() => [
          <Space key="toolbar">
            <Button
              key="export"
              icon={<DownloadOutlined />}
              onClick={handleExport}
            >
              导出
            </Button>
          </Space>,
        ]}
        request={(params) => queryChainMonitorList({ ...params, ...toGovernancePayload(scope) })}
        columns={columns}
        rowSelection={false}
        pagination={{ pageSize: 10 }}
      />

      <ChainDetailDrawer
        key="ChainDetailDrawer"
        visible={detailVisible}
        chainData={currentRow}
        onClose={() => {
          setDetailVisible(false);
          setCurrentRow(undefined);
        }}
      />
    </PageContainer>
  );
};

export default ChainMonitorList;
