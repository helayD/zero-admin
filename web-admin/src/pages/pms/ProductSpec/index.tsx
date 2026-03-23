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
import type { ProductSpecListItem } from './data.d';
import { addProductSpec, queryProductSpecList, removeProductSpec, updateProductSpec, updateProductSpecStatus } from './service';
import SpecValueModal from './components/SpecValueModal';
import GovernanceScopeBar from '@/pages/system/components/GovernanceScopeBar';
import { buildGovernanceScopeLabel, defaultGovernanceScope, type GovernanceScopeValue, toGovernancePayload } from '@/pages/system/components/governance';
import { readErrorMessage } from '@/pages/system/components/requestError';

const { confirm } = Modal;

const handleAdd = async (fields: ProductSpecListItem, scope: GovernanceScopeValue) => {
  const hide = message.loading('正在添加');
  try {
    await addProductSpec({ ...fields, ...toGovernancePayload(scope) });
    hide();
    message.success('添加成功');
    return true;
  } catch (error) {
    hide();
    message.error(readErrorMessage(error, '新增规格失败，请检查分类归属、作用域或名称冲突'));
    return false;
  }
};

const handleUpdate = async (fields: ProductSpecListItem, scope: GovernanceScopeValue) => {
  const hide = message.loading('正在更新');
  try {
    await updateProductSpec({ ...fields, ...toGovernancePayload(scope) });
    hide();
    message.success('更新成功');
    return true;
  } catch (error) {
    hide();
    message.error(readErrorMessage(error, '更新规格失败，请检查分类归属和当前作用域'));
    return false;
  }
};

const handleRemove = async (ids: number[], scope: GovernanceScopeValue) => {
  const hide = message.loading('正在删除');
  if (ids.length === 0) return true;
  try {
    await removeProductSpec(ids, toGovernancePayload(scope));
    hide();
    message.success('删除成功，即将刷新');
    return true;
  } catch (error) {
    hide();
    message.error(readErrorMessage(error, '删除规格失败，请先处理下游规格值引用'));
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
    await updateProductSpecStatus({ ids, status, ...toGovernancePayload(scope) });
    hide();
    message.success('更新状态成功');
    return true;
  } catch (error) {
    hide();
    message.error(readErrorMessage(error, '更新规格状态失败，请核对当前作用域'));
    return false;
  }
};

const ProductSpecList: React.FC = () => {
  const [addVisible, handleAddVisible] = useState(false);
  const [updateVisible, handleUpdateVisible] = useState(false);
  const [showDetail, setShowDetail] = useState(false);
  const actionRef = useRef<ActionType>();
  const [currentRow, setCurrentRow] = useState<ProductSpecListItem>();
  const [specVisible, handleSpecVisible] = useState(false);
  const [scope, setScope] = useState<GovernanceScopeValue>(defaultGovernanceScope);

  const showDeleteConfirm = (ids: number[]) => {
    confirm({
      title: '是否删除记录?',
      icon: <ExclamationCircleOutlined />,
      content: `当前主体：${buildGovernanceScopeLabel(scope)}。若规格下仍有规格值，将拒绝删除。`,
      onOk() { handleRemove(ids, scope).then(() => actionRef.current?.reloadAndRest?.()); },
    });
  };

  const showStatusConfirm = (ids: number[], status: number) => {
    confirm({
      title: `确定${status === 1 ? '启用' : '禁用'}吗？`,
      icon: <ExclamationCircleOutlined />,
      content: `当前主体：${buildGovernanceScopeLabel(scope)}。将影响 ${ids.length} 个商品规格。`,
      async onOk() { await handleStatus(ids, status, scope); actionRef.current?.clearSelected?.(); actionRef.current?.reload?.(); },
    });
  };

  const columns: ProColumns<ProductSpecListItem>[] = [
    { title: '', dataIndex: 'id', hideInSearch: true },
    { title: '分类ID', dataIndex: 'categoryId', hideInSearch: true },
    { title: '规格名称', dataIndex: 'name', render: (dom, entity) => <a onClick={() => { setCurrentRow(entity); setShowDetail(true); }}>{dom}</a> },
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
    { title: '设置', dataIndex: 'setting', valueType: 'option', render: (_, record) => <a onClick={() => { handleSpecVisible(true); setCurrentRow(record); }}><EditOutlined /> 设置规格值</a> },
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
        <GovernanceScopeBar value={scope} onChange={setScope} entityLabel="商品规格目录" />
        <Alert showIcon type="warning" message="Consequence Preview" description={`当前正在维护 ${buildGovernanceScopeLabel(scope)} 的商品规格。规格删除、启停和规格值维护会直接影响后续 SKU 建档。`} />
        <ProTable<ProductSpecListItem>
          headerTitle="商品规格管理"
          actionRef={actionRef}
          rowKey="id"
          search={{ labelWidth: 120 }}
          toolBarRender={() => [<Button type="primary" key="primary" onClick={() => handleAddVisible(true)}><PlusOutlined /> 新增</Button>]}
          request={(params) => queryProductSpecList({ ...params, ...toGovernancePayload(scope) })}
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
      <SpecValueModal onCancel={() => { handleSpecVisible(false); if (!showDetail) setCurrentRow(undefined); }} modalVisible={specVisible} specId={currentRow?.id || 0} scope={scope} />

      <Drawer width={600} open={showDetail} onClose={() => { setCurrentRow(undefined); setShowDetail(false); }} closable={false}>
        {currentRow?.id && <ProDescriptions<ProductSpecListItem> column={2} title="商品规格详情" request={async () => ({ data: currentRow || {} })} params={{ id: currentRow?.id }} columns={columns as ProDescriptionsItemProps<ProductSpecListItem>[]} />}
      </Drawer>
    </PageContainer>
  );
};

export default ProductSpecList;
