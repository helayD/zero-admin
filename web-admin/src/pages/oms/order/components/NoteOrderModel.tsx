import React, { useEffect, useState } from 'react';
import { Button, Col, Input, Modal, Row } from 'antd';
import { EditOutlined, ExclamationCircleFilled } from '@ant-design/icons';
import type { OrderListItem } from '@/pages/oms/order/data';

export interface UpdateFormProps {
  title: string;
  submitText: string;
  confirmTitle: string;
  confirmHint: string;
  visible: boolean;
  currentData: OrderListItem;
  initialRemark?: string;
  onCancel: () => void;
  onSubmit: (values: OrderListItem) => void;
}

const { confirm } = Modal;

const NoteOrderModel: React.FC<UpdateFormProps> = (props) => {
  const {
    onSubmit,
    onCancel,
    visible,
    currentData,
    title,
    submitText,
    confirmTitle,
    confirmHint,
    initialRemark,
  } = props;

  const [remark, setRemark] = useState<string>('');

  useEffect(() => {
    if (visible) {
      setRemark(initialRemark || currentData.note || '');
    }
  }, [visible, initialRemark, currentData.note]);

  const handleConfirm = () => {
    confirm({
      title: confirmTitle,
      icon: <ExclamationCircleFilled />,
      content: confirmHint,
      onOk() {
        onSubmit({
          id: currentData.id,
          note: remark,
          status: currentData.status,
        });
      },
    });
  };

  return (
    <Modal forceRender destroyOnClose title={title} visible={visible} onCancel={onCancel} footer={false}>
      <Row>
        <Col span={5} className="Col">
          操作备注：
        </Col>
        <Col span={19} className="Col">
          <Input.TextArea
            rows={4}
            placeholder="请输入备注"
            value={remark}
            onChange={(e) => setRemark(e.target.value)}
          />
        </Col>
      </Row>
      <Row style={{ marginTop: 30, marginBottom: 30 }}>
        <Col span={10} className="Col" />
        <Col span={14} className="Col">
          <Button type="primary" icon={<EditOutlined />} onClick={handleConfirm}>
            {submitText}
          </Button>
        </Col>
      </Row>
    </Modal>
  );
};

export default NoteOrderModel;
