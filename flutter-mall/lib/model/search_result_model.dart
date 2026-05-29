class SearchResultModel {
  final int code;
  final String message;
  final List<SearchResultItem> data;
  final int total;
  final bool empty;
  final String emptyHint;

  SearchResultModel({
    required this.code,
    required this.message,
    required this.data,
    required this.total,
    required this.empty,
    required this.emptyHint,
  });

  factory SearchResultModel.fromJson(Map<String, dynamic> json) {
    final dynamic rawData = json['data'];
    return SearchResultModel(
      code: json['code'] ?? 0,
      message: json['message'] ?? '',
      data: rawData is List
          ? rawData.map((e) => SearchResultItem.fromJson(e)).toList()
          : <SearchResultItem>[],
      total: json['total'] ?? 0,
      empty: json['empty'] ?? true,
      emptyHint: json['emptyHint'] ?? '',
    );
  }
}

class SearchResultItem {
  final int id;
  final String name;
  final String brief;
  final String price;
  final int originalPrice;
  final String mainPic;
  final int stock;
  final int sales;
  final int categoryId;
  final String categoryName;
  final int brandId;
  final String brandName;

  SearchResultItem({
    required this.id,
    required this.name,
    required this.brief,
    required this.price,
    required this.originalPrice,
    required this.mainPic,
    required this.stock,
    required this.sales,
    required this.categoryId,
    required this.categoryName,
    required this.brandId,
    required this.brandName,
  });

  factory SearchResultItem.fromJson(Map<String, dynamic> json) {
    return SearchResultItem(
      id: json['id'] ?? 0,
      name: json['name'] ?? '',
      brief: json['brief'] ?? '',
      price: json['price'] ?? '',
      originalPrice: json['originalPrice'] ?? 0,
      mainPic: json['mainPic'] ?? '',
      stock: json['stock'] ?? 0,
      sales: json['sales'] ?? 0,
      categoryId: json['categoryId'] ?? 0,
      categoryName: json['categoryName'] ?? '',
      brandId: json['brandId'] ?? 0,
      brandName: json['brandName'] ?? '',
    );
  }
}
