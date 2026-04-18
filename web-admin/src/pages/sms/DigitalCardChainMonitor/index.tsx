import React, { useRef, useState } from 'react';
import { Alert, Button, Input, message, Modal, Space, Tag } from 'antd';
import { PageContainer } from '@ant-design/pro-layout';
import type { ActionType, ProColumns } from '@ant-design/pro-table';
import ProTable from '@ant-design/pro-table';
import type {
  DigitalCardChainDetailData,
  DigitalCardChainItem,
  DigitalCardChainListParams,
} from './data';
import {
  CHAIN_STATUS_OPTIONS,
  MANUAL_REQUIRED_OPTIONS,
  MINT_STATUS_OPTIONS,
  TASK_STATUS_OPTIONS,
} from './data';
import { buildActionMeta, getChainStatusColor, getTaskStatusColor } from './helper';
import {
  escalateDigitalCardChain,
  freezeDigitalCardChain,
  queryDigitalCardChainActions,
  queryDigitalCardChainDetail,
  queryDigitalCardChainList,
  retryDigitalCardChain,
} from './service';
import GovernanceScopeBar from '@/pages/system/components/GovernanceScopeBar';
import {
  buildGovernanceScopeLabel,
  defaultGovernanceScope,
  toGovernancePayload,
  type GovernanceScopeValue,
} from '@/pages/system/components/governance';
import { readErrorMessage } from '@/pages/system/components/requestError';
import ChainDetailDrawer from './components/ChainDetailDrawer';

const { TextArea } = Input;

type ActionKind = 'retry' | 'freeze' | 'escalate' | '';

const DigitalCardChainMonitor: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const [scope, setScope] = useState<GovernanceScopeValue>(defaultGovernanceScope);
  const [detailOpen, setDetailOpen] = useState(false);
  const [detailLoading, setDetailLoading] = useState(false);
  const [detailData, setDetailData] = useState<DigitalCardChainDetailData>();
  const [availableActions, setAvailableActions] = useState<string[]>([]);
  const [actionKind, setActionKind] = useState<ActionKind>('');
  const [actionReason, setActionReason] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const scopeLabel = buildGovernanceScopeLabel(scope);
  const actionMeta = buildActionMeta(actionKind);

  const reloadList = () => {
    actionRef.current?.reload?.();
  };

  const loadDetail = async (row: DigitalCardChainItem) => {
    setDetailOpen(true);
    setDetailLoading(true);
    try {
      const [detailResp, actionsResp] = await Promise.all([
        queryDigitalCardChainDetail(row.taskId),
        queryDigitalCardChainActions(row.taskId),
      ]);
      setDetailData(detailResp?.data);
      setAvailableActions(actionsResp?.availableActions || []);
    } catch (error) {
      setDetailData(undefined);
      setAvailableActions([]);
      message.error(readErrorMessage(error, '加载链路详情失败'));
    } finally {
      setDetailLoading(false);
    }
  };

  const closeDetail = () => {
    setDetailOpen(false);
    setDetailLoading(false);
    setDetailData(undefined);
    setAvailableActions([]);
  };

  const openActionModal = (nextAction: Exclude<ActionKind, ''>) => {
    setActionKind(nextAction);
    setActionReason('');
  };

  const closeActionModal = () => {
    setActionKind('');
    setActionReason('');
  };

  const handleSubmitAction = async () => {
    if (!detailData?.item?.taskId) {
      message.warning('请先选择一条链路');
      return;
    }
    if (!actionReason.trim()) {
      message.warning('请填写处理原因');
      return;
    }

    const payload = {
      taskId: detailData.item.taskId,
      reason: actionReason.trim(),
      ...toGovernancePayload(scope),
    };

    setSubmitting(true);
    const hide = message.loading(actionMeta.loadingMessage);
    try {
      if (actionKind === 'retry') {
        await retryDigitalCardChain(payload);
      } else if (actionKind === 'freeze') {
        await freezeDigitalCardChain(payload);
      } else if (actionKind === 'escalate') {
        await escalateDigitalCardChain(payload);
      }
      hide();
      message.success(actionMeta.successMessage);
      closeActionModal();
      reloadList();
      await loadDetail(detailData.item);
    } catch (error) {
      hide();
      message.error(readErrorMessage(error, `${actionMeta.title}失败`));
    } finally {
      setSubmitting(false);
    }
  };

  const columns: ProColumns<DigitalCardChainItem>[] = [
    {
      title: '任务 ID',
      dataIndex: 'taskId',
      width: 88,
      hideInSearch: true,
    },
    {
      title: '资产编号',
      dataIndex: 'assetNo',
      width: 180,
      ellipsis: true,
      render: (_, row) => (
        <a
          onClick={() => {
            void loadDetail(row);
          }}
        >
          {row.assetNo || '-'}
        </a>
      ),
    },
    {
      title: '活动 ID',
      dataIndex: 'activityId',
      width: 90,
    },
    {
      title: '活动名称',
      dataIndex: 'activityName',
      ellipsis: true,
      width: 180,
    },
    {
      title: '会员 ID',
      dataIndex: 'memberId',
      width: 90,
    },
    {
      title: '模板 ID',
      dataIndex: 'templateId',
      width: 90,
    },
    {
      title: 'tokenId',
      dataIndex: 'tokenId',
      width: 180,
      ellipsis: true,
    },
    {
      title: '任务状态',
      dataIndex: 'taskStatus',
      width: 120,
      valueType: 'select',
      fieldProps: {
        options: TASK_STATUS_OPTIONS,
      },
      render: (_, row) => (
        <Tag color={getTaskStatusColor(row.taskStatus)}>
          {row.taskStatusText || row.taskStatus || '-'}
        </Tag>
      ),
    },
    {
      title: '发放状态',
      dataIndex: 'mintStatus',
      width: 130,
      valueType: 'select',
      fieldProps: {
        options: MINT_STATUS_OPTIONS,
      },
      render: (_, row) => row.mintStatusText || row.mintStatus || '-',
    },
    {
      title: '链上状态',
      dataIndex: 'chainStatus',
      width: 120,
      valueType: 'select',
      fieldProps: {
        options: CHAIN_STATUS_OPTIONS,
      },
      render: (_, row) => (
        <Tag color={getChainStatusColor(row.chainStatus)}>
          {row.chainStatusText || row.chainStatus || '-'}
        </Tag>
      ),
    },
    {
      title: '人工复核',
      dataIndex: 'manualRequired',
      width: 110,
      valueType: 'select',
      fieldProps: {
        options: MANUAL_REQUIRED_OPTIONS,
      },
      hideInTable: true,
    },
    {
      title: '人工复核',
      dataIndex: 'manualRequiredDisplay',
      hideInSearch: true,
      width: 100,
      render: (_, row) =>
        row.manualRequired ? <Tag color="gold">人工复核</Tag> : <Tag color="green">自动链路</Tag>,
    },
    {
      title: '当前资产阶段',
      dataIndex: 'assetStatusText',
      hideInSearch: true,
      width: 170,
    },
    {
      title: '重试次数',
      dataIndex: 'retryCount',
      hideInSearch: true,
      width: 88,
    },
    {
      title: '最近执行时间',
      dataIndex: 'lastExecuteAt',
      hideInSearch: true,
      width: 168,
    },
    {
      title: '最近错误',
      dataIndex: 'lastError',
      hideInSearch: true,
      ellipsis: true,
      width: 240,
      render: (_, row) =>
        row.lastError ? <span style={{ color: '#ff4d4f' }}>{row.lastError}</span> : '-',
    },
    {
      title: '时间范围',
      dataIndex: 'timeRange',
      hideInTable: true,
      valueType: 'dateTimeRange',
      search: {
        transform: (value: [string, string]) => ({
          startTime: value?.[0],
          endTime: value?.[1],
        }),
      },
    },
    {
      title: '操作',
      dataIndex: 'option',
      valueType: 'option',
      width: 96,
      render: (_, row) => [
        <a
          key="detail"
          onClick={() => {
            void loadDetail(row);
          }}
        >
          详情
        </a>,
      ],
    },
  ];

  return (
    <PageContainer>
      <Space direction="vertical" size={16} style={{ width: '100%' }}>
        <GovernanceScopeBar
          value={scope}
          onChange={setScope}
          entityLabel="数字卡片链路与补偿动作"
        />

        <Alert
          showIcon
          type="info"
          message={`当前工作台基于 ${scopeLabel} 读取 sms_card_mint_task，并与 sms_card_instance 的资产快照联动展示。`}
          description="重试、冻结、升级人工复核都只会作用于当前治理范围内的任务，不会跨租户或跨商户串写。"
        />

        <ProTable<DigitalCardChainItem, DigitalCardChainListParams>
          headerTitle="数字卡片链路工作台"
          rowKey="taskId"
          actionRef={actionRef}
          search={{
            labelWidth: 110,
            defaultCollapsed: false,
          }}
          toolBarRender={() => [
            <Button key="reload" onClick={reloadList}>
              刷新
            </Button>,
          ]}
          request={async (params) => {
            const response = await queryDigitalCardChainList({
              ...params,
              ...toGovernancePayload(scope),
              activityId: params.activityId ? Number(params.activityId) : undefined,
              memberId: params.memberId ? Number(params.memberId) : undefined,
              templateId: params.templateId ? Number(params.templateId) : undefined,
              manualRequired:
                params.manualRequired !== undefined ? Number(params.manualRequired) : 0,
            });

            return {
              data: response.data || [],
              success: response.success ?? true,
              total: response.total || 0,
            };
          }}
          columns={columns}
          pagination={{
            defaultPageSize: 20,
            showSizeChanger: true,
          }}
        />
      </Space>

      <ChainDetailDrawer
        open={detailOpen}
        loading={detailLoading}
        detail={detailData}
        availableActions={availableActions}
        onClose={closeDetail}
        onRetry={() => openActionModal('retry')}
        onFreeze={() => openActionModal('freeze')}
        onEscalate={() => openActionModal('escalate')}
      />

      <Modal
        title={actionMeta.title}
        open={actionKind !== ''}
        onCancel={closeActionModal}
        onOk={() => {
          void handleSubmitAction();
        }}
        confirmLoading={submitting}
      >
        <Space direction="vertical" size={12} style={{ width: '100%' }}>
          <Alert
            showIcon
            type="warning"
            message={`当前操作主体：${scopeLabel}`}
            description="操作原因会进入链路日志与后台审计，请填写能帮助后续追溯的说明。"
          />
          <TextArea
            rows={4}
            value={actionReason}
            placeholder={actionMeta.placeholder}
            onChange={(event) => setActionReason(event.target.value)}
          />
        </Space>
      </Modal>
    </PageContainer>
  );
};

export default DigitalCardChainMonitor;
