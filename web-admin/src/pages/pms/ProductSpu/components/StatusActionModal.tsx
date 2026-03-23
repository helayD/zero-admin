import { Alert, Form, Input, Modal, Radio, Space, Tag, Typography } from 'antd';
import type { FC } from 'react';
import { useEffect } from 'react';
import type { ProductSpuListItem } from '../data.d';

const { Paragraph, Text } = Typography;

export type ProductSpuStatusModalAction = 'publish' | 'verify' | 'recommend';

type StatusActionModalProps = {
  visible: boolean;
  action?: ProductSpuStatusModalAction;
  scopeLabel: string;
  targetCount: number;
  initialStatus?: number;
  currentProduct?: Pick<
    ProductSpuListItem,
    'name' | 'publishStatus' | 'verifyStatus' | 'recommendStatus'
  >;
  onCancel: () => void;
  onSubmit: (payload: { status: number; detail?: string }) => Promise<void> | void;
};

type StatusOption = {
  value: number;
  label: string;
  description: string;
};

const actionLabels: Record<ProductSpuStatusModalAction, string> = {
  publish: '上下架',
  verify: '审核',
  recommend: '推荐',
};

const actionOptions: Record<ProductSpuStatusModalAction, StatusOption[]> = {
  publish: [
    { value: 1, label: '上架', description: '商品会进入当前作用域下的可见候选集，前提是它已审核通过。' },
    { value: 0, label: '下架', description: '商品会立即退出前台曝光，下游推荐也会同步失效。' },
  ],
  verify: [
    { value: 1, label: '审核通过', description: '商品会进入可上架状态，但不会自动曝光。' },
    { value: 2, label: '审核驳回', description: '商品将保持后台可编辑，但前台不可见。' },
    { value: 0, label: '重新置为未审核', description: '用于撤回审核结论，商品仍不会自动曝光。' },
  ],
  recommend: [
    { value: 1, label: '加入推荐', description: '仅适用于已审核且已上架的商品。' },
    { value: 0, label: '取消推荐', description: '商品仍可继续售卖，但会退出推荐位。' },
  ],
};

function getDefaultStatus(action?: ProductSpuStatusModalAction, initialStatus?: number) {
  if (!action) {
    return 0;
  }
  const options = actionOptions[action];
  if (options.some((item) => item.value === initialStatus)) {
    return initialStatus as number;
  }
  return options[0].value;
}

function shouldRequireDetail(action?: ProductSpuStatusModalAction, status?: number) {
  return (
    (action === 'verify' && status === 2) ||
    (action === 'publish' && status === 0) ||
    (action === 'recommend' && status === 0)
  );
}

function renderCurrentState(currentProduct?: Pick<ProductSpuListItem, 'publishStatus' | 'verifyStatus' | 'recommendStatus'>) {
  if (!currentProduct) {
    return null;
  }

  return (
    <Space size={[8, 8]} wrap>
      <Tag color={currentProduct.verifyStatus === 1 ? 'green' : currentProduct.verifyStatus === 2 ? 'red' : 'default'}>
        审核：
        {currentProduct.verifyStatus === 1
          ? '通过'
          : currentProduct.verifyStatus === 2
            ? '驳回'
            : '未审核'}
      </Tag>
      <Tag color={currentProduct.publishStatus === 1 ? 'blue' : 'default'}>
        上架：
        {currentProduct.publishStatus === 1 ? '已上架' : '已下架'}
      </Tag>
      <Tag color={currentProduct.recommendStatus === 1 ? 'gold' : 'default'}>
        推荐：
        {currentProduct.recommendStatus === 1 ? '推荐中' : '未推荐'}
      </Tag>
    </Space>
  );
}

const StatusActionModal: FC<StatusActionModalProps> = ({
  visible,
  action,
  scopeLabel,
  targetCount,
  initialStatus,
  currentProduct,
  onCancel,
  onSubmit,
}) => {
  const [form] = Form.useForm<{ status: number; detail?: string }>();
  const selectedStatus = Form.useWatch('status', form);

  useEffect(() => {
    if (!visible || !action) {
      return;
    }
    form.setFieldsValue({
      status: getDefaultStatus(action, initialStatus),
      detail: '',
    });
  }, [action, form, initialStatus, visible]);

  const options = action ? actionOptions[action] : [];
  const currentOption = options.find((item) => item.value === selectedStatus) || options[0];

  return (
    <Modal
      visible={visible}
      destroyOnClose
      title={action ? `${actionLabels[action]}确认` : '状态确认'}
      okText="确认提交"
      cancelText="取消"
      onCancel={onCancel}
      onOk={async () => {
        const values = await form.validateFields();
        await onSubmit(values);
      }}
    >
      <Alert
        showIcon
        type="warning"
        style={{ marginBottom: 16 }}
        message={`当前主体：${scopeLabel}`}
        description={`本次将影响 ${targetCount} 个商品。请先确认状态变更与前台曝光范围，再提交。`}
      />
      {currentProduct?.name && (
        <Paragraph style={{ marginBottom: 12 }}>
          <Text strong>目标商品：</Text>
          {currentProduct.name}
        </Paragraph>
      )}
      {renderCurrentState(currentProduct)}
      <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
        <Form.Item
          name="status"
          label="目标状态"
          rules={[{ required: true, message: '请选择目标状态' }]}
        >
          <Radio.Group style={{ width: '100%' }}>
            <Space direction="vertical" style={{ width: '100%' }}>
              {options.map((item) => (
                <Radio key={item.value} value={item.value}>
                  {item.label}
                </Radio>
              ))}
            </Space>
          </Radio.Group>
        </Form.Item>
        {currentOption && (
          <Alert
            showIcon
            type="info"
            style={{ marginBottom: 16 }}
            message="Consequence Preview"
            description={currentOption.description}
          />
        )}
        <Form.Item
          name="detail"
          label="操作说明"
          rules={[
            {
              validator: async (_, value) => {
                if (shouldRequireDetail(action, selectedStatus) && !String(value || '').trim()) {
                  throw new Error('当前动作需要填写原因，方便后续追踪和恢复');
                }
              },
            },
          ]}
          extra="审核驳回、下架、取消推荐等动作建议填写清晰原因，便于运营、商户和客服追溯。"
        >
          <Input.TextArea
            rows={4}
            maxLength={200}
            placeholder="例如：库存盘点中暂时下架；规格信息缺失，驳回后请补齐主图与有效 SKU。"
          />
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default StatusActionModal;
