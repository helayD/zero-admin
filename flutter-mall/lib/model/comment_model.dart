//
// 商品评价数据模型（Story 8-2 Review Fix）
//
// 作者：David
// 日期：2026/04/02
//

// ==================== 评价列表响应 ====================

// CommentListModel 评价列表响应
class CommentListModel {
  int code;
  String message;
  List<CommentItem> data;
  int total;

  CommentListModel({
    required this.code,
    required this.message,
    required this.data,
    required this.total,
  });

  factory CommentListModel.fromJson(Map<String, dynamic> json) =>
      CommentListModel(
        code: json["code"] ?? 0,
        message: json["message"] ?? "",
        data: json["data"] != null
            ? List<CommentItem>.from(
                json["data"].map((x) => CommentItem.fromJson(x)),
              )
            : [],
        total: json["total"] ?? 0,
      );

  Map<String, dynamic> toJson() => {
    "code": code,
    "message": message,
    "data": List<dynamic>.from(data.map((x) => x.toJson())),
    "total": total,
  };

  bool get isSuccess => code == 0;
  bool get hasMore => data.length < total;
}

// CommentItem 评价列表项
class CommentItem {
  String id; // 评价ID
  int productId; // 商品ID
  String memberNickName; // 评价者昵称
  int memberId; // 会员ID
  String memberIcon; // 评价者头像
  int star; // 评分 0-5
  String content; // 评价内容
  String pics; // 图片地址，逗号分隔
  String productAttribute; // 购买商品属性
  int showStatus; // 审核状态
  int replayCount; // 回复数量
  String memberIp; // 评价IP
  String createTime; // 评价时间

  CommentItem({
    required this.id,
    required this.productId,
    required this.memberNickName,
    required this.memberId,
    required this.memberIcon,
    required this.star,
    required this.content,
    required this.pics,
    required this.productAttribute,
    required this.showStatus,
    required this.replayCount,
    required this.memberIp,
    required this.createTime,
  });

  factory CommentItem.fromJson(Map<String, dynamic> json) => CommentItem(
    id: json["id"] ?? "",
    productId: json["productId"] ?? 0,
    memberNickName: json["memberNickName"] ?? "",
    memberId: json["memberId"] ?? 0,
    memberIcon: json["memberIcon"] ?? "",
    star: json["star"] ?? 0,
    content: json["content"] ?? "",
    pics: json["pics"] ?? "",
    productAttribute: json["productAttribute"] ?? "",
    showStatus: json["showStatus"] ?? 0,
    replayCount: json["replayCount"] ?? 0,
    memberIp: json["memberIp"] ?? "",
    createTime: json["createTime"] ?? "",
  );

  Map<String, dynamic> toJson() => {
    "id": id,
    "productId": productId,
    "memberNickName": memberNickName,
    "memberId": memberId,
    "memberIcon": memberIcon,
    "star": star,
    "content": content,
    "pics": pics,
    "productAttribute": productAttribute,
    "showStatus": showStatus,
    "replayCount": replayCount,
    "memberIp": memberIp,
    "createTime": createTime,
  };

  // 获取图片列表
  List<String> get picList {
    if (pics.isEmpty) return [];
    return pics.split(',').where((s) => s.isNotEmpty).toList();
  }

  // 判断是否为当前用户评价
  bool isMyComment(int currentMemberId) => memberId == currentMemberId;
}

// ==================== 评价详情响应 ====================

// CommentDetailModel 评价详情响应
class CommentDetailModel {
  int code;
  String message;
  CommentItem data;
  List<CommentReplayItem> replays;

  CommentDetailModel({
    required this.code,
    required this.message,
    required this.data,
    required this.replays,
  });

  factory CommentDetailModel.fromJson(Map<String, dynamic> json) =>
      CommentDetailModel(
        code: json["code"] ?? 0,
        message: json["message"] ?? "",
        data: json["data"] != null
            ? CommentItem.fromJson(json["data"])
            : CommentItem(
                id: "",
                productId: 0,
                memberNickName: "",
                memberId: 0,
                memberIcon: "",
                star: 0,
                content: "",
                pics: "",
                productAttribute: "",
                showStatus: 0,
                replayCount: 0,
                memberIp: "",
                createTime: "",
              ),
        replays: json["replays"] != null
            ? List<CommentReplayItem>.from(
                json["replays"].map((x) => CommentReplayItem.fromJson(x)),
              )
            : [],
      );

  Map<String, dynamic> toJson() => {
    "code": code,
    "message": message,
    "data": data.toJson(),
    "replays": List<dynamic>.from(replays.map((x) => x.toJson())),
  };

  bool get isSuccess => code == 0;
}

// CommentReplayItem 回复项
// type 为 Dart 保留字，已重命名为 replyType
class CommentReplayItem {
  String id; // 回复ID
  String commentId; // 关联评价ID（MongoDB ObjectID）
  int replyType; // 评论人员类型：0->会员；1->管理员（原 type，Dart 保留字）
  String memberNickName; // 评论人员昵称
  String memberIcon; // 评论人员头像
  String content; // 内容
  String createTime; // 回复时间

  CommentReplayItem({
    required this.id,
    required this.commentId,
    required this.replyType,
    required this.memberNickName,
    required this.memberIcon,
    required this.content,
    required this.createTime,
  });

  factory CommentReplayItem.fromJson(Map<String, dynamic> json) =>
      CommentReplayItem(
        id: json["id"] ?? "",
        commentId: json["commentId"] ?? "",
        replyType: json["type"] ?? 0, // API 仍返回 "type" 字段名
        memberNickName: json["memberNickName"] ?? "",
        memberIcon: json["memberIcon"] ?? "",
        content: json["content"] ?? "",
        createTime: json["createTime"] ?? "",
      );

  Map<String, dynamic> toJson() => {
    "id": id,
    "commentId": commentId,
    "type": replyType, // API 期望 "type" 字段名
    "memberNickName": memberNickName,
    "memberIcon": memberIcon,
    "content": content,
    "createTime": createTime,
  };

  // 是否为管理员回复
  bool get isAdminReply => replyType == 1;
}

// ==================== 提交评价请求 ====================

// AddCommentReqData 提交评价请求数据
class AddCommentReqData {
  int productId; // 商品ID
  int orderId; // 订单ID
  int star; // 评分 0-5
  String content; // 评价内容
  String pics; // 图片地址，逗号分隔
  String memberNickName; // 评价者昵称
  String memberIcon; // 评价者头像
  String productName; // 商品名称
  String productAttribute; // 购买商品属性

  AddCommentReqData({
    required this.productId,
    required this.orderId,
    required this.star,
    required this.content,
    this.pics = "",
    required this.memberNickName,
    this.memberIcon = "",
    required this.productName,
    this.productAttribute = "",
  });

  Map<String, dynamic> toJson() => {
    "productId": productId,
    "orderId": orderId,
    "star": star,
    "content": content,
    "pics": pics,
    "memberNickName": memberNickName,
    "memberIcon": memberIcon,
    "productName": productName,
    "productAttribute": productAttribute,
  };
}

// AddCommentRespData 提交评价响应
class AddCommentRespData {
  int code;
  String message;

  AddCommentRespData({required this.code, required this.message});

  factory AddCommentRespData.fromJson(Map<String, dynamic> json) =>
      AddCommentRespData(
        code: json["code"] ?? 0,
        message: json["message"] ?? "",
      );

  bool get isSuccess => code == 0;
}
