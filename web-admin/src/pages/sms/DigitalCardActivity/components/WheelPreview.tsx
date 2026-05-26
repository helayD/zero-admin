import React from 'react';
import { Tag } from 'antd';
import type { DrawPoolTemplateData } from '../data.d';

export const WHEEL_COLORS = ['#f5a623', '#7ed321', '#4a90e2', '#9b59b6', '#e74c3c'];
export const SLOT_COUNT = 5;

export const WheelPreview: React.FC<{ slots: DrawPoolTemplateData[]; size?: number }> = ({
  slots,
  size = 160,
}) => {
  const cx = size / 2;
  const cy = size / 2;
  const r = size / 2 - 8;
  const total = slots.reduce((acc, s) => acc + (s.probability || 0), 0);

  const sectors = (() => {
    let angle = -Math.PI / 2;
    return slots.map((slot, i) => {
      const prob = slot.probability || 0;
      const sweep =
        total > 0 ? (prob / total) * 2 * Math.PI : (2 * Math.PI) / (slots.length || 1);
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
        label: slot.templateName || `格${slot.slotIndex || i + 1}`,
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
        <text x={cx} y={cy} textAnchor="middle" dominantBaseline="middle" fill="#bbb" fontSize={12}>
          暂无格位
        </text>
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
              <text
                x={s.textX}
                y={s.textY - 6}
                textAnchor="middle"
                dominantBaseline="middle"
                fill="white"
                fontSize={10}
                fontWeight="bold"
              >
                {slot_label(slots[i])}
              </text>
              <text
                x={s.textX}
                y={s.textY + 7}
                textAnchor="middle"
                dominantBaseline="middle"
                fill="white"
                fontSize={9}
              >
                {s.pct}%
              </text>
            </>
          )}
        </g>
      ))}
      <circle cx={cx} cy={cy} r={r * 0.28} fill="white" stroke="#eee" strokeWidth={1} />
      <text x={cx} y={cy} textAnchor="middle" dominantBaseline="middle" fill="#aaa" fontSize={10}>
        转盘
      </text>
    </svg>
  );
};

function slot_label(slot: DrawPoolTemplateData): string {
  return slot.rarity || `格${slot.slotIndex || 0}`;
}

export const ProbabilitySumTag: React.FC<{ slots: DrawPoolTemplateData[] }> = ({ slots }) => {
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
