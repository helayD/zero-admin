import React from 'react';
import { Modal } from 'antd';
import type { GovernanceScopeValue } from '@/pages/system/components/governance';
import AttributeList from '@/pages/pms/ProductAttribute';

export interface UpdateFormProps {
  onCancel: () => void;
  modalVisible: boolean;
  groupId: number;
  scope?: GovernanceScopeValue;
}

const AttributeModal: React.FC<UpdateFormProps> = (props) => {
  const { modalVisible, onCancel, scope } = props;

  return (
    <Modal width={1600} forceRender destroyOnClose onCancel={onCancel} open={modalVisible} footer={null}>
      {modalVisible && <AttributeList groupId={props.groupId} scope={scope} />}
    </Modal>
  );
};

export default AttributeModal;
