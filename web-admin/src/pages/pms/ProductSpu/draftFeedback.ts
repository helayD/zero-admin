import type {
  ProductSpuAttributeValueItem,
  ProductSpuDraftFormValues,
  ProductSpuSkuItem,
} from './data.d';

export type DraftFieldError = {
  name: (string | number)[];
  errors: string[];
};

function trimText(value?: string) {
  return `${value ?? ''}`.trim();
}

export function normalizeSkuSpecData(raw?: string): string {
  const trimmed = trimText(raw);
  if (!trimmed) {
    throw new Error('缺少规格主键');
  }

  let parsed: Record<string, unknown>;
  try {
    parsed = JSON.parse(trimmed) as Record<string, unknown>;
  } catch (error) {
    throw new Error('规格数据不是合法JSON');
  }

  const entries = Object.entries(parsed);
  if (entries.length === 0) {
    throw new Error('规格数据不能为空对象');
  }

  const normalizedEntries = entries.map(([key, value]) => {
    const normalizedKey = trimText(key);
    const normalizedValue = trimText(value === null || value === undefined ? '' : String(value));
    if (!normalizedKey || !normalizedValue) {
      throw new Error('规格键和值不能为空');
    }
    return `${normalizedKey}=${normalizedValue}`;
  });

  normalizedEntries.sort();
  return normalizedEntries.join('|');
}

export function hasDuplicateSkuCode(
  skuList: ProductSpuSkuItem[] | undefined,
  currentIndex: number,
  skuCode?: string,
): boolean {
  const normalizedCode = trimText(skuCode);
  if (!normalizedCode || !Array.isArray(skuList)) {
    return false;
  }

  return skuList.some(
    (item, index) => index !== currentIndex && trimText(item?.skuCode) === normalizedCode,
  );
}

export function getDuplicateSpecRowIndex(
  skuList: ProductSpuSkuItem[] | undefined,
  currentIndex: number,
  specData?: string,
): number {
  if (!Array.isArray(skuList)) {
    return -1;
  }

  const normalizedSpec = normalizeSkuSpecData(specData);
  for (let index = 0; index < skuList.length; index += 1) {
    if (index === currentIndex) {
      continue;
    }

    const candidate = trimText(skuList[index]?.specData);
    if (!candidate) {
      continue;
    }

    try {
      if (normalizeSkuSpecData(candidate) === normalizedSpec) {
        return index;
      }
    } catch (error) {
      continue;
    }
  }

  return -1;
}

export function hasDuplicateAttributeId(
  attributeValueList: ProductSpuAttributeValueItem[] | undefined,
  currentIndex: number,
  attributeId?: number,
): boolean {
  if (!attributeId || !Array.isArray(attributeValueList)) {
    return false;
  }

  return attributeValueList.some(
    (item, index) => index !== currentIndex && item?.attributeId === attributeId,
  );
}

function findAttributeIndex(values: ProductSpuDraftFormValues, attributeId: number): number {
  return (values.attributeValueList || []).findIndex((item) => item?.attributeId === attributeId);
}

function buildSkuRowFieldError(
  index: number,
  field: keyof ProductSpuSkuItem,
  message: string,
): DraftFieldError {
  return {
    name: ['skuList', index, field],
    errors: [message],
  };
}

function hasAnySaleableSku(skuList: ProductSpuSkuItem[] | undefined): boolean {
  if (!Array.isArray(skuList)) {
    return false;
  }

  return skuList.some((item) => Number(item?.price || 0) > 0 && Number(item?.stock || 0) > 0);
}

export function buildDraftClientValidationErrors(
  values: ProductSpuDraftFormValues = {},
): DraftFieldError[] {
  const errors: DraftFieldError[] = [];
  const skuList = values.skuList || [];

  if (!Array.isArray(skuList) || skuList.length === 0) {
    return [
      {
        name: ['skuList'],
        errors: ['至少需要一个有效SKU'],
      },
    ];
  }

  if (!hasAnySaleableSku(skuList)) {
    errors.push({
      name: ['skuList'],
      errors: ['至少一个SKU需要具备有效库存和售价'],
    });
  }

  if (
    values.promotionType === 2 &&
    (!values.memberPriceList || values.memberPriceList.length === 0)
  ) {
    errors.push({
      name: ['memberPriceList'],
      errors: ['当前促销类型为会员价，至少需要配置一条会员价'],
    });
  }

  if (values.promotionType === 3 && (!values.ladderList || values.ladderList.length === 0)) {
    errors.push({
      name: ['ladderList'],
      errors: ['当前促销类型为阶梯价，至少需要配置一条阶梯价'],
    });
  }

  if (values.promotionType === 4 && (!values.fullList || values.fullList.length === 0)) {
    errors.push({
      name: ['fullList'],
      errors: ['当前促销类型为满减价，至少需要配置一条满减规则'],
    });
  }

  return errors;
}

export function buildDraftFieldErrors(
  message: string,
  values: ProductSpuDraftFormValues = {},
): DraftFieldError[] {
  const normalizedMessage = trimText(message);
  if (!normalizedMessage) {
    return [];
  }

  const directFieldMap: { pattern: RegExp; name: (string | number)[] }[] = [
    {
      pattern: /^(商品名称不能为空)$/,
      name: ['name'],
    },
    {
      pattern: /^(商品货号不能为空)$/,
      name: ['productSn'],
    },
    {
      pattern: /^(商品主图不能为空)$/,
      name: ['mainPic'],
    },
    {
      pattern:
        /^(缺少有效商品分类ID|商品分类不存在|商品分类已失效，无法继续建档|当前主体无权将商品归属到该商品分类)$/,
      name: ['categoryId'],
    },
    {
      pattern:
        /^(缺少有效商品品牌ID|商品品牌不存在|商品品牌已失效，无法继续建档|当前主体无权将商品归属到该商品品牌)$/,
      name: ['brandId'],
    },
  ];

  for (const item of directFieldMap) {
    if (item.pattern.test(normalizedMessage)) {
      return [{ name: item.name, errors: [normalizedMessage] }];
    }
  }

  if (
    normalizedMessage === '至少需要一个有效SKU' ||
    normalizedMessage === '商品草稿不可保存：至少一个SKU需要具备有效库存' ||
    normalizedMessage === '至少一个SKU需要具备有效库存和售价'
  ) {
    return [{ name: ['skuList'], errors: [normalizedMessage] }];
  }

  const skuFieldMatchers: { regex: RegExp; field: keyof ProductSpuSkuItem }[] = [
    { regex: /^第(\d+)行SKU名称不能为空$/, field: 'name' },
    { regex: /^第(\d+)行SKU价格必须大于0$/, field: 'price' },
    { regex: /^第(\d+)行SKU促销价不能小于0$/, field: 'promotionPrice' },
    { regex: /^第(\d+)行SKU促销价不能高于销售价$/, field: 'promotionPrice' },
    { regex: /^第(\d+)行SKU库存不能小于0$/, field: 'stock' },
    { regex: /^第(\d+)行SKU预警库存不能小于0$/, field: 'lowStock' },
    { regex: /^第(\d+)行SKU预警库存不能大于可用库存$/, field: 'lowStock' },
    { regex: /^第(\d+)行SKU规格组合非法: .+$/, field: 'specData' },
    { regex: /^第(\d+)行SKU规格组合重复$/, field: 'specData' },
    { regex: /^第(\d+)行SKU编码重复$/, field: 'skuCode' },
    { regex: /^第(\d+)行SKU编码冲突: .+$/, field: 'skuCode' },
  ];

  for (const matcher of skuFieldMatchers) {
    const result = normalizedMessage.match(matcher.regex);
    if (result) {
      const rowIndex = Math.max(Number(result[1]) - 1, 0);
      return [buildSkuRowFieldError(rowIndex, matcher.field, normalizedMessage)];
    }
  }

  const duplicateAttributeMatch = normalizedMessage.match(/^商品属性重复填写: (\d+)$/);
  if (duplicateAttributeMatch) {
    const attributeId = Number(duplicateAttributeMatch[1]);
    const rowIndex = findAttributeIndex(values, attributeId);
    if (rowIndex >= 0) {
      return [
        {
          name: ['attributeValueList', rowIndex, 'attributeId'],
          errors: [normalizedMessage],
        },
      ];
    }
  }

  const emptyAttributeValueMatch = normalizedMessage.match(/^商品属性值不能为空: (\d+)$/);
  if (emptyAttributeValueMatch) {
    const attributeId = Number(emptyAttributeValueMatch[1]);
    const rowIndex = findAttributeIndex(values, attributeId);
    if (rowIndex >= 0) {
      return [
        {
          name: ['attributeValueList', rowIndex, 'value'],
          errors: [normalizedMessage],
        },
      ];
    }
  }

  if (
    normalizedMessage === '缺少有效商品属性ID' ||
    normalizedMessage === '商品属性不存在' ||
    normalizedMessage === '商品属性已失效，无法继续建档' ||
    normalizedMessage === '当前主体无权绑定该商品属性'
  ) {
    return [
      {
        name: ['attributeValueList', 0, 'attributeId'],
        errors: [normalizedMessage],
      },
    ];
  }

  return [];
}
