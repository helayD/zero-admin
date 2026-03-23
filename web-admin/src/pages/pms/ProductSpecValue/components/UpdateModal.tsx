import React, { useEffect, useState } from 'react';
import { Form, Input, InputNumber, Modal, Radio, Select, message } from 'antd';
import type { ProductSpecValueListItem } from '../data.d';
import type { GovernanceScopeValue } from '@/pages/system/components/governance';
import { toGovernancePayload } from '@/pages/system/components/governance';
import { queryProductSpecList } from '@/pages/pms/ProductSpec/service';
import type { ProductSpecListItem } from '@/pages/pms/ProductSpec/data';

export interface UpdateModalProps {
  onCancel: () => void;
  onSubmit: (values: ProductSpecValueListItem) => void;
  updateVisible: boolean;
  currentData: Partial<ProductSpecValueListItem>;
  scope?: GovernanceScopeValue;
}

const FormItem = Form.Item;
const formLayout = { labelCol: { span: 7 }, wrapperCol: { span: 13 } };

const UpdateModal: React.FC<UpdateModalProps> = ({ onSubmit, onCancel, updateVisible, currentData, scope }) => {
  const [form] = Form.useForm();
  const [specItems, setSpecItems] = useState<ProductSpecListItem[]>([]);

  useEffect(() => {
    if (form && !updateVisible) {
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
  }, [updateVisible, form, scope]);

  useEffect(() => {
    if (currentData) {
      form.setFieldsValue({ ...currentData });
    }
  }, [currentData, form]);

  return (
    <Modal forceRender destroyOnClose title="编辑" open={updateVisible} okText="保存" onOk={() => form.submit()} onCancel={onCancel}>
      <Form {...formLayout} form={form} onFinish={(values) => onSubmit?.(values as ProductSpecValueListItem)}>
        <FormItem name="id" hidden><Input /></FormItem>
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

export default UpdateModal;
