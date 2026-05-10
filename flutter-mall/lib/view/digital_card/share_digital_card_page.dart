import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/theme/app_theme.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/widgets/cached_image_widget.dart';
import 'package:qr_flutter/qr_flutter.dart';

/// 分享提货卡页面（Story 10.7 Review Fix CRITICAL-2 / HIGH-4）
///
/// 职责：
/// - 调用 `/api/digitalCard/asset/generateShareLink` 获取分享 token / 短链 / 有效期 / 可领取次数
/// - 用 qr_flutter 渲染二维码（扫码=H5 领取页）
/// - 展示链接，支持复制到剪贴板
/// - 展示有效期和可领取次数
///
/// 规格要求：
/// - 分享链接必须指向 `/h5/digital-card/claim?token=xxx`（由后端生成，此页不拼 URL）
/// - 未登录态不可进入（外部调用者保证）
/// - 失败场景需要明确的错误提示和重试
class ShareDigitalCardPage extends StatefulWidget {
  final int assetInstanceId;
  final String templateName;
  final String cardFaceImage;

  const ShareDigitalCardPage({
    super.key,
    required this.assetInstanceId,
    required this.templateName,
    required this.cardFaceImage,
  });

  @override
  State<ShareDigitalCardPage> createState() => _ShareDigitalCardPageState();
}

class _ShareDigitalCardPageState extends State<ShareDigitalCardPage> {
  bool _isLoading = true;
  String? _errorMessage;
  String? _shareLink;
  String? _token;
  String? _expireAt;
  int? _maxClaims;

  @override
  void initState() {
    super.initState();
    _generateLink();
  }

  Future<void> _generateLink() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });
    try {
      final Response response = await HttpUtil.post(
        generateShareLinkUrl,
        data: <String, dynamic>{
          'cardInstanceId': widget.assetInstanceId,
          // 不传 domain：由后端根据白名单首项兜底（Review Fix CRITICAL-1）
          'requestId': 'share-${DateTime.now().millisecondsSinceEpoch}',
        },
      );
      final Map<String, dynamic> data = _responseData(response.data);
      if (!mounted) return;
      setState(() {
        _shareLink = data['shareLink']?.toString();
        _token = data['token']?.toString();
        _expireAt = data['expireAt']?.toString();
        _maxClaims = (data['maxClaims'] as num?)?.toInt();
        _isLoading = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _errorMessage = _parseError(e);
        _isLoading = false;
      });
    }
  }

  Map<String, dynamic> _responseData(dynamic raw) {
    if (raw is Map && raw['data'] is Map) {
      return Map<String, dynamic>.from(raw['data'] as Map);
    }
    return <String, dynamic>{};
  }

  String _parseError(dynamic error) {
    if (error is DioException) {
      final int? code = error.response?.statusCode;
      final dynamic data = error.response?.data;
      if (data is Map && data['message'] != null) {
        return data['message'].toString();
      }
      if (code == 401) return '登录已过期，请重新登录';
      if (code == 429) return '操作过于频繁，请稍后再试';
      if (error.type == DioExceptionType.connectionTimeout ||
          error.type == DioExceptionType.receiveTimeout) {
        return '网络连接超时，请检查网络后重试';
      }
      if (error.type == DioExceptionType.connectionError) {
        return '网络连接失败，请检查网络设置';
      }
    }
    return '生成分享链接失败，请稍后重试';
  }

  Future<void> _copyLink() async {
    final String? link = _shareLink;
    if (link == null || link.isEmpty) return;
    await Clipboard.setData(ClipboardData(text: link));
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('分享链接已复制，可直接粘贴到微信发给朋友')),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        title: const Text('分享给朋友'),
      ),
      body: _buildBody(),
    );
  }

  Widget _buildBody() {
    if (_isLoading) {
      return const Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: <Widget>[
            CircularProgressIndicator(),
            SizedBox(height: AppSpacing.lg),
            Text('正在生成分享凭证...',
                style: TextStyle(color: AppColors.textSecondary)),
          ],
        ),
      );
    }
    if (_errorMessage != null || _shareLink == null) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(AppSpacing.xl),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              const Icon(Icons.error_outline,
                  size: 52, color: Color(0xFFB42318)),
              const SizedBox(height: AppSpacing.md),
              Text(
                _errorMessage ?? '生成失败',
                textAlign: TextAlign.center,
                style: Theme.of(context).textTheme.bodyMedium,
              ),
              const SizedBox(height: AppSpacing.lg),
              FilledButton(
                onPressed: _generateLink,
                child: const Text('重新生成'),
              ),
            ],
          ),
        ),
      );
    }

    return SafeArea(
      child: SingleChildScrollView(
        padding: const EdgeInsets.all(AppSpacing.xl),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.center,
          children: <Widget>[
            if (widget.cardFaceImage.isNotEmpty)
              ClipRRect(
                borderRadius: BorderRadius.circular(AppRadii.lg),
                child: CachedImageWidget(
                  120,
                  160,
                  widget.cardFaceImage,
                ),
              ),
            const SizedBox(height: AppSpacing.md),
            Text(
              widget.templateName.isEmpty ? '提货卡' : widget.templateName,
              style: Theme.of(context).textTheme.titleLarge,
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: AppSpacing.lg),
            Container(
              padding: const EdgeInsets.all(AppSpacing.lg),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(AppRadii.xl),
                border: Border.all(color: AppColors.border),
              ),
              child: Column(
                children: <Widget>[
                  Text(
                    '让朋友扫码领取',
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  const SizedBox(height: AppSpacing.md),
                  QrImageView(
                    data: _shareLink!,
                    version: QrVersions.auto,
                    size: 220,
                    gapless: false,
                    backgroundColor: Colors.white,
                  ),
                  const SizedBox(height: AppSpacing.md),
                  Text(
                    '扫码后在手机浏览器（含微信内置浏览器）完成登录或注册，即可领取这张提货卡。',
                    textAlign: TextAlign.center,
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: AppColors.textSecondary,
                        ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: AppSpacing.lg),
            _buildMetaCard(),
            const SizedBox(height: AppSpacing.lg),
            Container(
              padding: const EdgeInsets.all(AppSpacing.md),
              decoration: BoxDecoration(
                color: AppColors.surface,
                borderRadius: BorderRadius.circular(AppRadii.md),
                border: Border.all(color: AppColors.border),
              ),
              child: Row(
                children: <Widget>[
                  Expanded(
                    child: Text(
                      _shareLink!,
                      style: Theme.of(context).textTheme.bodySmall,
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                  IconButton(
                    tooltip: '复制链接',
                    onPressed: _copyLink,
                    icon: const Icon(Icons.copy_outlined),
                  ),
                ],
              ),
            ),
            const SizedBox(height: AppSpacing.md),
            SizedBox(
              width: double.infinity,
              child: FilledButton.icon(
                onPressed: _copyLink,
                icon: const Icon(Icons.link_outlined),
                label: const Text('复制链接分享'),
              ),
            ),
            const SizedBox(height: AppSpacing.sm),
            SizedBox(
              width: double.infinity,
              child: OutlinedButton.icon(
                onPressed: _generateLink,
                icon: const Icon(Icons.refresh_outlined),
                label: const Text('重新生成'),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildMetaCard() {
    return Container(
      padding: const EdgeInsets.all(AppSpacing.md),
      decoration: BoxDecoration(
        color: AppColors.surface,
        borderRadius: BorderRadius.circular(AppRadii.md),
        border: Border.all(color: AppColors.border),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          _buildLine('有效期至', _expireAt?.isNotEmpty == true ? _expireAt! : '24 小时'),
          const SizedBox(height: AppSpacing.xs),
          _buildLine(
            '可领取次数',
            _maxClaims != null ? '$_maxClaims 次' : '1 次',
          ),
          const SizedBox(height: AppSpacing.xs),
          _buildLine(
            '分享凭证',
            _token == null || _token!.length < 6
                ? '-'
                : '${_token!.substring(0, 4)}***${_token!.substring(_token!.length - 4)}',
          ),
        ],
      ),
    );
  }

  Widget _buildLine(String label, String value) {
    return Row(
      children: <Widget>[
        SizedBox(
          width: 80,
          child: Text(
            label,
            style: Theme.of(context).textTheme.bodySmall,
          ),
        ),
        Expanded(
          child: Text(
            value,
            style: Theme.of(context).textTheme.bodyMedium,
          ),
        ),
      ],
    );
  }
}
