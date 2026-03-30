// 客服工作台订单列表项
export interface CustomerServiceOrderItem {
  id: number;
  orderNo: string;
  orderStatus: number;
  orderStatusText: string;
  payStatus: number;
  payStatusText: string;
  returnStatus: number;
  returnStatusText: string;
  memberId: number;
  receiverName: string;
  receiverPhone: string;
  receiverAddress: string;
  totalAmount: number;
  payAmount: number;
  deliveryCompany: string;
  deliverySn: string;
  createTime: string;
  payTime?: string;
  deliveryTime?: string;
  receiveTime?: string;
  orderItemData?: OrderItemData[];
}

// 订单商品明细
export interface OrderItemData {
  id: number;
  orderId: number;
  orderNo: string;
  orderItemStatus: number;
  skuId: number;
  skuName: string;
  skuPic: string;
  skuPrice: number;
  skuQuantity: number;
  specData: string;
  skuTotalAmount: number;
  promotionAmount: number;
  couponAmount: number;
  pointsAmount: number;
  discountAmount: number;
  realAmount: number;
  createTime: string;
}

// 客服工作台订单列表请求
export interface CustomerServiceOrderListParams {
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
  orderNo?: string;
  orderStatus?: number;
  returnStatus?: number;
  receiverName?: string;
  receiverPhone?: string;
  startTime?: string;
  endTime?: string;
  current?: number;
  pageSize?: number;
}

// 公司退货地址项
export interface CompanyAddressItem {
  id: number;
  addressName: string;
  receiverName: string;
  phone: string;
  province: string;
  city: string;
  region: string;
  detailAddress: string;
  fullAddress: string;
  defaultStatus: number;
  sendStatus: number;
  receiveStatus: number;
  createTime: string;
}

// 公司退货地址列表请求
export interface CompanyAddressListParams {
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
  addressName?: string;
  name?: string;
  phone?: string;
  current?: number;
  pageSize?: number;
}

// 订单操作日志项
export interface OrderOperationLogItem {
  id: number;
  orderId: number;
  orderNo: string;
  operationType: number;
  operationTypeText: string;
  operatorId: number;
  operatorType: number;
  operatorTypeText: string;
  operatorNote: string;
  createTime: string;
}

// 订单操作日志列表请求
export interface OrderOperationLogListParams {
  orderId?: number;
  orderNo?: string;
  scopeType?: 'platform' | 'tenant' | 'merchant';
  platformId?: number;
  tenantId?: number;
  merchantId?: number;
  current?: number;
  pageSize?: number;
}

// 退货审核请求
export interface UpdateOrderReturnReq {
  id: number;
  companyAddressId?: number;
  status: number;
  handleNote?: string;
  receiveNote?: string;
  refundAmount?: number;
}
