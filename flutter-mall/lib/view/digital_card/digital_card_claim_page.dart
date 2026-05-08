import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/theme/app_theme.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/view/digital_card/digital_card_asset_detail_page.dart';
import 'package:flutter_mall/view/mine/login/login.dart';
import 'package:flutter_mall/widgets/cached_image_widget.dart';

class DigitalCardClaimPage extends StatefulWidget {
  final String token;
  final int? cardInstanceId;

  const DigitalCardClaimPage({
    super.key,
    required this.token,
    this.cardInstanceId,
  });

  @override
  State<DigitalCardClaimPage> createState() => _DigitalCardClaimPageState();
}

class _DigitalCardClaimPageState extends State<DigitalCardClaimPage> {
  bool _isLoading = true;
  bool _isClaiming = false;
  String? _errorMessage;
  Map<String, dynamic>? _claimInfo;

  @override
  void initState() {
    super.initState();
    _validateToken();
  }

  Future<void> _validateToken() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });
    try {
      final Response response = await HttpUtil.post(
        validateClaimTokenUrl,
        data: <String, dynamic>{
          'token': widget.token,
        },
      );
      final Map<String, dynamic> data = _responseData(response.data);
      if (!mounted) return;
      final bool valid = data['valid'] == true;
      if (!valid) {
        setState(() {
          _errorMessage = data['failureReason']?.toString() ?? '链接已失效或已被领取';
          _isLoading = false;
        });
        return;
      }
      setState(() {
        _claimInfo = data['token'] as Map<String, dynamic>? ?? <String, dynamic>{};
        _isLoading = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _errorMessage = _getErrorMessage(e);
        _isLoading = false;
      });
    }
  }

  Future<void> _claimCard() async {
    if (_isClaiming) return;

    // 检查登录状态
    final bool isLoggedIn = await _checkLoginStatus();
    if (!isLoggedIn) {
      // 未登录，引导到登录页，携带当前页面作为 redirect_uri
      if (!mounted) return;
      final bool? loginResult = await Navigator.of(context).push<bool>(
        MaterialPageRoute(
          builder: (_) => Login(
            redirectRoute: '/digital-card/claim?token=${widget.token}',
          ),
        ),
      );
      // 登录成功后重新尝试领取
      if (loginResult == true) {
        _claimCard();
      }
      return;
    }

    if (!mounted) return;
    // 显示确认对话框
    final bool? confirmed = await showDialog<bool>(
      context: context,
      builder: (BuildContext dialogContext) => AlertDialog(
        title: const Text('确认领取'),
        content: const Text('领取后该卡片将转入您的账户，确认领取吗？'),
        actions: <Widget>[
          TextButton(
            onPressed: () => Navigator.pop(dialogContext, false),
            child: const Text('取消'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(dialogContext, true),
            child: const Text('确认领取'),
          ),
        ],
      ),
    );
    if (confirmed != true) return;

    setState(() => _isClaiming = true);
    try {
      final Response response = await HttpUtil.post(
        claimDigitalCardUrl,
        data: <String, dynamic>{
          'token': widget.token,
          'requestId': 'claim-${DateTime.now().millisecondsSinceEpoch}',
        },
      );
      final Map<String, dynamic> data = _responseData(response.data);
      if (!mounted) return;
      
      // 显示成功对话框，延迟跳转
      if (!mounted) return;
      await showDialog<void>(
        context: context,
        barrierDismissible: false,
        builder: (BuildContext dialogContext) => AlertDialog(
          icon: const Icon(Icons.check_circle, color: Colors.green, size: 48),
          title: const Text('领取成功！'),
          content: const Text('卡片已成功转入您的账户'),
          actions: <Widget>[
            FilledButton(
              onPressed: () => Navigator.pop(dialogContext),
              child: const Text('查看卡片'),
            ),
          ],
        ),
      );
      
      if (!mounted) return;
      final int assetInstanceId = data['cardInstanceId'] as int? ?? widget.cardInstanceId ?? 0;
      if (assetInstanceId > 0) {
        Navigator.of(context).pushReplacement(
          MaterialPageRoute(
            builder: (_) => DigitalCardAssetDetailPage(
              assetInstanceId: assetInstanceId,
            ),
          ),
        );
      } else {
        Navigator.of(context).pop();
      }
    } catch (e) {
      if (!mounted) return;
      setState(() => _isClaiming = false);
      _showSnack(_getErrorMessage(e));
    }
  }

  String _getErrorMessage(dynamic error) {
    if (error is DioException) {
      if (error.response?.statusCode == 404) {
        return '链接已失效或已被领取';
      } else if (error.response?.statusCode == 429) {
        return '操作过于频繁，请稍后再试';
      } else if (error.response?.statusCode == 401) {
        return '登录已过期，请重新登录';
      } else if (error.type == DioExceptionType.connectionTimeout ||
          error.type == DioExceptionType.receiveTimeout) {
        return '网络连接超时，请检查网络后重试';
      } else if (error.type == DioExceptionType.connectionError) {
        return '网络连接失败，请检查网络设置';
      }
    }
    return '领取失败，请稍后重试';
  }

  Future<bool> _checkLoginStatus() async {
    try {
      // 尝试获取用户信息，如果失败则说明未登录
      await HttpUtil.get(memberInfoDataUrl);
      return true;
    } catch (e) {
      return false;
    }
  }

  Map<String, dynamic> _responseData(dynamic raw) {
    if (raw is Map && raw['data'] is Map) {
      return Map<String, dynamic>.from(raw['data'] as Map);
    }
    return <String, dynamic>{};
  }

  void _showSnack(String message) {
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(message)),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        title: const Text('领取卡片'),
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
            Text(
              '正在验证链接...',
              style: TextStyle(
                color: AppColors.textSecondary,
                fontSize: 14,
              ),
            ),
          ],
        ),
      );
    }
    if (_errorMessage != null) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(AppSpacing.xl),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              const Icon(Icons.error_outline, size: 52, color: Color(0xFFB42318)),
              const SizedBox(height: AppSpacing.md),
              Text(
                _errorMessage!,
                textAlign: TextAlign.center,
                style: Theme.of(context).textTheme.bodyMedium,
              ),
              const SizedBox(height: AppSpacing.lg),
              FilledButton(
                onPressed: _validateToken,
                child: const Text('重新加载'),
              ),
            ],
          ),
        ),
      );
    }

    final Map<String, dynamic> info = _claimInfo ?? <String, dynamic>{};
    final String templateName = info['templateName']?.toString() ?? '提货卡';
    final String cardFaceImage = info['cardFaceImage']?.toString() ?? '';
    final String senderName = info['senderName']?.toString() ?? '朋友';
    final bool canClaim = info['canClaim'] == true;
    final String? claimHint = info['claimHint']?.toString();

    return SafeArea(
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.xl),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.center,
          children: <Widget>[
            const SizedBox(height: AppSpacing.xl),
            Text(
              '$senderName 分享了一张卡片给你',
              style: Theme.of(context).textTheme.headlineSmall,
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: AppSpacing.lg),
            if (cardFaceImage.isNotEmpty)
              ClipRRect(
                borderRadius: BorderRadius.circular(AppRadii.xl),
                child: CachedImageWidget(
                  200,
                  280,
                  cardFaceImage,
                ),
              ),
            const SizedBox(height: AppSpacing.lg),
            Text(
              templateName,
              style: Theme.of(context).textTheme.titleLarge,
            ),
            const SizedBox(height: AppSpacing.md),
            if (claimHint != null)
              Text(
                claimHint,
                style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                      color: AppColors.textSecondary,
                    ),
                textAlign: TextAlign.center,
              ),
            const Spacer(),
            SizedBox(
              width: double.infinity,
              child: FilledButton.icon(
                onPressed: canClaim && !_isClaiming ? _claimCard : null,
                icon: _isClaiming
                    ? const SizedBox(
                        width: 16,
                        height: 16,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : const Icon(Icons.card_giftcard_outlined),
                label: Text(_isClaiming ? '领取中...' : '立即领取'),
              ),
            ),
            if (!canClaim) ...<Widget>[
              const SizedBox(height: AppSpacing.sm),
              Text(
                claimHint ?? '当前无法领取',
                style: Theme.of(context).textTheme.bodySmall?.copyWith(
                      color: AppColors.price,
                    ),
                textAlign: TextAlign.center,
              ),
            ],
            const SizedBox(height: AppSpacing.md),
            SizedBox(
              width: double.infinity,
              child: OutlinedButton(
                onPressed: () => Navigator.of(context).pop(),
                child: const Text('暂不领取'),
              ),
            ),
          ],
        ),
      ),
    );
  }
}