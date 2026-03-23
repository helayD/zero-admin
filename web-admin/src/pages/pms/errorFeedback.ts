export type CatalogActionError = {
  title: string;
  description: string;
};

const fallbackDescription = '请检查分类、品牌、属性、SKU 编码、价格和库存后重试。';

function pickMessage(error: any): string {
  return (
    error?.data?.message ||
    error?.data?.msg ||
    error?.info?.message ||
    error?.info?.msg ||
    error?.response?.data?.message ||
    error?.response?.data?.msg ||
    error?.message ||
    ''
  );
}

export function buildCatalogActionError(error: any, fallbackTitle: string): CatalogActionError {
  const rawMessage = `${pickMessage(error) || ''}`.trim();
  if (!rawMessage) {
    return {
      title: fallbackTitle,
      description: fallbackDescription,
    };
  }

  return {
    title: fallbackTitle,
    description: rawMessage,
  };
}
