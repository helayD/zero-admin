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

  factory DirectCheckoutParams.fromJson(Map<String, dynamic> json) {
    return DirectCheckoutParams(
      productId: json['productId'] ?? 0,
      productSkuId: json['productSkuId'] ?? 0,
      quantity: json['quantity'] ?? 1,
    );
  }
}
