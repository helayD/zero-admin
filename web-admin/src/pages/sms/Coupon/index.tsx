import {
  PlusOutlined,
  ExclamationCircleOutlined,
  DeleteOutlined,
  EditOutlined,
  EyeOutlined,
  CheckCircleOutlined,
  StopOutlined,
} from '@ant-design/icons';
import {Alert, Button, Divider, message, Drawer, Modal, Tag, Tooltip} from 'antd';
import React, {useState, useRef} from 'react';
import {PageContainer, FooterToolbar} from '@ant-design/pro-layout';
import ProTable from '@ant-design/pro-table';
import type {ProColumns, ActionType} from '@ant-design/pro-table';
import ProDescriptions from '@ant-design/pro-descriptions';
import type {ProDescriptionsItemProps} from '@ant-design/pro-descriptions';
import CreateCouponForm from './components/CreateCouponForm';
import UpdateCouponForm from './components/UpdateCouponForm';
import type {CouponListItem} from './data.d';
import {queryCoupon, updateCoupon, addCoupon, removeCoupon, updateCouponStatus} from './service';
import moment from "moment";
import CouponDetailForm from "@/pages/sms/Coupon/components/CouponDetailForm";
import GovernanceScopeBar from '@/pages/system/components/GovernanceScopeBar';
import {
  buildGovernanceScopeLabel,
  defaultGovernanceScope,
  type GovernanceScopeValue,
  toGovernancePayload,
} from '@/pages/system/components/governance';

const {confirm} = Modal;

/**
 * 添加节点
 * @param fields
 */
const handleAdd = async (fields: CouponListItem, scope: GovernanceScopeValue) => {
  const hide = message.loading('正在添加');
  try {
    await addCoupon({...fields, ...toGovernancePayload(scope)});
    hide();
    message.success('添加成功');
    return true;
  } catch (error) {
    hide();
    return false;
  }
};

/**
 * 更新节点
 * @param fields
 */
const handleUpdate = async (fields: CouponListItem, scope: GovernanceScopeValue) => {
  const hide = message.loading('正在更新');
  try {
    await updateCoupon({...fields, ...toGovernancePayload(scope)});
    hide();

    message.success('更新成功');
    return true;
  } catch (error) {
    hide();
    return false;
  }
};

/**
 *  删除节点
 * @param selectedRows
 */
const handleRemove = async (selectedRows: CouponListItem[], scope: GovernanceScopeValue) => {
  const hide = message.loading('正在删除');
  if (!selectedRows) return true;
  try {
    await removeCoupon({
      ids: selectedRows.map((row) => row.id as number),
      ...toGovernancePayload(scope),
    });
    hide();
    message.success('删除成功，即将刷新');
    return true;
  } catch (error) {
    hide();
    return false;
  }
};

const CouponList: React.FC = () => {
  const [createModalVisible, handleModalVisible] = useState<boolean>(false);
  const [updateModalVisible, handleUpdateModalVisible] = useState<boolean>(false);
  const [detailModalVisible, handleDetailModalVisible] = useState<boolean>(false);
  const [showDetail, setShowDetail] = useState<boolean>(false);
  const actionRef = useRef<ActionType>();
  const [currentRow, setCurrentRow] = useState<CouponListItem>();
  const [selectedRowsState, setSelectedRows] = useState<CouponListItem[]>([]);
  const [scope, setScope] = useState<GovernanceScopeValue>(defaultGovernanceScope);

  const showDeleteConfirm = (item: CouponListItem) => {
    confirm({
      title: '是否删除记录?',
      icon: <ExclamationCircleOutlined/>,
      content: `当前主体：${buildGovernanceScopeLabel(scope)}。删除后不可恢复，请确认。`,
      onOk() {
        handleRemove([item], scope).then(() => {
          actionRef.current?.reloadAndRest?.();
        });
      },
      onCancel() {
      },
    });
  };

  const columns: ProColumns<CouponListItem>[] = [
    {
      title: '编号',
      dataIndex: 'id',
      hideInSearch: true,
      width: 60,
    },
    {
      title: '优惠券名称',
      dataIndex: 'name',
      render: (dom, entity) => {
        return (
          <a
            onClick={() => {
              setCurrentRow(entity);
              setShowDetail(true);
            }}
          >
            {dom}
          </a>
        );
      },
    },
    {
      title: '优惠券类型',
      dataIndex: 'typeName',
      hideInSearch: true,
      render: (_, record) => record.typeName || '-',
    },
    {
      title: '优惠券码',
      dataIndex: 'code',
      hideInSearch: true,
    },
    {
      title: '面值',
      dataIndex: 'amount',
      hideInSearch: true,
      render: (_, record) => `¥${record.amount || 0}`,
    },
    {
      title: '使用门槛',
      dataIndex: 'minAmount',
      hideInSearch: true,
      render: (_, record) => record.minAmount ? `满¥${record.minAmount}` : '无门槛',
    },
    {
      title: '发放总量',
      dataIndex: 'totalCount',
      hideInSearch: true,
    },
    {
      title: '已领/已用',
      hideInSearch: true,
      render: (_, record) => `${record.receivedCount || 0}/${record.usedCount || 0}`,
    },
    {
      title: '每人限领',
      dataIndex: 'perLimit',
      hideInSearch: true,
    },
    {
      title: '有效期',
      hideInSearch: true,
      render: (_, entity) => {
        return (
          <>
            {moment(entity.startTime).format('YYYY-MM-DD')}
            {' ~ '}
            {moment(entity.endTime).format('YYYY-MM-DD')}
          </>
        );
      },
    },
    {
      title: '适用范围',
      dataIndex: 'scopeCount',
      hideInSearch: true,
      render: (_, record) => {
        if (!record.scopeCount || record.scopeCount === 0) {
          return <span style={{color: '#ff4d4f'}}>未配置</span>;
        }
        return `${record.scopeCount}条`;
      },
    },
    {
      title: '生效状态',
      dataIndex: 'effectiveStatus',
      hideInSearch: true,
      render: (_, record) => {
        const colorMap: Record<string, string> = {
          '生效中': 'green',
          '未开始': 'blue',
          '已过期': 'orange',
          '已结束': 'default',
          '已取消': 'red',
        };
        const color = colorMap[record.effectiveStatus || ''] || 'default';
        return <Tag color={color}>{record.effectiveStatus || '-'}</Tag>;
      },
    },
    {
      title: '适用范围摘要',
      dataIndex: 'scopeSummary',
      hideInSearch: true,
      render: (_, record) => {
        if (!record.scopeSummary) return '-';
        return <Tooltip title={record.affectedPaths || '无影响链路'}>{record.scopeSummary}</Tooltip>;
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      valueEnum: {
        0: {text: '草稿', status: 'Default'},
        1: {text: '进行中', status: 'Processing'},
        2: {text: '已结束', status: 'Default'},
        3: {text: '已取消', status: 'Error'},
      },
    },
    {
      title: '启用',
      dataIndex: 'isEnabled',
      hideInSearch: true,
      valueEnum: {
        0: {text: '停用', status: 'Error'},
        1: {text: '启用', status: 'Success'},
      },
    },
    {
      title: '操作',
      dataIndex: 'option',
      valueType: 'option',
      width: 280,
      render: (_, record) => (
        <>
          <a
            key="eye"
            onClick={() => {
              handleDetailModalVisible(true);
              setCurrentRow(record);
            }}
          >
            <EyeOutlined/> 查看
          </a>
          <Divider type="vertical"/>
          <a
            key="edit"
            onClick={() => {
              handleUpdateModalVisible(true);
              setCurrentRow(record);
            }}
          >
            <EditOutlined/> 编辑
          </a>
          {record.status === 0 && (
            <>
              <Divider type="vertical"/>
              <a
                key="publish"
                style={{color: '#52c41a'}}
                onClick={() => {
                  confirm({
                    title: '确认发布此优惠券？',
                    icon: <ExclamationCircleOutlined/>,
                    content: `发布后核心字段将不可修改。当前主体：${buildGovernanceScopeLabel(scope)}`,
                    onOk: async () => {
                      try {
                        await updateCouponStatus({
                          ids: [record.id as number],
                          status: 1,
                          ...toGovernancePayload(scope),
                        });
                        message.success('发布成功');
                        actionRef.current?.reload?.();
                      } catch (e) {
                        // 错误已由 request 拦截器处理
                      }
                    },
                  });
                }}
              >
                <CheckCircleOutlined/> 发布
              </a>
            </>
          )}
          {record.status === 1 && (
            <>
              <Divider type="vertical"/>
              <a
                key="cancel"
                style={{color: '#faad14'}}
                onClick={() => {
                  confirm({
                    title: '确认取消此优惠券？',
                    icon: <ExclamationCircleOutlined/>,
                    content: '取消后优惠券将不再可用。',
                    onOk: async () => {
                      try {
                        await updateCouponStatus({
                          ids: [record.id as number],
                          status: 3,
                          ...toGovernancePayload(scope),
                        });
                        message.success('已取消');
                        actionRef.current?.reload?.();
                      } catch (e) {
                        // 错误已由 request 拦截器处理
                      }
                    },
                  });
                }}
              >
                <StopOutlined/> 取消
              </a>
            </>
          )}
          <Divider type="vertical"/>
          <a
            key="delete"
            style={{color: '#ff4d4f'}}
            onClick={() => {
              showDeleteConfirm(record);
            }}
          >
            <DeleteOutlined/> 删除
          </a>
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
        entityLabel="优惠券"
        style={{ marginBottom: 16 }}
      />
      <Alert
        showIcon
        type="info"
        style={{ marginBottom: 16 }}
        message={`当前查询范围：${buildGovernanceScopeLabel(scope)}`}
        description="这里只展示当前主体可见的优惠券；跨主体的券不会再出现在列表和详情里。"
      />
      <ProTable<CouponListItem>
        headerTitle="优惠券列表"
        actionRef={actionRef}
        rowKey="id"
        search={{
          labelWidth: 120,
        }}
        toolBarRender={() => [
          <Button type="primary" onClick={() => handleModalVisible(true)}>
            <PlusOutlined/> 新建优惠券
          </Button>,
        ]}
        request={(params) => queryCoupon({ ...params, ...toGovernancePayload(scope) })}
        columns={columns}
        rowSelection={{
          onChange: (_, selectedRows) => setSelectedRows(selectedRows),
        }}
        pagination={{pageSize: 10}}
      />
      {selectedRowsState?.length > 0 && (
        <FooterToolbar
          extra={
            <div>
              已选择 <a style={{fontWeight: 600}}>{selectedRowsState.length}</a> 项&nbsp;&nbsp;
            </div>
          }
        >
          <Button
            onClick={async () => {
              await handleRemove(selectedRowsState, scope);
              setSelectedRows([]);
              actionRef.current?.reloadAndRest?.();
            }}
          >
            批量删除
          </Button>
        </FooterToolbar>
      )}

      <CreateCouponForm
        key={'CreateCouponForm'}
        scope={scope}
        onSubmit={async (value) => {
          const success = await handleAdd(value, scope);
          if (success) {
            handleModalVisible(false);
            setCurrentRow(undefined);
            if (actionRef.current) {
              actionRef.current.reload();
            }
          }
        }}
        onCancel={() => {
          handleModalVisible(false);
          if (!showDetail) {
            setCurrentRow(undefined);
          }
        }}
        createModalVisible={createModalVisible}
      />

      <UpdateCouponForm
        key={'UpdateCouponForm'}
        scope={scope}
        onSubmit={async (value) => {
          const success = await handleUpdate(value, scope);
          if (success) {
            handleUpdateModalVisible(false);
            setCurrentRow(undefined);
            if (actionRef.current) {
              actionRef.current.reload();
            }
          }
        }}
        onCancel={() => {
          handleUpdateModalVisible(false);
          if (!showDetail) {
            setCurrentRow(undefined);
          }
        }}
        updateModalVisible={updateModalVisible}
        id={currentRow?.id || 0}
      />

      <CouponDetailForm
        key={'CouponDetailForm'}
        scope={scope}
        onSubmit={async (value) => {
          const success = await handleUpdate(value, scope);
          if (success) {
            handleUpdateModalVisible(false);
            setCurrentRow(undefined);
            if (actionRef.current) {
              actionRef.current.reload();
            }
          }
        }}
        onCancel={() => {
          handleDetailModalVisible(false);
          if (!showDetail) {
            setCurrentRow(undefined);
          }
        }}
        detailModalVisible={detailModalVisible}
        id={currentRow?.id || 0}
      />

      <Drawer
        width={600}
        visible={showDetail}
        onClose={() => {
          setCurrentRow(undefined);
          setShowDetail(false);
        }}
        closable={false}
      >
        {currentRow?.id && (
          <ProDescriptions<CouponListItem>
            column={2}
            title={currentRow?.name}
            request={async () => ({
              data: currentRow || {},
            })}
            params={{
              id: currentRow?.id,
            }}
            columns={columns as ProDescriptionsItemProps<CouponListItem>[]}
          />
        )}
      </Drawer>
    </PageContainer>
  );
};

export default CouponList;
