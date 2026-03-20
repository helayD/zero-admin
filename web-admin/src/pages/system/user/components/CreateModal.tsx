import React, { useEffect, useState } from 'react';
import { Alert, Form, Input, Modal, Radio, Select, TreeSelect } from 'antd';
import type { JobList, UserListItem } from '../data.d';
import { queryDeptAndPostList } from '@/pages/system/user/service';
import { tree } from '@/utils/utils';
import type { GovernanceScopeValue } from '../../components/governance';
import { buildGovernanceScopeLabel, toGovernancePayload } from '../../components/governance';

export interface CreateFormProps {
  onCancel: () => void;
  onSubmit: (values: UserListItem) => void;
  createModalVisible: boolean;
  scope: GovernanceScopeValue;
}

const FormItem = Form.Item;

const formLayout = {
  labelCol: { span: 7 },
  wrapperCol: { span: 13 },
};

const CreateModal: React.FC<CreateFormProps> = (props) => {
  const [form] = Form.useForm();

  const [jobConf, setJobConf] = useState<JobList[]>([]);
  const [deptConf, setDeptConf] = useState<JobList[]>([]);

  const { onSubmit, onCancel, createModalVisible, scope } = props;

  useEffect(() => {
    if (form && !createModalVisible) {
      form.resetFields();
    } else {
      queryDeptAndPostList(toGovernancePayload(scope)).then((res) => {
        setJobConf(res.data.postList);
        setDeptConf(tree(res.data.deptList, 0, 'parentId'));
      });
    }
  }, [props.createModalVisible, scope]);

  const handleSubmit = () => {
    if (!form) return;
    form.submit();
  };

  const handleFinish = (values: UserListItem) => {
    if (onSubmit) {
      // values.deptId=Number(selectedKey)
      onSubmit(values);
    }
  };

  const renderContent = () => {
    return (
      <>
        <Alert
          showIcon
          type="info"
          style={{ marginBottom: 16 }}
          message={`当前主体：${buildGovernanceScopeLabel(scope)}`}
          description="部门和岗位会按当前主体过滤。若是租户管理员 bootstrap 场景，建议保持待激活并使用 bootstrap 角色模式。"
        />
        <FormItem name="deptId" label="部门" rules={[{ required: true, message: '请选择部门' }]}>
          <TreeSelect
            style={{ width: '100%' }}
            treeData={deptConf}
            placeholder="请选择部门"
            treeDefaultExpandAll
          />
        </FormItem>
        <FormItem name="postIds" label="职位" rules={[{ required: true, message: '请选择职位' }]}>
          <Select id="postIds" mode="multiple" allowClear placeholder={'请选择职位'}>
            {jobConf.map((r) => (
              <Select.Option key={r.id} value={r.id}>
                {r.postName}
              </Select.Option>
            ))}
          </Select>
        </FormItem>

        <FormItem
          name="userName"
          label="用户名"
          rules={[{ required: true, message: '请输入用户名' }]}
        >
          <Input id="create-name" placeholder={'请输入用户名'} />
        </FormItem>

        <FormItem name="password" label="密码" rules={[{ required: true, message: '请输入密码' }]}>
          <Input.Password id="create-nick_name" placeholder={'请输入密码'} />
        </FormItem>
        <FormItem name="nickName" label="昵称" rules={[{ required: true, message: '请输入昵称' }]}>
          <Input id="create-nick_name" placeholder={'请输入昵称'} />
        </FormItem>
        <FormItem
          name="mobile"
          label="手机号"
          rules={[{ required: true, message: '请输入手机号' }]}
        >
          <Input id="create-mobile" placeholder={'请输入手机号'} />
        </FormItem>
        <FormItem name="email" label="邮箱" rules={[{ required: true, message: '请输入邮箱' }]}>
          <Input id="create-email" placeholder={'请输入邮箱'} />
        </FormItem>
        <FormItem
          name="status"
          label="状态"
          initialValue={1}
          rules={[{ required: true, message: '请选择状态' }]}
        >
          <Radio.Group id="status">
            <Radio value={0}>禁用</Radio>
            <Radio value={1}>启用</Radio>
          </Radio.Group>
        </FormItem>
        <FormItem
          name="activationStatus"
          label="激活状态"
          initialValue={scope.scopeType === 'tenant' ? 'pending-activation' : 'active'}
        >
          <Select
            options={[
              { value: 'active', label: '已激活' },
              { value: 'pending-activation', label: '待激活' },
              { value: 'disabled', label: '已停用' },
              { value: 'archived', label: '已归档' },
            ]}
          />
        </FormItem>
        <FormItem
          name="roleMode"
          label="角色模式"
          initialValue={scope.scopeType === 'tenant' ? 'bootstrap' : 'manual'}
        >
          <Select
            options={[
              { value: 'manual', label: '手动治理' },
              { value: 'bootstrap', label: 'Bootstrap' },
            ]}
          />
        </FormItem>
        <FormItem name="remark" label="备注">
          <Input.TextArea rows={2} placeholder={'请输入备注'} />
        </FormItem>
      </>
    );
  };

  const modalFooter = { okText: '保存', onOk: handleSubmit, onCancel };

  return (
    <Modal forceRender destroyOnClose title="新建" open={createModalVisible} {...modalFooter}>
      <Form {...formLayout} form={form} onFinish={handleFinish}>
        {renderContent()}
      </Form>
    </Modal>
  );
};

export default CreateModal;
