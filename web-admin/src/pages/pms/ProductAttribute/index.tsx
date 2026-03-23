import {
  DeleteOutlined,
  EditOutlined,
  ExclamationCircleOutlined,
  PlusOutlined,
} from '@ant-design/icons';
import { Alert, Button, Divider, Drawer, message, Modal, Select, Space, Switch, Tag } from 'antd';
import React, { useMemo, useRef, useState } from 'react';
import type { ActionType, ProColumns } from '@ant-design/pro-table';
import ProTable from '@ant-design/pro-table';
import type { ProDescriptionsItemProps } from '@ant-design/pro-descriptions';
import ProDescriptions from '@ant-design/pro-descriptions';
import AddModal from './components/AddModal';
import UpdateModal from './components/UpdateModal';
import type { ProductAttributeListItem } from './data.d';
import {
  addProductAttribute,
  queryProductAttributeList,
  removeProductAttribute,
  updateProductAttribute,
  updateProductAttributeStatus,
} from './service';
import {
  buildGovernanceScopeLabel,
  defaultGovernanceScope,
  normalizeGovernanceScope,
  toGovernancePayload,
  type GovernanceScopeValue,
} from '@/pages/system/components/governance';
import { readErrorMessage } from '@/pages/system/components/requestError';

const { confirm } = Modal;

const handleAdd = async (fields: ProductAttributeListItem, scope: GovernanceScopeValue) => {
  const hide = message.loading('正在添加');
  try {
    await addProductAttribute({ ...fields, ...toGovernancePayload(scope) });
    hide();
    message.success('添加成功');
    return true;
  } catch (error) {
    hide();
    message.error(readErrorMessage(error, '新增属性失败，请检查属性分组、作用域或状态'));
    return false;
  }
};

const handleUpdate = async (fields: ProductAttributeListItem, scope: GovernanceScopeValue) => {
  const hide = message.loading('正在更新');
  try {
    await updateProductAttribute({ ...fields, ...toGovernancePayload(scope) });
    hide();
    message.success('更新成功');
    return true;
  } catch (error) {
    hide();
    message.error(readErrorMessage(error, '更新属性失败，请检查作用域、分组归属或名称冲突'));
    return false;
  }
};

const handleRemove = async (ids: number[], scope: GovernanceScopeValue) => {
  const hide = message.loading('正在删除');
  if (ids.length === 0) return true;
  try {
    await removeProductAttribute(ids, toGovernancePayload(scope));
    hide();
    message.success('删除成功，即将刷新');
    return true;
  } catch (error) {
    hide();
    message.error(readErrorMessage(error, '删除属性失败，请先解除分类绑定或商品引用'));
    return false;
  }
};

const handleStatus = async (ids: number[], status: number, scope: GovernanceScopeValue) => {
  const hide = message.loading('正在更新状态');
  if (ids.length === 0) {
    hide();
    return true;
  }
  try {
    await updateProductAttributeStatus({ ids, status, ...toGovernancePayload(scope) });
    hide();
    message.success('更新状态成功');
    return true;
  } catch (error) {
    hide();
    message.error(readErrorMessage(error, '更新属性状态失败，请核对当前作用域和引用状态'));
    return false;
  }
};

export interface SignProps {
  groupId: number;
  scope?: GovernanceScopeValue;
}

const ProductAttributeList: React.FC<SignProps> = ({ groupId, scope }) => {
  const [addVisible, handleAddVisible] = useState(false);
  const [updateVisible, handleUpdateVisible] = useState(false);
  const [showDetail, setShowDetail] = useState(false);
  const actionRef = useRef<ActionType>();
  const [currentRow, setCurrentRow] = useState<ProductAttributeListItem>();
  const currentScope = useMemo(() => normalizeGovernanceScope(scope || defaultGovernanceScope), [scope]);

  const showDeleteConfirm = (ids: number[]) => {
    confirm({
      title: '是否删除记录?',
      icon: <ExclamationCircleOutlined />,
      content: `当前主体：${buildGovernanceScopeLabel(currentScope)}。已被分类或商品引用的属性将被拒绝删除。`,
      onOk() {
        handleRemove(ids, currentScope).then(() => actionRef.current?.reloadAndRest?.());
      },
    });
  };

  const showStatusConfirm = (ids: number[], status: number) => {
    confirm({
      title: `确定${status === 1 ? '启用' : '禁用'}吗？`,
      icon: <ExclamationCircleOutlined />,
      content: `当前主体：${buildGovernanceScopeLabel(currentScope)}。将影响 ${ids.length} 个属性定义。`,
      async onOk() {
        await handleStatus(ids, status, currentScope);
        actionRef.current?.clearSelected?.();
        actionRef.current?.reload?.();
      },
    });
  };

  const columns: ProColumns<ProductAttributeListItem>[] = [
    { title: '主键id', dataIndex: 'id', hideInSearch: true },
    { title: '属性分组ID', dataIndex: 'groupId', hideInSearch: true },
    {
      title: '属性名称',
      dataIndex: 'name',
      render: (dom, entity) => (
        <a onClick={() => { setCurrentRow(entity); setShowDetail(true); }}>{dom}</a>
      ),
    },
    {
      title: '输入类型',
      dataIndex: 'inputType',
      renderFormItem: (_, row) => (
        <Select
          value={row.value}
          options={[{ value: 1, label: '手动输入' }, { value: 2, label: '单选' }, { value: 3, label: '多选' }]}
        />
      ),
      render: (_, entity) => ({ 1: <Tag color="success">手动输入</Tag>, 2: <Tag color="success">单选</Tag>, 3: <Tag>多选</Tag> }[entity.inputType] || <>未知{entity.inputType}</>),
    },
    {
      title: '值类型',
      dataIndex: 'valueType',
      renderFormItem: (_, row) => (
        <Select
          value={row.value}
          options={[{ value: 1, label: '文本' }, { value: 2, label: '数字' }, { value: 3, label: '日期' }]}
        />
      ),
      render: (_, entity) => ({ 1: <Tag color="success">文本</Tag>, 2: <Tag color="success">数字</Tag>, 3: <Tag>日期</Tag> }[entity.valueType] || <>未知{entity.valueType}</>),
    },
    { title: '可选值列表', dataIndex: 'inputList', hideInSearch: true },
    { title: '单位', dataIndex: 'unit', hideInSearch: true },
    { title: '是否必填', dataIndex: 'isRequired', hideInSearch: true },
    { title: '是否支持搜索', dataIndex: 'isSearchable', hideInSearch: true },
    { title: '是否显示', dataIndex: 'isShow', hideInSearch: true },
    { title: '排序', dataIndex: 'sort', hideInSearch: true },
    {
      title: '状态',
      dataIndex: 'status',
      renderFormItem: (_, row) => <Select value={row.value} options={[{ value: 1, label: '正常' }, { value: 0, label: '禁用' }]} />,
      render: (_, entity) => <Switch checked={entity.status === 1} onChange={(flag) => showStatusConfirm([entity.id], flag ? 1 : 0)} />,
    },
    { title: '创建人ID', dataIndex: 'createBy', hideInSearch: true, hideInTable: true },
    { title: '创建时间', dataIndex: 'createTime', hideInSearch: true },
    { title: '更新人ID', dataIndex: 'updateBy', hideInSearch: true, hideInTable: true },
    { title: '更新时间', dataIndex: 'updateTime', hideInSearch: true, hideInTable: true },
    {
      title: '操作',
      dataIndex: 'option',
      valueType: 'option',
      width: 220,
      render: (_, record) => (
        <>
          <a onClick={() => { handleUpdateVisible(true); setCurrentRow(record); }}><EditOutlined /> 编辑</a>
          <Divider type="vertical" />
          <a style={{ color: '#ff4d4f' }} onClick={() => showDeleteConfirm([record.id])}><DeleteOutlined /> 删除</a>
        </>
      ),
    },
  ];

  return (
    <>
      <Space direction="vertical" style={{ width: '100%' }} size={16}>
        <Alert
          showIcon
          type="warning"
          message={`当前治理范围：${buildGovernanceScopeLabel(currentScope)}`}
          description="这里只展示当前属性分组下、当前主体可维护的商品属性。删除失败时会明确提示分类绑定或商品引用原因。"
        />
        <ProTable<ProductAttributeListItem>
          headerTitle="商品属性管理"
          actionRef={actionRef}
          rowKey="id"
          search={{ labelWidth: 120 }}
          toolBarRender={() => [<Button type="primary" key="primary" onClick={() => handleAddVisible(true)}><PlusOutlined /> 新增</Button>]}
          request={async (params) => {
            const res = await queryProductAttributeList({ ...params, groupId, ...toGovernancePayload(currentScope) });
            if (res.code === '000000') {
              return { data: res.data, total: res.total, pageSize: res.pageSize, current: res.current };
            }
            message.error(res.message || res.msg || '加载商品属性失败');
            return { data: [], total: 0, success: false };
          }}
          columns={columns}
          rowSelection={{}}
          pagination={{ pageSize: 10 }}
          tableAlertRender={({ selectedRowKeys, selectedRows }) => {
            const ids = selectedRows.map((row) => row.id);
            return (
              <Space size={16}>
                <span>已选 {selectedRowKeys.length} 项</span>
                <Button icon={<EditOutlined />} style={{ borderRadius: '5px' }} onClick={async () => showStatusConfirm(ids, 1)}>批量启用</Button>
                <Button icon={<EditOutlined />} style={{ borderRadius: '5px' }} onClick={async () => showStatusConfirm(ids, 0)}>批量禁用</Button>
                <Button icon={<DeleteOutlined />} danger style={{ borderRadius: '5px' }} onClick={async () => showDeleteConfirm(ids)}>批量删除</Button>
              </Space>
            );
          }}
        />
      </Space>

      <AddModal onSubmit={async (value) => { const success = await handleAdd(value, currentScope); if (success) { handleAddVisible(false); setCurrentRow(undefined); actionRef.current?.reload?.(); } }} onCancel={() => { handleAddVisible(false); if (!showDetail) setCurrentRow(undefined); }} addVisible={addVisible} scope={currentScope} />
      <UpdateModal onSubmit={async (value) => { const success = await handleUpdate(value, currentScope); if (success) { handleUpdateVisible(false); setCurrentRow(undefined); actionRef.current?.reload?.(); } }} onCancel={() => { handleUpdateVisible(false); if (!showDetail) setCurrentRow(undefined); }} updateVisible={updateVisible} currentData={currentRow || {}} scope={currentScope} />

      <Drawer width={600} open={showDetail} onClose={() => { setCurrentRow(undefined); setShowDetail(false); }} closable={false}>
        {currentRow?.id && <ProDescriptions<ProductAttributeListItem> column={2} title="商品属性详情" request={async () => ({ data: currentRow || {} })} params={{ id: currentRow?.id }} columns={columns as ProDescriptionsItemProps<ProductAttributeListItem>[]} />}
      </Drawer>
    </>
  );
};

export default ProductAttributeList;
