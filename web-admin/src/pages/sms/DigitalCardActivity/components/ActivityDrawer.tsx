import React, { useEffect } from 'react';
import {
  Button,
  Collapse,
  DatePicker,
  Drawer,
  Form,
  Input,
  InputNumber,
  Select,
  Space,
  Switch,
} from 'antd';
import type { DrawActivityFormValues } from '../data.d';
import { toFormInitialValues } from '../helper';

const { RangePicker } = DatePicker;
const { Panel } = Collapse;
const circulationPlaceholder = '至少包含：禁止集中竞价、禁止连续挂牌、禁止收益承诺';

type ActivityDrawerProps = {
  visible: boolean;
  title: string;
  current?: DrawActivityFormValues;
  onSubmit: (values: DrawActivityFormValues) => Promise<boolean>;
  onCancel: () => void;
};

const statusOptions = [
  { label: '草稿/下线', value: 0 },
  { label: '已发布', value: 1 },
  { label: '已归档', value: 2 },
];

const auditStatusOptions = [
  { label: '缺失', value: 0 },
  { label: '待审批', value: 1 },
  { label: '已通过', value: 2 },
  { label: '已拒绝', value: 3 },
];

const assetStatusOptions = [
  { label: '缺失', value: 0 },
  { label: '处理中', value: 1 },
  { label: '已通过', value: 2 },
  { label: '已拒绝', value: 3 },
];

const ActivityDrawer: React.FC<ActivityDrawerProps> = ({
  visible,
  title,
  current,
  onSubmit,
  onCancel,
}) => {
  const [form] = Form.useForm<DrawActivityFormValues>();
  const templates = Form.useWatch('templates', form) || [];

  useEffect(() => {
    form.setFieldsValue(toFormInitialValues(current));
  }, [current, form, visible]);

  return (
    <Drawer
      title={title}
      width={960}
      visible={visible}
      destroyOnClose
      onClose={onCancel}
      extra={
        <Space>
          <Button onClick={onCancel}>取消</Button>
          <Button
            type="primary"
            onClick={async () => {
              const values = await form.validateFields();
              const success = await onSubmit(values);
              if (success) {
                form.resetFields();
              }
            }}
          >
            保存
          </Button>
        </Space>
      }
    >
      <Form form={form} layout="vertical">
        <Collapse defaultActiveKey={['base', 'pool', 'template', 'compliance']}>
          <Panel header="活动基础信息" key="base">
            <Form.Item name="activityCode" label="活动编码" rules={[{ required: true, message: '请输入活动编码' }]}>
              <Input placeholder="如 DRAW20260416001" />
            </Form.Item>
            <Form.Item name="name" label="活动名称" rules={[{ required: true, message: '请输入活动名称' }]}>
              <Input />
            </Form.Item>
            <Form.Item name="activeTime" label="活动时间" rules={[{ required: true, message: '请选择活动时间' }]}>
              <RangePicker showTime style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="ruleSummary" label="规则摘要">
              <Input.TextArea rows={2} />
            </Form.Item>
            <Form.Item name="participantConditionSummary" label="参与条件摘要">
              <Input.TextArea rows={2} />
            </Form.Item>
            <Form.Item name="consumeRuleSummary" label="消耗规则摘要">
              <Input.TextArea rows={2} />
            </Form.Item>
            <Space size={24} style={{ display: 'flex' }}>
              <Form.Item name="realNameRequired" label="实名要求" valuePropName="checked" getValueFromEvent={(checked) => (checked ? 1 : 0)}>
                <Switch checkedChildren="需要" unCheckedChildren="可选" />
              </Form.Item>
              <Form.Item name="isEnabled" label="启用状态" valuePropName="checked" getValueFromEvent={(checked) => (checked ? 1 : 0)}>
                <Switch checkedChildren="启用" unCheckedChildren="停用" />
              </Form.Item>
            </Space>
            <Space size={16} style={{ display: 'flex' }}>
              <Form.Item name="status" label="活动状态">
                <Select options={statusOptions} style={{ width: 180 }} />
              </Form.Item>
              <Form.Item name="auditStatus" label="审批状态">
                <Select options={auditStatusOptions} style={{ width: 180 }} />
              </Form.Item>
            </Space>
          </Panel>

          <Panel header="Flutter 首页显著入口" key="home">
            <Form.Item name={['homeEntry', 'showOnHome']} label="首页显著展示" valuePropName="checked" getValueFromEvent={(checked) => (checked ? 1 : 0)}>
              <Switch checkedChildren="展示" unCheckedChildren="隐藏" />
            </Form.Item>
            <Form.Item name={['homeEntry', 'homeEntryTitle']} label="入口主文案">
              <Input placeholder="例如：限时数字卡片抽赏" />
            </Form.Item>
            <Form.Item name={['homeEntry', 'homeEntrySubtitle']} label="入口副文案">
              <Input placeholder="例如：完成实名后即可参与" />
            </Form.Item>
            <Form.Item name={['homeEntry', 'homeEntryImage']} label="封面图 URL">
              <Input />
            </Form.Item>
            <Space size={16} style={{ display: 'flex' }}>
              <Form.Item name={['homeEntry', 'homeEntrySort']} label="排序优先级">
                <InputNumber min={0} precision={0} />
              </Form.Item>
              <Form.Item name={['homeEntry', 'landingTargetType']} label="落地页类型">
                <Select
                  style={{ width: 200 }}
                  options={[
                    { label: '活动页', value: 'activity' },
                    { label: '专题页', value: 'subject' },
                  ]}
                />
              </Form.Item>
            </Space>
            <Form.Item name={['homeEntry', 'landingTargetValue']} label="落地页目标值">
              <Input placeholder="活动编码或专题编码" />
            </Form.Item>
          </Panel>

          <Panel header="卡片模板配置" key="template">
            <Form.List name="templates">
              {(fields, { add, remove }) => (
                <>
                  {fields.map((field) => (
                    <Space key={field.key} direction="vertical" style={{ width: '100%', marginBottom: 16 }}>
                      <Space align="start" style={{ display: 'flex', width: '100%' }}>
                        <Form.Item
                          {...field}
                          name={[field.name, 'templateName']}
                          fieldKey={[field.fieldKey || field.name, 'templateName']}
                          label="模板名称"
                          rules={[{ required: true, message: '请输入模板名称' }]}
                        >
                          <Input />
                        </Form.Item>
                        <Form.Item
                          {...field}
                          name={[field.name, 'templateCode']}
                          fieldKey={[field.fieldKey || field.name, 'templateCode']}
                          label="模板编码"
                          rules={[{ required: true, message: '请输入模板编码' }]}
                        >
                          <Input />
                        </Form.Item>
                        <Form.Item
                          {...field}
                          name={[field.name, 'issueLimit']}
                          fieldKey={[field.fieldKey || field.name, 'issueLimit']}
                          label="发行上限"
                          rules={[{ required: true, message: '请输入发行上限' }]}
                        >
                          <InputNumber min={1} precision={0} />
                        </Form.Item>
                        <Button danger style={{ marginTop: 30 }} onClick={() => remove(field.name)}>
                          删除模板
                        </Button>
                      </Space>
                      <Space align="start" style={{ display: 'flex', width: '100%' }}>
                        <Form.Item {...field} name={[field.name, 'rarity']} label="稀有度">
                          <Input placeholder="SSR / SR / R" />
                        </Form.Item>
                        <Form.Item {...field} name={[field.name, 'copyrightOwner']} label="版权归属">
                          <Input />
                        </Form.Item>
                        <Form.Item {...field} name={[field.name, 'copyrightProofSummary']} label="版权凭证摘要">
                          <Input />
                        </Form.Item>
                      </Space>
                      <Form.Item {...field} name={[field.name, 'cardFaceImage']} label="卡面资源 URL">
                        <Input />
                      </Form.Item>
                      <Form.Item {...field} name={[field.name, 'displayCopy']} label="展示文案">
                        <Input.TextArea rows={2} />
                      </Form.Item>
                      <Form.Item {...field} name={[field.name, 'circulationLimitSummary']} label="默认流转限制">
                        <Input.TextArea rows={2} placeholder={circulationPlaceholder} />
                      </Form.Item>
                    </Space>
                  ))}
                  <Button type="dashed" onClick={() => add({ displayStatus: 1, contentAuditStatus: 0, status: 0, auditStatus: 0 })} block>
                    新增卡片模板
                  </Button>
                </>
              )}
            </Form.List>
          </Panel>

          <Panel header="卡池配置" key="pool">
            <Form.List name="pools">
              {(fields, { add, remove }) => (
                <>
                  {fields.map((field) => (
                    <Space key={field.key} direction="vertical" style={{ width: '100%', marginBottom: 24 }}>
                      <Space align="start" style={{ display: 'flex', width: '100%' }}>
                        <Form.Item
                          {...field}
                          name={[field.name, 'poolName']}
                          label="卡池名称"
                          rules={[{ required: true, message: '请输入卡池名称' }]}
                        >
                          <Input />
                        </Form.Item>
                        <Form.Item
                          {...field}
                          name={[field.name, 'poolCode']}
                          label="卡池编码"
                          rules={[{ required: true, message: '请输入卡池编码' }]}
                        >
                          <Input />
                        </Form.Item>
                        <Form.Item {...field} name={[field.name, 'sort']} label="排序">
                          <InputNumber min={0} precision={0} />
                        </Form.Item>
                        <Button danger style={{ marginTop: 30 }} onClick={() => remove(field.name)}>
                          删除卡池
                        </Button>
                      </Space>
                      <Form.Item {...field} name={[field.name, 'probabilityRule']} label="卡池概率披露规则">
                        <Input placeholder="例如：概率以 0-1 小数表达，卡池总和必须为 1" />
                      </Form.Item>
                      <Form.List name={[field.name, 'templates']}>
                        {(mappingFields, mappingOperator) => (
                          <>
                            {mappingFields.map((mappingField) => (
                              <Space key={mappingField.key} align="start" style={{ display: 'flex', width: '100%' }}>
                                <Form.Item
                                  {...mappingField}
                                  name={[mappingField.name, 'templateCode']}
                                  label="模板编码"
                                  rules={[{ required: true, message: '请选择模板编码' }]}
                                >
                                  <Select
                                    style={{ width: 180 }}
                                    options={(templates || []).map((item: any) => ({
                                      label: `${item.templateName || item.templateCode || '未命名模板'} (${item.templateCode || '-'})`,
                                      value: item.templateCode,
                                    }))}
                                  />
                                </Form.Item>
                                <Form.Item {...mappingField} name={[mappingField.name, 'rarity']} label="稀有度">
                                  <Input style={{ width: 120 }} />
                                </Form.Item>
                                <Form.Item
                                  {...mappingField}
                                  name={[mappingField.name, 'probability']}
                                  label="概率"
                                  rules={[{ required: true, message: '请输入概率' }]}
                                >
                                  <InputNumber min={0.0001} max={1} step={0.01} />
                                </Form.Item>
                                <Form.Item
                                  {...mappingField}
                                  name={[mappingField.name, 'saleLimit']}
                                  label="发售数量"
                                  rules={[{ required: true, message: '请输入发售数量' }]}
                                >
                                  <InputNumber min={1} precision={0} />
                                </Form.Item>
                                <Form.Item {...mappingField} name={[mappingField.name, 'remainingLimit']} label="剩余可发">
                                  <InputNumber min={0} precision={0} />
                                </Form.Item>
                                <Form.Item {...mappingField} name={[mappingField.name, 'configLimit']} label="配置上限">
                                  <InputNumber min={0} precision={0} />
                                </Form.Item>
                                <Button danger style={{ marginTop: 30 }} onClick={() => mappingOperator.remove(mappingField.name)}>
                                  删除
                                </Button>
                              </Space>
                            ))}
                            <Button type="dashed" onClick={() => mappingOperator.add({ probability: 0, saleLimit: 1, remainingLimit: 1, configLimit: 0 })} block>
                              新增卡池模板映射
                            </Button>
                          </>
                        )}
                      </Form.List>
                    </Space>
                  ))}
                  <Button type="dashed" onClick={() => add({ sort: 0, status: 0, templates: [] })} block>
                    新增卡池
                  </Button>
                </>
              )}
            </Form.List>
          </Panel>

          <Panel header="合规与发布预检" key="compliance">
            <Form.Item name="probabilityRule" label="概率披露方式">
              <Input />
            </Form.Item>
            <Form.Item name="complianceRuleSummary" label="合规规则摘要">
              <Input.TextArea rows={2} />
            </Form.Item>
            <Form.Item name="circulationLimitSummary" label="活动流转限制">
              <Input.TextArea rows={2} placeholder={circulationPlaceholder} />
            </Form.Item>
            <Form.Item name="approvalRecordRef" label="审批记录引用">
              <Input placeholder="例如 APPROVAL-2026-0001" />
            </Form.Item>
            <Space size={16} style={{ display: 'flex' }}>
              <Form.Item name="copyrightStatus" label="版权校验状态">
                <Select options={assetStatusOptions} style={{ width: 180 }} />
              </Form.Item>
              <Form.Item name="contentAuditStatus" label="内容审核状态">
                <Select options={assetStatusOptions} style={{ width: 180 }} />
              </Form.Item>
            </Space>
          </Panel>
        </Collapse>
      </Form>
    </Drawer>
  );
};

export default ActivityDrawer;
