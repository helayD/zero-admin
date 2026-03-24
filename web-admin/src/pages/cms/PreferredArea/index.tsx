import {
  Alert,
  Button,
  Divider,
  Drawer,
  message,
  Modal,
  Select,
  Switch,
  Tag,
} from 'antd';
import { DeleteOutlined, EditOutlined, ExclamationCircleOutlined, PlusOutlined } from '@ant-design/icons';
import React, { useRef, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import type { ActionType, ProColumns } from '@ant-design/pro-table';
import ProTable from '@ant-design/pro-table';
import type { ProDescriptionsItemProps } from '@ant-design/pro-descriptions';
import ProDescriptions from '@ant-design/pro-descriptions';
import type { PreferredAreaListItem } from './data.d';
import CreatePreferredAreaForm from './components/CreatePostForm';
import UpdatePreferredAreaForm from './components/UpdatePostForm';
import {
  addPreferredArea,
  queryPreferredAreaList,
  removePreferredArea,
  updatePreferredArea,
  updatePreferredAreaStatus,
} from './service';
import GovernanceScopeBar from '@/pages/system/components/GovernanceScopeBar';
import {
  buildGovernanceScopeLabel,
  defaultGovernanceScope,
  type GovernanceScopeValue,
  toGovernancePayload,
} from '@/pages/system/components/governance';

const { confirm } = Modal;

const handleAdd = async (fields: PreferredAreaListItem, scope: GovernanceScopeValue) => {
  const hide = message.loading('正在新增优选专区');
  try {
    await addPreferredArea({ ...fields, ...toGovernancePayload(scope) });
    hide();
    message.success('新增成功');
    return true;
  } catch (error) {
    hide();
    return false;
  }
};

const handleUpdate = async (fields: PreferredAreaListItem, scope: GovernanceScopeValue) => {
  const hide = message.loading('正在更新优选专区');
  try {
    await updatePreferredArea({ ...fields, ...toGovernancePayload(scope) });
    hide();
    message.success('更新成功');
    return true;
  } catch (error) {
    hide();
    return false;
  }
};

const handleRemove = async (ids: number[], scope: GovernanceScopeValue) => {
  const hide = message.loading('正在删除优选专区');
  try {
    await removePreferredArea(ids, toGovernancePayload(scope));
    hide();
    message.success('删除成功');
    return true;
  } catch (error) {
    hide();
    return false;
  }
};

const handleStatus = async (row: PreferredAreaListItem, scope: GovernanceScopeValue) => {
  const hide = message.loading('正在更新显示状态');
  try {
    await updatePreferredAreaStatus({
      ids: [row.id as number],
      showStatus: row.showStatus || 0,
      ...toGovernancePayload(scope),
    });
    hide();
    message.success('状态更新成功');
    return true;
  } catch (error) {
    hide();
    return false;
  }
};

const PreferredAreaList: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const [showDetail, setShowDetail] = useState<boolean>(false);
  const [createVisible, setCreateVisible] = useState<boolean>(false);
  const [updateVisible, setUpdateVisible] = useState<boolean>(false);
  const [currentRow, setCurrentRow] = useState<PreferredAreaListItem>();
  const [scope, setScope] = useState<GovernanceScopeValue>(defaultGovernanceScope);

  const showDeleteConfirm = (record: PreferredAreaListItem) => {
    confirm({
      title: '是否删除优选专区?',
      icon: <ExclamationCircleOutlined />,
      content: `当前主体：${buildGovernanceScopeLabel(scope)}。删除后不可恢复，请确认。`,
      onOk() {
        return handleRemove([record.id as number], scope).then(() => {
          actionRef.current?.reloadAndRest?.();
        });
      },
    });
  };

  const showStatusConfirm = (record: PreferredAreaListItem, newShowStatus: number) => {
    confirm({
      title: '是否更新显示状态?',
      icon: <ExclamationCircleOutlined />,
      content: `当前主体：${buildGovernanceScopeLabel(scope)}。将影响优选专区「${record.name || record.id}」。`,
      async onOk() {
        const success = await handleStatus({ ...record, showStatus: newShowStatus }, scope);
        if (success) {
          actionRef.current?.reload?.();
        }
      },
    });
  };

  const columns: ProColumns<PreferredAreaListItem>[] = [
    {
      title: '编号',
      dataIndex: 'id',
      hideInSearch: true,
    },
    {
      title: '专区名称',
      dataIndex: 'name',
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
      title: '副标题',
      dataIndex: 'subTitle',
      hideInSearch: true,
    },
    {
      title: '显示状态',
      dataIndex: 'showStatus',
      renderFormItem: (text, row) => (
        <Select
          value={row.value}
          options={[
            { value: 1, label: '显示' },
            { value: 0, label: '隐藏' },
          ]}
        />
      ),
      render: (_, entity) => (
        <Switch
          checked={entity.showStatus === 1}
          onChange={(checked) => {
            showStatusConfirm(entity, checked ? 1 : 0);
          }}
        />
      ),
    },
    {
      title: '排序',
      dataIndex: 'sort',
      hideInSearch: true,
    },
    {
      title: '创建时间',
      dataIndex: 'createTime',
      hideInSearch: true,
    },
    {
      title: '更新时间',
      dataIndex: 'updateTime',
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
            key="edit"
            onClick={() => {
              setCurrentRow(record);
              setUpdateVisible(true);
            }}
          >
            <EditOutlined /> 编辑
          </a>
          <Divider type="vertical" />
          <a
            key="delete"
            style={{ color: '#ff4d4f' }}
            onClick={() => {
              showDeleteConfirm(record);
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
        entityLabel="优选专区"
        style={{ marginBottom: 16 }}
      />
      <Alert
        showIcon
        type="info"
        style={{ marginBottom: 16 }}
        message={`当前治理范围：${buildGovernanceScopeLabel(scope)}`}
        description="新增、编辑、上下线与删除都会带上当前主体范围，避免把内容错误写入其他租户或商户。"
      />
      <ProTable<PreferredAreaListItem>
        headerTitle="优选专区管理"
        actionRef={actionRef}
        rowKey="id"
        search={{
          labelWidth: 120,
        }}
        toolBarRender={() => [
          <Button key="create" type="primary" onClick={() => setCreateVisible(true)}>
            <PlusOutlined /> 新建优选专区
          </Button>,
        ]}
        request={(params) => queryPreferredAreaList({ ...params, ...toGovernancePayload(scope) })}
        columns={columns}
        rowSelection={false}
        pagination={{ pageSize: 10 }}
        tableAlertRender={false}
      />

      <CreatePreferredAreaForm
        createModalVisible={createVisible}
        onCancel={() => setCreateVisible(false)}
        onSubmit={async (value) => {
          const success = await handleAdd(value, scope);
          if (success) {
            setCreateVisible(false);
            actionRef.current?.reload?.();
          }
        }}
      />

      <UpdatePreferredAreaForm
        updateModalVisible={updateVisible}
        currentData={currentRow || {}}
        onCancel={() => setUpdateVisible(false)}
        onSubmit={async (value) => {
          const success = await handleUpdate(value, scope);
          if (success) {
            setUpdateVisible(false);
            setCurrentRow(undefined);
            actionRef.current?.reload?.();
          }
        }}
      />

      <Drawer
        width={720}
        visible={showDetail}
        onClose={() => {
          setCurrentRow(undefined);
          setShowDetail(false);
        }}
        closable={false}
      >
        {currentRow?.id && (
          <ProDescriptions<PreferredAreaListItem>
            column={2}
            title="优选专区详情"
            request={async () => ({
              data: currentRow || {},
            })}
            params={{
              id: currentRow?.id,
            }}
            columns={[
              ...(columns.filter((column) => column.dataIndex !== 'option') as ProDescriptionsItemProps<PreferredAreaListItem>[]),
              {
                title: '显示标签',
                dataIndex: 'showStatusLabel',
                render: () =>
                  currentRow.showStatus === 1 ? <Tag color="success">显示</Tag> : <Tag>隐藏</Tag>,
              },
              {
                title: '图片',
                dataIndex: 'pic',
                render: () =>
                  currentRow.pic ? (
                    <img src={currentRow.pic} alt="专区图片" style={{ maxWidth: 200 }} />
                  ) : (
                    '-'
                  ),
              },
            ]}
          />
        )}
      </Drawer>
    </PageContainer>
  );
};

export default PreferredAreaList;
