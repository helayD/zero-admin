import React, { useEffect, useState } from 'react';
import { Alert, Input, Modal, message } from 'antd';

type ActionReasonModalProps = {
  visible: boolean;
  title: string;
  description: string;
  confirmText: string;
  placeholder: string;
  danger?: boolean;
  requireReason?: boolean;
  onCancel: () => void;
  onSubmit: (reason: string) => Promise<boolean | void>;
};

const ActionReasonModal: React.FC<ActionReasonModalProps> = ({
  visible,
  title,
  description,
  confirmText,
  placeholder,
  danger,
  requireReason,
  onCancel,
  onSubmit,
}) => {
  const [reason, setReason] = useState('');
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (visible) {
      setReason('');
    }
  }, [visible]);

  return (
    <Modal
      title={title}
      visible={visible}
      destroyOnClose
      onCancel={onCancel}
      okText={confirmText}
      confirmLoading={submitting}
      okButtonProps={danger ? { danger: true } : undefined}
      onOk={async () => {
        if (requireReason && !reason.trim()) {
          message.warning('请先填写原因说明');
          return;
        }
        setSubmitting(true);
        try {
          const success = await onSubmit(reason.trim());
          if (success !== false) {
            onCancel();
          }
        } finally {
          setSubmitting(false);
        }
      }}
    >
      <Alert showIcon type={danger ? 'warning' : 'info'} message={description} style={{ marginBottom: 16 }} />
      <Input.TextArea
        rows={4}
        value={reason}
        placeholder={placeholder}
        onChange={(event) => setReason(event.target.value)}
      />
    </Modal>
  );
};

export default ActionReasonModal;
