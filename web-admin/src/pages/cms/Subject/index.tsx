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
import type { SubjectListItem } from './data.d';
import CreatePostForm from './components/CreatePostForm';
import UpdatePostForm from './components/UpdatePostForm';
import {
  addSubject,
  querySubjectList,
  removeSubject,
  updateSubject,
  updateSubjectStatus,
} from './service';
import GovernanceScopeBar from '@/pages/system/components/GovernanceScopeBar';
import {
  buildGovernanceScopeLabel,
  defaultGovernanceScope,
  type GovernanceScopeValue,
  toGovernancePayload,
} from '@/pages/system/components/governance';

const { confirm } = Modal;

const handleAdd = async (fields: SubjectListItem, scope: GovernanceScopeValue) => {
  const hide = message.loading('正在新增专题');
  try {
    await addSubject({ ...fields, ...toGovernancePayload(scope) });
    hide();
    message.success('新增成功');
    return true;
  } catch (error) {
    hide();
    return false;
  }
};

const handleUpdate = async (fields: SubjectListItem, scope: GovernanceScopeValue) => {
  const hide = message.loading('正在更新专题');
  try {
    await updateSubject({ ...fields, ...toGovernancePayload(scope) });
    hide();
    message.success('更新成功');
    return true;
  } catch (error) {
    hide();
    return false;
  }
};

const handleRemove = async (ids: number[], scope: GovernanceScopeValue) => {
  const hide = message.loading('正在删除专题');
  try {
    await removeSubject(ids, toGovernancePayload(scope));
    hide();
    message.success('删除成功');
    return true;
  } catch (error) {
    hide();
    return false;
  }
};

const handleStatus = async (row: SubjectListItem, scope: GovernanceScopeValue) => {
  const hide = message.loading('正在更新专题状态');
  try {
    await updateSubjectStatus({
      ids: [row.id as number],
      showStatus: row.showStatus || 0,
      recommendStatus: row.recommendStatus || 0,
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

const SubjectList: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const [showDetail, setShowDetail] = useState<boolean>(false);
  const [createVisible, setCreateVisible] = useState<boolean>(false);
  const [updateVisible, setUpdateVisible] = useState<boolean>(false);
  const [currentRow, setCurrentRow] = useState<SubjectListItem>();
  const [scope, setScope] = useState<GovernanceScopeValue>(defaultGovernanceScope);

  const showDeleteConfirm = (record: SubjectListItem) => {
    confirm({
      title: '是否删除专题?',
      icon: <ExclamationCircleOutlined />,
      content: `当前主体：${buildGovernanceScopeLabel(scope)}。删除后不可恢复，请确认。`,
      onOk() {
        return handleRemove([record.id as number], scope).then(() => {
          actionRef.current?.reloadAndRest?.();
        });
      },
    });
  };

  const showStatusConfirm = (record: SubjectListItem, patch: Partial<SubjectListItem>, label: string) => {
    if (label === '显示状态' && patch.showStatus === 1) {
      const missing: string[] = [];
      if (!record.title) missing.push('专题标题');
      if (!record.categoryId) missing.push('专题分类');
      if (!record.pic) missing.push('主图');
      if (missing.length > 0) {
        Modal.warning({
          title: '发布校验未通过',
          content: `以下必填字段缺失，无法发布：${missing.join('、')}。请先编辑补齐后再尝试发布。`,
        });
        return;
      }
    }
    confirm({
      title: `是否更新${label}?`,
      icon: <ExclamationCircleOutlined />,
      content: `当前主体：${buildGovernanceScopeLabel(scope)}。将影响专题「${record.title || record.id}」。`,
      async onOk() {
        const success = await handleStatus({ ...record, ...patch }, scope);
        if (success) {
          actionRef.current?.reload?.();
        }
      },
    });
  };

  const columns: ProColumns<SubjectListItem>[] = [
    {
      title: '编号',
      dataIndex: 'id',
      hideInSearch: true,
    },
    {
      title: '专题标题',
      dataIndex: 'title',
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
      title: '专题分类',
      dataIndex: 'categoryName',
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
            showStatusConfirm(entity, { showStatus: checked ? 1 : 0 }, '显示状态');
          }}
        />
      ),
    },
    {
      title: '推荐状态',
      dataIndex: 'recommendStatus',
      renderFormItem: (text, row) => (
        <Select
          value={row.value}
          options={[
            { value: 1, label: '推荐' },
            { value: 0, label: '不推荐' },
          ]}
        />
      ),
      render: (_, entity) => (
        <Switch
          checked={entity.recommendStatus === 1}
          onChange={(checked) => {
            showStatusConfirm(entity, { recommendStatus: checked ? 1 : 0 }, '推荐状态');
          }}
        />
      ),
    },
    {
      title: '关联商品数',
      dataIndex: 'productCount',
      hideInSearch: true,
    },
    {
      title: '生效状态',
      dataIndex: 'effectiveStatus',
      hideInSearch: true,
      render: (_, record) => {
        const colorMap: Record<string, string> = {
          '已发布': 'green',
          '未发布': 'default',
        };
        const text = record.effectiveStatus || '-';
        return <Tag color={colorMap[text] || 'default'}>{text}</Tag>;
      },
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
        entityLabel="专题"
        style={{ marginBottom: 16 }}
      />
      <Alert
        showIcon
        type="info"
        style={{ marginBottom: 16 }}
        message={`当前治理范围：${buildGovernanceScopeLabel(scope)}`}
        description="新增、编辑、上下线与删除都会带上当前主体范围，避免把内容错误写入其他租户或商户。"
      />
      <ProTable<SubjectListItem>
        headerTitle="专题管理"
        actionRef={actionRef}
        rowKey="id"
        search={{
          labelWidth: 120,
        }}
        toolBarRender={() => [
          <Button key="create" type="primary" onClick={() => setCreateVisible(true)}>
            <PlusOutlined /> 新建专题
          </Button>,
        ]}
        request={(params) => querySubjectList({ ...params, ...toGovernancePayload(scope) })}
        columns={columns}
        rowSelection={false}
        pagination={{ pageSize: 10 }}
        tableAlertRender={false}
      />

      <CreatePostForm
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

      <UpdatePostForm
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
          <ProDescriptions<SubjectListItem>
            column={2}
            title="专题详情"
            request={async () => ({
              data: currentRow || {},
            })}
            params={{
              id: currentRow?.id,
            }}
            columns={[
              ...(columns.filter((column) => column.dataIndex !== 'option') as ProDescriptionsItemProps<SubjectListItem>[]),
              {
                title: '显示标签',
                dataIndex: 'showStatusLabel',
                render: () =>
                  currentRow.showStatus === 1 ? <Tag color="success">显示</Tag> : <Tag>隐藏</Tag>,
              },
              {
                title: '推荐标签',
                dataIndex: 'recommendStatusLabel',
                render: () =>
                  currentRow.recommendStatus === 1 ? (
                    <Tag color="processing">推荐</Tag>
                  ) : (
                    <Tag>普通</Tag>
                  ),
              },
              {
                title: '摘要',
                dataIndex: 'description',
                span: 2,
                render: () => currentRow.description || '-',
              },
              {
                title: '正文',
                dataIndex: 'content',
                span: 2,
                render: () =>
                  currentRow.content ? (
                    <div
                      style={{ maxHeight: 300, overflow: 'auto' }}
                      dangerouslySetInnerHTML={{ __html: currentRow.content }}
                    />
                  ) : (
                    '-'
                  ),
              },
              {
                title: '画册图片',
                dataIndex: 'albumPics',
                span: 2,
                render: () =>
                  currentRow.albumPics
                    ? currentRow.albumPics.split(',').map((url, idx) => (
                        <img
                          key={idx}
                          src={url.trim()}
                          alt={`画册图片${idx + 1}`}
                          style={{ maxWidth: 120, marginRight: 8, marginBottom: 8 }}
                        />
                      ))
                    : '-',
              },
            ]}
          />
        )}
      </Drawer>
    </PageContainer>
  );
};

export default SubjectList;
