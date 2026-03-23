import React, { useEffect, useState } from 'react';
import { Form, Input, InputNumber, Modal, Radio, Select, message } from 'antd';
import type { ProductSpecValueListItem } from '../data.d';
import type { GovernanceScopeValue } from '@/pages/system/components/governance';
import { toGovernancePayload } from '@/pages/system/components/governance';
import type { ProductSpecListItem } from '@/pages/pms/ProductSpec/data';
import { queryProductSpecList } from '@/pages/pms/ProductSpec/service';

export interface AddModalProps {
  onCancel: () => void;
  onSubmit: (values: ProductSpecValueListItem) => void;
  addVisible: boolean;
  scope?: GovernanceScopeValue;
}

const FormItem = Form.Item;
const formLayout = { labelCol: { span: 7 }, wrapperCol: { span: 13 } };

const AddModal: React.FC<AddModalProps> = ({ onSubmit, onCancel, addVisible, scope }) => {
  const [form] = Form.useForm();
  const [specItems, setSpecItems] = useState<ProductSpecListItem[]>([]);

  useEffect(() => {
    if (form && !addVisible) {
      form.resetFields();
      return;
    }
    queryProductSpecList({ pageSize: 100, current: 1, ...toGovernancePayload(scope) }).then((res) => {
      if (res.code === '000000') {
        setSpecItems(res.data || []);
      } else {
        message.error(res.message || res.msg || '加载商品规格失败');
      }
    });
  }, [addVisible, form, scope]);

  return (
    <Modal forceRender destroyOnClose title="新增" open={addVisible} okText="保存" onOk={() => form.submit()} onCancel={onCancel}>
      <Form {...formLayout} form={form} onFinish={(values) => onSubmit?.(values)}>
        <FormItem name="specId" label="规格" rules={[{ required: true, message: '请选择规格!' }]}> 
          <Select placeholder="请选择规格">
            {specItems.map((item) => (
              <Select.Option key={item.id} value={item.id}>{item.name}</Select.Option>
            ))}
          </Select>
        </FormItem>
        <FormItem name="value" label="规格值" rules={[{ required: true, message: '请输入规格值!' }]}><Input placeholder="请输入规格值!" /></FormItem>
        <FormItem name="sort" label="排序" rules={[{ required: true, message: '请输入排序!' }]}><InputNumber style={{ width: 255 }} /></FormItem>
        <FormItem name="status" label="状态" rules={[{ required: true, message: '请选择状态!' }]}><Radio.Group><Radio value={0}>禁用</Radio><Radio value={1}>正常</Radio></Radio.Group></FormItem>
      </Form>
    </Modal>
  );
};

export default AddModal;
