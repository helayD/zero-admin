import React, { useRef, useState } from 'react';
import { Button, Form, Input, InputNumber, message, Modal, Select, Space, Tag } from 'antd';
import { PageContainer } from '@ant-design/pro-layout';
import type { ActionType, ProColumns } from '@ant-design/pro-table';
import ProTable from '@ant-design/pro-table';
import { PlusOutlined } from '@ant-design/icons';
import type {
  DigitalCardPhysicalFulfillmentDetailData,
  DigitalCardPhysicalFulfillmentItem,
  DigitalCardPhysicalFulfillmentListParams,
} from './data';
import {
  FULFILLMENT_STATUS_OPTIONS,
  PRODUCTION_STATUS_OPTIONS,
  SHIPPING_STATUS_OPTIONS,
} from './data';
import { buildPhysicalActionMeta, getFulfillmentStatusColor } from './helper';
import {
  ensureDigitalCardPhysicalFulfillment,
  markDigitalCardPhysicalFulfillmentException,
  queryDigitalCardPhysicalFulfillmentDetail,
  queryDigitalCardPhysicalFulfillmentList,
  requestDigitalCardPhysicalReissue,
  shipDigitalCardPhysicalFulfillment,
  updateDigitalCardPhysicalProductionStatus,
} from './service';
import type { ExceptionPayload, ProductionPayload, ShipPayload } from './service';
import GovernanceScopeBar from '@/pages/system/components/GovernanceScopeBar';
import {
  defaultGovernanceScope,
  toGovernancePayload,
  type GovernanceScopeValue,
} from '@/pages/system/components/governance';
import { readErrorMessage } from '@/pages/system/components/requestError';
import PhysicalFulfillmentDetailDrawer from './components/PhysicalFulfillmentDetailDrawer';

type ActionKind = 'ensure' | 'production' | 'ship' | 'exception' | 'reissue' | '';

const DigitalCardPhysicalFulfillmentWorkbench: React.FC = () => {
  const [form] = Form.useForm();
  const actionRef = useRef<ActionType>();
  const [scope, setScope] = useState<GovernanceScopeValue>(defaultGovernanceScope);
  const [detailOpen, setDetailOpen] = useState(false);
  const [detailLoading, setDetailLoading] = useState(false);
  const [detailData, setDetailData] = useState<DigitalCardPhysicalFulfillmentDetailData>();
  const [actionKind, setActionKind] = useState<ActionKind>('');
  const [submitting, setSubmitting] = useState(false);

  const actionMeta = buildPhysicalActionMeta(actionKind);

  const reloadList = () => {
    actionRef.current?.reload?.();
  };

  const loadDetail = async (row: DigitalCardPhysicalFulfillmentItem) => {
    setDetailOpen(true);
    setDetailLoading(true);
    try {
      const detailResp = await queryDigitalCardPhysicalFulfillmentDetail(row.fulfillmentId);
      setDetailData(detailResp?.data);
    } catch (error) {
      setDetailData(undefined);
      message.error(readErrorMessage(error, '加载履约详情失败'));
    } finally {
      setDetailLoading(false);
    }
  };

  const closeDetail = () => {
    setDetailOpen(false);
    setDetailLoading(false);
    setDetailData(undefined);
  };

  const openActionModal = (
    nextAction: Exclude<ActionKind, ''>,
    row?: DigitalCardPhysicalFulfillmentItem,
  ) => {
    setActionKind(nextAction);
    form.resetFields();
    if (nextAction === 'ensure') {
      form.setFieldsValue({ assetInstanceId: row?.assetInstanceId });
    } else {
      form.setFieldsValue({
        fulfillmentId: row?.fulfillmentId ?? detailData?.item?.fulfillmentId,
        productionBatchNo: row?.productionBatchNo ?? detailData?.item?.productionBatchNo,
      });
    }
  };

  const closeActionModal = () => {
    setActionKind('');
    form.resetFields();
  };

  const refreshAfterAction = async () => {
    reloadList();
    const latestItem = detailData?.item;
    if (latestItem?.fulfillmentId && actionKind !== 'ensure') {
      await loadDetail(latestItem);
    }
  };

  const handleSubmitAction = async () => {
    const values = await form.validateFields();
    const governancePayload = toGovernancePayload(scope);

    setSubmitting(true);
    const hide = message.loading(actionMeta.loadingMessage);
    try {
      if (actionKind === 'ensure') {
        await ensureDigitalCardPhysicalFulfillment({
          assetInstanceId: Number(values.assetInstanceId),
          ...governancePayload,
        });
      } else if (actionKind === 'production') {
        await updateDigitalCardPhysicalProductionStatus({
          fulfillmentId: Number(values.fulfillmentId),
          productionBatchNo: values.productionBatchNo,
          productionStatus: values.productionStatus,
          reason: values.reason,
          ...governancePayload,
        } as ProductionPayload);
      } else if (actionKind === 'ship') {
        await shipDigitalCardPhysicalFulfillment({
          fulfillmentId: Number(values.fulfillmentId),
          carrierCode: values.carrierCode,
          carrierName: values.carrierName,
          trackingNo: values.trackingNo,
          reason: values.reason,
          ...governancePayload,
        } as ShipPayload);
      } else if (actionKind === 'exception') {
        await markDigitalCardPhysicalFulfillmentException({
          fulfillmentId: Number(values.fulfillmentId),
          failureCode: values.failureCode,
          reason: values.reason,
          ...governancePayload,
        } as ExceptionPayload);
      } else if (actionKind === 'reissue') {
        await requestDigitalCardPhysicalReissue({
          fulfillmentId: Number(values.fulfillmentId),
          failureCode: values.failureCode,
          reason: values.reason,
          ...governancePayload,
        } as ExceptionPayload);
      }

      hide();
      message.success(actionMeta.successMessage);
      closeActionModal();
      await refreshAfterAction();
    } catch (error) {
      hide();
      message.error(readErrorMessage(error, `${actionMeta.title}失败`));
    } finally {
      setSubmitting(false);
    }
  };

  const columns: ProColumns<DigitalCardPhysicalFulfillmentItem>[] = [
    {
      title: '履约单号',
      dataIndex: 'fulfillmentNo',
      width: 180,
      ellipsis: true,
      render: (_, row) => (
        <a
          onClick={() => {
            void loadDetail(row);
          }}
        >
          {row.fulfillmentNo || '-'}
        </a>
      ),
    },
    {
      title: '资产编号',
      dataIndex: 'assetNo',
      width: 160,
      ellipsis: true,
    },
    {
      title: '会员 ID',
      dataIndex: 'memberId',
      width: 88,
    },
    {
      title: '活动 ID',
      dataIndex: 'activityId',
      width: 88,
    },
    {
      title: '活动名称',
      dataIndex: 'activityName',
      hideInSearch: true,
      width: 170,
      ellipsis: true,
    },
    {
      title: '模板 ID',
      dataIndex: 'templateId',
      width: 88,
    },
    {
      title: '模板名称',
      dataIndex: 'templateName',
      hideInSearch: true,
      width: 160,
      ellipsis: true,
    },
    {
      title: '履约状态',
      dataIndex: 'fulfillmentStatus',
      width: 130,
      valueType: 'select',
      fieldProps: { options: FULFILLMENT_STATUS_OPTIONS },
      render: (_, row) => (
        <Tag color={getFulfillmentStatusColor(row.fulfillmentStatus)}>
          {row.fulfillmentStatusText || row.fulfillmentStatus || '-'}
        </Tag>
      ),
    },
    {
      title: '制作状态',
      dataIndex: 'productionStatus',
      width: 120,
      valueType: 'select',
      fieldProps: { options: PRODUCTION_STATUS_OPTIONS },
      render: (_, row) => row.productionStatusText || row.productionStatus || '-',
    },
    {
      title: '物流状态',
      dataIndex: 'shippingStatus',
      width: 120,
      valueType: 'select',
      fieldProps: { options: SHIPPING_STATUS_OPTIONS },
      render: (_, row) => row.shippingStatusText || row.shippingStatus || '-',
    },
    {
      title: '制作批次',
      dataIndex: 'productionBatchNo',
      width: 150,
      ellipsis: true,
    },
    {
      title: '物流单号',
      dataIndex: 'trackingNo',
      width: 160,
      ellipsis: true,
    },
    {
      title: '异常码',
      dataIndex: 'failureCode',
      width: 120,
      ellipsis: true,
    },
    {
      title: '发货时间',
      dataIndex: 'shippedAt',
      hideInSearch: true,
      width: 168,
    },
    {
      title: '签收时间',
      dataIndex: 'signedAt',
      hideInSearch: true,
      width: 168,
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
    {
      title: '操作',
      valueType: 'option',
      width: 180,
      fixed: 'right',
      render: (_, row) => [
        <a
          key="detail"
          onClick={() => {
            void loadDetail(row);
          }}
        >
          详情
        </a>,
        <a key="production" onClick={() => openActionModal('production', row)}>
          制作
        </a>,
        <a key="ship" onClick={() => openActionModal('ship', row)}>
          发货
        </a>,
      ],
    },
  ];

  const renderActionFields = () => {
    if (actionKind === 'ensure') {
      return (
        <Form.Item
          name="assetInstanceId"
          label="资产实例 ID"
          rules={[{ required: true, message: '请输入资产实例 ID' }]}
        >
          <InputNumber min={1} style={{ width: '100%' }} />
        </Form.Item>
      );
    }

    return (
      <>
        <Form.Item
          name="fulfillmentId"
          label="履约单 ID"
          rules={[{ required: true, message: '请输入履约单 ID' }]}
        >
          <InputNumber min={1} style={{ width: '100%' }} />
        </Form.Item>
        {actionKind === 'production' ? (
          <>
            <Form.Item name="productionBatchNo" label="制作批次">
              <Input placeholder="例如：BATCH-20260426-01" />
            </Form.Item>
            <Form.Item
              name="productionStatus"
              label="制作状态"
              rules={[{ required: true, message: '请选择制作状态' }]}
            >
              <Select options={PRODUCTION_STATUS_OPTIONS.filter((item) => item.value !== '')} />
            </Form.Item>
          </>
        ) : null}
        {actionKind === 'ship' ? (
          <>
            <Form.Item
              name="carrierCode"
              label="承运商编码"
              rules={[{ required: true, message: '请输入承运商编码' }]}
            >
              <Input placeholder="例如：SF" />
            </Form.Item>
            <Form.Item name="carrierName" label="承运商名称">
              <Input placeholder="例如：顺丰速运" />
            </Form.Item>
            <Form.Item
              name="trackingNo"
              label="物流单号"
              rules={[{ required: true, message: '请输入物流单号' }]}
            >
              <Input />
            </Form.Item>
          </>
        ) : null}
        {actionKind === 'exception' || actionKind === 'reissue' ? (
          <Form.Item name="failureCode" label="异常码">
            <Input placeholder="例如：shipping_exception" />
          </Form.Item>
        ) : null}
        <Form.Item
          name="reason"
          label="处理原因"
          rules={[{ required: true, message: '请填写处理原因' }]}
        >
          <Input.TextArea rows={4} />
        </Form.Item>
      </>
    );
  };

  return (
    <PageContainer header={{ title: '实体卡履约工作台' }}>
      <Space direction="vertical" size={16} style={{ width: '100%' }}>
        <GovernanceScopeBar value={scope} onChange={setScope} />
        <ProTable<DigitalCardPhysicalFulfillmentItem, DigitalCardPhysicalFulfillmentListParams>
          rowKey="fulfillmentId"
          actionRef={actionRef}
          columns={columns}
          scroll={{ x: 1700 }}
          search={{ labelWidth: 96 }}
          request={async (params) => {
            const response = await queryDigitalCardPhysicalFulfillmentList({
              ...params,
              ...toGovernancePayload(scope),
            });
            return response as any;
          }}
          toolBarRender={() => [
            <Button
              key="ensure"
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => openActionModal('ensure')}
            >
              初始化履约单
            </Button>,
          ]}
          pagination={{ showSizeChanger: true }}
        />
      </Space>

      <PhysicalFulfillmentDetailDrawer
        open={detailOpen}
        loading={detailLoading}
        detail={detailData}
        onClose={closeDetail}
        onUpdateProduction={() => openActionModal('production', detailData?.item)}
        onShip={() => openActionModal('ship', detailData?.item)}
        onException={() => openActionModal('exception', detailData?.item)}
        onReissue={() => openActionModal('reissue', detailData?.item)}
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
        <Form form={form} layout="vertical" preserve={false}>
          {renderActionFields()}
        </Form>
      </Modal>
    </PageContainer>
  );
};

export default DigitalCardPhysicalFulfillmentWorkbench;
