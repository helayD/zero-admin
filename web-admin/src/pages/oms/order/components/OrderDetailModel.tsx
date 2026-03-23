import React, { useEffect, useMemo, useState } from 'react';
import { Button, Card, message, Modal, Space, Steps } from 'antd';
import { DeleteOutlined, EditOutlined, ExclamationCircleOutlined } from '@ant-design/icons';
import type { OrderListItem } from '../data.d';
import '../index.less';
import OperationInfo from '@/pages/oms/order/components/OperationInfo';
import ProductInfo from '@/pages/oms/order/components/ProductInfo';
import ReceiveInfo from '@/pages/oms/order/components/ReceiveInfo';
import BaseInfo from '@/pages/oms/order/components/BaseInfo';
import CostInfo from '@/pages/oms/order/components/CostInfo';
import NoteOrderModel from '@/pages/oms/order/components/NoteOrderModel';
import OrderTrackingModel from '@/pages/oms/order/components/OrderTrackingModel';
import DeliveryModel from '@/pages/oms/order/components/DeliveryModel';
import { queryOrderDetail } from '@/pages/oms/order/service';
import type { GovernanceScopeValue } from '@/pages/system/components/governance';
import { buildGovernanceScopeLabel, toGovernancePayload } from '@/pages/system/components/governance';

export interface UpdateFormProps {
  onCancel: () => void;
  onRefresh: () => void;
  updateModalVisible: boolean;
  currentData: OrderListItem;
  scope: GovernanceScopeValue;
  onDeleteOrder: (values: OrderListItem) => Promise<boolean>;
  onCloseOrder: (values: OrderListItem) => Promise<boolean>;
  onDeliveryOrder: (values: OrderListItem) => Promise<boolean>;
  onNoteOrder: (values: OrderListItem) => Promise<boolean>;
}

const steps = [
  { title: '提交订单' },
  { title: '支付订单' },
  { title: '平台发货' },
  { title: '确认收货' },
  { title: '完成评价' },
];

const { confirm } = Modal;

const statusMeta = (status?: number) => {
  switch (status) {
    case 0:
      return { current: 0, message: '当前订单状态: 待付款' };
    case 1:
      return { current: 2, message: '当前订单状态: 待发货' };
    case 2:
      return { current: 3, message: '当前订单状态: 已发货' };
    case 3:
      return { current: 4, message: '当前订单状态: 已完成' };
    case 4:
      return { current: 1, message: '当前订单状态: 已关闭' };
    default:
      return { current: 0, message: '当前订单状态: 无效订单' };
  }
};

const OrderDetailModel: React.FC<UpdateFormProps> = (props) => {
  const {
    onCancel,
    onRefresh,
    updateModalVisible,
    currentData,
    scope,
    onDeleteOrder,
    onCloseOrder,
    onDeliveryOrder,
    onNoteOrder,
  } = props;

  const [detailData, setDetailData] = useState<OrderListItem>(currentData);
  const [closeOrderModelVisible, setCloseOrderModelVisible] = useState<boolean>(false);
  const [noteOrderModelVisible, setNoteOrderModelVisible] = useState<boolean>(false);
  const [deliveryModelVisible, setDeliveryModelVisible] = useState<boolean>(false);
  const [orderTrackingModalVisible, setOrderTrackingModalVisible] = useState<boolean>(false);

  useEffect(() => {
    if (!updateModalVisible || !currentData?.id) {
      return;
    }

    setDetailData(currentData);
    queryOrderDetail(currentData.id, toGovernancePayload(scope))
      .then((res) => {
        if (res?.code === '000000') {
          setDetailData(res.data);
          return;
        }
        message.error(res?.message || '获取订单详情失败');
      })
      .catch(() => {
        message.error('获取订单详情失败');
      });
  }, [updateModalVisible, currentData, scope]);

  const { current, message: statusMsg } = useMemo(() => statusMeta(detailData.status), [detailData.status]);
  const items = steps.map((item) => ({ key: item.title, title: item.title }));

  const scopeLabel = buildGovernanceScopeLabel(scope);

  const showDeleteConfirm = () => {
    confirm({
      title: '是否删除订单?',
      icon: <ExclamationCircleOutlined />,
      content: `当前主体：${scopeLabel}。删除后不可恢复，请确认。`,
      async onOk() {
        const success = await onDeleteOrder(detailData);
        if (success) {
          onRefresh();
        }
      },
    });
  };

  return (
    <>
      <Modal forceRender destroyOnClose title="订单详情" visible={updateModalVisible} onCancel={onCancel} footer={false} width={1200}>
        <Space direction="vertical" size="large" style={{ display: 'flex' }}>
          <Steps current={current} items={items} />
          <Card
            title={statusMsg}
            extra={
              <Space>
                {detailData.status === 0 && (
                  <Button
                    type="primary"
                    danger
                    icon={<DeleteOutlined />}
                    onClick={() => setCloseOrderModelVisible(true)}
                  >
                    关闭订单
                  </Button>
                )}
                {detailData.status === 4 && (
                  <Button type="primary" danger icon={<DeleteOutlined />} onClick={showDeleteConfirm}>
                    删除订单
                  </Button>
                )}
                {detailData.status === 1 && (
                  <Button
                    icon={<EditOutlined />}
                    onClick={() => setDeliveryModelVisible(true)}
                    style={{ background: '#1677ff', color: 'white' }}
                  >
                    订单发货
                  </Button>
                )}
                {(detailData.status === 2 || detailData.status === 3) && (
                  <Button type="primary" icon={<EditOutlined />} onClick={() => setOrderTrackingModalVisible(true)}>
                    订单跟踪
                  </Button>
                )}
                <Button icon={<EditOutlined />} onClick={() => setNoteOrderModelVisible(true)}>
                  备注订单
                </Button>
              </Space>
            }
          >
            <Card type="inner" title="基本信息">
              <BaseInfo currentData={detailData} />
            </Card>
            <Card style={{ marginTop: 16 }} type="inner" title="收货人信息">
              <ReceiveInfo currentData={detailData} />
            </Card>
            <Card style={{ marginTop: 16 }} type="inner" title="商品信息">
              <ProductInfo currentData={detailData.listOrderItemData} totalAmount={detailData.totalAmount || 0} />
            </Card>
            <Card style={{ marginTop: 16 }} type="inner" title="费用信息">
              <CostInfo currentData={detailData} />
            </Card>
            <Card style={{ marginTop: 16 }} type="inner" title="操作信息">
              <OperationInfo currentData={detailData.listOperateHistoryData} />
            </Card>
          </Card>
        </Space>
      </Modal>

      <NoteOrderModel
        title="关闭订单"
        submitText="关闭订单"
        confirmTitle="确认关闭订单"
        confirmHint={`当前主体：${scopeLabel}。关闭后用户将无法继续支付。`}
        visible={closeOrderModelVisible}
        currentData={detailData}
        onCancel={() => setCloseOrderModelVisible(false)}
        onSubmit={async (value) => {
          const success = await onCloseOrder(value);
          if (success) {
            setCloseOrderModelVisible(false);
            setDetailData({ ...detailData, status: 4, note: value.note });
            onRefresh();
          }
        }}
      />

      <NoteOrderModel
        title="备注订单"
        submitText="保存备注"
        confirmTitle="确认保存订单备注"
        confirmHint={`当前主体：${scopeLabel}。本次备注只会写入该主体可管理的订单。`}
        visible={noteOrderModelVisible}
        currentData={detailData}
        initialRemark={detailData.note}
        onCancel={() => setNoteOrderModelVisible(false)}
        onSubmit={async (value) => {
          const success = await onNoteOrder(value);
          if (success) {
            setNoteOrderModelVisible(false);
            setDetailData({ ...detailData, note: value.note });
            onRefresh();
          }
        }}
      />

      <OrderTrackingModel
        key="OrderTrackingModel"
        onCancel={() => {
          setOrderTrackingModalVisible(false);
        }}
        orderTrackingModalVisible={orderTrackingModalVisible}
        currentData={detailData}
      />

      <DeliveryModel
        key="DeliveryModel"
        onSubmit={async (value) => {
          const success = await onDeliveryOrder({ ...detailData, ...value });
          if (success) {
            setDetailData({
              ...detailData,
              status: 2,
              deliveryCompany: value.deliveryCompany,
              deliverySn: value.deliverySn,
            });
            setDeliveryModelVisible(false);
            onRefresh();
          }
        }}
        onCancel={() => {
          setDeliveryModelVisible(false);
        }}
        deliveryModelVisible={deliveryModelVisible}
        currentData={detailData}
      />
    </>
  );
};

export default OrderDetailModel;
