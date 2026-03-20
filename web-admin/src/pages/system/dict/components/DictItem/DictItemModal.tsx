import React from 'react';
import { Modal } from 'antd';
import { DictTypeListItem } from '../../data';
import DictItemLIst from '@/pages/system/dict/components/DictItem/DictItemLIst';
import type { GovernanceScopeValue } from '../../../components/governance';

export interface UpdateFormProps {
  onCancel: () => void;
  onSubmit: (values: DictTypeListItem) => void;
  dictItemModalVisible: boolean;
  currentData: Partial<DictTypeListItem>;
  scope: GovernanceScopeValue;
}

const DictItemModal: React.FC<UpdateFormProps> = (props) => {
  const { dictItemModalVisible, onCancel } = props;

  return (
    <Modal
      width={1600}
      forceRender
      destroyOnClose
      title={'配置字典项 - ' + props.currentData.dictName}
      onCancel={onCancel}
      open={dictItemModalVisible}
      footer={null}
    >
      {dictItemModalVisible && (
        <DictItemLIst
          dictType={props.currentData.dictType}
          dictTypeId={props.currentData.id}
          dictItemModalVisible={dictItemModalVisible}
          scope={props.scope}
        />
      )}
    </Modal>
  );
};

export default DictItemModal;
