import { DeleteOutlined, EditOutlined, ExclamationCircleOutlined, PlusOutlined } from '@ant-design/icons';
import { Alert, Button, Divider, Drawer, message, Modal, Select, Space, Switch } from 'antd';
import React, { useMemo, useRef, useState } from 'react';
import type { ActionType, ProColumns } from '@ant-design/pro-table';
import ProTable from '@ant-design/pro-table';
import type { ProDescriptionsItemProps } from '@ant-design/pro-descriptions';
import ProDescriptions from '@ant-design/pro-descriptions';
import AddModal from './components/AddModal';
import UpdateModal from './components/UpdateModal';
import type { ProductSpecValueListItem } from './data.d';
import { addProductSpecValue, queryProductSpecValueList, removeProductSpecValue, updateProductSpecValue, updateProductSpecValueStatus } from './service';
import { buildGovernanceScopeLabel, defaultGovernanceScope, normalizeGovernanceScope, toGovernancePayload, type GovernanceScopeValue } from '@/pages/system/components/governance';
import { readErrorMessage } from '@/pages/system/components/requestError';

const { confirm } = Modal;

const handleAdd = async (fields: ProductSpecValueListItem, scope: GovernanceScopeValue) => {
  const hide = message.loading('正在添加');
  try {
    await addProductSpecValue({ ...fields, ...toGovernancePayload(scope) });
    hide();
    message.success('添加成功');
    return true;
  } catch (error) {
    hide();
    message.error(readErrorMessage(error, '新增规格值失败，请检查规格归属、作用域或状态'));
    return false;
  }
};

const handleUpdate = async (fields: ProductSpecValueListItem, scope: GovernanceScopeValue) => {
  const hide = message.loading('正在更新');
  try {
    await updateProductSpecValue({ ...fields, ...toGovernancePayload(scope) });
    hide();
    message.success('更新成功');
    return true;
  } catch (error) {
    hide();
    message.error(readErrorMessage(error, '更新规格值失败，请检查规格归属和当前作用域'));
    return false;
  }
};

const handleRemove = async (ids: number[], scope: GovernanceScopeValue) => {
  const hide = message.loading('正在删除');
  if (ids.length === 0) return true;
  try {
    await removeProductSpecValue(ids, toGovernancePayload(scope));
    hide();
    message.success('删除成功，即将刷新');
    return true;
  } catch (error) {
    hide();
    message.error(readErrorMessage(error, '删除规格值失败，请核对当前作用域'));
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
    await updateProductSpecValueStatus({ ids, status, ...toGovernancePayload(scope) });
    hide();
    message.success('更新状态成功');
    return true;
  } catch (error) {
    hide();
    message.error(readErrorMessage(error, '更新规格值状态失败，请核对当前作用域'));
    return false;
  }
};

export interface SignProps {
  specId: number;
  scope?: GovernanceScopeValue;
}

const ProductSpecValueList: React.FC<SignProps> = ({ specId, scope }) => {
  const [addVisible, handleAddVisible] = useState(false);
  const [updateVisible, handleUpdateVisible] = useState(false);
  const [showDetail, setShowDetail] = useState(false);
  const actionRef = useRef<ActionType>();
  const [currentRow, setCurrentRow] = useState<ProductSpecValueListItem>();
  const currentScope = useMemo(() => normalizeGovernanceScope(scope || defaultGovernanceScope), [scope]);

  const showDeleteConfirm = (ids: number[]) => {
    confirm({
      title: '是否删除记录?',
      icon: <ExclamationCircleOutlined />,
      content: `当前主体：${buildGovernanceScopeLabel(currentScope)}。删除后不可恢复，请确认。`,
      onOk() { handleRemove(ids, currentScope).then(() => actionRef.current?.reloadAndRest?.()); },
    });
  };

  const showStatusConfirm = (ids: number[], status: number) => {
    confirm({
      title: `确定${status === 1 ? '启用' : '禁用'}吗？`,
      icon: <ExclamationCircleOutlined />,
      content: `当前主体：${buildGovernanceScopeLabel(currentScope)}。将影响 ${ids.length} 个规格值。`,
      async onOk() { await handleStatus(ids, status, currentScope); actionRef.current?.clearSelected?.(); actionRef.current?.reload?.(); },
    });
  };

  const columns: ProColumns<ProductSpecValueListItem>[] = [
    { title: '规格值id', dataIndex: 'id', hideInSearch: true },
    { title: '规格ID', dataIndex: 'specId', hideInSearch: true },
    { title: '规格值', dataIndex: 'value' },
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
        <Alert showIcon type="warning" message={`当前治理范围：${buildGovernanceScopeLabel(currentScope)}`} description="这里只展示当前规格下、当前主体可维护的规格值。后续 SKU 建档会直接消费这些选项。" />
        <ProTable<ProductSpecValueListItem>
          headerTitle="商品规格值管理"
          actionRef={actionRef}
          rowKey="id"
          search={{ labelWidth: 120 }}
          toolBarRender={() => [<Button type="primary" key="primary" onClick={() => handleAddVisible(true)}><PlusOutlined /> 新增</Button>]}
          request={async (params) => {
            const res = await queryProductSpecValueList({ ...params, specId, ...toGovernancePayload(currentScope) });
            if (res.code === '000000') {
              return { data: res.data, total: res.total, pageSize: res.pageSize, current: res.current };
            }
            message.error(res.message || res.msg || '加载规格值失败');
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
        {currentRow?.id && <ProDescriptions<ProductSpecValueListItem> column={2} title="商品规格值详情" request={async () => ({ data: currentRow || {} })} params={{ id: currentRow?.id }} columns={columns as ProDescriptionsItemProps<ProductSpecValueListItem>[]} />}
      </Drawer>
    </>
  );
};

export default ProductSpecValueList;
