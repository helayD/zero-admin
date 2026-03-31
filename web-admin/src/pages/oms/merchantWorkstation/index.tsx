import { Drawer, Tag } from 'antd';
import React, { useRef, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import type { ActionType, ProColumns } from '@ant-design/pro-table';
import ProTable from '@ant-design/pro-table';
import MerchantDeliveryConfirm from './components/MerchantDeliveryConfirm';
import MerchantOrderDetail from './components/MerchantOrderDetail';
import type { MerchantOrderItem } from './data';
import { queryMerchantOrderList } from './service';

const getOrderStatusColor = (status: number) => {
  switch (status) {
    case 1: return 'default';
    case 2: return 'processing';
    case 3: return 'blue';
    case 4: return 'success';
    case 5: return 'warning';
    case 6: return 'purple';
    case 7: return 'error';
    default: return 'default';
  }
};

const getReturnStatusColor = (status: number) => {
  switch (status) {
    case 0: return 'default';
    case 1: return 'warning';
    case 2: return 'processing';
    case 3: return 'error';
    default: return 'default';
  }
};

const MerchantWorkstation: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [currentRow, setCurrentRow] = useState<MerchantOrderItem>();
  const [openReturnModal, setOpenReturnModal] = useState(false);
  const [deliveryConfirmVisible, setDeliveryConfirmVisible] = useState(false);

  const columns: ProColumns<MerchantOrderItem>[] = [
    {
      title: '订单编号',
      dataIndex: 'orderNo',
      hideInSearch: true,
      render: (dom, entity) => {
        return <a onClick={() => {
          setCurrentRow(entity);
          setDrawerVisible(true);
        }}>{dom}</a>;
      },
    },
    {
      title: '订单编号',
      dataIndex: 'orderNo',
      hideInTable: true,
      paramType: 'form',
    },
    {
      title: '会员ID',
      dataIndex: 'memberId',
      hideInSearch: true,
      width: 100,
    },
    {
      title: '订单状态',
      dataIndex: 'orderStatus',
      valueType: 'select',
      valueEnum: {
        1: { text: '待支付', status: 'Default' },
        2: { text: '已支付', status: 'Processing' },
        3: { text: '已发货', status: 'Processing' },
        4: { text: '已完成', status: 'Success' },
        5: { text: '已取消', status: 'Warning' },
        6: { text: '已退款', status: 'Error' },
        7: { text: '售后中', status: 'Error' },
      },
      hideInSearch: true,
      width: 100,
      render: (_, record) => (
        <Tag color={getOrderStatusColor(record.orderStatus)}>{record.orderStatusText}</Tag>
      ),
    },
    {
      title: '订单状态',
      dataIndex: 'orderStatus',
      valueType: 'select',
      hideInTable: true,
      valueEnum: {
        1: '待支付',
        2: '已支付',
        3: '已发货',
        4: '已完成',
        5: '已取消',
        6: '已退款',
        7: '售后中',
      },
    },
    {
      title: '售后状态',
      dataIndex: 'returnStatus',
      valueType: 'select',
      valueEnum: {
        0: { text: '无售后', status: 'Default' },
        1: { text: '待审核', status: 'Warning' },
        2: { text: '审核通过', status: 'Processing' },
        3: { text: '审核拒绝', status: 'Error' },
      },
      hideInSearch: true,
      width: 100,
      render: (_, record) => (
        <Tag color={getReturnStatusColor(record.returnStatus)}>{record.returnStatusText}</Tag>
      ),
    },
    {
      title: '售后状态',
      dataIndex: 'returnStatus',
      valueType: 'select',
      hideInTable: true,
      valueEnum: {
        0: '无售后',
        1: '待审核',
        2: '审核通过',
        3: '审核拒绝',
      },
    },
    {
      title: '收货人',
      dataIndex: 'receiverName',
      hideInSearch: true,
      width: 100,
    },
    {
      title: '收货人',
      dataIndex: 'receiverName',
      hideInTable: true,
      paramType: 'form',
    },
    {
      title: '联系电话',
      dataIndex: 'receiverPhone',
      hideInSearch: true,
      width: 120,
    },
    {
      title: '联系电话',
      dataIndex: 'receiverPhone',
      hideInTable: true,
      paramType: 'form',
    },
    {
      title: '订单金额',
      dataIndex: 'payAmount',
      hideInSearch: true,
      width: 100,
      render: (_, record) => `¥${record.payAmount?.toFixed(2) ?? '0.00'}`,
    },
    {
      title: '下单时间',
      dataIndex: 'createTime',
      hideInSearch: true,
      width: 160,
      valueType: 'dateTime',
    },
    {
      title: '下单时间',
      dataIndex: 'createTime',
      hideInTable: true,
      valueType: 'dateRange',
      search: {
        transform: (value) => {
          return {
            startTime: value[0],
            endTime: value[1],
          };
        },
      },
    },
    {
      title: '操作',
      dataIndex: 'option',
      valueType: 'option',
      width: 200,
      render: (_, record) => [
        <a
          key="detail"
          onClick={() => {
            setCurrentRow(record);
            setDrawerVisible(true);
          }}
        >
          查看详情
        </a>,
        record.orderStatus === 2 && !record.deliverySn ? (
          <a
            key="delivery"
            onClick={() => {
              setCurrentRow(record);
              setDeliveryConfirmVisible(true);
            }}
          >
            确认发货
          </a>
        ) : null,
        record.returnStatus !== 0 ? (
          <a
            key="handle"
            onClick={() => {
              setCurrentRow(record);
              setOpenReturnModal(true);
              setDrawerVisible(true);
            }}
          >
            处理退货
          </a>
        ) : null,
      ],
    },
  ];

  return (
    <PageContainer>
      <ProTable<MerchantOrderItem>
        headerTitle="商户订单与售后处理"
        actionRef={actionRef}
        rowKey="id"
        search={{
          labelWidth: 80,
        }}
        request={async (params) => {
          const res = await queryMerchantOrderList({
            orderNo: params.orderNo,
            orderStatus: params.orderStatus,
            returnStatus: params.returnStatus,
            receiverName: params.receiverName,
            receiverPhone: params.receiverPhone,
            startTime: (params as any).startTime,
            endTime: (params as any).endTime,
            current: params.current,
            pageSize: params.pageSize,
          });
          return {
            data: res.data || [],
            success: res.success,
            total: res.total || 0,
          };
        }}
        columns={columns}
        pagination={{
          pageSize: 20,
          defaultPageSize: 20,
          showSizeChanger: true,
          showQuickJumper: true,
        }}
      />

      <Drawer
        title="订单详情"
        width={780}
        open={drawerVisible}
        onClose={() => {
          setDrawerVisible(false);
          setCurrentRow(undefined);
          setOpenReturnModal(false);
          setDeliveryConfirmVisible(false);
        }}
        destroyOnClose
      >
        {currentRow && (
          <MerchantOrderDetail
            order={currentRow}
            onSuccess={() => {
              actionRef.current?.reload();
            }}
            autoOpenDeliveryConfirm={false}
            autoOpenReturnModal={openReturnModal}
          />
        )}
      </Drawer>

      {currentRow && (
        <MerchantDeliveryConfirm
          order={currentRow}
          visible={deliveryConfirmVisible}
          onCancel={() => {
            setDeliveryConfirmVisible(false);
            setCurrentRow(undefined);
          }}
          onSuccess={() => {
            setDeliveryConfirmVisible(false);
            actionRef.current?.reload();
          }}
        />
      )}

      {currentRow && (
        <MerchantReturnHandler
          order={currentRow}
          onSuccess={() => {
            actionRef.current?.reload();
          }}
        />
      )}
    </PageContainer>
  );
};

export default MerchantWorkstation;
