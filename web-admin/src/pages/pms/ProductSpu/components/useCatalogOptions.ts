import { message } from 'antd';
import { useEffect, useMemo, useState } from 'react';
import type { GovernanceScopeValue } from '@/pages/system/components/governance';
import { toGovernancePayload } from '@/pages/system/components/governance';
import type { ProductAttributeListItem } from '@/pages/pms/ProductAttribute/data.d';
import { queryProductAttributeList } from '@/pages/pms/ProductAttribute/service';
import type { ProductBrandListItem } from '@/pages/pms/ProductBrand/data.d';
import { queryProductBrandList } from '@/pages/pms/ProductBrand/service';
import type { ProductCategoryListItem } from '@/pages/pms/ProductCategory/data';
import { queryProductCategoryList } from '@/pages/pms/ProductCategory/service';
import { tree } from '@/utils/utils';

export type SelectOption = {
  label: string;
  value: number;
};

type CategoryTreeOption = {
  id: number;
  value: number;
  label: string;
  title: string;
  parentId: number;
  disabled?: boolean;
  children?: CategoryTreeOption[];
};

type CatalogOptionsResult = {
  attributeOptions: SelectOption[];
  brandOptions: SelectOption[];
  categoryNameById: Record<number, string>;
  categoryPathById: Record<number, number[]>;
  categoryTreeOptions: CategoryTreeOption[];
};

const emptyCatalogOptions: CatalogOptionsResult = {
  attributeOptions: [],
  brandOptions: [],
  categoryNameById: {},
  categoryPathById: {},
  categoryTreeOptions: [],
};

const buildCategoryPathMap = (items: ProductCategoryListItem[]) => {
  const byId = items.reduce<Record<number, ProductCategoryListItem>>((acc, item) => {
    acc[item.id] = item;
    return acc;
  }, {});

  const pathById = items.reduce<Record<number, number[]>>((acc, item) => {
    const path: number[] = [];
    let current: ProductCategoryListItem | undefined = item;

    while (current) {
      path.unshift(current.id);
      if (!current.parentId) {
        break;
      }
      current = byId[current.parentId];
    }

    acc[item.id] = path;
    return acc;
  }, {});

  const categoryNameById = items.reduce<Record<number, string>>((acc, item) => {
    acc[item.id] = item.name;
    return acc;
  }, {});

  return { categoryNameById, pathById };
};

export const useCatalogOptions = (
  visible: boolean,
  scope?: GovernanceScopeValue,
): CatalogOptionsResult => {
  const [catalogOptions, setCatalogOptions] = useState<CatalogOptionsResult>(emptyCatalogOptions);

  const scopePayload = useMemo(() => toGovernancePayload(scope), [scope]);

  useEffect(() => {
    if (!visible) {
      return;
    }

    let cancelled = false;

    Promise.all([
      queryProductCategoryList({ pageSize: 200, current: 1, ...scopePayload }),
      queryProductBrandList({ pageSize: 200, current: 1, ...scopePayload }),
      queryProductAttributeList({ pageSize: 200, current: 1, ...scopePayload }),
    ])
      .then(([categoryResp, brandResp, attributeResp]) => {
        if (cancelled) {
          return;
        }

        const categoryItems: ProductCategoryListItem[] =
          categoryResp?.code === '000000' && Array.isArray(categoryResp.data)
            ? categoryResp.data
            : [];
        const brandItems: ProductBrandListItem[] =
          brandResp?.code === '000000' && Array.isArray(brandResp.data) ? brandResp.data : [];
        const attributeItems: ProductAttributeListItem[] =
          attributeResp?.code === '000000' && Array.isArray(attributeResp.data)
            ? attributeResp.data
            : [];

        if (categoryResp?.code !== '000000') {
          message.error(categoryResp?.message || categoryResp?.msg || '加载商品分类失败');
        }
        if (brandResp?.code !== '000000') {
          message.error(brandResp?.message || brandResp?.msg || '加载商品品牌失败');
        }
        if (attributeResp?.code !== '000000') {
          message.error(attributeResp?.message || attributeResp?.msg || '加载商品属性失败');
        }

        const categoryTreeOptions = tree(
          categoryItems.map((item) => ({
            value: item.id,
            id: item.id,
            label: item.isEnabled === 1 ? item.name : `${item.name}（已停用）`,
            title: item.isEnabled === 1 ? item.name : `${item.name}（已停用）`,
            parentId: item.parentId,
            disabled: item.isEnabled !== 1,
          })),
          0,
          'parentId',
        ) as CategoryTreeOption[];

        const { categoryNameById, pathById } = buildCategoryPathMap(categoryItems);

        const enabledBrands = brandItems.filter((item) => item.isEnabled === 1);
        const enabledAttributes = attributeItems.filter((item) => item.status === 1);

        setCatalogOptions({
          attributeOptions: enabledAttributes.map((item) => ({
            label: item.name,
            value: item.id,
          })),
          brandOptions: enabledBrands.map((item) => ({
            label: item.name,
            value: item.id,
          })),
          categoryNameById,
          categoryPathById: pathById,
          categoryTreeOptions,
        });
      })
      .catch(() => {
        if (!cancelled) {
          message.error('加载商品建档目录候选项失败');
        }
      });

    return () => {
      cancelled = true;
    };
  }, [scopePayload, visible]);

  return catalogOptions;
};
