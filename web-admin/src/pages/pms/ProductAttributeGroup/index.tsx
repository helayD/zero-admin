import { DeleteOutlined, EditOutlined, ExclamationCircleOutlined, PlusOutlined } from '@ant-design/icons';
import { Alert, Button, Divider, Drawer, message, Modal, Select, Space, Switch } from 'antd';
import React, { useRef, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import type { ActionType, ProColumns } from '@ant-design/pro-table';
import ProTable from '@ant-design/pro-table';
import type { ProDescriptionsItemProps } from '@ant-design/pro-descriptions';
import ProDescriptions from '@ant-design/pro-descriptions';
import AddModal from './components/AddModal';
import UpdateModal from './components/UpdateModal';
import type { ProductAttributeGroupListItem } from './data.d';
import {
  addProductAttributeGroup,
  queryProductAttributeGroupList,
  removeProductAttributeGroup,
  updateProductAttributeGroup,
  updateProductAttributeGroupStatus,
} from './service';
import AttributeModal from './components/AttributeModal';
import GovernanceScopeBar from '@/pages/system/components/GovernanceScopeBar';
import { buildGovernanceScopeLabel, defaultGovernanceScope, type GovernanceScopeValue, toGovernancePayload } from '@/pages/system/components/governance';
import { readCatalogErrorMessage } from '@/pages/pms/scopedCatalog';

const { confirm } = Modal;

const handleAdd = async (fields: ProductAttributeGroupListItem, scope: GovernanceScopeValue) => {
  const hide = message.loading('正在添加');
  try {
    await addProductAttributeGroup({ ...fields, ...toGovernancePayload(scope) });
    hide();
    message.success('添加成功');
    return true;
  } catch (error) {
    hide();
    message.error(readCatalogErrorMessage(error, '新增属性分组失败，请检查分类归属、作用域或名称冲突'));
    return false;
  }
};

const handleUpdate = async (fields: ProductAttributeGroupListItem, scope: GovernanceScopeValue) => {
  const hide = message.loading('正在更新');
  try {
    await updateProductAttributeGroup({ ...fields, ...toGovernancePayload(scope) });
    hide();
    message.success('更新成功');
    return true;
  } catch (error) {
    hide();
    message.error(readCatalogErrorMessage(error, '更新属性分组失败，请检查分类归属和当前作用域'));
    return false;
  }
};

const handleRemove = async (ids: number[], scope: GovernanceScopeValue) => {
  const hide = message.loading('正在删除');
  if (ids.length === 0) return true;
  try {
    await removeProductAttributeGroup(ids, toGovernancePayload(scope));
    hide();
    message.success('删除成功，即将刷新');
    return true;
  } catch (error) {
    hide();
    message.error(readCatalogErrorMessage(error, '删除属性分组失败，请先处理下游属性引用'));
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
    await updateProductAttributeGroupStatus({ ids, status, ...toGovernancePayload(scope) });
    hide();
    message.success('更新状态成功');
    return true;
  } catch (error) {
    hide();
    message.error(readCatalogErrorMessage(error, '更新属性分组状态失败，请核对当前作用域'));
    return false;
  }
};

const ProductAttributeGroupList: React.FC = () => {
  const [addVisible, handleAddVisible] = useState(false);
  const [updateVisible, handleUpdateVisible] = useState(false);
  const [showDetail, setShowDetail] = useState(false);
  const actionRef = useRef<ActionType>();
  const [currentRow, setCurrentRow] = useState<ProductAttributeGroupListItem>();
  const [attributeVisible, handleAttributeVisible] = useState(false);
  const [scope, setScope] = useState<GovernanceScopeValue>(defaultGovernanceScope);

  const showDeleteConfirm = (ids: number[]) => {
    confirm({
      title: '是否删除记录?',
      icon: <ExclamationCircleOutlined />,
      content: `当前主体：${buildGovernanceScopeLabel(scope)}。若分组下仍有属性，将拒绝删除。`,
      onOk() { handleRemove(ids, scope).then(() => actionRef.current?.reloadAndRest?.()); },
    });
  };

  const showStatusConfirm = (ids: number[], status: number) => {
    confirm({
      title: `确定${status === 1 ? '启用' : '禁用'}吗？`,
      icon: <ExclamationCircleOutlined />,
      content: `当前主体：${buildGovernanceScopeLabel(scope)}。将影响 ${ids.length} 个属性分组。`,
      async onOk() { await handleStatus(ids, status, scope); actionRef.current?.clearSelected?.(); actionRef.current?.reload?.(); },
    });
  };

  const columns: ProColumns<ProductAttributeGroupListItem>[] = [
    { title: '主键id', dataIndex: 'id', hideInSearch: true },
    { title: '分类ID', dataIndex: 'categoryId', hideInSearch: true },
    { title: '分组名称', dataIndex: 'name', render: (dom, entity) => <a onClick={() => { setCurrentRow(entity); setShowDetail(true); }}>{dom}</a> },
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
    { title: '更新时间', dataIndex: 'updateTime', hideInSearch: true },
    { title: '设置', dataIndex: 'setting', valueType: 'option', render: (_, record) => <a onClick={() => { handleAttributeVisible(true); setCurrentRow(record); }}><EditOutlined /> 设置属性</a> },
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
    <PageContainer>
      <Space direction="vertical" style={{ width: '100%' }} size={16}>
        <GovernanceScopeBar value={scope} onChange={setScope} entityLabel="商品属性分组目录" />
        <Alert showIcon type="warning" message="Consequence Preview" description={`当前正在维护 ${buildGovernanceScopeLabel(scope)} 的商品属性分组。分组删除、启停和属性维护会直接影响后续商品属性建档。`} />
        <ProTable<ProductAttributeGroupListItem>
          headerTitle="商品属性分组管理"
          actionRef={actionRef}
          rowKey="id"
          search={{ labelWidth: 120 }}
          toolBarRender={() => [<Button type="primary" key="primary" onClick={() => handleAddVisible(true)}><PlusOutlined /> 新增</Button>]}
          request={(params) => queryProductAttributeGroupList({ ...params, ...toGovernancePayload(scope) })}
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

      <AddModal onSubmit={async (value) => { const success = await handleAdd(value, scope); if (success) { handleAddVisible(false); setCurrentRow(undefined); actionRef.current?.reload?.(); } }} onCancel={() => { handleAddVisible(false); if (!showDetail) setCurrentRow(undefined); }} addVisible={addVisible} scope={scope} />
      <UpdateModal onSubmit={async (value) => { const success = await handleUpdate(value, scope); if (success) { handleUpdateVisible(false); setCurrentRow(undefined); actionRef.current?.reload?.(); } }} onCancel={() => { handleUpdateVisible(false); if (!showDetail) setCurrentRow(undefined); }} updateVisible={updateVisible} currentData={currentRow || {}} scope={scope} />
      <AttributeModal onCancel={() => { handleAttributeVisible(false); if (!showDetail) setCurrentRow(undefined); }} modalVisible={attributeVisible} groupId={currentRow?.id || 0} scope={scope} />

      <Drawer width={600} open={showDetail} onClose={() => { setCurrentRow(undefined); setShowDetail(false); }} closable={false}>
        {currentRow?.id && <ProDescriptions<ProductAttributeGroupListItem> column={2} title="商品属性分组详情" request={async () => ({ data: currentRow || {} })} params={{ id: currentRow?.id }} columns={columns as ProDescriptionsItemProps<ProductAttributeGroupListItem>[]} />}
      </Drawer>
    </PageContainer>
  );
};

export default ProductAttributeGroupList;
