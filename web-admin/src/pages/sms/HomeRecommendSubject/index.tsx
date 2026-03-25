import { EditOutlined, ExclamationCircleOutlined, PlusOutlined } from '@ant-design/icons';
import { Alert, Button, Drawer, message, Modal, Select, Switch } from 'antd';
import React, { useRef, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import type { ActionType, ProColumns } from '@ant-design/pro-table';
import ProTable from '@ant-design/pro-table';
import type { ProDescriptionsItemProps } from '@ant-design/pro-descriptions';
import ProDescriptions from '@ant-design/pro-descriptions';
import CreateForm from './components/CreateForm';
import SetSortForm from './components/SetSortForm';
import type { RecommendSubjectListItem } from './data.d';
import {
  addRecommendSubject,
  queryRecommendSubjectList,
  removeRecommendSubject,
  updateRecommendSubjectSort,
} from './service';
import GovernanceScopeBar from '@/pages/system/components/GovernanceScopeBar';
import {
  buildGovernanceScopeLabel,
  defaultGovernanceScope,
  type GovernanceScopeValue,
  toGovernancePayload,
} from '@/pages/system/components/governance';

const { confirm } = Modal;

/**
 * 添加节点
 * @param subjectIds
 */
const handleAdd = async (subjectIds: number[]) => {
  const hide = message.loading('正在添加');
  if (subjectIds.length <= 0) {
    hide();
    return true;
  }
  try {
    await addRecommendSubject(subjectIds);
    hide();
    message.success('添加成功');
    return true;
  } catch (error) {
    hide();
    return false;
  }
};

/**
 * 更新节点排序
 * @param fields
 */
const handleUpdateSubjectSor = async (fields: RecommendSubjectListItem) => {
  const hide = message.loading('正在更新');
  try {
    await updateRecommendSubjectSort(fields);
    hide();

    message.success('更新成功');
    return true;
  } catch (error) {
    hide();
    return false;
  }
};

/**
 * 更新推荐状态
 * @param subjectIds
 */
const handleStatus = async (subjectIds: number[]) => {
  const hide = message.loading('正在更新专题推荐状态');
  if (subjectIds.length == 0) {
    hide();
    return true;
  }
  try {
    await removeRecommendSubject(subjectIds);
    hide();
    message.success('更新专题推荐状态成功');
    return true;
  } catch (error) {
    hide();
    return false;
  }
};

const RecommendSubjectList: React.FC = () => {
  const [createModalVisible, handleModalVisible] = useState<boolean>(false);
  const [updateModalVisible, handleUpdateModalVisible] = useState<boolean>(false);
  const [showDetail, setShowDetail] = useState<boolean>(false);
  const actionRef = useRef<ActionType>();
  const [currentRow, setCurrentRow] = useState<RecommendSubjectListItem>();
  const [scope, setScope] = useState<GovernanceScopeValue>(defaultGovernanceScope);

  const showStatusConfirm = (item: RecommendSubjectListItem, status: number) => {
    confirm({
      title: `确定${status == 1 ? '推荐' : '不推荐'}${item.subjectName}专题吗？`,
      icon: <ExclamationCircleOutlined />,
      async onOk() {
        await handleStatus([item.id]);
        actionRef.current?.reload?.();
      },
      onCancel() {},
    });
  };

  const columns: ProColumns<RecommendSubjectListItem>[] = [
    {
      title: '编号',
      dataIndex: 'id',
      hideInSearch: true,
    },
    {
      title: '专题名称',
      dataIndex: 'title',
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
      title: '推荐状态',
      dataIndex: 'recommendStatus',
      hideInSearch: true,
      renderFormItem: (text, row) => {
        return (
          <Select
            value={row.value}
            options={[
              { value: '1', label: '推荐' },
              { value: '0', label: '不推荐' },
            ]}
          />
        );
      },
      render: (dom, entity) => {
        return (
          <Switch
            checked={entity.recommendStatus == 1}
            onChange={(flag) => {
              showStatusConfirm(entity, flag ? 1 : 0);
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
      title: '操作',
      dataIndex: 'option',
      valueType: 'option',
      width: 220,
      render: (_, record) => (
        <>
          <a
            key="sort"
            onClick={() => {
              handleUpdateModalVisible(true);
              setCurrentRow(record);
            }}
          >
            <EditOutlined /> 设置排序
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
        entityLabel="专题推荐"
        style={{ marginBottom: 16 }}
      />
      <Alert
        showIcon
        type="info"
        style={{ marginBottom: 16 }}
        message={`当前查询范围：${buildGovernanceScopeLabel(scope)}`}
        description="专题推荐列表和“选择专题”弹窗会共用同一治理范围，避免把别的主体专题误加进首页推荐。"
      />
      <ProTable<RecommendSubjectListItem>
        headerTitle="专题推荐列表"
        actionRef={actionRef}
        rowKey="id"
        search={{
          labelWidth: 120,
        }}
        toolBarRender={() => [
          <Button type="primary" key={'selectId'} onClick={() => handleModalVisible(true)}>
            <PlusOutlined /> 选择专题
          </Button>,
        ]}
        request={(params) => queryRecommendSubjectList({ ...params, ...toGovernancePayload(scope) })}
        columns={columns}
        rowSelection={{}}
        pagination={{ pageSize: 10 }}
        tableAlertRender={false}
      />

      <CreateForm
        key={'CreateForm'}
        onSubmit={async (value) => {
          const success = await handleAdd(value);
          if (success) {
            handleModalVisible(false);
            setCurrentRow(undefined);
            if (actionRef.current) {
              actionRef.current.reload();
            }
          }
        }}
        onCancel={() => {
          handleModalVisible(false);
          if (!showDetail) {
            setCurrentRow(undefined);
          }
        }}
        createModalVisible={createModalVisible}
        scope={scope}
      />

      <SetSortForm
        key={'SetSortModal'}
        onSubmit={async (value) => {
          const success = await handleUpdateSubjectSor(value);
          if (success) {
            handleUpdateModalVisible(false);
            setCurrentRow(undefined);
            if (actionRef.current) {
              actionRef.current.reload();
            }
          }
        }}
        onCancel={() => {
          handleUpdateModalVisible(false);
          if (!showDetail) {
            setCurrentRow(undefined);
          }
        }}
        updateModalVisible={updateModalVisible}
        values={currentRow || {}}
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
          <ProDescriptions<RecommendSubjectListItem>
            column={2}
            title={currentRow?.subjectName}
            request={async () => ({
              data: currentRow || {},
            })}
            params={{
              id: currentRow?.id,
            }}
            columns={columns as ProDescriptionsItemProps<RecommendSubjectListItem>[]}
          />
        )}
      </Drawer>
    </PageContainer>
  );
};

export default RecommendSubjectList;
