import {DeleteOutlined, EditOutlined, ExclamationCircleOutlined, PlusOutlined} from '@ant-design/icons';
import {Alert, Button, Divider, Drawer, message, Modal, Select, Space, Switch} from 'antd';
import React, {useRef, useState} from 'react';
import type {ActionType, ProColumns} from '@ant-design/pro-table';
import ProTable from '@ant-design/pro-table';
import type {ProDescriptionsItemProps} from '@ant-design/pro-descriptions';
import ProDescriptions from '@ant-design/pro-descriptions';
import AddModal from './components/AddModal';
import UpdateModal from './components/UpdateModal';
import type { ProductSkuListItem} from './data.d';
import {addProductSku, queryProductSkuList, removeProductSku, updateProductSku} from './service';
import { buildCatalogActionError, type CatalogActionError } from '@/pages/pms/errorFeedback';
import { defaultGovernanceScope, type GovernanceScopeValue, toGovernancePayload } from '@/pages/system/components/governance';

const {confirm} = Modal;

/**
 * 添加商品SKU
 * @param fields
 */
const handleAdd = async (
  fields: ProductSkuListItem,
  scope: GovernanceScopeValue,
  spuId: number,
  onError?: (error: CatalogActionError) => void,
) => {
  const hide = message.loading('正在添加');
  try {
    await addProductSku({...fields, spuId, ...toGovernancePayload(scope)});
    hide();
    message.success('添加成功');
    return true;
  } catch (error) {
    hide();
    const catalogError = buildCatalogActionError(error, '商品 SKU 建档失败');
    onError?.(catalogError);
    message.error(catalogError.description);
    return false;
  }
};

/**
 * 更新商品SKU
 * @param fields
 */
const handleUpdate = async (
  fields: ProductSkuListItem,
  scope: GovernanceScopeValue,
  onError?: (error: CatalogActionError) => void,
) => {
  const hide = message.loading('正在更新');
  try {
    await updateProductSku({
      data: [fields],
      ...toGovernancePayload(scope),
    });
    hide();

    message.success('更新成功');
    return true;
  } catch (error) {
    hide();
    const catalogError = buildCatalogActionError(error, '商品 SKU 更新失败');
    onError?.(catalogError);
    message.error(catalogError.description);
    return false;
  }
};

/**
 *  删除商品SKU
 * @param ids
 */
const handleRemove = async (ids: number[], scope: GovernanceScopeValue) => {
  const hide = message.loading('正在删除');
  if (ids.length === 0) return true;
  try {
    await removeProductSku(ids, toGovernancePayload(scope));
    hide();
    message.success('删除成功，即将刷新');
    return true;
  } catch (error) {
    hide();
    return false;
  }
};

/**
 * 更新商品SKU状态
 * @param rows
 * @param field
 * @param status
 */
const handleStatus = async (
  rows: ProductSkuListItem[],
  field: 'publishStatus' | 'verifyStatus',
  status: number,
  scope: GovernanceScopeValue,
) => {
  const hide = message.loading('正在更新状态');
  if (rows.length == 0) {
    hide();
    return true;
  }
  try {
    await updateProductSku({
      data: rows.map((row) => ({
        ...row,
        [field]: status,
      })),
      ...toGovernancePayload(scope),
    });
    hide();
    message.success('更新状态成功');
    return true;
  } catch (error) {
    hide();
    return false;
  }
};

export interface SignProps {
  spuId: number;
  scope?: GovernanceScopeValue;
}
const ProductSkuList: React.FC<SignProps> = (props) => {
  const [addVisible, handleAddVisible] = useState<boolean>(false);
  const [updateVisible, handleUpdateVisible] = useState<boolean>(false);
  const [showDetail, setShowDetail] = useState<boolean>(false);
  const actionRef = useRef<ActionType>();
  const [currentRow, setCurrentRow] = useState<ProductSkuListItem>();
  const [submitError, setSubmitError] = useState<CatalogActionError>();
  const effectiveScope = props.scope || defaultGovernanceScope;

  const showDeleteConfirm = (ids: number[]) => {
    confirm({
      title: '是否删除记录?',
      icon: <ExclamationCircleOutlined/>,
      content: `当前主体：${effectiveScope.scopeLabel || '默认范围'}。删除后不可恢复，请确认。`,
      onOk() {
        handleRemove(ids, effectiveScope).then(() => {
          actionRef.current?.reloadAndRest?.();
        });
      },
      onCancel() {
      },
    });
  };

  const showStatusConfirm = (
    rows: ProductSkuListItem[],
    field: 'publishStatus' | 'verifyStatus',
    status: number,
  ) => {
    confirm({
      title: `确定${status == 1 ? '启用' : '停用'}${field === 'publishStatus' ? '上架' : '审核'}状态吗？`,
      icon: <ExclamationCircleOutlined/>,
      content: `当前主体：${effectiveScope.scopeLabel || '默认范围'}。将影响 ${rows.length} 个 SKU。`,
      async onOk() {
        await handleStatus(rows, field, status, effectiveScope)
        actionRef.current?.clearSelected?.();
        actionRef.current?.reload?.();
      },
      onCancel() {
      },
    });
  };

  const columns: ProColumns<ProductSkuListItem>[] = [
    {
      title: '编号',
      dataIndex: 'id',
      hideInSearch: true,
    },
    {
      title: '商品SpuId',
      dataIndex: 'spuId',
      hideInSearch: true,
      hideInTable: true,
    },
    {
      title: 'SKU名称',
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
      title: 'SKU编码',
      dataIndex: 'skuCode',
      hideInSearch: true,
    },
    {
      title: '主图',
      dataIndex: 'mainPic',
      hideInSearch: true,
      valueType: 'image',
      fieldProps: { width: 100, height: 80 },
    },
    {
      title: '图片集',
      dataIndex: 'albumPics',
      hideInSearch: true,
      hideInTable: true,
    },
    {
      title: '价格',
      dataIndex: 'price',
      hideInSearch: true,
    },
    {
      title: '单品促销价格',
      dataIndex: 'promotionPrice',
      hideInSearch: true,
    },
    {
      title: '促销时间',
      dataIndex: 'promotionStartTime',
      hideInSearch: true,
      render: (dom, entity) => {
        return <>{entity.promotionStartTime}至{entity.promotionEndTime}</>;
      },
    },
    {
      title: '库存',
      dataIndex: 'stock',
      hideInSearch: true,
    },
    {
      title: '预警库存',
      dataIndex: 'lowStock',
      hideInSearch: true,
    },
    {
      title: '规格数据',
      dataIndex: 'specData',
      hideInSearch: true,
    },
    {
      title: '重量(kg)',
      dataIndex: 'weight',
      hideInSearch: true,
    },
    {
      title: '上架状态',
      dataIndex: 'publishStatus',
      renderFormItem: (text, row, index) => {
        return (
          <Select
            value={row.value}
            options={[
              { value: 0, label: '下架' },
              { value: 1, label: '上架' },
            ]}
          />
        );
      },
      render: (dom, entity) => {
        return (
          <Switch
            checked={entity.publishStatus == 1}
            onChange={(flag) => {
              showStatusConfirm([entity], 'publishStatus', flag ? 1 : 0);
            }}
          />
        );
      },
    },
    {
      title: '审核状态',
      dataIndex: 'verifyStatus',
      renderFormItem: (text, row, index) => {
        return (
          <Select
            value={row.value}
            options={[
              { value: 0, label: '未审核' },
              { value: 1, label: '审核通过' },
              { value: 2, label: '审核不通过' },
            ]}
          />
        );
      },
      render: (dom, entity) => {
        return (
          <Switch
            checked={entity.verifyStatus == 1}
            onChange={(flag) => {
              showStatusConfirm([entity], 'verifyStatus', flag ? 1 : 0);
            }}
          />
        );
      },
    },
    {
      title: '排序',
      dataIndex: 'sort',
      hideInSearch: true,
    },
    {
      title: '销量',
      dataIndex: 'sales',
      hideInSearch: true,
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
      hideInTable: true,
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
      hideInTable: true,
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
              showDeleteConfirm([record.id as number]);
            }}
          >
            <DeleteOutlined /> 删除
          </a>
        </>
      ),
    },
  ];

return (
    <>
      {submitError && (
        <Alert
          showIcon
          closable
          type="error"
          style={{ marginBottom: 16 }}
          message={submitError.title}
          description={submitError.description}
          onClose={() => setSubmitError(undefined)}
        />
      )}
      <ProTable<ProductSkuListItem>
        headerTitle="商品SKU管理"
        actionRef={actionRef}
        rowKey="id"
        search={ {
          labelWidth: 120,
        } }
        toolBarRender={() => [
          <Button type="primary" key="primary" onClick={() => handleAddVisible(true)}>
            <PlusOutlined/> 新增
          </Button>,
        ]}
        request={async (params) => {
          return queryProductSkuList({
            ...params,
            spuId: props.spuId,
            ...toGovernancePayload(effectiveScope),
          }).then((res) => {
            if (res.code === '000000') {
              return {
                data: res.data,
                total: res.total,
                pageSize: res.pageSize,
                current: res.current,
              };
            } else {
              return message.error(res.msg);
            }
          });
        }}
        columns={columns}
        rowSelection={ {} }
        pagination={ {pageSize: 10}}
        tableAlertRender={ ({
                             selectedRowKeys,
                             selectedRows,
                           }) => {
          const ids = selectedRows.map((row) => row.id as number);
          return (
            <Space size={16}>
              <span>已选 {selectedRowKeys.length} 项</span>
              <Button
                icon={<EditOutlined/>}
                style={ {borderRadius: '5px'}}
                onClick={async () => {
                  showStatusConfirm(selectedRows, 'publishStatus', 1)
                }}
              >批量上架</Button>
              <Button
                icon={<EditOutlined/>}
                style={ {borderRadius: '5px'} }
                onClick={async () => {
                  showStatusConfirm(selectedRows, 'publishStatus', 0)
                }}
              >批量下架</Button>
              <Button
                icon={<DeleteOutlined/>}
                danger
                style={ {borderRadius: '5px'} }
                onClick={async () => {
                  showDeleteConfirm(ids);
                }}
              >批量删除</Button>
            </Space>
          );
        }}
      />


      <AddModal
        key={'AddModal'}
        onSubmit={async (value) => {
          setSubmitError(undefined);
          const success = await handleAdd(value, effectiveScope, props.spuId, setSubmitError);
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
          setSubmitError(undefined);
          const success = await handleUpdate({
            ...currentRow,
            ...value,
            spuId: value.spuId || currentRow?.spuId || props.spuId,
          }, effectiveScope, setSubmitError);
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
        currentData={currentRow || {} }
      />

      <Drawer
        width={600}
        visible={showDetail}
        onClose={() => {
          setCurrentRow(undefined);
          setShowDetail(false)
        }}
        closable={false}
      >
        {currentRow?.id && (
          <ProDescriptions<ProductSkuListItem>
            column={2}
            title={"商品SKU详情"}
            request={async () => ({
              data: currentRow || {},
            })}
            params={ {
              id: currentRow?.id,
            }}
            columns={columns as ProDescriptionsItemProps<ProductSkuListItem>[]}
          />
        )}
      </Drawer>
    </>
  );
};

export default ProductSkuList;
