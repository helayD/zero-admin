import {
  DeleteOutlined,
  EditOutlined,
  ExclamationCircleOutlined,
  PlusOutlined,
} from '@ant-design/icons';
import { Alert, Button, Divider, Drawer, message, Modal, Select, Space, Switch, Tag } from 'antd';
import React, { useRef, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import type { ActionType, ProColumns } from '@ant-design/pro-table';
import ProTable from '@ant-design/pro-table';
import type { ProDescriptionsItemProps } from '@ant-design/pro-descriptions';
import ProDescriptions from '@ant-design/pro-descriptions';
import AddModal from './components/AddModal';
import UpdateModal from './components/UpdateModal';
import type { NoticeListItem } from './data.d';
import {
  addNotice,
  queryNoticeList,
  removeNotice,
  updateNotice,
  updateNoticeStatus,
} from './service';
import GovernanceScopeBar from '../components/GovernanceScopeBar';
import {
  buildGovernanceScopeLabel,
  defaultGovernanceScope,
  governanceScopeColor,
  type GovernanceScopeValue,
  toGovernancePayload,
} from '../components/governance';

const { confirm } = Modal;

const handleAdd = async (fields: NoticeListItem, scope: GovernanceScopeValue) => {
  const hide = message.loading('正在添加');
  try {
    await addNotice({ ...fields, ...toGovernancePayload(scope) });
    hide();
    message.success('添加成功');
    return true;
  } catch (error) {
    hide();
    return false;
  }
};

const handleUpdate = async (fields: NoticeListItem, scope: GovernanceScopeValue) => {
  const hide = message.loading('正在更新');
  try {
    await updateNotice({ ...fields, ...toGovernancePayload(scope) });
    hide();
    message.success('更新成功');
    return true;
  } catch (error) {
    hide();
    return false;
  }
};

const handleRemove = async (ids: number[]) => {
  const hide = message.loading('正在删除');
  if (ids.length === 0) return true;
  try {
    await removeNotice(ids);
    hide();
    message.success('删除成功，即将刷新');
    return true;
  } catch (error) {
    hide();
    return false;
  }
};

const handleStatus = async (ids: number[], status: number) => {
  const hide = message.loading('正在更新状态');
  if (ids.length === 0) {
    hide();
    return true;
  }
  try {
    await updateNoticeStatus({ ids, status });
    hide();
    message.success('更新状态成功');
    return true;
  } catch (error) {
    hide();
    return false;
  }
};

const NoticeList: React.FC = () => {
  const [createModalVisible, setCreateModalVisible] = useState<boolean>(false);
  const [updateModalVisible, setUpdateModalVisible] = useState<boolean>(false);
  const [showDetail, setShowDetail] = useState<boolean>(false);
  const [scope, setScope] = useState<GovernanceScopeValue>(defaultGovernanceScope);
  const actionRef = useRef<ActionType>();
  const [currentRow, setCurrentRow] = useState<NoticeListItem>();

  const showDeleteConfirm = (ids: number[]) => {
    confirm({
      title: '是否删除记录?',
      icon: <ExclamationCircleOutlined />,
      content: '删除的记录不能恢复,请确认!',
      onOk() {
        handleRemove(ids).then(() => {
          actionRef.current?.reloadAndRest?.();
        });
      },
    });
  };

  const showStatusConfirm = (items: NoticeListItem[], status: number) => {
    confirm({
      title: `确定${status === 1 ? '启用' : '关闭'}通知吗？`,
      icon: <ExclamationCircleOutlined />,
      onOk: async () => {
        await handleStatus(
          items.map((item) => item.id),
          status,
        );
        actionRef.current?.reload?.();
      },
    });
  };

  const columns: ProColumns<NoticeListItem>[] = [
    {
      title: '通知标题',
      dataIndex: 'noticeTitle',
      render: (dom, entity) => (
        <a
          onClick={() => {
            setCurrentRow(entity);
            setShowDetail(true);
          }}
        >
          {dom}
        </a>
      ),
    },
    {
      title: '通知类型',
      dataIndex: 'noticeType',
      renderFormItem: (_, row) => (
        <Select
          value={row.value}
          options={[
            { value: 1, label: '通知' },
            { value: 2, label: '公告' },
          ]}
        />
      ),
      render: (_, entity) => (entity.noticeType === 2 ? '公告' : '通知'),
    },
    {
      title: '治理范围',
      dataIndex: 'scopeLabel',
      hideInSearch: true,
      render: (_, entity) => (
        <Tag color={governanceScopeColor(entity.scopeType as any)}>
          {entity.scopeLabel || buildGovernanceScopeLabel(entity)}
        </Tag>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      renderFormItem: (_, row) => (
        <Select
          value={row.value}
          options={[
            { value: 1, label: '正常' },
            { value: 0, label: '关闭' },
          ]}
        />
      ),
      render: (_, entity) => (
        <Switch
          checked={entity.status === 1}
          onChange={(flag) => showStatusConfirm([entity], flag ? 1 : 0)}
        />
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'createTime',
      hideInSearch: true,
      valueType: 'dateTime',
    },
    {
      title: '备注',
      dataIndex: 'remark',
      hideInTable: true,
      hideInSearch: true,
      valueType: 'textarea',
    },
    {
      title: '通知内容',
      dataIndex: 'noticeContent',
      hideInTable: true,
      hideInSearch: true,
      valueType: 'textarea',
    },
    {
      title: '操作',
      dataIndex: 'option',
      valueType: 'option',
      width: 220,
      render: (_, record) => (
        <>
          <a
            onClick={() => {
              setUpdateModalVisible(true);
              setCurrentRow(record);
            }}
          >
            <EditOutlined /> 编辑
          </a>
          <Divider type="vertical" />
          <a style={{ color: '#ff4d4f' }} onClick={() => showDeleteConfirm([record.id])}>
            <DeleteOutlined /> 删除
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
        entityLabel="通知和公告"
        style={{ marginBottom: 16 }}
      />
      <Alert
        showIcon
        type="info"
        style={{ marginBottom: 16 }}
        message={`Consequence Preview · 当前维护 ${buildGovernanceScopeLabel(scope)} 的通知入口`}
        description="通知会直接影响后台运营和治理提醒。若只是某个租户的本地公告，请先切到租户级后再创建。"
      />
      <ProTable<NoticeListItem>
        headerTitle="通知管理"
        actionRef={actionRef}
        rowKey="id"
        search={{ labelWidth: 120 }}
        toolBarRender={() => [
          <Button type="primary" key="create" onClick={() => setCreateModalVisible(true)}>
            <PlusOutlined /> 新建
          </Button>,
        ]}
        request={(params) => queryNoticeList({ ...params, ...toGovernancePayload(scope) })}
        columns={columns}
        rowSelection={{}}
        pagination={{ pageSize: 10 }}
        tableAlertRender={({ selectedRowKeys, selectedRows }) => {
          const ids = selectedRows.map((row) => row.id);
          return (
            <Space size={16}>
              <span>已选 {selectedRowKeys.length} 项</span>
              <Button danger onClick={() => showDeleteConfirm(ids)}>
                批量删除
              </Button>
            </Space>
          );
        }}
      />

      <AddModal
        open={createModalVisible}
        onCancel={() => setCreateModalVisible(false)}
        onSubmit={async (value) => {
          const success = await handleAdd(value, scope);
          if (success) {
            setCreateModalVisible(false);
            setCurrentRow(undefined);
            actionRef.current?.reload?.();
          }
        }}
      />

      <UpdateModal
        open={updateModalVisible}
        currentData={currentRow || {}}
        onCancel={() => setUpdateModalVisible(false)}
        onSubmit={async (value) => {
          const success = await handleUpdate(value, scope);
          if (success) {
            setUpdateModalVisible(false);
            setCurrentRow(undefined);
            actionRef.current?.reload?.();
          }
        }}
      />

      <Drawer
        width={720}
        open={showDetail}
        onClose={() => {
          setCurrentRow(undefined);
          setShowDetail(false);
        }}
        closable={false}
      >
        {currentRow?.id && (
          <ProDescriptions<NoticeListItem>
            column={2}
            title="通知详情"
            request={async () => ({
              data: currentRow || {},
            })}
            params={{ id: currentRow.id }}
            columns={columns as ProDescriptionsItemProps<NoticeListItem>[]}
          />
        )}
      </Drawer>
    </PageContainer>
  );
};

export default NoticeList;
