import React, { useEffect } from 'react';
import { Form, Input, Modal, Radio, Select } from 'antd';
import type { NoticeListItem } from '../data.d';

export interface CreateFormProps {
  onCancel: () => void;
  onSubmit: (values: NoticeListItem) => void;
  open: boolean;
}

const FormItem = Form.Item;

const formLayout = {
  labelCol: { span: 6 },
  wrapperCol: { span: 16 },
};

const AddModal: React.FC<CreateFormProps> = ({ onSubmit, onCancel, open }) => {
  const [form] = Form.useForm();

  useEffect(() => {
    if (!open) {
      form.resetFields();
    }
  }, [form, open]);

  return (
    <Modal
      forceRender
      destroyOnClose
      title="新建通知"
      open={open}
      okText="保存"
      onOk={() => form.submit()}
      onCancel={onCancel}
    >
      <Form {...formLayout} form={form} onFinish={onSubmit}>
        <FormItem
          name="noticeTitle"
          label="通知标题"
          rules={[{ required: true, message: '请输入通知标题' }]}
        >
          <Input placeholder="请输入通知标题" />
        </FormItem>
        <FormItem
          name="noticeType"
          label="通知类型"
          initialValue={1}
          rules={[{ required: true, message: '请选择通知类型' }]}
        >
          <Select
            options={[
              { value: 1, label: '通知' },
              { value: 2, label: '公告' },
            ]}
          />
        </FormItem>
        <FormItem
          name="status"
          label="状态"
          initialValue={1}
          rules={[{ required: true, message: '请选择状态' }]}
        >
          <Radio.Group>
            <Radio value={1}>正常</Radio>
            <Radio value={0}>关闭</Radio>
          </Radio.Group>
        </FormItem>
        <FormItem
          name="noticeContent"
          label="通知内容"
          rules={[{ required: true, message: '请输入通知内容' }]}
        >
          <Input.TextArea rows={6} placeholder="请输入通知内容" />
        </FormItem>
        <FormItem name="remark" label="备注">
          <Input.TextArea rows={2} placeholder="请输入备注" />
        </FormItem>
      </Form>
    </Modal>
  );
};

export default AddModal;
