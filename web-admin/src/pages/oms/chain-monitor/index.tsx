import React, { useRef, useState } from 'react';
import {
  Alert,
  Button,
  Input,
  message,
  Modal,
  Space,
} from 'antd';
import { DownloadOutlined } from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-layout';
import type { ActionType, ProColumns } from '@ant-design/pro-table';
import ProTable from '@ant-design/pro-table';
import type { ChainMonitorItem } from './data.d';
import {
  queryChainMonitorList,
  exportChainMonitorList,
  retryChain,
  replayChain,
  pauseChain,
  escalateChain,
  queryChainActions,
} from './service';
import ChainDetailDrawer from './components/ChainDetailDrawer';
import GovernanceScopeBar from '@/pages/system/components/GovernanceScopeBar';
import {
  buildGovernanceScopeLabel,
  defaultGovernanceScope,
  toGovernancePayload,
  type GovernanceScopeValue,
} from '@/pages/system/components/governance';

const { TextArea } = Input;

const ChainMonitorList: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const [scope, setScope] = useState<GovernanceScopeValue>(defaultGovernanceScope);
  const [detailVisible, setDetailVisible] = useState<boolean>(false);
  const [currentRow, setCurrentRow] = useState<ChainMonitorItem>();
  const [availableActions, setAvailableActions] = useState<string[]>([]);

  // 各干预动作的确认弹窗
  const [retryModalVisible, setRetryModalVisible] = useState(false);
  const [retryRemark, setRetryRemark] = useState('');
  const [replayModalVisible, setReplayModalVisible] = useState(false);
  const [replayReason, setReplayReason] = useState('');
  const [pauseModalVisible, setPauseModalVisible] = useState(false);
  const [pauseReason, setPauseReason] = useState('');
  const [escalateModalVisible, setEscalateModalVisible] = useState(false);
  const [escalateReason, setEscalateReason] = useState('');

  const scopeLabel = buildGovernanceScopeLabel(scope);

  // 打开详情抽屉时查询可用动作
  const handleOpenDetail = async (row: ChainMonitorItem) => {
    setCurrentRow(row);
    setDetailVisible(true);
    try {
      const res = await queryChainActions(row.entityId);
      if (res.success && res.availableActions) {
        setAvailableActions(res.availableActions);
      } else {
        setAvailableActions([]);
      }
    } catch {
      setAvailableActions([]);
    }
  };

  const handleCloseDetail = () => {
    setDetailVisible(false);
    setCurrentRow(undefined);
    setAvailableActions([]);
  };

  // 刷新列表
  const reloadList = () => {
    actionRef.current?.reload?.();
  };

  // 重试
  const handleRetry = async () => {
    if (!currentRow) return;
    setRetryModalVisible(false);
    const hide = message.loading('正在重试链路...');
    try {
      const res = await retryChain(currentRow.entityId, retryRemark);
      hide();
      if (res.success) {
        message.success(`链路重试成功，当前重试次数：${res.newRetryCount}`);
        setRetryRemark('');
        handleCloseDetail();
        reloadList();
      } else {
        message.error(res.message || '重试失败');
      }
    } catch {
      hide();
      message.error('重试失败，请稍后重试');
    }
  };

  // 回放
  const handleReplay = async () => {
    if (!currentRow || !replayReason.trim()) {
      message.warning('请填写回放原因');
      return;
    }
    setReplayModalVisible(false);
    const hide = message.loading('正在回放链路...');
    try {
      const res = await replayChain(currentRow.entityId, replayReason);
      hide();
      if (res.success) {
        message.success('链路回放成功');
        setReplayReason('');
        handleCloseDetail();
        reloadList();
      } else {
        message.error(res.message || '回放失败');
      }
    } catch {
      hide();
      message.error('回放失败，请稍后重试');
    }
  };

  // 暂停
  const handlePause = async () => {
    if (!currentRow || !pauseReason.trim()) {
      message.warning('请填写暂停原因');
      return;
    }
    setPauseModalVisible(false);
    const hide = message.loading('正在暂停链路...');
    try {
      const res = await pauseChain(currentRow.entityId, pauseReason);
      hide();
      if (res.success) {
        message.success('链路已暂停');
        setPauseReason('');
        handleCloseDetail();
        reloadList();
      } else {
        message.error(res.message || '暂停失败');
      }
    } catch {
      hide();
      message.error('暂停失败，请稍后重试');
    }
  };

  // 升级
  const handleEscalate = async () => {
    if (!currentRow || !escalateReason.trim()) {
      message.warning('请填写升级原因');
      return;
    }
    setEscalateModalVisible(false);
    const hide = message.loading('正在升级链路...');
    try {
      const res = await escalateChain(currentRow.entityId, escalateReason);
      hide();
      if (res.success) {
        message.success('链路已升级为需人工介入');
        setEscalateReason('');
        handleCloseDetail();
        reloadList();
      } else {
        message.error(res.message || '升级失败');
      }
    } catch {
      hide();
      message.error('升级失败，请稍后重试');
    }
  };

  const handleExport = async () => {
    const hide = message.loading('正在导出...');
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
      message.success('导出成功');
    } catch {
      hide();
      message.error('导出失败');
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
            handleOpenDetail(entity);
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
      title: '暂停',
      dataIndex: 'paused',
      hideInSearch: true,
      width: 70,
      render: (_, entity) => {
        if (entity.paused === 1) {
          return <span style={{ color: '#faad14' }}>已暂停</span>;
        }
        return '-';
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
        description="按租户/商户筛选仅返回该主体下的链路数据；点击业务编号可查看链路详情和干预操作。"
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
        availableActions={availableActions}
        onClose={handleCloseDetail}
        onRetry={(orderId) => {
          setCurrentRow({ ...currentRow!, entityId: orderId } as ChainMonitorItem);
          setRetryModalVisible(true);
        }}
        onReplay={(orderId) => {
          setCurrentRow({ ...currentRow!, entityId: orderId } as ChainMonitorItem);
          setReplayModalVisible(true);
        }}
        onPause={(orderId) => {
          setCurrentRow({ ...currentRow!, entityId: orderId } as ChainMonitorItem);
          setPauseModalVisible(true);
        }}
        onEscalate={(orderId) => {
          setCurrentRow({ ...currentRow!, entityId: orderId } as ChainMonitorItem);
          setEscalateModalVisible(true);
        }}
      />

      {/* 重试确认弹窗 */}
      <Modal
        title="确认重试链路"
        open={retryModalVisible}
        onOk={handleRetry}
        onCancel={() => {
          setRetryModalVisible(false);
          setRetryRemark('');
        }}
        okText="确认重试"
        cancelText="取消"
      >
        <p>
          确定重试此链路？
          {currentRow && (
            <>
              <br />
              当前重试次数：<strong>{currentRow.retryCount}</strong>
            </>
          )}
        </p>
        <div style={{ marginTop: 12 }}>
          <span>备注（可选）：</span>
          <TextArea
            rows={2}
            placeholder="请输入重试备注"
            value={retryRemark}
            onChange={(e) => setRetryRemark(e.target.value)}
          />
        </div>
      </Modal>

      {/* 回放确认弹窗 */}
      <Modal
        title="确认回放链路"
        open={replayModalVisible}
        onOk={handleReplay}
        onCancel={() => {
          setReplayModalVisible(false);
          setReplayReason('');
        }}
        okText="确认回放"
        cancelText="取消"
      >
        <p>确定回放此链路？回放将重新执行完整补偿链路。</p>
        <div style={{ marginTop: 12 }}>
          <span style={{ color: '#ff4d4f' }}>* </span>
          <span>回放原因（必填）：</span>
          <TextArea
            rows={2}
            placeholder="请输入回放原因"
            value={replayReason}
            onChange={(e) => setReplayReason(e.target.value)}
          />
        </div>
      </Modal>

      {/* 暂停确认弹窗 */}
      <Modal
        title="确认暂停链路"
        open={pauseModalVisible}
        onOk={handlePause}
        onCancel={() => {
          setPauseModalVisible(false);
          setPauseReason('');
        }}
        okText="确认暂停"
        cancelText="取消"
      >
        <p>确定暂停此链路？暂停后 Job 将跳过此链路。</p>
        <div style={{ marginTop: 12 }}>
          <span style={{ color: '#ff4d4f' }}>* </span>
          <span>暂停原因（必填）：</span>
          <TextArea
            rows={2}
            placeholder="请输入暂停原因"
            value={pauseReason}
            onChange={(e) => setPauseReason(e.target.value)}
          />
        </div>
      </Modal>

      {/* 升级确认弹窗 */}
      <Modal
        title="确认升级链路"
        open={escalateModalVisible}
        onOk={handleEscalate}
        onCancel={() => {
          setEscalateModalVisible(false);
          setEscalateReason('');
        }}
        okText="确认升级"
        cancelText="取消"
      >
        <p>确定升级此链路为需人工介入？升级后链路将进入人工处理队列。</p>
        <div style={{ marginTop: 12 }}>
          <span style={{ color: '#ff4d4f' }}>* </span>
          <span>升级原因（必填）：</span>
          <TextArea
            rows={2}
            placeholder="请输入升级原因"
            value={escalateReason}
            onChange={(e) => setEscalateReason(e.target.value)}
          />
        </div>
      </Modal>
    </PageContainer>
  );
};

export default ChainMonitorList;
