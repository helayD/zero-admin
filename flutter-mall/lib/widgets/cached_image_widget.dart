import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/material.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/theme/app_theme.dart';

///
/// 缓存图片组件
///
/// 作者：David
/// 日期：2023/11/21 17:17
///
class CachedImageWidget extends StatelessWidget {
  //宽度
  final double width;

  //高度
  final double height;

  //图片地址
  final String url;

  //图片填充模式
  final BoxFit? fit;

  //自定义兜底
  final Widget? fallback;

  //构造方法
  const CachedImageWidget(
    this.width,
    this.height,
    this.url, {
    super.key,
    this.fit = BoxFit.cover,
    this.fallback,
  });

  String get _normalizedUrl => url.trim();

  bool get _canLoadRemoteImage {
    if (_normalizedUrl.isEmpty) {
      return false;
    }
    final Uri? uri = Uri.tryParse(_normalizedUrl);
    return uri != null && uri.hasScheme && uri.hasAuthority;
  }

  Widget _buildLoadingPlaceholder() {
    return Container(
      width: width,
      height: height,
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: AppColors.surfaceMuted,
        borderRadius: BorderRadius.circular(AppRadii.md),
      ),
      child: const SizedBox(
        width: 22,
        height: 22,
        child: CircularProgressIndicator(strokeWidth: 2),
      ),
    );
  }

  Widget _buildDefaultFallback() {
    return Container(
      width: width,
      height: height,
      decoration: BoxDecoration(
        gradient: LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [
            AppColors.primarySoft,
            AppColors.surfaceMuted,
            Colors.white.withValues(alpha: 0.96),
          ],
        ),
        borderRadius: BorderRadius.circular(AppRadii.md),
      ),
      child: LayoutBuilder(
        builder: (context, constraints) {
          final bool ultraCompact =
              constraints.maxWidth < 84 || constraints.maxHeight < 84;
          final bool compact = ultraCompact ||
              constraints.maxWidth < 132 ||
              constraints.maxHeight < 120;

          if (ultraCompact) {
            return Center(
              child: Container(
                width: 30,
                height: 30,
                decoration: BoxDecoration(
                  color: Colors.white.withValues(alpha: 0.84),
                  borderRadius: BorderRadius.circular(10),
                ),
                child: const Icon(
                  Icons.image_not_supported_outlined,
                  color: AppColors.primaryDark,
                  size: 18,
                ),
              ),
            );
          }

          if (compact) {
            return Padding(
              padding: const EdgeInsets.all(AppSpacing.md),
              child: Center(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Container(
                      width: 34,
                      height: 34,
                      decoration: BoxDecoration(
                        color: Colors.white.withValues(alpha: 0.84),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: const Icon(
                        Icons.image_not_supported_outlined,
                        color: AppColors.primaryDark,
                        size: 18,
                      ),
                    ),
                    const SizedBox(height: AppSpacing.xs),
                    const Text(
                      '图片暂未就绪',
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      textAlign: TextAlign.center,
                      style: TextStyle(
                        fontSize: 11,
                        fontWeight: FontWeight.w700,
                        color: AppColors.textPrimary,
                      ),
                    ),
                  ],
                ),
              ),
            );
          }

          return Padding(
            padding: const EdgeInsets.all(AppSpacing.lg),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Container(
                  width: 40,
                  height: 40,
                  decoration: BoxDecoration(
                    color: Colors.white.withValues(alpha: 0.82),
                    borderRadius: BorderRadius.circular(14),
                  ),
                  child: const Icon(
                    Icons.image_not_supported_outlined,
                    color: AppColors.primaryDark,
                    size: 22,
                  ),
                ),
                const Spacer(),
                const Text(
                  '图片暂未就绪',
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: TextStyle(
                    fontSize: 15,
                    fontWeight: FontWeight.w700,
                    color: AppColors.textPrimary,
                  ),
                ),
                const SizedBox(height: AppSpacing.xs),
                const Text(
                  '当前展示默认占位内容',
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                  style: TextStyle(
                    fontSize: 12,
                    color: AppColors.textSecondary,
                  ),
                ),
              ],
            ),
          );
        },
      ),
    );
  }

  Widget _buildFallback() {
    return fallback ?? _buildDefaultFallback();
  }

  @override
  Widget build(BuildContext context) {
    if (!_canLoadRemoteImage) {
      return _buildFallback();
    }

    if (kIsWeb) {
      //Web端使用Image.network+代理（Flutter 3.29移除了HTML renderer，CanvasKit需要CORS支持）
      return Container(
        width: width,
        height: height,
        alignment: Alignment.center,
        child: Image.network(
          proxyImageUrl(_normalizedUrl),
          fit: fit,
          width: width,
          height: height,
          loadingBuilder: (context, child, loadingProgress) {
            if (loadingProgress == null) {
              return child;
            }
            return _buildLoadingPlaceholder();
          },
          errorBuilder: (context, error, stackTrace) {
            return _buildFallback();
          },
        ),
      );
    }
    return Container(
      width: width,
      height: height,
      alignment: Alignment.center,
      //使用CachedNetworkImage组件
      child: CachedNetworkImage(
        //图片地址
        imageUrl: _normalizedUrl,
        //填充方式
        fit: fit,
        width: width,
        height: height,
        //等待提示
        placeholder: (BuildContext context, String url) {
          return _buildLoadingPlaceholder();
        },
        errorWidget: (BuildContext context, String url, Object error) {
          return _buildFallback();
        },
      ),
    );
  }
}
