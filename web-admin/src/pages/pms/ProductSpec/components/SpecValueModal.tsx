import React from 'react';
import { Modal } from 'antd';
import type { GovernanceScopeValue } from '@/pages/system/components/governance';
import ProductSpecValueList from '@/pages/pms/ProductSpecValue';

export interface UpdateFormProps {
  onCancel: () => void;
  modalVisible: boolean;
  specId: number;
  scope?: GovernanceScopeValue;
}

const SpecValueModal: React.FC<UpdateFormProps> = (props) => {
  const { modalVisible, onCancel, scope } = props;

  return (
    <Modal width={1600} forceRender destroyOnClose onCancel={onCancel} open={modalVisible} footer={null}>
      {modalVisible && <ProductSpecValueList specId={props.specId} scope={scope} />}
    </Modal>
  );
};

export default SpecValueModal;
