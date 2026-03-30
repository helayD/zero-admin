import React from 'react';
import { Descriptions, Tag } from 'antd';
import type { CompanyAddressItem } from '../data.d';

export interface CompanyAddressListProps {
  addresses: CompanyAddressItem[];
}

const CompanyAddressList: React.FC<CompanyAddressListProps> = (props) => {
  const { addresses } = props;

  if (!addresses || addresses.length === 0) {
    return null;
  }

  return (
    <>
      {addresses.map((addr) => (
        <Descriptions.Item key={addr.id} label="公司退货地址">
          <Tag color={addr.defaultStatus === 1 ? 'blue' : 'default'}>
            {addr.addressName} | {addr.receiverName} | {addr.phone}
          </Tag>
          <br />
          <span style={{ color: '#666', fontSize: 12 }}>
            {addr.province} {addr.city} {addr.region} {addr.detailAddress}
          </span>
          {addr.defaultStatus === 1 && (
            <Tag color="blue" style={{ marginLeft: 8 }}>默认</Tag>
          )}
        </Descriptions.Item>
      ))}
    </>
  );
};

export default CompanyAddressList;
