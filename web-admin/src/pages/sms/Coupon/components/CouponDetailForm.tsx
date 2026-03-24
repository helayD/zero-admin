import React, {useEffect, useRef, useState} from 'react';
import {Card, Descriptions, Modal, Select, Tag} from 'antd';
import {CouponDetailData, CouponHistoryListItem, CouponListItem} from '../data.d';
import {queryCouponDetail, queryCouponHistoryList} from "@/pages/sms/Coupon/service";

import ProTable, {type ActionType, type ProColumns} from "@ant-design/pro-table";
import moment from "moment/moment";
import type { GovernanceScopeValue } from '@/pages/system/components/governance';

export interface CreateFormProps {
  onCancel: () => void;
  onSubmit: (values: CouponListItem) => void;
  detailModalVisible: boolean;
  id: number;
  scope: GovernanceScopeValue;
}

const CouponDetailForm: React.FC<CreateFormProps> = (props) => {
  const actionRef = useRef<ActionType>();

  const [flag, setFlag] = useState<boolean>(false)
  const statusMap: Record<number, string> = {
    0: '草稿',
    1: '进行中',
    2: '已结束',
    3: '已取消',
  };
  const scopeTypeMap: Record<number, string> = {
    0: '全场通用',
    1: '指定分类',
    2: '指定商品',
  };

  const [couponDetail, setCouponDetail] = useState<CouponDetailData>({
    amount: 0,
    code: "",
    endTime: "",
    id: 0,
    minAmount: 0,
    name: "",
    description: "",
    perLimit: 0,
    totalCount: 0,
    receivedCount: 0,
    startTime: "",
    typeId: 0,
    usedCount: 0,
    status: 0,
    isEnabled: 0,
    createBy: 0,
    createTime: "",
    updateBy: 0,
    updateTime: "",
    scopeType: 0,
  });


  const {
    detailModalVisible,
    id,
    onCancel,
    scope,
  } = props;

  useEffect(() => {

    if (detailModalVisible) {
      queryCouponDetail(id, scope).then((res) => {
        setCouponDetail(res.data)
        actionRef.current?.reloadAndRest?.();
        setFlag(moment().isBefore(moment(res.data.endTime)))
      });
    }
  }, [props.detailModalVisible]);


  const columns: ProColumns<CouponHistoryListItem>[] = [
    {
      title: '编号',
      dataIndex: 'id',
      hideInSearch: true,
    },
    {
      title: '优惠码',
      dataIndex: 'couponCode',
      hideInSearch: true,
    },
    {
      title: '领取会员',
      dataIndex: 'memberNickName',
      hideInSearch: true,
    },
    {
      title: '领取方式',
      dataIndex: 'getType',
      hideInSearch: true,
      // 获取类型：0->后台赠送；1->主动获取
      render: (dom, entity) => {
        switch (entity.getType) {
          case 0:
            return <Tag>后台赠送</Tag>;
          case 1:
            return <Tag>主动获取</Tag>;
        }
        return <>未知{entity.getType}</>;
      },
    },
    {
      title: '当前状态',
      dataIndex: 'useStatus',
      renderFormItem: (text, row, index) => {
        // 使用状态：0->未使用；1->已使用；2->已过期
        return <Select
          value={row.value}
          options={[
            {value: '0', label: '未使用'},
            {value: '1', label: '已使用'},
            {value: '2', label: '已过期'},
          ]}
        />

      },
      render: (dom, entity) => {
        switch (entity.useStatus) {
          case 0:
            return <Tag>未使用</Tag>;
          case 1:
            return <Tag color={'success'}>已使用</Tag>;
          case 2:
            return <Tag color={'grey'}>已过期</Tag>;
        }
        return <>未知{entity.useStatus}</>;
      },
    },
    {
      title: '使用时间',
      dataIndex: 'useTime',
      hideInSearch: true,

    },
    {
      title: '订单号',
      dataIndex: 'orderId',
    },
  ];


  return (
    <Modal
      forceRender
      destroyOnClose
      title="优惠券详情"
      visible={detailModalVisible}
      footer={false}
      width={1200}
      onCancel={onCancel}
    >

      <Card type="inner" title="优惠券基本信息">
        <Descriptions column={4}>
          <Descriptions.Item label="名称">
            {couponDetail.name}
          </Descriptions.Item>
          <Descriptions.Item label="优惠券类型">
            {couponDetail.typeName || couponDetail.typeId}
          </Descriptions.Item>
          <Descriptions.Item label="优惠券码">
            {couponDetail.code || '-'}
          </Descriptions.Item>
          <Descriptions.Item label="适用范围">
            {scopeTypeMap[couponDetail.scopeType] || '全场通用'}
          </Descriptions.Item>
          <Descriptions.Item label="面值">
            ¥{couponDetail.amount}
          </Descriptions.Item>
          <Descriptions.Item label="使用门槛">
            {couponDetail.minAmount ? `满¥${couponDetail.minAmount}` : '无门槛'}
          </Descriptions.Item>
          <Descriptions.Item label="状态">
            {statusMap[couponDetail.status] || '未知'}
            {flag ? '' : '（已过期）'}
          </Descriptions.Item>
          <Descriptions.Item label="启用">
            {couponDetail.isEnabled === 1 ? '启用' : '停用'}
          </Descriptions.Item>

          <Descriptions.Item label="发放总量">
            {couponDetail.totalCount}
          </Descriptions.Item>
          <Descriptions.Item label="已领取">
            {couponDetail.receivedCount}
          </Descriptions.Item>
          <Descriptions.Item label="待领取">
            {couponDetail.totalCount - couponDetail.receivedCount}
          </Descriptions.Item>
          <Descriptions.Item label="已使用">
            {couponDetail.usedCount}
          </Descriptions.Item>
          <Descriptions.Item label="每人限领">
            {couponDetail.perLimit}张
          </Descriptions.Item>
          <Descriptions.Item label="有效期">
            {moment(couponDetail.startTime).format('YYYY-MM-DD')}
            {' ~ '}
            {moment(couponDetail.endTime).format('YYYY-MM-DD')}
          </Descriptions.Item>
          <Descriptions.Item label="说明" span={2}>
            {couponDetail.description || '-'}
          </Descriptions.Item>
        </Descriptions>
      </Card>
      <Card
        style={{marginTop: 16}}
        type="inner"
        title="优惠券领取详情"
      >
        <ProTable<CouponHistoryListItem>
          toolBarRender={false}
          actionRef={actionRef}
          rowKey="id"
          search={{
            labelWidth: 80,
            span: 8
          }}
          request={(params)=>{
            return queryCouponHistoryList({...params,couponId:id})
          }}
          columns={columns}
          rowSelection={false}
          pagination={{pageSize: 6}}
        />
      </Card>

    </Modal>
  );
};

export default CouponDetailForm;
