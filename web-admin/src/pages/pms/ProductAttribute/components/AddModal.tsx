import React, { useEffect, useState } from 'react';
import { Form, Input, InputNumber, Modal, Radio, Select, message } from 'antd';
import type { ProductAttributeListItem } from '../data.d';
import type { GovernanceScopeValue } from '@/pages/system/components/governance';
import { toGovernancePayload } from '@/pages/system/components/governance';
import type { ProductAttributeGroupListItem } from '@/pages/pms/ProductAttributeGroup/data';
import { queryProductAttributeGroupList } from '@/pages/pms/ProductAttributeGroup/service';

export interface AddModalProps {
  onCancel: () => void;
  onSubmit: (values: ProductAttributeListItem) => void;
  addVisible: boolean;
  scope?: GovernanceScopeValue;
}

const FormItem = Form.Item;
const formLayout = { labelCol: { span: 7 }, wrapperCol: { span: 13 } };

const AddModal: React.FC<AddModalProps> = (props) => {
  const [form] = Form.useForm();
  const [groupItems, setGroupItems] = useState<ProductAttributeGroupListItem[]>([]);
  const { onSubmit, onCancel, addVisible, scope } = props;

  useEffect(() => {
    if (form && !addVisible) {
      form.resetFields();
      return;
    }
    queryProductAttributeGroupList({ pageSize: 100, current: 1, ...toGovernancePayload(scope) }).then((res) => {
      if (res.code === '000000') {
        setGroupItems(res.data || []);
      } else {
        message.error(res.message || res.msg || '加载属性分组失败');
      }
    });
  }, [addVisible, form, scope]);

  return (
    <Modal forceRender destroyOnClose title="新增" open={addVisible} okText="保存" onOk={() => form.submit()} onCancel={onCancel}>
      <Form {...formLayout} form={form} onFinish={(values) => onSubmit?.(values)}>
        <FormItem name="groupId" label="属性分组" rules={[{ required: true, message: '请选择属性分组!' }]}> 
          <Select placeholder="请选择属性分组">
            {groupItems.map((item) => (
              <Select.Option key={item.id} value={item.id}>{item.name}</Select.Option>
            ))}
          </Select>
        </FormItem>
        <FormItem name="name" label="属性名称" rules={[{ required: true, message: '请输入属性名称!' }]}>
          <Input placeholder="请输入属性名称!" />
        </FormItem>
        <FormItem name="inputType" label="输入类型" rules={[{ required: true, message: '请选择输入类型!' }]}>
          <Radio.Group>
            <Radio value={1}>手动输入</Radio>
            <Radio value={2}>单选</Radio>
            <Radio value={3}>多选</Radio>
          </Radio.Group>
        </FormItem>
        <FormItem name="valueType" label="值类型" rules={[{ required: true, message: '请选择值类型!' }]}>
          <Radio.Group>
            <Radio value={1}>文本</Radio>
            <Radio value={2}>数字</Radio>
            <Radio value={3}>日期</Radio>
          </Radio.Group>
        </FormItem>
        <FormItem name="inputList" label="可选值列表" rules={[{ required: true, message: '请输入可选值列表，用逗号分隔!' }]}>
          <Input placeholder="请输入可选值列表，用逗号分隔!" />
        </FormItem>
        <FormItem name="unit" label="单位" rules={[{ required: true, message: '请输入单位!' }]}>
          <Input placeholder="请输入单位!" />
        </FormItem>
        <FormItem name="isRequired" label="是否必填" rules={[{ required: true, message: '请选择是否必填!' }]}>
          <Radio.Group><Radio value={0}>否</Radio><Radio value={1}>是</Radio></Radio.Group>
        </FormItem>
        <FormItem name="isSearchable" label="是否支持搜索" rules={[{ required: true, message: '请选择是否支持搜索!' }]}>
          <Radio.Group><Radio value={0}>否</Radio><Radio value={1}>是</Radio></Radio.Group>
        </FormItem>
        <FormItem name="isShow" label="是否显示" rules={[{ required: true, message: '请选择是否显示!' }]}>
          <Radio.Group><Radio value={0}>否</Radio><Radio value={1}>是</Radio></Radio.Group>
        </FormItem>
        <FormItem name="sort" label="排序" rules={[{ required: true, message: '请输入排序!' }]}>
          <InputNumber style={{ width: 255 }} />
        </FormItem>
        <FormItem name="status" label="状态" rules={[{ required: true, message: '请选择状态!' }]}>
          <Radio.Group><Radio value={0}>禁用</Radio><Radio value={1}>正常</Radio></Radio.Group>
        </FormItem>
      </Form>
    </Modal>
  );
};

export default AddModal;
