import {
  DeleteOutlined,
  EditOutlined,
  ExclamationCircleOutlined,
  PlusOutlined,
} from '@ant-design/icons';
import { Alert, Button, Divider, Drawer, message, Modal, Select, Space, Switch } from 'antd';
import React, { useRef, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import type { ActionType, ProColumns } from '@ant-design/pro-table';
import ProTable from '@ant-design/pro-table';
import type { ProDescriptionsItemProps } from '@ant-design/pro-descriptions';
import ProDescriptions from '@ant-design/pro-descriptions';
import AddModal from './components/AddModal';
import UpdateModal from './components/UpdateModal';
import type { ProductBrandListItem } from './data.d';
import {
  addProductBrand,
  queryProductBrandDetail,
  queryProductBrandList,
  removeProductBrand,
  updateProductBrand,
  updateProductBrandRecommendStatus,
  updateProductBrandStatus,
} from './service';
import GovernanceScopeBar from '@/pages/system/components/GovernanceScopeBar';
import {
  defaultGovernanceScope,
  type GovernanceScopeValue,
  toGovernancePayload,
} from '@/pages/system/components/governance';
import { readErrorMessage } from '@/pages/system/components/requestError';

const { confirm } = Modal;

/**
 * 添加商品品牌
 * @param fields
 */
const handleAdd = async (fields: ProductBrandListItem, scope: GovernanceScopeValue) => {
  const hide = message.loading('正在添加');
  try {
    await addProductBrand({ ...fields, ...toGovernancePayload(scope) });
    hide();
    message.success('添加成功');
    return true;
  } catch (error) {
    hide();
    message.error(readErrorMessage(error, '添加品牌失败，请检查当前主体范围与品牌字段'));
    return false;
  }
};

/**
 * 更新商品品牌
 * @param fields
 */
const handleUpdate = async (fields: ProductBrandListItem, scope: GovernanceScopeValue) => {
  const hide = message.loading('正在更新');
  try {
    await updateProductBrand({ ...fields, ...toGovernancePayload(scope) });
    hide();

    message.success('更新成功');
    return true;
  } catch (error) {
    hide();
    message.error(readErrorMessage(error, '更新品牌失败，请检查当前主体范围与品牌状态'));
    return false;
  }
};

/**
 *  删除商品品牌
 * @param ids
 */
const handleRemove = async (ids: number[], scope: GovernanceScopeValue) => {
  const hide = message.loading('正在删除');
  if (ids.length === 0) return true;
  try {
    await removeProductBrand(ids, toGovernancePayload(scope));
    hide();
    message.success('删除成功，即将刷新');
    return true;
  } catch (error) {
    hide();
    message.error(readErrorMessage(error, '删除品牌失败，请先解除被引用关系后重试'));
    return false;
  }
};

/**
 * 更新商品品牌状态
 * @param ids
 * @param status
 * @param t
 */
const handleStatus = async (ids: number[], status: number, t: number, scope: GovernanceScopeValue) => {
  const hide = message.loading('正在更新状态');
  if (ids.length == 0) {
    hide();
    return true;
  }
  try {
    if (t == 1) {
      await updateProductBrandStatus({ ids: ids, status: status, ...toGovernancePayload(scope) });
    } else {
      await updateProductBrandRecommendStatus({ ids: ids, status: status, ...toGovernancePayload(scope) });
    }
    hide();
    message.success('更新状态成功');
    return true;
  } catch (error) {
    hide();
    message.error(readErrorMessage(error, '更新品牌状态失败，请核对当前品牌状态后重试'));
    return false;
  }
};

const ProductBrandList: React.FC = () => {
  const [addVisible, handleAddVisible] = useState<boolean>(false);
  const [updateVisible, handleUpdateVisible] = useState<boolean>(false);
  const [showDetail, setShowDetail] = useState<boolean>(false);
  const actionRef = useRef<ActionType>();
  const [currentRow, setCurrentRow] = useState<ProductBrandListItem>();
  const [scope, setScope] = useState<GovernanceScopeValue>(defaultGovernanceScope);

  const showDeleteConfirm = (ids: number[]) => {
    confirm({
      title: '是否删除记录?',
      icon: <ExclamationCircleOutlined />,
      content: `当前主体：${scope.scopeLabel || '默认范围'}。删除后不可恢复，请确认。`,
      onOk() {
        handleRemove(ids, scope).then(() => {
          actionRef.current?.reloadAndRest?.();
        });
      },
      onCancel() {},
    });
  };

  const showStatusConfirm = (ids: number[], status: number, t: number) => {
    confirm({
      title: `确定${status == 1 ? '启用' : '禁用'}吗？`,
      icon: <ExclamationCircleOutlined />,
      content: `当前主体：${scope.scopeLabel || '默认范围'}。将影响 ${ids.length} 条品牌记录。`,
      async onOk() {
        await handleStatus(ids, status, t, scope);
        actionRef.current?.clearSelected?.();
        actionRef.current?.reload?.();
      },
      onCancel() {},
    });
  };

  const columns: ProColumns<ProductBrandListItem>[] = [
    {
      title: '主键',
      dataIndex: 'id',
      hideInSearch: true,
    },
    {
      title: '品牌名称',
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
      title: '品牌logo',
      dataIndex: 'logo',
      hideInSearch: true,
      valueType: 'image',
      fieldProps: { width: 100, height: 80 },
    },
    {
      title: '专区大图',
      dataIndex: 'bigPic',
      hideInSearch: true,
      valueType: 'image',
      fieldProps: { width: 100, height: 80 },
    },
    {
      title: '描述',
      dataIndex: 'description',
      hideInSearch: true,
    },
    {
      title: '首字母',
      dataIndex: 'firstLetter',
      hideInSearch: true,
    },
    {
      title: '排序',
      dataIndex: 'sort',
      hideInSearch: true,
    },
    {
      title: '推荐状态',
      dataIndex: 'recommendStatus',
      renderFormItem: (text, row, index) => {
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
            checked={entity.recommendStatus == 1}
            onChange={(flag) => {
              showStatusConfirm([entity.id], flag ? 1 : 0, 0);
            }}
          />
        );
      },
    },
    {
      title: '产品数量',
      dataIndex: 'productCount',
      hideInSearch: true,
    },
    {
      title: '产品评论数量',
      dataIndex: 'productCommentCount',
      hideInSearch: true,
    },
    {
      title: '是否启用',
      dataIndex: 'isEnabled',
      renderFormItem: (text, row, index) => {
        return (
          <Select
            value={row.value}
            options={[
              { value: '1', label: '是' },
              { value: '0', label: '否' },
            ]}
          />
        );
      },
      render: (dom, entity) => {
        return (
          <Switch
            checked={entity.isEnabled == 1}
            onChange={(flag) => {
              showStatusConfirm([entity.id], flag ? 1 : 0, 1);
            }}
          />
        );
      },
    },
    {
      title: '创建人ID',
      dataIndex: 'createBy',
      hideInSearch: true,
      hideInTable: true,
    },
    {
      title: '创建时间',
      dataIndex: 'createTime',
      hideInSearch: true,
    },
    {
      title: '更新人ID',
      dataIndex: 'updateBy',
      hideInSearch: true,
      hideInTable: true,
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
            key="sort"
            onClick={() => {
              handleUpdateVisible(true);
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
      <Space direction="vertical" style={{ width: '100%' }} size={16}>
        <GovernanceScopeBar value={scope} onChange={setScope} entityLabel="商品品牌目录" />
        <Alert
          showIcon
          type="warning"
          message="Consequence Preview"
          description={`当前正在维护 ${scope.scopeLabel || '默认范围'} 的商品品牌目录。启停、推荐状态和删除都会直接影响后续商品建档可复用的品牌候选项。`}
        />
      <ProTable<ProductBrandListItem>
        headerTitle="商品品牌管理"
        actionRef={actionRef}
        rowKey="id"
        search={{
          labelWidth: 120,
        }}
        toolBarRender={() => [
          <Button type="primary" key="primary" onClick={() => handleAddVisible(true)}>
            <PlusOutlined /> 新增
          </Button>,
        ]}
        request={async (params) =>
          queryProductBrandList({
            ...params,
            ...toGovernancePayload(scope),
          })
        }
        columns={columns}
        rowSelection={{}}
        pagination={{ pageSize: 10 }}
        tableAlertRender={false}
      />
      </Space>

      <AddModal
        key={'AddModal'}
        onSubmit={async (value) => {
          const success = await handleAdd(value, scope);
          if (success) {
            handleAddVisible(false);
            setCurrentRow(undefined);
            if (actionRef.current) {
              actionRef.current.reload();
            }
          }
        }}
        onCancel={() => {
          handleAddVisible(false);
          if (!showDetail) {
            setCurrentRow(undefined);
          }
        }}
        addVisible={addVisible}
      />

      <UpdateModal
        key={'UpdateModal'}
        onSubmit={async (value) => {
          const success = await handleUpdate(value, scope);
          if (success) {
            handleUpdateVisible(false);
            setCurrentRow(undefined);
            if (actionRef.current) {
              actionRef.current.reload();
            }
          }
        }}
        onCancel={() => {
          handleUpdateVisible(false);
          if (!showDetail) {
            setCurrentRow(undefined);
          }
        }}
        updateVisible={updateVisible}
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
          <ProDescriptions<ProductBrandListItem>
            column={2}
            title={'商品品牌详情'}
            request={async () => {
              if (!currentRow?.id) {
                return { data: currentRow || {} };
              }
              const detail = await queryProductBrandDetail(currentRow.id, toGovernancePayload(scope));
              return {
                data: detail?.data || currentRow || {},
              };
            }}
            params={{
              id: currentRow?.id,
            }}
            columns={columns as ProDescriptionsItemProps<ProductBrandListItem>[]}
          />
        )}
      </Drawer>
    </PageContainer>
  );
};

export default ProductBrandList;
