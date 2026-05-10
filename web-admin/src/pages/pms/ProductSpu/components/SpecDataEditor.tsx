import React, { useEffect, useState } from 'react';
import { Button, Input, Space } from 'antd';
import { MinusCircleOutlined, PlusOutlined } from '@ant-design/icons';

/**
 * SKU 规格数据结构化编辑器
 * 对外仍以 JSON 字符串作为 value/onChange 协议（向后端透传不变），
 * 内部以 [{key, value}] 数组让用户行编辑，避免手写 JSON 语法错误。
 */

type Pair = { key: string; value: string };

const parseJson = (raw?: string): Pair[] => {
  if (!raw) return [];
  try {
    const trimmed = String(raw).trim();
    if (!trimmed) return [];
    const obj = JSON.parse(trimmed);
    if (typeof obj !== 'object' || obj === null || Array.isArray(obj)) return [];
    return Object.entries(obj).map(([k, v]) => ({
      key: String(k),
      value: v === null || v === undefined ? '' : String(v),
    }));
  } catch {
    return [];
  }
};

const buildJson = (pairs: Pair[]): string => {
  const obj: Record<string, string> = {};
  pairs.forEach((p) => {
    const k = (p.key ?? '').trim();
    const v = (p.value ?? '').trim();
    if (k) obj[k] = v;
  });
  if (Object.keys(obj).length === 0) return '';
  return JSON.stringify(obj);
};

interface Props {
  value?: string;
  onChange?: (val: string) => void;
}

const SpecDataEditor: React.FC<Props> = ({ value, onChange }) => {
  const [pairs, setPairs] = useState<Pair[]>(() => parseJson(value));

  // 当外部 value 变化（form.setFieldsValue 回填、reset 等），同步内部状态
  useEffect(() => {
    const parsed = parseJson(value);
    // 仅当生成的 JSON 与当前一致时跳过，避免把用户正在编辑的空 key 行覆盖掉
    if (buildJson(parsed) !== buildJson(pairs)) {
      setPairs(parsed);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [value]);

  const emit = (next: Pair[]) => {
    setPairs(next);
    onChange?.(buildJson(next));
  };

  const updatePair = (idx: number, patch: Partial<Pair>) => {
    emit(pairs.map((p, i) => (i === idx ? { ...p, ...patch } : p)));
  };

  const addPair = () => {
    emit([...pairs, { key: '', value: '' }]);
  };

  const removePair = (idx: number) => {
    emit(pairs.filter((_, i) => i !== idx));
  };

  return (
    <div>
      {pairs.length === 0 && (
        <div style={{ marginBottom: 8, color: '#999', fontSize: 12 }}>
          暂无规格属性，点击下方按钮添加，例如「颜色 = 黑色」「容量 = 128G」
        </div>
      )}
      {pairs.map((p, idx) => (
        <Space
          key={idx}
          align="baseline"
          style={{ display: 'flex', marginBottom: 8, width: '100%' }}
          size={8}
        >
          <Input
            placeholder="属性名（如 颜色）"
            value={p.key}
            style={{ width: 160 }}
            maxLength={20}
            onChange={(e) => updatePair(idx, { key: e.target.value })}
          />
          <span style={{ color: '#999' }}>=</span>
          <Input
            placeholder="属性值（如 黑色）"
            value={p.value}
            style={{ width: 220 }}
            maxLength={40}
            onChange={(e) => updatePair(idx, { value: e.target.value })}
          />
          <MinusCircleOutlined
            onClick={() => removePair(idx)}
            style={{ color: '#ff4d4f', cursor: 'pointer', fontSize: 16 }}
          />
        </Space>
      ))}
      <Button
        type="dashed"
        size="small"
        icon={<PlusOutlined />}
        onClick={addPair}
        style={{ marginTop: 4 }}
      >
        新增规格属性
      </Button>
    </div>
  );
};

export default SpecDataEditor;
