import React, { useEffect } from 'react';
import { Form, Input, InputNumber, Modal, Radio } from 'antd';
import type { SubjectListItem } from '../data.d';

export interface UpdateFormProps {
  onCancel: () => void;
  onSubmit: (values: SubjectListItem) => void;
  updateModalVisible: boolean;
  currentData: Partial<SubjectListItem>;
}

const FormItem = Form.Item;

const formLayout = {
  labelCol: { span: 6 },
  wrapperCol: { span: 16 },
};

const UpdatePostForm: React.FC<UpdateFormProps> = (props) => {
  const [form] = Form.useForm();
  const { onSubmit, onCancel, updateModalVisible, currentData } = props;

  useEffect(() => {
    if (form && !updateModalVisible) {
      form.resetFields();
    }
  }, [updateModalVisible, form]);

  useEffect(() => {
    if (currentData && updateModalVisible) {
      form.setFieldsValue({
        ...currentData,
      });
    }
  }, [currentData, updateModalVisible, form]);

  return (
    <Modal
      forceRender
      destroyOnClose
      title="编辑专题"
      visible={updateModalVisible}
      okText="保存"
      onOk={() => form.submit()}
      onCancel={onCancel}
      width={760}
    >
      <Form {...formLayout} form={form} onFinish={onSubmit}>
        <FormItem name="id" hidden>
          <Input />
        </FormItem>
        <FormItem name="title" label="专题标题" rules={[{ required: true, message: '请输入专题标题' }]}>
          <Input placeholder="请输入专题标题" />
        </FormItem>
        <FormItem
          name="categoryName"
          label="专题分类"
          rules={[{ required: true, message: '请输入专题分类名称' }]}
        >
          <Input placeholder="请输入专题分类名称" />
        </FormItem>
        <FormItem
          name="categoryId"
          label="分类 ID"
          rules={[{ required: true, message: '请输入专题分类 ID' }]}
        >
          <InputNumber style={{ width: '100%' }} min={1} />
        </FormItem>
        <FormItem name="pic" label="主图链接">
          <Input placeholder="请输入专题主图 URL" />
        </FormItem>
        <FormItem name="productCount" label="关联商品数" initialValue={0}>
          <InputNumber style={{ width: '100%' }} min={0} />
        </FormItem>
        <FormItem name="sort" label="排序" initialValue={0}>
          <InputNumber style={{ width: '100%' }} min={0} />
        </FormItem>
        <FormItem name="showStatus" label="显示状态">
          <Radio.Group>
            <Radio value={1}>显示</Radio>
            <Radio value={0}>隐藏</Radio>
          </Radio.Group>
        </FormItem>
        <FormItem name="recommendStatus" label="推荐状态">
          <Radio.Group>
            <Radio value={1}>推荐</Radio>
            <Radio value={0}>普通</Radio>
          </Radio.Group>
        </FormItem>
        <FormItem name="description" label="摘要">
          <Input.TextArea rows={3} placeholder="请输入专题摘要" />
        </FormItem>
        <FormItem name="content" label="正文">
          <Input.TextArea rows={6} placeholder="请输入专题正文" />
        </FormItem>
      </Form>
    </Modal>
  );
};

export default UpdatePostForm;
