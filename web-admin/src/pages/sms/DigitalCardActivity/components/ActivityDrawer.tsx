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
  Tag,
} from 'antd';
import type { DrawActivityFormValues, DrawPoolTemplateData } from '../data.d';
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

const WHEEL_COLORS = ['#f5a623', '#7ed321', '#4a90e2', '#9b59b6', '#e74c3c'];
const SLOT_COUNT = 5;

const WheelPreview: React.FC<{ slots: DrawPoolTemplateData[] }> = ({ slots }) => {
  const size = 160;
  const cx = size / 2;
  const cy = size / 2;
  const r = size / 2 - 8;
  const total = slots.reduce((acc, s) => acc + (s.probability || 0), 0);

  const sectors = (() => {
    let angle = -Math.PI / 2;
    return slots.map((slot, i) => {
      const prob = slot.probability || 0;
      const sweep = total > 0 ? (prob / total) * 2 * Math.PI : (2 * Math.PI) / (slots.length || 1);
      const endAngle = angle + sweep;
      const x1 = cx + r * Math.cos(angle);
      const y1 = cy + r * Math.sin(angle);
      const x2 = cx + r * Math.cos(endAngle);
      const y2 = cy + r * Math.sin(endAngle);
      const largeArc = sweep > Math.PI ? 1 : 0;
      const midAngle = angle + sweep / 2;
      const textR = r * 0.65;
      const result = {
        path: `M ${cx},${cy} L ${x1.toFixed(2)},${y1.toFixed(2)} A ${r},${r} 0 ${largeArc},1 ${x2.toFixed(2)},${y2.toFixed(2)} Z`,
        textX: cx + textR * Math.cos(midAngle),
        textY: cy + textR * Math.sin(midAngle),
        color: WHEEL_COLORS[i % WHEEL_COLORS.length],
        pct: total > 0 ? Math.round((prob / total) * 100) : 0,
        slotNum: i + 1,
        sweep,
      };
      angle = endAngle;
      return result;
    });
  })();

  if (slots.length === 0) {
    return (
      <svg width={size} height={size}>
        <circle cx={cx} cy={cy} r={r} fill="#f5f5f5" stroke="#d9d9d9" strokeWidth={1} />
        <text x={cx} y={cy} textAnchor="middle" dominantBaseline="middle" fill="#bbb" fontSize={12}>暂无格位</text>
      </svg>
    );
  }

  return (
    <svg width={size} height={size}>
      {sectors.map((s, i) => (
        <g key={i}>
          <path d={s.path} fill={s.color} stroke="white" strokeWidth={2} />
          {s.sweep > 0.3 && (
            <>
              <text x={s.textX} y={s.textY - 6} textAnchor="middle" dominantBaseline="middle" fill="white" fontSize={10} fontWeight="bold">
                {s.slotNum}
              </text>
              <text x={s.textX} y={s.textY + 7} textAnchor="middle" dominantBaseline="middle" fill="white" fontSize={9}>
                {s.pct}%
              </text>
            </>
          )}
        </g>
      ))}
      <circle cx={cx} cy={cy} r={r * 0.28} fill="white" stroke="#eee" strokeWidth={1} />
      <text x={cx} y={cy} textAnchor="middle" dominantBaseline="middle" fill="#aaa" fontSize={10}>转盘</text>
    </svg>
  );
};

const ProbabilitySumTag: React.FC<{ slots: DrawPoolTemplateData[] }> = ({ slots }) => {
  const sum = slots.reduce((acc, s) => acc + (s.probability || 0), 0);
  const valid = Math.abs(sum - 1) <= 0.0001;
  return (
    <div style={{ marginTop: 8 }}>
      <span style={{ fontSize: 12, color: '#666' }}>概率总和：</span>
      <strong style={{ color: valid ? '#52c41a' : '#ff4d4f' }}>{sum.toFixed(4)}</strong>
      <Tag color={valid ? 'success' : 'error'} style={{ marginLeft: 8 }}>
        {valid ? '✓ 总和 = 1，每次必中' : '✗ 总和必须等于 1'}
      </Tag>
    </div>
  );
};

const ActivityDrawer: React.FC<ActivityDrawerProps> = ({
  visible,
  title,
  current,
  onSubmit,
  onCancel,
}) => {
  const [form] = Form.useForm<DrawActivityFormValues>();
  const templates = Form.useWatch('templates', form) || [];
  const pools: any[] = Form.useWatch('pools', form) || [];

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
            <Space size={16} style={{ display: 'flex' }}>
              <Form.Item name="consumeType" label="消耗类型" initialValue="points">
                <Select style={{ width: 160 }} options={[{ label: '积分消耗', value: 'points' }, { label: '免费', value: 'free' }]} />
              </Form.Item>
              <Form.Item name="consumeAmount" label="每次消耗积分数" initialValue={1}>
                <InputNumber min={0} style={{ width: 160 }} />
              </Form.Item>
            </Space>
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

          <Panel header={`大转盘卡池配置（每池固定 ${SLOT_COUNT} 格位，总概率必须等于 1）`} key="pool">
            <Form.List name="pools">
              {(fields, { add, remove }) => (
                <>
                  {fields.map((field) => {
                    const currentSlots = (pools[field.name]?.templates || []) as DrawPoolTemplateData[];
                    return (
                      <div
                        key={field.key}
                        style={{ marginBottom: 24, border: '1px solid #f0f0f0', padding: 16, borderRadius: 8, background: '#fafafa' }}
                      >
                        <Space align="start" style={{ display: 'flex', width: '100%', marginBottom: 4 }}>
                          <Form.Item
                            {...field}
                            name={[field.name, 'poolName']}
                            label="卡池名称"
                            rules={[{ required: true, message: '请输入卡池名称' }]}
                          >
                            <Input style={{ width: 180 }} />
                          </Form.Item>
                          <Form.Item
                            {...field}
                            name={[field.name, 'poolCode']}
                            label="卡池编码"
                            rules={[{ required: true, message: '请输入卡池编码' }]}
                          >
                            <Input style={{ width: 160 }} />
                          </Form.Item>
                          <Form.Item {...field} name={[field.name, 'sort']} label="排序">
                            <InputNumber min={0} precision={0} style={{ width: 80 }} />
                          </Form.Item>
                          <Button danger style={{ marginTop: 30 }} onClick={() => remove(field.name)}>
                            删除卡池
                          </Button>
                        </Space>
                        <Form.Item {...field} name={[field.name, 'probabilityRule']} label="概率披露规则">
                          <Input placeholder="例如：5 格位概率以小数表达，总和必须等于 1，每次旋转必然中奖" />
                        </Form.Item>

                        <div style={{ display: 'flex', gap: 16, alignItems: 'flex-start' }}>
                          <div style={{ flex: 1 }}>
                            <div style={{ display: 'flex', gap: 8, marginBottom: 6, color: '#888', fontSize: 12, paddingLeft: 4 }}>
                              <span style={{ minWidth: 58 }}>格位</span>
                              <span style={{ width: 200 }}>卡片模板</span>
                              <span style={{ width: 90 }}>中奖概率</span>
                              <span style={{ width: 82 }}>发售数量</span>
                              <span style={{ width: 82 }}>剩余可发</span>
                            </div>
                            <Form.List name={[field.name, 'templates']}>
                              {(slotFields) => (
                                <>
                                  {slotFields.map((slotField, slotIdx) => (
                                    <div
                                      key={slotField.key}
                                      style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 6 }}
                                    >
                                      <Tag
                                        color={WHEEL_COLORS[slotIdx % WHEEL_COLORS.length]}
                                        style={{ minWidth: 52, textAlign: 'center', fontSize: 12, margin: 0 }}
                                      >
                                        格位 {slotIdx + 1}
                                      </Tag>
                                      <Form.Item name={[slotField.name, 'slotIndex']} hidden noStyle>
                                        <InputNumber />
                                      </Form.Item>
                                      <Form.Item
                                        name={[slotField.name, 'templateCode']}
                                        style={{ marginBottom: 0 }}
                                        rules={[{ required: true, message: '请选择模板' }]}
                                      >
                                        <Select
                                          style={{ width: 200 }}
                                          placeholder="选择卡片模板"
                                          options={(templates || []).map((item: any) => ({
                                            label: `${item.templateName || '未命名'} (${item.templateCode || '-'})`,
                                            value: item.templateCode,
                                          }))}
                                        />
                                      </Form.Item>
                                      <Form.Item
                                        name={[slotField.name, 'probability']}
                                        style={{ marginBottom: 0 }}
                                        rules={[
                                          { required: true, message: '请输入概率' },
                                          { type: 'number', min: 0.0001, max: 1, message: '范围 0.0001~1' },
                                        ]}
                                      >
                                        <InputNumber min={0.0001} max={1} step={0.01} style={{ width: 90 }} />
                                      </Form.Item>
                                      <Form.Item name={[slotField.name, 'saleLimit']} style={{ marginBottom: 0 }}>
                                        <InputNumber min={1} precision={0} style={{ width: 82 }} placeholder="发售数" />
                                      </Form.Item>
                                      <Form.Item name={[slotField.name, 'remainingLimit']} style={{ marginBottom: 0 }}>
                                        <InputNumber min={0} precision={0} style={{ width: 82 }} placeholder="剩余可发" />
                                      </Form.Item>
                                    </div>
                                  ))}
                                </>
                              )}
                            </Form.List>
                            <ProbabilitySumTag slots={currentSlots} />
                          </div>
                          <div style={{ flexShrink: 0, textAlign: 'center' }}>
                            <div style={{ fontSize: 11, color: '#888', marginBottom: 4 }}>概率预览</div>
                            <WheelPreview slots={currentSlots} />
                          </div>
                        </div>
                      </div>
                    );
                  })}
                  <Button
                    type="dashed"
                    onClick={() =>
                      add({
                        sort: 0,
                        status: 0,
                        probabilityRule: '',
                        templates: Array.from({ length: SLOT_COUNT }, (_, i) => ({
                          slotIndex: i + 1,
                          probability: parseFloat((1 / SLOT_COUNT).toFixed(4)),
                          saleLimit: 1,
                          remainingLimit: 1,
                          configLimit: 0,
                          templateCode: '',
                          rarity: '',
                        })),
                      })
                    }
                    block
                  >
                    新增转盘卡池（自动初始化 {SLOT_COUNT} 格位）
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
