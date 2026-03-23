import {
  buildDraftClientValidationErrors,
  buildDraftFieldErrors,
  getDuplicateSpecRowIndex,
  hasDuplicateAttributeId,
  hasDuplicateSkuCode,
  normalizeSkuSpecData,
} from './draftFeedback';

describe('draftFeedback', () => {
  it('normalizes spec data regardless of key order', () => {
    expect(normalizeSkuSpecData('{"容量":"128G","颜色":"黑色"}')).toBe('容量=128G|颜色=黑色');
    expect(normalizeSkuSpecData('{"颜色":"黑色","容量":"128G"}')).toBe('容量=128G|颜色=黑色');
  });

  it('detects duplicate sku code, spec and attribute ids', () => {
    const skuList = [
      { skuCode: 'SKU-001', specData: '{"颜色":"黑色","容量":"128G"}' },
      { skuCode: 'SKU-002', specData: '{"颜色":"白色","容量":"256G"}' },
      { skuCode: 'SKU-001', specData: '{"容量":"128G","颜色":"黑色"}' },
    ];

    expect(hasDuplicateSkuCode(skuList, 2, 'SKU-001')).toBe(true);
    expect(getDuplicateSpecRowIndex(skuList, 2, '{"颜色":"黑色","容量":"128G"}')).toBe(0);
    expect(
      hasDuplicateAttributeId(
        [
          { attributeId: 101, value: '黑色' },
          { attributeId: 102, value: '128G' },
          { attributeId: 101, value: '白色' },
        ],
        2,
        101,
      ),
    ).toBe(true);
  });

  it('maps backend sku and attribute errors back to form fields', () => {
    const values = {
      attributeValueList: [
        { attributeId: 301, value: '黑色' },
        { attributeId: 302, value: '' },
      ],
    };

    expect(buildDraftFieldErrors('第2行SKU预警库存不能大于可用库存')).toEqual([
      {
        name: ['skuList', 1, 'lowStock'],
        errors: ['第2行SKU预警库存不能大于可用库存'],
      },
    ]);

    expect(buildDraftFieldErrors('商品属性值不能为空: 302', values)).toEqual([
      {
        name: ['attributeValueList', 1, 'value'],
        errors: ['商品属性值不能为空: 302'],
      },
    ]);

    expect(buildDraftFieldErrors('商品草稿不可保存：至少一个SKU需要具备有效库存')).toEqual([
      {
        name: ['skuList'],
        errors: ['商品草稿不可保存：至少一个SKU需要具备有效库存'],
      },
    ]);
  });

  it('builds client-side validation errors for promotion-dependent drafts', () => {
    expect(
      buildDraftClientValidationErrors({
        skuList: [
          { name: '黑色/128G', price: 0, stock: 0, lowStock: 0, specData: '{"颜色":"黑色"}' },
        ],
      }),
    ).toEqual([
      {
        name: ['skuList'],
        errors: ['至少一个SKU需要具备有效库存和售价'],
      },
    ]);

    expect(
      buildDraftClientValidationErrors({
        promotionType: 2,
        skuList: [
          { name: '黑色/128G', price: 99, stock: 10, lowStock: 1, specData: '{"颜色":"黑色"}' },
        ],
      }),
    ).toEqual([
      {
        name: ['memberPriceList'],
        errors: ['当前促销类型为会员价，至少需要配置一条会员价'],
      },
    ]);

    expect(
      buildDraftClientValidationErrors({
        promotionType: 3,
        skuList: [
          { name: '黑色/128G', price: 99, stock: 10, lowStock: 1, specData: '{"颜色":"黑色"}' },
        ],
      }),
    ).toEqual([
      {
        name: ['ladderList'],
        errors: ['当前促销类型为阶梯价，至少需要配置一条阶梯价'],
      },
    ]);
  });
});
