import React, { useEffect, useState } from 'react';
import { Form, Input, InputNumber, message, Modal, Radio, TreeSelect } from 'antd';
import type { ProductAttributeGroupListItem } from '../data.d';
import type { GovernanceScopeValue } from '@/pages/system/components/governance';
import { toGovernancePayload } from '@/pages/system/components/governance';
import { queryProductCategoryList } from '@/pages/pms/ProductCategory/service';
import { tree } from '@/utils/utils';
import type { ProductCategoryListItem } from '@/pages/pms/ProductCategory/data';

export interface UpdateModalProps {
  onCancel: () => void;
  onSubmit: (values: ProductAttributeGroupListItem) => void;
  updateVisible: boolean;
  currentData: Partial<ProductAttributeGroupListItem>;
  scope?: GovernanceScopeValue;
}

const FormItem = Form.Item;
const formLayout = { labelCol: { span: 7 }, wrapperCol: { span: 13 } };

const UpdateModal: React.FC<UpdateModalProps> = ({ onSubmit, onCancel, updateVisible, currentData, scope }) => {
  const [form] = Form.useForm();
  const [categoryListItem, setCategoryListItem] = useState<ProductCategoryListItem[]>([]);

  useEffect(() => {
    if (form && !updateVisible) {
      form.resetFields();
      return;
    }
    queryProductCategoryList({ pageSize: 100, current: 1, ...toGovernancePayload(scope) }).then((res) => {
      if (res.code === '000000') {
        const map = (res.data || []).map((item: { id: number; name: string; parentId: number }) => ({ value: item.id, id: item.id, label: item.name, title: item.name, parentId: item.parentId }));
        setCategoryListItem(tree(map, 0, 'parentId'));
      } else {
        message.error(res.message || res.msg || '加载商品分类失败');
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
      <Form {...formLayout} form={form} onFinish={(values) => onSubmit?.(values as ProductAttributeGroupListItem)}>
        <FormItem name="id" hidden><Input /></FormItem>
        <FormItem name="categoryId" label="商品分类" rules={[{ required: true, message: '请选择商品分类!' }]}>
          <TreeSelect style={{ width: '100%' }} treeData={categoryListItem} placeholder="请选择商品分类" treeDefaultExpandAll />
        </FormItem>
        <FormItem name="name" label="分组名称" rules={[{ required: true, message: '请输入分组名称!' }]}><Input placeholder="请输入分组名称!" /></FormItem>
        <FormItem name="sort" label="排序" rules={[{ required: true, message: '请输入排序!' }]}><InputNumber style={{ width: 255 }} /></FormItem>
        <FormItem name="status" label="状态" rules={[{ required: true, message: '请选择状态!' }]}><Radio.Group><Radio value={0}>禁用</Radio><Radio value={1}>正常</Radio></Radio.Group></FormItem>
      </Form>
    </Modal>
  );
};

export default UpdateModal;
