import React, { useRef, useState } from 'react';
import { Alert, Button, Descriptions, Drawer, message, Modal, Space, Tag } from 'antd';
import type { ActionType, ProColumns } from '@ant-design/pro-table';
import ProTable from '@ant-design/pro-table';
import { PageContainer } from '@ant-design/pro-layout';
import { DeleteOutlined, EditOutlined, EyeOutlined, PlusOutlined } from '@ant-design/icons';
import GovernanceScopeBar from '@/pages/system/components/GovernanceScopeBar';
import {
  buildGovernanceScopeLabel,
  defaultGovernanceScope,
  type GovernanceScopeValue,
  toGovernancePayload,
} from '@/pages/system/components/governance';
import { readErrorMessage } from '@/pages/system/components/requestError';
import type {
  DrawActivityFormValues,
  DrawActivityListItem,
  PreviewDrawActivityPublishReadinessData,
} from './data.d';
import {
  addDrawActivity,
  previewDrawActivityPublishReadiness,
  queryDrawActivityDetail,
  queryDrawActivityList,
  removeDrawActivity,
  updateDrawActivity,
  updateDrawActivityStatus,
} from './service';
import ActivityDrawer from './components/ActivityDrawer';
import { WheelPreview, WHEEL_COLORS } from './components/WheelPreview';
import { buildPreviewMessages } from './helper';

const { confirm, info } = Modal;

const statusMap: Record<number, { text: string; color: string }> = {
  0: { text: '草稿/下线', color: 'default' },
  1: { text: '已发布', color: 'green' },
  2: { text: '已归档', color: 'purple' },
};

const auditMap: Record<number, { text: string; color: string }> = {
  0: { text: '缺失', color: 'default' },
  1: { text: '待审批', color: 'orange' },
  2: { text: '已通过', color: 'green' },
  3: { text: '已拒绝', color: 'red' },
};

const readinessMap: Record<number, { text: string; color: string }> = {
  0: { text: '缺少信息', color: 'red' },
  1: { text: '可发布', color: 'green' },
  2: { text: '待审批', color: 'orange' },
  3: { text: '已下线', color: 'default' },
};

const DigitalCardActivity: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const [scope, setScope] = useState<GovernanceScopeValue>(defaultGovernanceScope);
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [drawerTitle, setDrawerTitle] = useState('新建抽卡活动');
  const [currentItem, setCurrentItem] = useState<DrawActivityFormValues | undefined>(undefined);
  const [detailVisible, setDetailVisible] = useState(false);
  const [detailItem, setDetailItem] = useState<DrawActivityListItem | undefined>(undefined);

  const loadDetail = async (record: DrawActivityListItem) => {
    const result = await queryDrawActivityDetail(record.id as number, toGovernancePayload(scope));
    return result?.data as DrawActivityListItem;
  };

  const handleAdd = async (values: DrawActivityFormValues) => {
    try {
      await addDrawActivity({ ...values, ...toGovernancePayload(scope) });
      message.success('抽卡活动创建成功');
      setDrawerVisible(false);
      setCurrentItem(undefined);
      actionRef.current?.reload();
      return true;
    } catch (error) {
      message.error(readErrorMessage(error, '创建抽卡活动失败，请检查卡池、模板与概率配置'));
      return false;
    }
  };

  const handleUpdate = async (values: DrawActivityFormValues) => {
    try {
      await updateDrawActivity({ ...values, ...toGovernancePayload(scope) });
      message.success('抽卡活动更新成功');
      setDrawerVisible(false);
      setCurrentItem(undefined);
      actionRef.current?.reload();
      return true;
    } catch (error) {
      message.error(readErrorMessage(error, '更新抽卡活动失败，请检查合规字段与主体范围'));
      return false;
    }
  };

  const handleDelete = (record: DrawActivityListItem) => {
    confirm({
      title: '是否删除当前抽卡活动？',
      content: `活动「${record.name}」主体：${buildGovernanceScopeLabel(record)}。删除后会一并下线该活动下的卡池配置。`,
      onOk: async () => {
        try {
          await removeDrawActivity({
            ids: [record.id as number],
            ...toGovernancePayload(scope),
          });
          message.success('删除成功');
          actionRef.current?.reload();
        } catch (error) {
          message.error(readErrorMessage(error, '删除抽卡活动失败'));
        }
      },
    });
  };

  const handlePreview = async (record: DrawActivityListItem) => {
    try {
      const result = await previewDrawActivityPublishReadiness(
        record.id as number,
        toGovernancePayload(scope),
      );
      const preview = result?.data as PreviewDrawActivityPublishReadinessData;
      const messages = buildPreviewMessages(preview);
      if (preview?.readyToPublish) {
        confirm({
          title: '发布预检通过',
          content: `活动「${record.name}」主体：${buildGovernanceScopeLabel(record)}。确认将活动发布上线吗？`,
          onOk: async () => {
            await updateDrawActivityStatus({
              ids: [record.id as number],
              status: 1,
              ...toGovernancePayload(scope),
            });
            message.success('抽卡活动已发布');
            actionRef.current?.reload();
          },
        });
        return;
      }
      info({
        title: `发布预检未通过 · ${preview?.readinessLabel || '缺少信息'}`,
        width: 720,
        content: (
          <div>
            <p>{preview?.summary || '请先处理以下失败项后再发布。'}</p>
            <ul>
              {messages.map((item) => (
                <li key={item}>{item}</li>
              ))}
            </ul>
          </div>
        ),
      });
    } catch (error) {
      message.error(readErrorMessage(error, '执行发布预检失败'));
    }
  };

  const columns: ProColumns<DrawActivityListItem>[] = [
    {
      title: '活动名称',
      dataIndex: 'name',
      render: (_, record) => (
        <a
          onClick={async () => {
            const detail = await loadDetail(record);
            setDetailItem(detail);
            setDetailVisible(true);
          }}
        >
          {record.name}
        </a>
      ),
    },
    {
      title: '活动编码',
      dataIndex: 'activityCode',
    },
    {
      title: '主体范围',
      dataIndex: 'scopeType',
      render: (_, record) => buildGovernanceScopeLabel(record),
    },
    {
      title: '可发布状态',
      dataIndex: 'publishReadiness',
      render: (_, record) => {
        const target = readinessMap[record.publishReadiness || 0] || readinessMap[0];
        return <Tag color={target.color}>{record.readinessLabel || target.text}</Tag>;
      },
    },
    {
      title: '审批状态',
      dataIndex: 'auditStatus',
      render: (_, record) => {
        const target = auditMap[record.auditStatus || 0] || auditMap[0];
        return <Tag color={target.color}>{target.text}</Tag>;
      },
    },
    {
      title: '活动状态',
      dataIndex: 'status',
      render: (_, record) => {
        const target = statusMap[record.status || 0] || statusMap[0];
        return <Tag color={target.color}>{target.text}</Tag>;
      },
    },
    {
      title: '最近失败摘要',
      dataIndex: 'publishFailureSummary',
      hideInSearch: true,
      render: (value) => value || '-',
    },
    {
      title: '首页入口',
      dataIndex: 'homeEntryTitle',
      hideInSearch: true,
      render: (_, record) => (record.showOnHome ? record.homeEntryTitle || '已配置' : '未展示'),
    },
    {
      title: '更新时间',
      dataIndex: 'updateTime',
      hideInSearch: true,
    },
    {
      title: '操作',
      valueType: 'option',
      width: 280,
      render: (_, record) => [
        <a
          key="view"
          onClick={async () => {
            const detail = await loadDetail(record);
            setDetailItem(detail);
            setDetailVisible(true);
          }}
        >
          <EyeOutlined /> 查看
        </a>,
        <a
          key="edit"
          onClick={async () => {
            const detail = await loadDetail(record);
            setCurrentItem(detail as DrawActivityFormValues);
            setDrawerTitle('编辑抽卡活动');
            setDrawerVisible(true);
          }}
        >
          <EditOutlined /> 编辑
        </a>,
        <a key="preview" onClick={() => handlePreview(record)}>
          预检发布
        </a>,
        <a
          key="offline"
          onClick={async () => {
            try {
              await updateDrawActivityStatus({
                ids: [record.id as number],
                status: 0,
                ...toGovernancePayload(scope),
              });
              message.success('已切换为草稿/下线状态');
              actionRef.current?.reload();
            } catch (error) {
              message.error(readErrorMessage(error, '切换活动状态失败'));
            }
          }}
        >
          下线
        </a>,
        <a key="delete" onClick={() => handleDelete(record)} style={{ color: '#ff4d4f' }}>
          <DeleteOutlined /> 删除
        </a>,
      ],
    },
  ];

  return (
    <PageContainer>
      <Space direction="vertical" style={{ width: '100%' }} size={16}>
        <GovernanceScopeBar value={scope} onChange={setScope} entityLabel="抽卡活动配置" />
        <Alert
          showIcon
          type="warning"
          message="Consequence Preview"
          description={`当前正在维护 ${buildGovernanceScopeLabel(scope)} 的抽卡活动、卡池与模板配置。发布前会由服务端统一执行合规、版权、审批和概率预检，页面不会本地伪造“可发布”状态。`}
        />
        <ProTable<DrawActivityListItem>
          actionRef={actionRef}
          rowKey="id"
          headerTitle="抽卡活动工作台"
          columns={columns}
          request={(params) => queryDrawActivityList({ ...params, ...toGovernancePayload(scope) })}
          search={{ labelWidth: 120 }}
          pagination={{ pageSize: 10 }}
          toolBarRender={() => [
            <Button
              key="new"
              type="primary"
              onClick={() => {
                setDrawerTitle('新建抽卡活动');
                setCurrentItem(undefined);
                setDrawerVisible(true);
              }}
            >
              <PlusOutlined /> 新建活动
            </Button>,
          ]}
        />
      </Space>

      <ActivityDrawer
        visible={drawerVisible}
        title={drawerTitle}
        current={currentItem}
        onCancel={() => {
          setDrawerVisible(false);
          setCurrentItem(undefined);
        }}
        onSubmit={currentItem?.id ? handleUpdate : handleAdd}
      />

      <Drawer
        width={720}
        visible={detailVisible}
        destroyOnClose
        onClose={() => {
          setDetailVisible(false);
          setDetailItem(undefined);
        }}
        title="抽卡活动详情"
      >
        {detailItem && (
          <Descriptions bordered column={1} size="small">
            <Descriptions.Item label="活动名称">{detailItem.name}</Descriptions.Item>
            <Descriptions.Item label="活动编码">{detailItem.activityCode}</Descriptions.Item>
            <Descriptions.Item label="主体范围">
              {buildGovernanceScopeLabel(detailItem)}
            </Descriptions.Item>
            <Descriptions.Item label="可发布状态">
              {detailItem.readinessLabel || readinessMap[detailItem.publishReadiness || 0]?.text}
            </Descriptions.Item>
            <Descriptions.Item label="活动时间">
              {detailItem.startTime} ~ {detailItem.endTime}
            </Descriptions.Item>
            <Descriptions.Item label="规则摘要">
              {detailItem.ruleSummary || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="首页入口">
              {detailItem.homeEntry?.showOnHome ? detailItem.homeEntry?.homeEntryTitle || '已启用' : '未启用'}
            </Descriptions.Item>
            <Descriptions.Item label="模板数量">{detailItem.templates?.length || 0}</Descriptions.Item>
            <Descriptions.Item label="卡池数量">{detailItem.pools?.length || 0}</Descriptions.Item>
            {(detailItem.pools || []).map((pool, poolIdx) => (
              <Descriptions.Item
                key={pool.id || poolIdx}
                label={`卡池 ${poolIdx + 1}：${pool.poolName}`}
              >
                <div style={{ display: 'flex', gap: 16, alignItems: 'flex-start' }}>
                  <div style={{ flex: 1 }}>
                    <div
                      style={{
                        display: 'flex',
                        gap: 8,
                        marginBottom: 4,
                        color: '#888',
                        fontSize: 12,
                      }}
                    >
                      <span style={{ minWidth: 48 }}>格位</span>
                      <span style={{ flex: 1 }}>模板名称</span>
                      <span style={{ width: 50 }}>稀有度</span>
                      <span style={{ width: 60 }}>概率</span>
                      <span style={{ width: 60 }}>剩余可发</span>
                    </div>
                    {(pool.templates || [])
                      .slice()
                      .sort((a, b) => (a.slotIndex || 0) - (b.slotIndex || 0))
                      .map((tpl, tplIdx) => (
                        <div
                          key={tpl.id || tplIdx}
                          style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}
                        >
                          <Tag
                            color={WHEEL_COLORS[(tpl.slotIndex - 1) % WHEEL_COLORS.length]}
                            style={{ minWidth: 44, textAlign: 'center', margin: 0 }}
                          >
                            格{tpl.slotIndex}
                          </Tag>
                          <span style={{ flex: 1, fontSize: 13 }}>{tpl.templateName || '-'}</span>
                          <Tag style={{ width: 50, textAlign: 'center', margin: 0 }}>
                            {tpl.rarity || '-'}
                          </Tag>
                          <span style={{ width: 60, color: '#595959', fontSize: 13 }}>
                            {((tpl.probability || 0) * 100).toFixed(1)}%
                          </span>
                          <span style={{ width: 60, color: '#595959', fontSize: 13 }}>
                            {tpl.remainingLimit ?? '-'}
                          </span>
                        </div>
                      ))}
                  </div>
                  <div style={{ flexShrink: 0, textAlign: 'center' }}>
                    <div style={{ fontSize: 11, color: '#888', marginBottom: 4 }}>概率分布</div>
                    <WheelPreview
                      slots={(pool.templates || [])
                        .slice()
                        .sort((a, b) => (a.slotIndex || 0) - (b.slotIndex || 0))}
                      size={140}
                    />
                  </div>
                </div>
              </Descriptions.Item>
            ))}
            <Descriptions.Item label="最近失败摘要">
              {detailItem.publishFailureSummary || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="预检失败项">
              {(detailItem.readinessItems || []).length > 0
                ? (detailItem.readinessItems || []).map((item) => (
                    <div key={`${item.code}-${item.field}`}>
                      {item.blocking ? '阻断' : '提示'}：{item.message}
                    </div>
                  ))
                : '当前无失败项'}
            </Descriptions.Item>
          </Descriptions>
        )}
      </Drawer>
    </PageContainer>
  );
};

export default DigitalCardActivity;
