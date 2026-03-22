import { Alert, Drawer, Select, Tag } from 'antd';
import React, { useRef, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import type { ActionType, ProColumns } from '@ant-design/pro-table';
import ProTable from '@ant-design/pro-table';
import type { ProDescriptionsItemProps } from '@ant-design/pro-descriptions';
import ProDescriptions from '@ant-design/pro-descriptions';
import type { SubjectListItem } from './data.d';
import { querySubjectList } from './service';
import GovernanceScopeBar from '@/pages/system/components/GovernanceScopeBar';
import {
  buildGovernanceScopeLabel,
  defaultGovernanceScope,
  type GovernanceScopeValue,
  toGovernancePayload,
} from '@/pages/system/components/governance';

const SubjectList: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const [showDetail, setShowDetail] = useState<boolean>(false);
  const [currentRow, setCurrentRow] = useState<SubjectListItem>();
  const [scope, setScope] = useState<GovernanceScopeValue>(defaultGovernanceScope);

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
      render: (_, entity) =>
        entity.showStatus === 1 ? <Tag color="success">显示</Tag> : <Tag>隐藏</Tag>,
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
      render: (_, entity) =>
        entity.recommendStatus === 1 ? <Tag color="processing">推荐</Tag> : <Tag>普通</Tag>,
    },
    {
      title: '关联商品数',
      dataIndex: 'productCount',
      hideInSearch: true,
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
        message={`当前查询范围：${buildGovernanceScopeLabel(scope)}`}
        description="这里只展示当前主体可见的专题内容；列表、详情和搜索筛选会共享同一治理范围。"
      />
      <ProTable<SubjectListItem>
        headerTitle="专题管理"
        actionRef={actionRef}
        rowKey="id"
        search={{
          labelWidth: 120,
        }}
        toolBarRender={false}
        request={(params) => querySubjectList({ ...params, ...toGovernancePayload(scope) })}
        columns={columns}
        rowSelection={{}}
        pagination={{ pageSize: 10 }}
        tableAlertRender={false}
      />

      <Drawer
        width={600}
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
            columns={columns as ProDescriptionsItemProps<SubjectListItem>[]}
          />
        )}
      </Drawer>
    </PageContainer>
  );
};

export default SubjectList;
