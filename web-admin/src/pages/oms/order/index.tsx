import {
  CloseOutlined,
  DeleteOutlined,
  EditOutlined,
  ExclamationCircleOutlined,
  RiseOutlined,
  StepForwardOutlined,
} from '@ant-design/icons';
import { Alert, Divider, message, Modal } from 'antd';
import React, { useRef, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import type { ActionType, ProColumns } from '@ant-design/pro-table';
import ProTable from '@ant-design/pro-table';
import OrderDetailModel from './components/OrderDetailModel';
import type { OrderListItem } from './data.d';
import { closeOrder, delivery, queryOrderList, removeOrder, updateNote } from './service';
import NoteOrderModel from '@/pages/oms/order/components/NoteOrderModel';
import DeliveryModel from '@/pages/oms/order/components/DeliveryModel';
import OrderTrackingModel from '@/pages/oms/order/components/OrderTrackingModel';
import GovernanceScopeBar from '@/pages/system/components/GovernanceScopeBar';
import {
  buildGovernanceScopeLabel,
  defaultGovernanceScope,
  type GovernanceScopeValue,
  toGovernancePayload,
} from '@/pages/system/components/governance';

const { confirm } = Modal;

type OrderMutationAction = 'close' | 'delivery' | 'note';

const handleMutation = async (
  action: OrderMutationAction,
  fields: OrderListItem,
  scope: GovernanceScopeValue,
) => {
  const hide = message.loading('正在提交订单操作');
  try {
    if (action === 'delivery') {
      await delivery({
        orderId: fields.id as number,
        deliveryCompany: fields.deliveryCompany || '',
        deliverySn: fields.deliverySn || '',
        ...toGovernancePayload(scope),
      });
    }

    if (action === 'close') {
      await closeOrder({
        ids: [fields.id as number],
        note: fields.note || '',
        ...toGovernancePayload(scope),
      });
    }

    if (action === 'note') {
      await updateNote({
        id: fields.id as number,
        status: fields.status,
        note: fields.note || '',
        ...toGovernancePayload(scope),
      });
    }

    hide();
    message.success('订单操作成功');
    return true;
  } catch (error) {
    hide();
    message.error('订单操作失败，请稍后重试');
    return false;
  }
};

const handleRemove = async (selectedRows: OrderListItem[], scope: GovernanceScopeValue) => {
  const hide = message.loading('正在删除订单');
  if (!selectedRows.length) {
    hide();
    return true;
  }
  try {
    await removeOrder(
      selectedRows.map((row) => row.id as number),
      toGovernancePayload(scope),
    );
    hide();
    message.success('删除成功，即将刷新');
    return true;
  } catch (error) {
    hide();
    message.error('删除失败，请重试');
    return false;
  }
};

const OrderList: React.FC = () => {
  const [detailVisible, setDetailVisible] = useState<boolean>(false);
  const [closeOrderModelVisible, setCloseOrderModelVisible] = useState<boolean>(false);
  const [deliveryModelVisible, setDeliveryModelVisible] = useState<boolean>(false);
  const [orderTrackingModalVisible, setOrderTrackingModalVisible] = useState<boolean>(false);
  const actionRef = useRef<ActionType>();
  const [currentRow, setCurrentRow] = useState<OrderListItem>();
  const [scope, setScope] = useState<GovernanceScopeValue>(defaultGovernanceScope);

  const scopeLabel = buildGovernanceScopeLabel(scope);

  const showDeleteConfirm = (item: OrderListItem) => {
    confirm({
      title: '是否删除记录?',
      icon: <ExclamationCircleOutlined />,
      content: `当前主体：${scopeLabel}。删除后不可恢复，请确认。`,
      onOk() {
        return handleRemove([item], scope).then(() => {
          actionRef.current?.reloadAndRest?.();
        });
      },
    });
  };

  const columns: ProColumns<OrderListItem>[] = [
    {
      title: '编号',
      dataIndex: 'id',
      hideInSearch: true,
    },
    {
      title: '订单编号',
      dataIndex: 'orderSn',
      render: (dom, entity) => {
        return (
          <a
            onClick={() => {
              setCurrentRow(entity);
              setDetailVisible(true);
            }}
          >
            {dom}
          </a>
        );
      },
    },
    {
      title: '提交时间',
      dataIndex: 'createTime',
      hideInSearch: true,
    },
    {
      title: '用户帐号',
      dataIndex: 'memberUserName',
      hideInSearch: true,
    },
    {
      title: '订单总金额',
      dataIndex: 'totalAmount',
      hideInSearch: true,
    },
    {
      title: '应付金额',
      dataIndex: 'payAmount',
      hideInSearch: true,
      hideInTable: true,
    },
    {
      title: '运费金额',
      dataIndex: 'freightAmount',
      hideInSearch: true,
      hideInTable: true,
    },
    {
      title: '促销优惠金额',
      dataIndex: 'promotionAmount',
      hideInSearch: true,
      hideInTable: true,
    },
    {
      title: '积分抵扣金额',
      dataIndex: 'integrationAmount',
      hideInSearch: true,
      hideInTable: true,
    },
    {
      title: '优惠券抵扣金额',
      dataIndex: 'couponAmount',
      hideInSearch: true,
      hideInTable: true,
    },
    {
      title: '支付方式',
      dataIndex: 'payType',
      valueEnum: {
        0: { text: '未支付', status: 'Error' },
        1: { text: '支付宝', status: 'Success' },
        2: { text: '微信', status: 'Success' },
      },
    },
    {
      title: '来源',
      dataIndex: 'sourceType',
      valueEnum: {
        0: { text: 'PC订单', status: 'Error' },
        1: { text: 'app订单', status: 'Success' },
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      valueEnum: {
        0: { text: '待付款', status: 'Success' },
        1: { text: '待发货', status: 'Success' },
        2: { text: '已发货', status: 'Success' },
        3: { text: '已完成', status: 'Success' },
        4: { text: '已关闭', status: 'Error' },
        5: { text: '无效订单', status: 'Error' },
      },
    },
    {
      title: '操作',
      dataIndex: 'option',
      valueType: 'option',
      render: (_, record) => (
        <>
          <a
            key="view"
            onClick={() => {
              setDetailVisible(true);
              setCurrentRow(record);
            }}
          >
            <EditOutlined /> 查看订单
          </a>
          <Divider type="vertical" />
          {record.status === 0 && (
            <a
              key="close"
              onClick={() => {
                setCloseOrderModelVisible(true);
                setCurrentRow(record);
              }}
            >
              <CloseOutlined /> 关闭订单
            </a>
          )}
          {record.status === 4 && (
            <>
              <Divider type="vertical" />
              <a
                key="delete"
                style={{ color: '#ff4d4f' }}
                onClick={() => {
                  showDeleteConfirm(record);
                  setCurrentRow(record);
                }}
              >
                <DeleteOutlined /> 删除订单
              </a>
            </>
          )}
          {record.status === 1 && (
            <>
              <Divider type="vertical" />
              <a
                key="delivery"
                onClick={() => {
                  setDeliveryModelVisible(true);
                  setCurrentRow(record);
                }}
              >
                <StepForwardOutlined /> 订单发货
              </a>
            </>
          )}
          {(record.status === 2 || record.status === 3) && (
            <>
              <Divider type="vertical" />
              <a
                key="tracking"
                onClick={() => {
                  setOrderTrackingModalVisible(true);
                  setCurrentRow(record);
                }}
              >
                <RiseOutlined /> 订单跟踪
              </a>
            </>
          )}
        </>
      ),
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
        entityLabel="订单"
        style={{ marginBottom: 16 }}
      />
      <Alert
        showIcon
        type="info"
        style={{ marginBottom: 16 }}
        message={`当前查询范围：${scopeLabel}`}
        description="发货、关单、删除和备注都会按当前主体写入；如果切换到其他租户或商户视角，请先确认影响范围。"
      />
      <ProTable<OrderListItem>
        headerTitle="订单列表"
        actionRef={actionRef}
        rowKey="id"
        search={{
          labelWidth: 120,
        }}
        toolBarRender={false}
        request={(params) => queryOrderList({ ...params, ...toGovernancePayload(scope) })}
        columns={columns}
        rowSelection={false}
        pagination={{ pageSize: 10 }}
      />

      <OrderDetailModel
        key="OrderDetailModel"
        onRefresh={() => {
          actionRef.current?.reload?.();
        }}
        onCancel={() => {
          setDetailVisible(false);
          setCurrentRow(undefined);
        }}
        updateModalVisible={detailVisible}
        currentData={currentRow || { id: 0 }}
        scope={scope}
        onDeleteOrder={async (value) => {
          const success = await handleRemove([value], scope);
          if (success) {
            setDetailVisible(false);
            setCurrentRow(undefined);
          }
          return success;
        }}
        onCloseOrder={(value) => handleMutation('close', value, scope)}
        onDeliveryOrder={(value) => handleMutation('delivery', value, scope)}
        onNoteOrder={(value) => handleMutation('note', value, scope)}
      />

      <NoteOrderModel
        title="关闭订单"
        submitText="关闭订单"
        confirmTitle="确认关闭订单"
        confirmHint={`当前主体：${scopeLabel}。关闭后用户将无法继续支付。`}
        visible={closeOrderModelVisible}
        currentData={currentRow || { id: 0 }}
        onCancel={() => {
          setCloseOrderModelVisible(false);
          setCurrentRow(undefined);
        }}
        onSubmit={async (value) => {
          const success = await handleMutation('close', value, scope);
          if (success) {
            setCloseOrderModelVisible(false);
            setCurrentRow(undefined);
            actionRef.current?.reload?.();
          }
        }}
      />

      <DeliveryModel
        key="DeliveryModel"
        onSubmit={async (value) => {
          const success = await handleMutation('delivery', value, scope);
          if (success) {
            setDeliveryModelVisible(false);
            setCurrentRow(undefined);
            actionRef.current?.reload?.();
          }
        }}
        onCancel={() => {
          setDeliveryModelVisible(false);
          setCurrentRow(undefined);
        }}
        deliveryModelVisible={deliveryModelVisible}
        currentData={currentRow || { id: 0 }}
      />

      <OrderTrackingModel
        key="OrderTrackingModel"
        onCancel={() => {
          setOrderTrackingModalVisible(false);
          setCurrentRow(undefined);
        }}
        orderTrackingModalVisible={orderTrackingModalVisible}
        currentData={currentRow || { id: 0 }}
      />
    </PageContainer>
  );
};

export default OrderList;
