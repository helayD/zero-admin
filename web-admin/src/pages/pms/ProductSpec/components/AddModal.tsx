import React, { useEffect, useState } from 'react';
import { Form, Input, InputNumber, message, Modal, Radio, TreeSelect } from 'antd';
import type { ProductSpecListItem } from '../data.d';
import type { GovernanceScopeValue } from '@/pages/system/components/governance';
import { toGovernancePayload } from '@/pages/system/components/governance';
import { queryProductCategoryList } from '@/pages/pms/ProductCategory/service';
import { tree } from '@/utils/utils';
import type { ProductCategoryListItem } from '@/pages/pms/ProductCategory/data';

export interface AddModalProps {
  onCancel: () => void;
  onSubmit: (values: ProductSpecListItem) => void;
  addVisible: boolean;
  scope?: GovernanceScopeValue;
}

const FormItem = Form.Item;
const formLayout = { labelCol: { span: 7 }, wrapperCol: { span: 13 } };

const AddModal: React.FC<AddModalProps> = ({ onSubmit, onCancel, addVisible, scope }) => {
  const [form] = Form.useForm();
  const [categoryListItem, setCategoryListItem] = useState<ProductCategoryListItem[]>([]);

  useEffect(() => {
    if (form && !addVisible) {
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
  }, [addVisible, form, scope]);

  return (
    <Modal forceRender destroyOnClose title="新增" open={addVisible} okText="保存" onOk={() => form.submit()} onCancel={onCancel}>
      <Form {...formLayout} form={form} onFinish={(values) => onSubmit?.(values)}>
        <FormItem name="categoryId" label="商品分类" rules={[{ required: true, message: '请选择商品分类!' }]}>
          <TreeSelect style={{ width: '100%' }} treeData={categoryListItem} placeholder="请选择商品分类" treeDefaultExpandAll />
        </FormItem>
        <FormItem name="name" label="规格名称" rules={[{ required: true, message: '请输入规格名称!' }]}><Input placeholder="请输入规格名称!" /></FormItem>
        <FormItem name="sort" label="排序" rules={[{ required: true, message: '请输入排序!' }]}><InputNumber style={{ width: 255 }} /></FormItem>
        <FormItem name="status" label="状态" rules={[{ required: true, message: '请选择状态!' }]}><Radio.Group><Radio value={0}>禁用</Radio><Radio value={1}>正常</Radio></Radio.Group></FormItem>
      </Form>
    </Modal>
  );
};

export default AddModal;
