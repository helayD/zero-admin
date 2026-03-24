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
import type { SubjectCategoryListItem } from './data.d';
import CreateSubjectCategoryForm from './components/CreateSubjectCategoryForm';
import UpdateSubjectCategoryForm from './components/UpdateSubjectCategoryForm';
import {
  addSubjectCategory,
  querySubjectCategoryList,
  removeSubjectCategory,
  updateSubjectCategory,
  updateSubjectCategoryStatus,
} from './service';
import GovernanceScopeBar from '@/pages/system/components/GovernanceScopeBar';
import {
  buildGovernanceScopeLabel,
  defaultGovernanceScope,
  type GovernanceScopeValue,
  toGovernancePayload,
} from '@/pages/system/components/governance';

const { confirm } = Modal;

const handleAdd = async (fields: SubjectCategoryListItem, scope: GovernanceScopeValue) => {
  const hide = message.loading('正在新增专题分类');
  try {
    await addSubjectCategory({ ...fields, ...toGovernancePayload(scope) });
    hide();
    message.success('新增成功');
    return true;
  } catch (error) {
    hide();
    return false;
  }
};

const handleUpdate = async (fields: SubjectCategoryListItem, scope: GovernanceScopeValue) => {
  const hide = message.loading('正在更新专题分类');
  try {
    await updateSubjectCategory({ ...fields, ...toGovernancePayload(scope) });
    hide();
    message.success('更新成功');
    return true;
  } catch (error) {
    hide();
    return false;
  }
};

const handleRemove = async (ids: number[], scope: GovernanceScopeValue) => {
  const hide = message.loading('正在删除专题分类');
  try {
    await removeSubjectCategory(ids, toGovernancePayload(scope));
    hide();
    message.success('删除成功');
    return true;
  } catch (error) {
    hide();
    return false;
  }
};

const handleStatus = async (row: SubjectCategoryListItem, scope: GovernanceScopeValue) => {
  const hide = message.loading('正在更新显示状态');
  try {
    await updateSubjectCategoryStatus({
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

const SubjectCategoryList: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const [showDetail, setShowDetail] = useState<boolean>(false);
  const [createVisible, setCreateVisible] = useState<boolean>(false);
  const [updateVisible, setUpdateVisible] = useState<boolean>(false);
  const [currentRow, setCurrentRow] = useState<SubjectCategoryListItem>();
  const [scope, setScope] = useState<GovernanceScopeValue>(defaultGovernanceScope);

  const showDeleteConfirm = (record: SubjectCategoryListItem) => {
    confirm({
      title: '是否删除专题分类?',
      icon: <ExclamationCircleOutlined />,
      content: `当前主体：${buildGovernanceScopeLabel(scope)}。删除后不可恢复，请确认。`,
      onOk() {
        return handleRemove([record.id as number], scope).then(() => {
          actionRef.current?.reloadAndRest?.();
        });
      },
    });
  };

  const showStatusConfirm = (record: SubjectCategoryListItem, newShowStatus: number) => {
    confirm({
      title: '是否更新显示状态?',
      icon: <ExclamationCircleOutlined />,
      content: `当前主体：${buildGovernanceScopeLabel(scope)}。将影响专题分类「${record.name || record.id}」。`,
      async onOk() {
        const success = await handleStatus({ ...record, showStatus: newShowStatus }, scope);
        if (success) {
          actionRef.current?.reload?.();
        }
      },
    });
  };

  const columns: ProColumns<SubjectCategoryListItem>[] = [
    {
      title: '编号',
      dataIndex: 'id',
      hideInSearch: true,
    },
    {
      title: '分类名称',
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
      title: '图标',
      dataIndex: 'icon',
      hideInSearch: true,
      render: (_, entity) =>
        entity.icon ? <img src={entity.icon} alt="图标" style={{ maxWidth: 32 }} /> : '-',
    },
    {
      title: '专题数量',
      dataIndex: 'subjectCount',
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
        entityLabel="专题分类"
        style={{ marginBottom: 16 }}
      />
      <Alert
        showIcon
        type="info"
        style={{ marginBottom: 16 }}
        message={`当前治理范围：${buildGovernanceScopeLabel(scope)}`}
        description="新增、编辑、上下线与删除都会带上当前主体范围，避免把内容错误写入其他租户或商户。"
      />
      <ProTable<SubjectCategoryListItem>
        headerTitle="专题分类管理"
        actionRef={actionRef}
        rowKey="id"
        search={{
          labelWidth: 120,
        }}
        toolBarRender={() => [
          <Button key="create" type="primary" onClick={() => setCreateVisible(true)}>
            <PlusOutlined /> 新建专题分类
          </Button>,
        ]}
        request={(params) => querySubjectCategoryList({ ...params, ...toGovernancePayload(scope) })}
        columns={columns}
        rowSelection={false}
        pagination={{ pageSize: 10 }}
        tableAlertRender={false}
      />

      <CreateSubjectCategoryForm
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

      <UpdateSubjectCategoryForm
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
          <ProDescriptions<SubjectCategoryListItem>
            column={2}
            title="专题分类详情"
            request={async () => ({
              data: currentRow || {},
            })}
            params={{
              id: currentRow?.id,
            }}
            columns={[
              ...(columns.filter((column) => column.dataIndex !== 'option') as ProDescriptionsItemProps<SubjectCategoryListItem>[]),
              {
                title: '显示标签',
                dataIndex: 'showStatusLabel',
                render: () =>
                  currentRow.showStatus === 1 ? <Tag color="success">显示</Tag> : <Tag>隐藏</Tag>,
              },
            ]}
          />
        )}
      </Drawer>
    </PageContainer>
  );
};

export default SubjectCategoryList;
