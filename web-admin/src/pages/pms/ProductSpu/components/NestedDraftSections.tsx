import { MinusCircleOutlined, PlusOutlined } from '@ant-design/icons';
import { Button, Card, Divider, Form, Input, InputNumber, Radio, Select, Space } from 'antd';
import type { FC } from 'react';
import {
  getDuplicateSpecRowIndex,
  hasDuplicateAttributeId,
  hasDuplicateSkuCode,
} from '../draftFeedback';
import type { SelectOption } from './useCatalogOptions';

const listCardStyle = {
  marginBottom: 12,
};

const statusOptions = [
  { label: '禁用', value: 0 },
  { label: '启用', value: 1 },
];

const verifyStatusOptions = [
  { label: '未审核', value: 0 },
  { label: '审核通过', value: 1 },
  { label: '审核不通过', value: 2 },
];

const priceInputStyle = { width: 160 };
const compactInputStyle = { width: 120 };

type NestedDraftSectionsProps = {
  attributeOptions?: SelectOption[];
};

const NestedDraftSections: FC<NestedDraftSectionsProps> = ({ attributeOptions = [] }) => {
  return (
    <>
      <Divider orientation="left">SKU 建档</Divider>
      <Form.List
        name="skuList"
        rules={[
          {
            validator: async (_, value) => {
              if (!Array.isArray(value) || value.length === 0) {
                throw new Error('至少需要一个有效SKU');
              }
            },
          },
        ]}
      >
        {(fields, { add, remove }, { errors }) => (
          <>
            {fields.map((field, index) => (
              <Card
                key={field.key}
                size="small"
                title={`SKU #${index + 1}`}
                style={listCardStyle}
                extra={
                  <Button type="link" danger onClick={() => remove(field.name)}>
                    <MinusCircleOutlined /> 删除
                  </Button>
                }
              >
                <Form.Item name={[field.name, 'id']} hidden>
                  <Input />
                </Form.Item>
                <Form.Item name={[field.name, 'spuId']} hidden>
                  <Input />
                </Form.Item>
                <Space direction="vertical" style={{ width: '100%' }} size={8}>
                  <Form.Item
                    name={[field.name, 'name']}
                    label="SKU 名称"
                    rules={[{ required: true, message: '请输入 SKU 名称' }]}
                  >
                    <Input placeholder="例如：黑色 / 128G" />
                  </Form.Item>
                  <Form.Item
                    name={[field.name, 'skuCode']}
                    label="SKU 编码"
                    dependencies={['skuList']}
                    rules={[
                      ({ getFieldValue }) => ({
                        validator: async (_, value) => {
                          const skuList = getFieldValue(['skuList']);
                          if (hasDuplicateSkuCode(skuList, field.name, value)) {
                            throw new Error('当前草稿内已存在相同SKU编码');
                          }
                        },
                      }),
                    ]}
                  >
                    <Input placeholder="留空则由后端兜底生成" />
                  </Form.Item>
                  <Form.Item
                    name={[field.name, 'specData']}
                    label="规格数据"
                    dependencies={['skuList']}
                    rules={[
                      { required: true, message: '请输入规格数据' },
                      ({ getFieldValue }) => ({
                        validator: async (_, value) => {
                          try {
                            const skuList = getFieldValue(['skuList']);
                            const duplicateIndex = getDuplicateSpecRowIndex(
                              skuList,
                              field.name,
                              value,
                            );
                            if (duplicateIndex >= 0) {
                              throw new Error(`当前规格组合与第${duplicateIndex + 1}行SKU重复`);
                            }
                          } catch (error) {
                            if (error instanceof Error) {
                              throw error;
                            }
                            throw new Error('规格数据不是合法JSON');
                          }
                        },
                      }),
                    ]}
                  >
                    <Input.TextArea rows={2} placeholder='例如：{"颜色":"黑色","容量":"128G"}' />
                  </Form.Item>
                  <Form.Item name={[field.name, 'mainPic']} label="SKU 主图">
                    <Input placeholder="图片 URL" />
                  </Form.Item>
                  <Form.Item name={[field.name, 'albumPics']} label="SKU 图集">
                    <Input placeholder="多个图片 URL 用逗号分隔" />
                  </Form.Item>
                  <Space size={12} wrap>
                    <Form.Item
                      name={[field.name, 'price']}
                      label="售价"
                      rules={[
                        { required: true, message: '请输入售价' },
                        {
                          validator: async (_, value) => {
                            if (value === undefined || value === null) {
                              return;
                            }
                            if (value <= 0) {
                              throw new Error('SKU 价格必须大于0');
                            }
                          },
                        },
                      ]}
                    >
                      <InputNumber min={0} precision={2} style={priceInputStyle} />
                    </Form.Item>
                    <Form.Item
                      name={[field.name, 'promotionPrice']}
                      label="促销价"
                      dependencies={[['skuList', field.name, 'price']]}
                      rules={[
                        ({ getFieldValue }) => ({
                          validator: async (_, value) => {
                            if (value === undefined || value === null) {
                              return;
                            }
                            if (value < 0) {
                              throw new Error('SKU 促销价不能小于0');
                            }
                            const price = getFieldValue(['skuList', field.name, 'price']);
                            if (price !== undefined && price !== null && value > price) {
                              throw new Error('SKU 促销价不能高于销售价');
                            }
                          },
                        }),
                      ]}
                    >
                      <InputNumber min={0} precision={2} style={priceInputStyle} />
                    </Form.Item>
                    <Form.Item
                      name={[field.name, 'stock']}
                      label="库存"
                      rules={[
                        { required: true, message: '请输入库存' },
                        {
                          validator: async (_, value) => {
                            if (value === undefined || value === null) {
                              return;
                            }
                            if (value < 0) {
                              throw new Error('SKU 库存不能小于0');
                            }
                          },
                        },
                      ]}
                    >
                      <InputNumber min={0} style={compactInputStyle} />
                    </Form.Item>
                    <Form.Item
                      name={[field.name, 'lowStock']}
                      label="预警库存"
                      dependencies={[['skuList', field.name, 'stock']]}
                      rules={[
                        { required: true, message: '请输入预警库存' },
                        ({ getFieldValue }) => ({
                          validator: async (_, value) => {
                            if (value === undefined || value === null) {
                              return;
                            }
                            if (value < 0) {
                              throw new Error('SKU 预警库存不能小于0');
                            }
                            const stock = getFieldValue(['skuList', field.name, 'stock']);
                            if (stock !== undefined && stock !== null && value > stock) {
                              throw new Error('SKU 预警库存不能大于可用库存');
                            }
                          },
                        }),
                      ]}
                    >
                      <InputNumber min={0} style={compactInputStyle} />
                    </Form.Item>
                    <Form.Item name={[field.name, 'weight']} label="重量(kg)">
                      <InputNumber min={0} precision={2} style={compactInputStyle} />
                    </Form.Item>
                    <Form.Item name={[field.name, 'sort']} label="排序">
                      <InputNumber min={0} style={compactInputStyle} />
                    </Form.Item>
                  </Space>
                  <Space size={12} wrap>
                    <Form.Item name={[field.name, 'publishStatus']} label="上架状态">
                      <Radio.Group options={statusOptions} optionType="button" />
                    </Form.Item>
                    <Form.Item name={[field.name, 'verifyStatus']} label="审核状态">
                      <Radio.Group options={verifyStatusOptions} optionType="button" />
                    </Form.Item>
                  </Space>
                </Space>
              </Card>
            ))}
            <Form.ErrorList errors={errors} />
            <Button type="dashed" block onClick={() => add({ publishStatus: 0, verifyStatus: 0 })}>
              <PlusOutlined /> 新增 SKU
            </Button>
          </>
        )}
      </Form.List>

      <Divider orientation="left">会员价</Divider>
      <Form.List name="memberPriceList">
        {(fields, { add, remove }, { errors }) => (
          <>
            {fields.map((field, index) => (
              <Card
                key={field.key}
                size="small"
                title={`会员价 #${index + 1}`}
                style={listCardStyle}
                extra={
                  <Button type="link" danger onClick={() => remove(field.name)}>
                    <MinusCircleOutlined /> 删除
                  </Button>
                }
              >
                <Form.Item name={[field.name, 'id']} hidden>
                  <Input />
                </Form.Item>
                <Space size={12} wrap>
                  <Form.Item
                    name={[field.name, 'memberLevelId']}
                    label="会员等级 ID"
                    rules={[{ required: true, message: '请输入会员等级 ID' }]}
                  >
                    <InputNumber min={1} style={priceInputStyle} />
                  </Form.Item>
                  <Form.Item
                    name={[field.name, 'memberLevelName']}
                    label="会员等级名称"
                    rules={[{ required: true, message: '请输入会员等级名称' }]}
                  >
                    <Input placeholder="例如：黄金会员" />
                  </Form.Item>
                  <Form.Item
                    name={[field.name, 'memberPrice']}
                    label="会员价"
                    rules={[
                      { required: true, message: '请输入会员价' },
                      {
                        validator: async (_, value) => {
                          if (value === undefined || value === null) {
                            return;
                          }
                          if (value <= 0) {
                            throw new Error('会员价必须大于0');
                          }
                        },
                      },
                    ]}
                  >
                    <InputNumber min={0} precision={2} style={priceInputStyle} />
                  </Form.Item>
                </Space>
              </Card>
            ))}
            <Form.ErrorList errors={errors} />
            <Button type="dashed" block onClick={() => add({})}>
              <PlusOutlined /> 新增会员价
            </Button>
          </>
        )}
      </Form.List>

      <Divider orientation="left">阶梯价</Divider>
      <Form.List name="ladderList">
        {(fields, { add, remove }, { errors }) => (
          <>
            {fields.map((field, index) => (
              <Card
                key={field.key}
                size="small"
                title={`阶梯价 #${index + 1}`}
                style={listCardStyle}
                extra={
                  <Button type="link" danger onClick={() => remove(field.name)}>
                    <MinusCircleOutlined /> 删除
                  </Button>
                }
              >
                <Form.Item name={[field.name, 'id']} hidden>
                  <Input />
                </Form.Item>
                <Space size={12} wrap>
                  <Form.Item
                    name={[field.name, 'count']}
                    label="满足件数"
                    rules={[{ required: true, message: '请输入满足件数' }]}
                  >
                    <InputNumber min={1} style={priceInputStyle} />
                  </Form.Item>
                  <Form.Item
                    name={[field.name, 'discount']}
                    label="折扣"
                    rules={[{ required: true, message: '请输入折扣' }]}
                  >
                    <InputNumber min={0} style={priceInputStyle} />
                  </Form.Item>
                  <Form.Item
                    name={[field.name, 'price']}
                    label="阶梯价"
                    rules={[
                      { required: true, message: '请输入阶梯价' },
                      {
                        validator: async (_, value) => {
                          if (value === undefined || value === null) {
                            return;
                          }
                          if (value <= 0) {
                            throw new Error('阶梯价必须大于0');
                          }
                        },
                      },
                    ]}
                  >
                    <InputNumber min={0} precision={2} style={priceInputStyle} />
                  </Form.Item>
                </Space>
              </Card>
            ))}
            <Form.ErrorList errors={errors} />
            <Button type="dashed" block onClick={() => add({})}>
              <PlusOutlined /> 新增阶梯价
            </Button>
          </>
        )}
      </Form.List>

      <Divider orientation="left">满减</Divider>
      <Form.List name="fullList">
        {(fields, { add, remove }, { errors }) => (
          <>
            {fields.map((field, index) => (
              <Card
                key={field.key}
                size="small"
                title={`满减 #${index + 1}`}
                style={listCardStyle}
                extra={
                  <Button type="link" danger onClick={() => remove(field.name)}>
                    <MinusCircleOutlined /> 删除
                  </Button>
                }
              >
                <Form.Item name={[field.name, 'id']} hidden>
                  <Input />
                </Form.Item>
                <Space size={12} wrap>
                  <Form.Item
                    name={[field.name, 'fullPrice']}
                    label="满额"
                    rules={[
                      { required: true, message: '请输入满额' },
                      {
                        validator: async (_, value) => {
                          if (value === undefined || value === null) {
                            return;
                          }
                          if (value <= 0) {
                            throw new Error('满额必须大于0');
                          }
                        },
                      },
                    ]}
                  >
                    <InputNumber min={0} precision={2} style={priceInputStyle} />
                  </Form.Item>
                  <Form.Item
                    name={[field.name, 'reducePrice']}
                    label="减额"
                    dependencies={[['fullList', field.name, 'fullPrice']]}
                    rules={[
                      { required: true, message: '请输入减额' },
                      ({ getFieldValue }) => ({
                        validator: async (_, value) => {
                          if (value === undefined || value === null) {
                            return;
                          }
                          if (value <= 0) {
                            throw new Error('减额必须大于0');
                          }
                          const fullPrice = getFieldValue(['fullList', field.name, 'fullPrice']);
                          if (fullPrice !== undefined && fullPrice !== null && value > fullPrice) {
                            throw new Error('减额不能高于满额');
                          }
                        },
                      }),
                    ]}
                  >
                    <InputNumber min={0} precision={2} style={priceInputStyle} />
                  </Form.Item>
                </Space>
              </Card>
            ))}
            <Form.ErrorList errors={errors} />
            <Button type="dashed" block onClick={() => add({})}>
              <PlusOutlined /> 新增满减
            </Button>
          </>
        )}
      </Form.List>

      <Divider orientation="left">属性值</Divider>
      <Form.List name="attributeValueList">
        {(fields, { add, remove }) => (
          <>
            {fields.map((field, index) => (
              <Card
                key={field.key}
                size="small"
                title={`属性值 #${index + 1}`}
                style={listCardStyle}
                extra={
                  <Button type="link" danger onClick={() => remove(field.name)}>
                    <MinusCircleOutlined /> 删除
                  </Button>
                }
              >
                <Form.Item name={[field.name, 'id']} hidden>
                  <Input />
                </Form.Item>
                <Form.Item name={[field.name, 'spuId']} hidden>
                  <Input />
                </Form.Item>
                <Space size={12} wrap>
                  <Form.Item
                    name={[field.name, 'attributeId']}
                    label="属性 ID"
                    dependencies={['attributeValueList']}
                    rules={[
                      { required: true, message: '请选择商品属性' },
                      ({ getFieldValue }) => ({
                        validator: async (_, value) => {
                          const attributeValueList = getFieldValue(['attributeValueList']);
                          if (hasDuplicateAttributeId(attributeValueList, field.name, value)) {
                            throw new Error('当前草稿内已重复绑定该属性');
                          }
                        },
                      }),
                    ]}
                  >
                    <Select
                      showSearch
                      allowClear
                      optionFilterProp="label"
                      options={attributeOptions}
                      placeholder="请选择商品属性"
                      style={priceInputStyle}
                    />
                  </Form.Item>
                  <Form.Item
                    name={[field.name, 'value']}
                    label="属性值"
                    rules={[{ required: true, message: '请输入属性值' }]}
                  >
                    <Input placeholder="例如：黑色" />
                  </Form.Item>
                  <Form.Item
                    name={[field.name, 'status']}
                    label="状态"
                    rules={[{ required: true, message: '请选择属性状态' }]}
                  >
                    <Radio.Group options={statusOptions} optionType="button" />
                  </Form.Item>
                </Space>
              </Card>
            ))}
            <Button type="dashed" block onClick={() => add({ status: 1 })}>
              <PlusOutlined /> 新增属性值
            </Button>
          </>
        )}
      </Form.List>
    </>
  );
};

export default NestedDraftSections;
