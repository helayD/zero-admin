import React, { useEffect } from 'react';
import { Form, Input, InputNumber, Modal, Radio } from 'antd';
import type { SubjectCategoryListItem } from '../data.d';

export interface CreateFormProps {
  onCancel: () => void;
  onSubmit: (values: SubjectCategoryListItem) => void;
  createModalVisible: boolean;
}

const FormItem = Form.Item;

const formLayout = {
  labelCol: { span: 6 },
  wrapperCol: { span: 16 },
};

const CreateSubjectCategoryForm: React.FC<CreateFormProps> = (props) => {
  const [form] = Form.useForm();
  const { onSubmit, onCancel, createModalVisible } = props;

  useEffect(() => {
    if (form && !createModalVisible) {
      form.resetFields();
    }
  }, [createModalVisible, form]);

  return (
    <Modal
      forceRender
      destroyOnClose
      title="新增专题分类"
      visible={createModalVisible}
      okText="保存"
      onOk={() => form.submit()}
      onCancel={onCancel}
      width={760}
    >
      <Form {...formLayout} form={form} onFinish={onSubmit}>
        <FormItem name="name" label="分类名称" rules={[{ required: true, message: '请输入分类名称' }]}>
          <Input placeholder="请输入分类名称" />
        </FormItem>
        <FormItem name="icon" label="图标链接">
          <Input placeholder="请输入图标 URL" />
        </FormItem>
        <FormItem name="sort" label="排序" initialValue={0}>
          <InputNumber style={{ width: '100%' }} min={0} />
        </FormItem>
        <FormItem name="showStatus" label="显示状态" initialValue={1}>
          <Radio.Group>
            <Radio value={1}>显示</Radio>
            <Radio value={0}>隐藏</Radio>
          </Radio.Group>
        </FormItem>
      </Form>
    </Modal>
  );
};

export default CreateSubjectCategoryForm;
