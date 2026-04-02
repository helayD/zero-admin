import { CheckCircleOutlined, DeleteOutlined, EyeOutlined, StopOutlined } from '@ant-design/icons';
import { Button, DatePicker, Drawer, Image, message, Modal, Select, Space, Tag, Typography } from 'antd';
import React, { useRef, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import type { ActionType, ProColumns } from '@ant-design/pro-table';
import ProTable from '@ant-design/pro-table';
import type { ProDescriptionsItemProps } from '@ant-design/pro-descriptions';
import ProDescriptions from '@ant-design/pro-descriptions';
import type { CommentDetailData, CommentListItem } from './data.d';
import { batchUpdateComment, queryCommentDetail, queryCommentList, updateComment } from './service';
import GovernanceScopeBar from '@/pages/system/components/GovernanceScopeBar';
import {
  buildGovernanceScopeLabel,
  governanceScopeColor,
  toGovernancePayload,
  type GovernanceScopeValue,
} from '@/pages/system/components/governance';

const { Text, Paragraph } = Typography;
const { RangePicker } = DatePicker;

const showStatusOptions = [
  { label: '全部', value: -1 },
  { label: '待审核', value: 0 },
  { label: '已通过', value: 1 },
];

const renderShowStatusTag = (status: number) => {
  switch (status) {
    case 0:
      return <Tag color="orange">待审核</Tag>;
    case 1:
      return <Tag color="green">已通过</Tag>;
    default:
      return <Tag>未知</Tag>;
  }
};

const renderStarRating = (star: number) => (
  <Space size={2}>
    {Array.from({ length: 5 }, (_, i) => (
      <span key={i} style={{ color: i < star ? '#FF9900' : '#E0E0E0', fontSize: 14 }}>
        ★
      </span>
    ))}
  </Space>
);

const ProductCommentPage: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const [scope, setScope] = useState<GovernanceScopeValue>({});
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [selectedComment, setSelectedComment] = useState<CommentDetailData | null>(null);
  const [detailLoading, setDetailLoading] = useState(false);
  // Review Fix M-3: 批量选择状态
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);

  const handleScopeChange = (newScope: GovernanceScopeValue) => {
    setScope(newScope);
  };

  const openDetail = async (id: string) => {
    setDetailLoading(true);
    setDrawerVisible(true);
    try {
      const res = await queryCommentDetail({ id }, toGovernancePayload(scope));
      if (res.code === 0) {
        setSelectedComment(res.data);
      } else {
        message.error(res.message || '获取评价详情失败');
      }
    } catch (e) {
      message.error('获取评价详情失败');
    } finally {
      setDetailLoading(false);
    }
  };

  const handleUpdateStatus = async (id: string, showStatus: number, action: string) => {
    try {
      const res = await updateComment(
        { id, showStatus, updateBy: 'admin' },
        toGovernancePayload(scope),
      );
      if (res.code === 0) {
        message.success(`评价${action}成功`);
        actionRef.current?.reload();
        if (selectedComment?.id === id) {
          openDetail(id);
        }
      } else {
        message.error(res.message || `${action}失败`);
      }
    } catch (e) {
      message.error(`${action}失败`);
    }
  };

  const confirmUpdateStatus = (id: string, showStatus: number, action: string) => {
    Modal.confirm({
      title: `确认${action}`,
      content: `确定要${action}该评价吗？`,
      okText: '确认',
      cancelText: '取消',
      onOk: () => handleUpdateStatus(id, showStatus, action),
    });
  };

  // Review Fix M-3: 批量通过待审核评价
  const handleBatchApprove = async () => {
    if (selectedRowKeys.length === 0) {
      message.warning('请先选择要操作的评价');
      return;
    }
    Modal.confirm({
      title: '批量通过审核',
      content: `确定要通过选中的 ${selectedRowKeys.length} 条评价吗？`,
      okText: '确认',
      cancelText: '取消',
      onOk: async () => {
        try {
          const res = await batchUpdateComment(
            selectedRowKeys as string[],
            1, // showStatus=1 通过
            toGovernancePayload(scope),
          );
          if (res.code === 0) {
            message.success(`已通过 ${selectedRowKeys.length} 条评价`);
            setSelectedRowKeys([]);
            actionRef.current?.reload();
          } else {
            message.error(res.message || '批量操作失败');
          }
        } catch {
          message.error('批量操作失败');
        }
      },
    });
  };

  // Review Fix M-3: 批量屏蔽评价
  const handleBatchBlock = async () => {
    if (selectedRowKeys.length === 0) {
      message.warning('请先选择要操作的评价');
      return;
    }
    Modal.confirm({
      title: '批量屏蔽评价',
      content: `确定要屏蔽选中的 ${selectedRowKeys.length} 条评价吗？`,
      okText: '确认',
      cancelText: '取消',
      onOk: async () => {
        try {
          const res = await batchUpdateComment(
            selectedRowKeys as string[],
            0, // showStatus=0 屏蔽
            toGovernancePayload(scope),
          );
          if (res.code === 0) {
            message.success(`已屏蔽 ${selectedRowKeys.length} 条评价`);
            setSelectedRowKeys([]);
            actionRef.current?.reload();
          } else {
            message.error(res.message || '批量操作失败');
          }
        } catch {
          message.error('批量操作失败');
        }
      },
    });
  };

  const columns: ProColumns<CommentListItem>[] = [
    {
      title: '评价ID',
      dataIndex: 'id',
      width: 120,
      copyable: true,
      hideInSearch: true,
    },
    {
      title: '商品名称/ID',
      dataIndex: 'productId',
      hideInSearch: true,
      width: 180,
      render: (_, record) => (
        <Space direction="vertical" size={2}>
          <Text>{record.productAttribute || '商品评价'}</Text>
          <Text type="secondary" style={{ fontSize: 12 }}>
            ID: {record.productId}
          </Text>
        </Space>
      ),
    },
    {
      title: '评价用户',
      dataIndex: 'memberNickName',
      width: 120,
      hideInSearch: true,
      render: (_, record) => (
        <Space>
          <Text>{record.memberNickName || '匿名用户'}</Text>
        </Space>
      ),
    },
    {
      title: '评分',
      dataIndex: 'star',
      width: 100,
      hideInSearch: true,
      render: (_, record) => renderStarRating(record.star),
    },
    {
      title: '评价内容',
      dataIndex: 'content',
      width: 200,
      ellipsis: true,
      hideInSearch: true,
      render: (_, record) => (
        <Paragraph
          ellipsis={{ rows: 2, expandable: true, symbol: '展开' }}
          style={{ margin: 0, maxWidth: 200 }}
        >
          {record.content}
        </Paragraph>
      ),
    },
    {
      title: '评价图片',
      dataIndex: 'pics',
      width: 120,
      hideInSearch: true,
      render: (_, record) => {
        const pics = record.pics?.split(',').filter((p) => p.trim()) || [];
        if (pics.length === 0) return <Text type="secondary">无</Text>;
        return (
          <Image.PreviewGroup>
            <Space>
              {pics.slice(0, 2).map((pic, idx) => (
                <Image
                  key={idx}
                  src={pic}
                  width={40}
                  height={40}
                  style={{ objectFit: 'cover', borderRadius: 4 }}
                />
              ))}
              {pics.length > 2 && (
                <Tag>+{pics.length - 2}</Tag>
              )}
            </Space>
          </Image.PreviewGroup>
        );
      },
    },
    {
      title: '状态',
      dataIndex: 'showStatus',
      width: 100,
      hideInSearch: true,
      render: (_, record) => renderShowStatusTag(record.showStatus),
    },
    {
      title: '回复数',
      dataIndex: 'replayCount',
      width: 80,
      hideInSearch: true,
      render: (_, record) => (
        <Text type={record.replayCount > 0 ? 'processing' : 'secondary'}>
          {record.replayCount}
        </Text>
      ),
    },
    {
      title: '评价时间',
      dataIndex: 'createTime',
      width: 160,
      hideInSearch: true,
      valueType: 'dateTime',
      render: (_, record) => (
        <Text type="secondary" style={{ fontSize: 12 }}>
          {record.createTime}
        </Text>
      ),
    },
    {
      title: '操作',
      valueType: 'option',
      width: 160,
      fixed: 'right',
      render: (_, record) => (
        <Space>
          <Button
            type="text"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => openDetail(record.id)}
          >
            详情
          </Button>
          {record.showStatus === 0 ? (
            <Button
              type="text"
              size="small"
              icon={<CheckCircleOutlined />}
              onClick={() => confirmUpdateStatus(record.id, 1, '通过审核')}
            >
              通过
            </Button>
          ) : (
            <Button
              type="text"
              size="small"
              danger
              icon={<StopOutlined />}
              onClick={() => confirmUpdateStatus(record.id, 0, '屏蔽')}
            >
              屏蔽
            </Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <PageContainer>
      <GovernanceScopeBar scope={scope} onScopeChange={handleScopeChange} />
      <ProTable<CommentListItem>
        headerTitle="商品评价列表"
        actionRef={actionRef}
        rowKey="id"
        scroll={{ x: 1200 }}
        search={{
          labelWidth: 'auto',
        }}
        rowSelection={{
          selectedRowKeys,
          onChange: (keys) => setSelectedRowKeys(keys),
        }}
        toolBarRender={() => [
          // Review Fix M-3: 批量操作按钮
          <Button
            key="batchApprove"
            type="primary"
            icon={<CheckCircleOutlined />}
            disabled={selectedRowKeys.length === 0}
            onClick={handleBatchApprove}
          >
            批量通过{selectedRowKeys.length > 0 ? ` (${selectedRowKeys.length})` : ''}
          </Button>,
          <Button
            key="batchBlock"
            danger
            icon={<StopOutlined />}
            disabled={selectedRowKeys.length === 0}
            onClick={handleBatchBlock}
          >
            批量屏蔽{selectedRowKeys.length > 0 ? ` (${selectedRowKeys.length})` : ''}
          </Button>,
          <Select
            key="showStatus"
            placeholder="筛选状态"
            options={showStatusOptions}
            style={{ width: 120 }}
            allowClear
            onChange={(value) => {
              actionRef.current?.reloadAndRest?.();
            }}
          />,
        ]}
        request={async (params, sort) => {
          const pageSize = params.pageSize || 20;
          const current = params.current || 1;
          const res = await queryCommentList(
            {
              productId: params.productId as unknown as number,
              showStatus: params.showStatus as unknown as number,
              pageSize,
              current,
            },
            toGovernancePayload(scope),
          );
          return {
            data: res.data?.data?.list || [],
            total: res.data?.data?.pagination?.total || 0,
            success: res.code === 0,
          };
        }}
        columns={columns}
        pagination={{
          defaultPageSize: 20,
          showSizeChanger: true,
          showQuickJumper: true,
        }}
      />

      {/* 评价详情抽屉 */}
      <Drawer
        title="评价详情"
        width={640}
        open={drawerVisible}
        onClose={() => {
          setDrawerVisible(false);
          setSelectedComment(null);
        }}
        extra={
          selectedComment && (
            <Space>
              {selectedComment.showStatus === 0 ? (
                <Button
                  type="primary"
                  icon={<CheckCircleOutlined />}
                  onClick={() =>
                    handleUpdateStatus(selectedComment.id, 1, '通过审核')
                  }
                >
                  通过审核
                </Button>
              ) : (
                <Button
                  danger
                  icon={<StopOutlined />}
                  onClick={() =>
                    handleUpdateStatus(selectedComment.id, 0, '屏蔽')
                  }
                >
                  屏蔽
                </Button>
              )}
            </Space>
          )
        }
      >
        {detailLoading ? (
          <div style={{ textAlign: 'center', padding: 40 }}>加载中...</div>
        ) : selectedComment ? (
          <ProDescriptions column={1} bordered size="small">
            <ProDescriptions.Item
              label="评价ID"
              span={1}
            >
              {selectedComment.id}
            </ProDescriptions.Item>
            <ProDescriptions.Item label="商品信息">
              <Space direction="vertical" size={2}>
                <Text>{selectedComment.productAttribute || '商品评价'}</Text>
                <Text type="secondary">商品ID: {selectedComment.productId}</Text>
              </Space>
            </ProDescriptions.Item>
            <ProDescriptions.Item label="评价用户">
              <Space>
                <Text>{selectedComment.memberNickName || '匿名用户'}</Text>
                <Text type="secondary">| ID: {selectedComment.memberId}</Text>
              </Space>
            </ProDescriptions.Item>
            <ProDescriptions.Item label="评分">
              {renderStarRating(selectedComment.star)}
            </ProDescriptions.Item>
            <ProDescriptions.Item label="评价内容" span={1}>
              <Paragraph style={{ margin: 0 }}>{selectedComment.content}</Paragraph>
            </ProDescriptions.Item>
            <ProDescriptions.Item label="评价图片">
              {selectedComment.pics ? (
                <Image.PreviewGroup>
                  <Space wrap>
                    {selectedComment.pics.split(',').filter((p) => p.trim()).map((pic, idx) => (
                      <Image
                        key={idx}
                        src={pic}
                        width={80}
                        height={80}
                        style={{ objectFit: 'cover', borderRadius: 4 }}
                      />
                    ))}
                  </Space>
                </Image.PreviewGroup>
              ) : (
                <Text type="secondary">无</Text>
              )}
            </ProDescriptions.Item>
            <ProDescriptions.Item label="状态">
              {renderShowStatusTag(selectedComment.showStatus)}
            </ProDescriptions.Item>
            <ProDescriptions.Item label="评价时间">
              {selectedComment.createTime}
            </ProDescriptions.Item>
            <ProDescriptions.Item label="评价IP">
              {selectedComment.memberIp || '-'}
            </ProDescriptions.Item>
            {selectedComment.updateBy && (
              <ProDescriptions.Item label="处理人">
                {selectedComment.updateBy}
              </ProDescriptions.Item>
            )}
            <ProDescriptions.Item label="回复数">
              {selectedComment.replayCount}
            </ProDescriptions.Item>
            {/* Review Fix M-4: 展示回复列表 */}
            {selectedComment.replays && selectedComment.replays.length > 0 && (
              <ProDescriptions.Item label="回复列表" span={1}>
                <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
                  {selectedComment.replays.map((reply) => (
                    <div
                      key={reply.id}
                      style={{
                        // reply.type: 0=会员回复, 1=管理员回复
                        background: reply.type === 1 ? '#f0f5ff' : '#fafafa',
                        border: '1px solid #f0f0f0',
                        borderRadius: 6,
                        padding: '10px 12px',
                      }}
                    >
                      <Space style={{ marginBottom: 4 }}>
                        {/* reply.type: 0=会员, 1=管理员 */}
                        <Tag color={reply.type === 1 ? 'blue' : 'green'}>
                          {reply.type === 1 ? '管理员' : '会员'}
                        </Tag>
                        <Text strong style={{ fontSize: 13 }}>
                          {reply.memberNickName || '匿名用户'}
                        </Text>
                        <Text type="secondary" style={{ fontSize: 12 }}>
                          {reply.createTime}
                        </Text>
                      </Space>
                      <Paragraph style={{ margin: '4px 0 0', fontSize: 13 }}>
                        {reply.content}
                      </Paragraph>
                    </div>
                  ))}
                </div>
              </ProDescriptions.Item>
            )}
          </ProDescriptions>
        ) : (
          <div style={{ textAlign: 'center', padding: 40 }}>
            暂无数据
          </div>
        )}
      </Drawer>
    </PageContainer>
  );
};

export default ProductCommentPage;
