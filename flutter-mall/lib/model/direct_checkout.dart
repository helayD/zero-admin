class DirectCheckoutParams {
  final int productId;
  final int productSkuId;
  final int quantity;

  const DirectCheckoutParams({
    required this.productId,
    required this.productSkuId,
    this.quantity = 1,
  });

  Map<String, dynamic> toJson() => {
        "productId": productId,
        "productSkuId": productSkuId,
        "quantity": quantity,
      };
}
