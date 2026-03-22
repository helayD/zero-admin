import React from 'react';
import { Modal } from 'antd';
import ProductSkuList from '@/pages/pms/ProductSku';
import type { GovernanceScopeValue } from '@/pages/system/components/governance';

export interface UpdateFormProps {
  onCancel: () => void;
  modalVisible: boolean;
  spuId: number;
  scope?: GovernanceScopeValue;
}

const AttributeModal: React.FC<UpdateFormProps> = (props) => {
  const { modalVisible, onCancel } = props;

  return (
    <Modal
      width={1600}
      forceRender
      destroyOnClose
      onCancel={onCancel}
      visible={modalVisible}
      footer={null}
    >
      {modalVisible && <ProductSkuList spuId={props.spuId} scope={props.scope}></ProductSkuList>}
    </Modal>
  );
};

export default AttributeModal;
