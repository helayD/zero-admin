import 'package:flutter/foundation.dart';

import '../config/service_url.dart';
import '../model/comment_model.dart';
import '../utils/http_util.dart';

///
/// 商品评价状态管理（Story 8-2）
///
/// 作者：刘飞华
/// 日期：2026/04/02
///
class CommentProvider with ChangeNotifier, DiagnosticableTreeMixin {
  // 评价列表（按商品ID索引）
  final Map<int, List<CommentItem>> _commentListCache = {};

  // 当前商品评价总数
  final Map<int, int> _commentTotalCache = {};

  // 评价详情（按评价ID索引）
  final Map<String, CommentDetailModel> _commentDetailCache = {};

  // 加载状态（按商品ID索引）
  final Map<int, bool> _loadingCache = {};

  // 是否有更多数据（按商品ID索引）
  final Map<int, bool> _hasMoreCache = {};

  // 当前页码（按商品ID索引）
  final Map<int, int> _currentPageCache = {};

  // 获取评价列表
  List<CommentItem> getCommentList(int productId) {
    return _commentListCache[productId] ?? [];
  }

  // 获取评价总数
  int getCommentTotal(int productId) {
    return _commentTotalCache[productId] ?? 0;
  }

  // 获取评价详情
  CommentDetailModel? getCommentDetail(String commentId) {
    return _commentDetailCache[commentId];
  }

  // 是否正在加载
  bool isLoading(int productId) {
    return _loadingCache[productId] ?? false;
  }

  // 是否有更多数据
  bool hasMore(int productId) {
    return _hasMoreCache[productId] ?? true;
  }

  // 获取当前页码
  int getCurrentPage(int productId) {
    return _currentPageCache[productId] ?? 1;
  }

  void syncCommentPage(
    int productId, {
    required List<CommentItem> data,
    required int total,
    required int page,
    bool append = false,
  }) {
    setState(() {
      final existing = append
          ? (_commentListCache[productId] ?? [])
          : <CommentItem>[];
      final merged = append ? [...existing, ...data] : data;
      _commentListCache[productId] = merged;
      _commentTotalCache[productId] = total;
      _currentPageCache[productId] = page;
      _hasMoreCache[productId] = merged.length < total;
    });
  }

  /// 查询商品评价列表
  Future<void> queryCommentList(
    int productId, {
    bool refresh = false,
    bool loadMore = false,
  }) async {
    final page = refresh ? 1 : (getCurrentPage(productId) + 1);

    // 缓存加载状态
    _loadingCache[productId] = true;
    notifyListeners();

    try {
      final pageSize = 20;
      final result = await HttpUtil.get(
        queryCommentListUrl,
        queryParameters: {
          "productId": productId,
          "pageNum": page,
          "pageSize": pageSize,
        },
      );
      final model = CommentListModel.fromJson(result.data);

      setState(() {
        final existing = loadMore
            ? (_commentListCache[productId] ?? [])
            : <CommentItem>[];
        final merged = loadMore ? [...existing, ...model.data] : model.data;
        _commentListCache[productId] = merged;
        _currentPageCache[productId] = page;
        _hasMoreCache[productId] = merged.length < model.total;
        _commentTotalCache[productId] = model.total;
        _loadingCache[productId] = false;
      });
    } catch (e) {
      setState(() {
        _loadingCache[productId] = false;
      });
      rethrow;
    }
  }

  /// 查询评价详情（含回复）
  Future<CommentDetailModel?> queryCommentDetail(String commentId) async {
    try {
      final result = await HttpUtil.get(
        queryCommentDetailUrl,
        queryParameters: {"id": commentId},
      );
      final model = CommentDetailModel.fromJson(result.data);

      if (model.isSuccess) {
        setState(() {
          _commentDetailCache[commentId] = model;
        });
      }
      return model;
    } catch (e) {
      rethrow;
    }
  }

  /// 添加评价到缓存（提交成功后调用）
  void addComment(int productId, CommentItem comment) {
    setState(() {
      final list = _commentListCache[productId] ?? [];
      _commentListCache[productId] = [comment, ...list];
      _commentTotalCache[productId] = (_commentTotalCache[productId] ?? 0) + 1;
    });
  }

  /// 清除指定商品的评价缓存
  void clearCache(int productId) {
    setState(() {
      _commentListCache.remove(productId);
      _commentTotalCache.remove(productId);
      _loadingCache.remove(productId);
      _hasMoreCache.remove(productId);
      _currentPageCache.remove(productId);
    });
  }

  void setState(VoidCallback callback) {
    callback();
    notifyListeners();
  }

  @override
  void debugFillProperties(DiagnosticPropertiesBuilder properties) {
    super.debugFillProperties(properties);
    properties.add(
      DiagnosticsProperty<Map<int, List<CommentItem>>>(
        'commentListCache',
        _commentListCache,
      ),
    );
    properties.add(
      DiagnosticsProperty<Map<int, int>>(
        'commentTotalCache',
        _commentTotalCache,
      ),
    );
  }
}

///
/// 商品评价提交状态管理（Story 8-2）
///
/// 作者：刘飞华
/// 日期：2026/04/02
///
class CommentUploadProvider with ChangeNotifier, DiagnosticableTreeMixin {
  // 提交状态
  CommentUploadState _uploadState = CommentUploadState.idle;
  CommentUploadState get uploadState => _uploadState;

  // 错误信息
  String? _errorMessage;
  String? get errorMessage => _errorMessage;

  // 成功消息
  String? _successMessage;
  String? get successMessage => _successMessage;

  /// 提交评价
  Future<bool> submitComment(AddCommentReqData req) async {
    setState(() {
      _uploadState = CommentUploadState.submitting;
      _errorMessage = null;
      _successMessage = null;
    });

    try {
      final result = await HttpUtil.post(addCommentUrl, data: req.toJson());
      final resp = AddCommentRespData.fromJson(result.data);

      if (resp.isSuccess) {
        setState(() {
          _uploadState = CommentUploadState.success;
          _successMessage = resp.message.isNotEmpty ? resp.message : "评价提交成功";
        });
        return true;
      } else {
        setState(() {
          _uploadState = CommentUploadState.error;
          _errorMessage = resp.message.isNotEmpty ? resp.message : "评价提交失败";
        });
        return false;
      }
    } catch (e) {
      setState(() {
        _uploadState = CommentUploadState.error;
        _errorMessage = "网络异常，请检查网络后重试";
      });
      return false;
    }
  }

  /// 重置状态
  void reset() {
    setState(() {
      _uploadState = CommentUploadState.idle;
      _errorMessage = null;
      _successMessage = null;
    });
  }

  void setState(VoidCallback callback) {
    callback();
    notifyListeners();
  }

  @override
  void debugFillProperties(DiagnosticPropertiesBuilder properties) {
    super.debugFillProperties(properties);
    properties.add(EnumProperty('uploadState', _uploadState));
  }
}

// 评价提交状态枚举
enum CommentUploadState {
  idle, // 空闲
  submitting, // 提交中
  success, // 提交成功
  error, // 提交失败
}
