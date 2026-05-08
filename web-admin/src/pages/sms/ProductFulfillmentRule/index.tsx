import {
  PlusOutlined,
  ExclamationCircleOutlined,
  DeleteOutlined,
  EditOutlined,
  StopOutlined,
  CheckCircleOutlined,
} from '@ant-design/icons';
import { Alert, Button, Divider, message, Modal, Space, Switch, Tag } from 'antd';
import React, { useState, useRef } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import ProTable from '@ant-design/pro-table';
import type { ProColumns, ActionType } from '@ant-design/pro-table';
import type { ProductFulfillmentRuleListItem } from './data.d';
import { RefundPolicyTextMap } from './data.d';
import {
  queryProductFulfillmentRuleList,
  deleteProductFulfillmentRule,
  updateProductFulfillmentRuleStatus,
  checkProductFulfillmentRuleBinding,
} from './service';
import GovernanceScopeBar from '@/pages/system/components/GovernanceScopeBar';
import {
  buildGovernanceScopeLabel,
  defaultGovernanceScope,
  governanceScopeColor,
  type GovernanceScopeValue,
  toGovernancePayload,
} from '@/pages/system/components/governance';

const { confirm } = Modal;

const renderStatusTag = (status: number) =>
  status === 1 ? <Tag color="green">启用</Tag> : <Tag color="red">禁用</Tag>;

const renderRefundPolicyTag = (policy: string) => {
  const text = RefundPolicyTextMap[policy] || policy;
  const color = policy === 'freeze_card' ? 'orange' : policy === 'recycle_card' ? 'red' : 'blue';
  return <Tag color={color}>{text}</Tag>;
};

const renderScopeSource = (record: Pick<
  ProductFulfillmentRuleListItem,
  'scopeType' | 'platformId' | 'tenantId' | 'merchantId'
>) => {
  const scopeLabel = buildGovernanceScopeLabel({
    scopeType: record.scopeType,
    platformId: record.platformId,
    tenantId: record.tenantId,
    merchantId: record.merchantId,
  });
  return <Tag color={governanceScopeColor(record.scopeType)}>{scopeLabel}</Tag>;
};

const ProductFulfillmentRuleList: React.FC = () => {
  const [showDetail, setShowDetail] = useState<boolean>(false);
  const actionRef = useRef<ActionType>();
  const [currentRow, setCurrentRow] = useState<ProductFulfillmentRuleListItem>();
  const [scope, setScope] = useState<GovernanceScopeValue>(defaultGovernanceScope);

  const handleDelete = async (record: ProductFulfillmentRuleListItem) => {
    try {
      const bindingResult = await checkProductFulfillmentRuleBinding(
        record.id,
        toGovernancePayload(scope),
      );
      if (bindingResult.data?.canDelete === false) {
        Modal.warning({
          title: '无法删除',
          content: bindingResult.data.message || '该规则已绑定商品，无法删除',
        });
        return;
      }
    } catch (error) {
      // 继续尝试删除
    }

    confirm({
      title: '是否删除该发卡规则?',
      icon: <ExclamationCircleOutlined />,
      content: `规则名称：${record.ruleName}。删除后不可恢复，请确认。`,
      onOk() {
        return deleteProductFulfillmentRule(record.id, toGovernancePayload(scope)).then(() => {
          message.success('删除成功，即将刷新');
          actionRef.current?.reload?.();
        });
      },
    });
  };

  const handleStatusChange = async (record: ProductFulfillmentRuleListItem, checked: boolean) => {
    const newStatus = checked ? 1 : 0;
    if (newStatus === 0) {
      try {
        const bindingResult = await checkProductFulfillmentRuleBinding(
          record.id,
          toGovernancePayload(scope),
        );
        if (bindingResult.data?.canDisable === false) {
          Modal.warning({
            title: '无法禁用',
            content: bindingResult.data.message || '该规则已绑定商品，无法禁用',
          });
          return;
        }
      } catch (error) {
        // 继续尝试更新
      }
    }

    try {
      await updateProductFulfillmentRuleStatus({
        id: record.id,
        ruleStatus: newStatus,
        ...toGovernancePayload(scope),
      });
      message.success(`规则已${checked ? '启用' : '禁用'}`);
      actionRef.current?.reload?.();
    } catch (error) {
      message.error('状态更新失败');
    }
  };

  const columns: ProColumns<ProductFulfillmentRuleListItem>[] = [
    {
      title: '规则ID',
      dataIndex: 'id',
      hideInSearch: true,
    },
    {
      title: '规则名称',
      dataIndex: 'ruleName',
    },
    {
      title: '关联卡片模板',
      dataIndex: 'cardTemplateName',
      hideInSearch: true,
    },
    {
      title: '有效期（天）',
      dataIndex: 'expireDays',
      hideInSearch: true,
    },
    {
      title: '可转赠',
      dataIndex: 'transferable',
      hideInSearch: true,
      render: (_, entity) => (
        entity.transferable === 1 ? (
          <Tag color="green">是{entity.transferLimit > 0 ? `（限${entity.transferLimit}次）` : ''}</Tag>
        ) : (
          <Tag>否</Tag>
        )
      ),
    },
    {
      title: '退款处置',
      dataIndex: 'refundPolicy',
      hideInSearch: true,
      render: (_, entity) => renderRefundPolicyTag(entity.refundPolicy),
    },
    {
      title: '绑定商品数',
      dataIndex: 'bindingCount',
      hideInSearch: true,
      render: (_, entity) => (
        <span>{entity.bindingCount} 个商品</span>
      ),
    },
    {
      title: '作用域来源',
      dataIndex: 'scopeType',
      hideInSearch: true,
      render: (_, entity) => renderScopeSource(entity),
    },
    {
      title: '状态',
      dataIndex: 'ruleStatus',
      valueType: 'select',
      valueEnum: {
        1: { text: '启用', status: 'Success' },
        0: { text: '禁用', status: 'Error' },
      },
      render: (_, entity) => (
        <Space size={8}>
          {renderStatusTag(entity.ruleStatus)}
          <Switch
            checked={entity.ruleStatus === 1}
            checkedChildren={<CheckCircleOutlined />}
            unCheckedChildren={<StopOutlined />}
            onChange={(checked) => handleStatusChange(entity, checked)}
          />
        </Space>
      ),
    },
    {
      title: '创建人',
      dataIndex: 'createBy',
      hideInSearch: true,
    },
    {
      title: '创建时间',
      dataIndex: 'createTime',
      hideInSearch: true,
      valueType: 'dateTime',
    },
    {
      title: '操作',
      dataIndex: 'option',
      valueType: 'option',
      width: 200,
      render: (_, record) => (
        <>
          <a
            onClick={() => {
              setCurrentRow(record);
              setShowDetail(true);
            }}
          >
            <EditOutlined /> 编辑
          </a>
          <Divider type="vertical" />
          <a
            style={{ color: '#ff4d4f' }}
            onClick={() => handleDelete(record)}
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
        entityLabel="发卡规则"
        style={{ marginBottom: 16 }}
      />
      <Alert
        showIcon
        type="info"
        style={{ marginBottom: 16 }}
        message={`当前查询范围：${buildGovernanceScopeLabel(scope)}`}
        description="管理提货卡商品的发卡规则，包括卡片模板、有效期、转赠规则和退款处置策略。"
      />
      <ProTable<ProductFulfillmentRuleListItem>
        headerTitle="发卡规则管理"
        actionRef={actionRef}
        rowKey="id"
        search={{
          labelWidth: 120,
        }}
        toolBarRender={() => [
          <Button
            type="primary"
            key="primary"
            onClick={() => {
              setCurrentRow(undefined);
              setShowDetail(true);
            }}
          >
            <PlusOutlined /> 新增规则
          </Button>,
        ]}
        request={(params) =>
          queryProductFulfillmentRuleList({
            ...params,
            ...toGovernancePayload(scope),
          })
        }
        columns={columns}
        pagination={{ pageSize: 10 }}
      />

      {/* TODO: 添加创建/编辑表单弹窗 */}
      {showDetail && currentRow && (
        <Modal
          title="编辑发卡规则"
          visible={showDetail}
          onCancel={() => {
            setShowDetail(false);
            setCurrentRow(undefined);
          }}
          footer={null}
          width={800}
        >
          <p>编辑功能待实现：{currentRow.ruleName}</p>
        </Modal>
      )}
      {showDetail && !currentRow && (
        <Modal
          title="新增发卡规则"
          visible={showDetail}
          onCancel={() => {
            setShowDetail(false);
          }}
          footer={null}
          width={800}
        >
          <p>新增功能待实现</p>
        </Modal>
      )}
    </PageContainer>
  );
};

export default ProductFulfillmentRuleList;
