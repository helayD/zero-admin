import React, { useRef, useState } from 'react';
import { Alert, Input, message, Modal, Space, Tag } from 'antd';
import { history } from 'umi';
import { PageContainer } from '@ant-design/pro-layout';
import type { ActionType, ProColumns } from '@ant-design/pro-table';
import ProTable from '@ant-design/pro-table';
import type { DigitalCardAssetDetailData, DigitalCardAssetItem, DigitalCardAssetListParams } from './data';
import {
  CHAIN_STATUS_OPTIONS,
  COMPLIANCE_STATUS_OPTIONS,
  DISPLAY_STATUS_OPTIONS,
  MINT_STATUS_OPTIONS,
} from './data';
import {
  buildAssetActionMeta,
  getChainStatusColor,
  getComplianceStatusColor,
  getDisplayStatusColor,
} from './helper';
import {
  offlineDigitalCardAssetDisplay,
  queryDigitalCardAssetDetail,
  queryDigitalCardAssetList,
  recycleDigitalCardAsset,
  reviewDigitalCardAssetCompliance,
} from './service';
import GovernanceScopeBar from '@/pages/system/components/GovernanceScopeBar';
import {
  buildGovernanceScopeLabel,
  defaultGovernanceScope,
  toGovernancePayload,
  type GovernanceScopeValue,
} from '@/pages/system/components/governance';
import { readErrorMessage } from '@/pages/system/components/requestError';
import AssetDetailDrawer from './components/AssetDetailDrawer';

const { TextArea } = Input;

type ActionKind = 'review' | 'offline' | 'recycle' | '';

const DigitalCardAssetWorkbench: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const [scope, setScope] = useState<GovernanceScopeValue>(defaultGovernanceScope);
  const [detailOpen, setDetailOpen] = useState(false);
  const [detailLoading, setDetailLoading] = useState(false);
  const [detailData, setDetailData] = useState<DigitalCardAssetDetailData>();
  const [actionKind, setActionKind] = useState<ActionKind>('');
  const [actionReason, setActionReason] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const scopeLabel = buildGovernanceScopeLabel(scope);
  const actionMeta = buildAssetActionMeta(actionKind);

  const reloadList = () => {
    actionRef.current?.reload?.();
  };

  const loadDetail = async (row: DigitalCardAssetItem) => {
    setDetailOpen(true);
    setDetailLoading(true);
    try {
      const detailResp = await queryDigitalCardAssetDetail(row.assetInstanceId);
      setDetailData(detailResp?.data);
    } catch (error) {
      setDetailData(undefined);
      message.error(readErrorMessage(error, '加载资产详情失败'));
    } finally {
      setDetailLoading(false);
    }
  };

  const closeDetail = () => {
    setDetailOpen(false);
    setDetailLoading(false);
    setDetailData(undefined);
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
    if (!detailData?.item?.assetInstanceId) {
      message.warning('请先选择一条资产');
      return;
    }
    if (!actionReason.trim()) {
      message.warning('请填写处理原因');
      return;
    }

    const payload = {
      assetInstanceId: detailData.item.assetInstanceId,
      reason: actionReason.trim(),
      ...toGovernancePayload(scope),
    };

    setSubmitting(true);
    const hide = message.loading(actionMeta.loadingMessage);
    try {
      if (actionKind === 'review') {
        await reviewDigitalCardAssetCompliance(payload);
      } else if (actionKind === 'offline') {
        await offlineDigitalCardAssetDisplay(payload);
      } else if (actionKind === 'recycle') {
        await recycleDigitalCardAsset(payload);
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

  const openChainMonitor = () => {
    const taskId = detailData?.mintTaskSummary?.taskId;
    if (!taskId) {
      message.info('当前资产没有可关联的发放任务');
      return;
    }
    history.push('/sms/digitalCardChain/list');
    message.info(`已打开数字卡片链路页，请按任务 ID ${taskId} 继续处理 retry / freeze / escalate`);
  };

  const columns: ProColumns<DigitalCardAssetItem>[] = [
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
      width: 88,
    },
    {
      title: '活动名称',
      dataIndex: 'activityName',
      width: 170,
      ellipsis: true,
    },
    {
      title: '会员 ID',
      dataIndex: 'memberId',
      width: 88,
    },
    {
      title: '模板 ID',
      dataIndex: 'templateId',
      width: 88,
    },
    {
      title: '模板名称',
      dataIndex: 'templateName',
      width: 160,
      ellipsis: true,
    },
    {
      title: 'tokenId',
      dataIndex: 'tokenId',
      width: 180,
      ellipsis: true,
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
      title: '展示状态',
      dataIndex: 'displayStatus',
      width: 120,
      valueType: 'select',
      fieldProps: {
        options: DISPLAY_STATUS_OPTIONS,
      },
      render: (_, row) => (
        <Tag color={getDisplayStatusColor(row.displayStatus)}>
          {row.displayStatusText || row.displayStatus || '-'}
        </Tag>
      ),
    },
    {
      title: '合规状态',
      dataIndex: 'complianceStatus',
      width: 120,
      valueType: 'select',
      fieldProps: {
        options: COMPLIANCE_STATUS_OPTIONS,
      },
      render: (_, row) => (
        <Tag color={getComplianceStatusColor(row.complianceStatus)}>
          {row.complianceStatusText || row.complianceStatus || '-'}
        </Tag>
      ),
    },
    {
      title: '获取时间',
      dataIndex: 'obtainedAt',
      hideInSearch: true,
      width: 168,
    },
    {
      title: '最近原因摘要',
      dataIndex: 'latestReasonSummary',
      hideInSearch: true,
      width: 220,
      ellipsis: true,
    },
    {
      title: '开始时间',
      dataIndex: 'startTime',
      valueType: 'dateTime',
      hideInTable: true,
    },
    {
      title: '结束时间',
      dataIndex: 'endTime',
      valueType: 'dateTime',
      hideInTable: true,
    },
  ];

  return (
    <PageContainer
      header={{
        title: '数字卡片资产工作台',
        subTitle: '围绕资产查询、合规处置与链路协同的统一入口',
      }}
    >
      <Space direction="vertical" size={16} style={{ width: '100%' }}>
        <GovernanceScopeBar value={scope} onChange={setScope} />
        <Alert
          type="info"
          showIcon
          message={`当前查询范围：${scopeLabel}`}
          description="资产级动作只改变展示与合规状态；发放异常请前往数字卡片链路页复用 retry / freeze / escalate。"
        />

        <ProTable<DigitalCardAssetItem, DigitalCardAssetListParams>
          rowKey="assetInstanceId"
          actionRef={actionRef}
          columns={columns}
          scroll={{ x: 1600 }}
          search={{ labelWidth: 96 }}
          request={async (params) => {
            const response = await queryDigitalCardAssetList({
              ...params,
              ...toGovernancePayload(scope),
            });
            return response as any;
          }}
          toolBarRender={false}
          pagination={{
            showSizeChanger: true,
          }}
        />
      </Space>

      <AssetDetailDrawer
        open={detailOpen}
        loading={detailLoading}
        detail={detailData}
        onClose={closeDetail}
        onReview={() => openActionModal('review')}
        onOffline={() => openActionModal('offline')}
        onRecycle={() => openActionModal('recycle')}
        onOpenChainMonitor={openChainMonitor}
      />

      <Modal
        title={actionMeta.title}
        open={actionKind !== ''}
        onCancel={closeActionModal}
        onOk={() => {
          void handleSubmitAction();
        }}
        confirmLoading={submitting}
        destroyOnClose
      >
        <Space direction="vertical" size={12} style={{ width: '100%' }}>
          <Alert
            type={actionKind === 'recycle' ? 'warning' : 'info'}
            showIcon
            message={`作用范围：${scopeLabel}`}
            description="该动作会落库到资产实例并写入 sms_card_asset_log，不会改变资产归属真相。"
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

export default DigitalCardAssetWorkbench;
