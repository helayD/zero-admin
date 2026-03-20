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
import type { PostListItem } from './data.d';
import { addPost, queryPostList, removePost, updatePost, updatePostStatus } from './service';
import GovernanceScopeBar from '../components/GovernanceScopeBar';
import {
  buildGovernanceScopeLabel,
  defaultGovernanceScope,
  governanceScopeColor,
  type GovernanceScopeValue,
  toGovernancePayload,
} from '../components/governance';

const { confirm } = Modal;

/**
 * 添加节点
 * @param fields
 */
const handleAdd = async (fields: PostListItem, scope: GovernanceScopeValue) => {
  const hide = message.loading('正在添加');
  try {
    await addPost({ ...fields, ...toGovernancePayload(scope) });
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
const handleUpdate = async (fields: PostListItem, scope: GovernanceScopeValue) => {
  const hide = message.loading('正在更新');
  try {
    await updatePost({ ...fields, ...toGovernancePayload(scope) });
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
 * @param ids
 */
const handleRemove = async (ids: number[]) => {
  const hide = message.loading('正在删除');
  if (ids.length === 0) return true;
  try {
    await removePost(ids);
    hide();
    message.success('删除成功，即将刷新');
    return true;
  } catch (error) {
    hide();
    return false;
  }
};

/**
 * 更新状态
 * @param ids
 * @param status
 */
const handleStatus = async (ids: number[], status: number) => {
  const hide = message.loading('正在更新状态');
  if (ids.length == 0) {
    hide();
    return true;
  }
  try {
    await updatePostStatus({ ids, status });
    hide();
    message.success('更新状态成功');
    return true;
  } catch (error) {
    hide();
    return false;
  }
};

const PostList: React.FC = () => {
  const [createModalVisible, handleModalVisible] = useState<boolean>(false);
  const [updateModalVisible, handleUpdateModalVisible] = useState<boolean>(false);
  const [showDetail, setShowDetail] = useState<boolean>(false);
  const actionRef = useRef<ActionType>();
  const [currentRow, setCurrentRow] = useState<PostListItem>();
  const [scope, setScope] = useState<GovernanceScopeValue>(defaultGovernanceScope);

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
      onCancel() {},
    });
  };

  const showStatusConfirm = (item: PostListItem[], status: number) => {
    confirm({
      title: `确定${status == 1 ? '启用' : '禁用'}岗位吗？`,
      icon: <ExclamationCircleOutlined />,
      async onOk() {
        await handleStatus(
          item.map((x) => x.id),
          status,
        );
        actionRef.current?.reload?.();
      },
      onCancel() {},
    });
  };

  const columns: ProColumns<PostListItem>[] = [
    {
      title: '岗位编号',
      dataIndex: 'id',
      hideInSearch: true,
    },
    {
      title: '岗位编码',
      dataIndex: 'postCode',
    },
    {
      title: '岗位名称',
      dataIndex: 'postName',
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
      title: '岗位排序',
      dataIndex: 'sort',
      hideInSearch: true,
    },
    {
      title: '状态',
      dataIndex: 'status',
      renderFormItem: (_, row) => {
        return (
          <Select
            value={row.value}
            options={[
              { value: '1', label: '正常' },
              { value: '0', label: '禁用' },
            ]}
          />
        );
      },
      render: (dom, entity) => {
        return (
          <Switch
            checked={entity.status == 1}
            onChange={(flag) => {
              showStatusConfirm([entity], flag ? 1 : 0);
            }}
          />
        );
      },
    },
    {
      title: '备注',
      dataIndex: 'remark',
      valueType: 'textarea',
      hideInSearch: true,
    },
    {
      title: '创建者',
      dataIndex: 'createBy',
      hideInSearch: true,
    },
    {
      title: '创建时间',
      dataIndex: 'createTime',
      valueType: 'dateTime',
      hideInSearch: true,
    },
    {
      title: '更新者',
      dataIndex: 'updateBy',
      hideInSearch: true,
    },
    {
      title: '更新时间',
      dataIndex: 'updateTime',
      valueType: 'dateTime',
      hideInSearch: true,
    },
    {
      title: '操作',
      dataIndex: 'option',
      valueType: 'option',
      width: 220,
      render: (_, record) => (
        <>
          <a
            key="sort"
            onClick={() => {
              handleUpdateModalVisible(true);
              setCurrentRow(record);
            }}
          >
            <EditOutlined /> 编辑
          </a>
          <Divider type="vertical" />
          <a
            key="delete"
            style={{ color: '#ff4d4f' }}
            onClick={() => {
              showDeleteConfirm([record.id]);
            }}
          >
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
        entityLabel="岗位元数据"
        style={{ marginBottom: 16 }}
      />
      <Alert
        showIcon
        type="info"
        style={{ marginBottom: 16 }}
        message={`Consequence Preview · 当前维护 ${buildGovernanceScopeLabel(scope)} 的岗位定义`}
        description="岗位会被后续后台用户绑定和复用。若你希望做平台默认岗位，请保持平台级；若只服务单一租户，请切到租户级后再保存。"
      />
      <ProTable<PostListItem>
        headerTitle="岗位管理"
        actionRef={actionRef}
        rowKey="id"
        search={{
          labelWidth: 120,
        }}
        toolBarRender={() => [
          <Button type="primary" key="primary" onClick={() => handleModalVisible(true)}>
            <PlusOutlined /> 新增
          </Button>,
        ]}
        request={(params) => queryPostList({ ...params, ...toGovernancePayload(scope) })}
        columns={columns}
        rowSelection={{}}
        pagination={{ pageSize: 10 }}
        tableAlertRender={({ selectedRowKeys, selectedRows, onCleanSelected }) => {
          const ids = selectedRows.map((row) => row.id);
          return (
            <Space size={16}>
              <span>已选 {selectedRowKeys.length} 项</span>
              <Button
                icon={<EditOutlined />}
                style={{ borderRadius: '5px' }}
                onClick={async () => {
                  await handleStatus(ids, 1);
                  onCleanSelected();
                  actionRef.current?.reload?.();
                }}
              >
                批量启用
              </Button>
              <Button
                icon={<EditOutlined />}
                style={{ borderRadius: '5px' }}
                onClick={async () => {
                  await handleStatus(ids, 0);
                  onCleanSelected();
                  actionRef.current?.reload?.();
                }}
              >
                批量禁用
              </Button>
              <Button
                icon={<DeleteOutlined />}
                danger
                style={{ borderRadius: '5px' }}
                onClick={async () => {
                  showDeleteConfirm(ids);
                }}
              >
                批量删除
              </Button>
            </Space>
          );
        }}
      />

      <AddModal
        key={'CreatePostForm'}
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
        open={createModalVisible}
      />

      <UpdateModal
        key={'UpdatePostForm'}
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
        open={updateModalVisible}
        currentData={currentRow || {}}
      />

      <Drawer
        width={600}
        open={showDetail}
        onClose={() => {
          setCurrentRow(undefined);
          setShowDetail(false);
        }}
        closable={false}
      >
        {currentRow?.id && (
          <ProDescriptions<PostListItem>
            column={2}
            title={'岗位详情'}
            request={async () => ({
              data: currentRow || {},
            })}
            params={{
              id: currentRow?.id,
            }}
            columns={columns as ProDescriptionsItemProps<PostListItem>[]}
          />
        )}
      </Drawer>
    </PageContainer>
  );
};

export default PostList;
