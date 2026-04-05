import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/material.dart';
import 'package:flutter_mall/config/service_url.dart';

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

  //构造方法
  const CachedImageWidget(this.width, this.height, this.url, {super.key, this.fit = BoxFit.cover});

  @override
  Widget build(BuildContext context) {
    if (kIsWeb) {
      //Web端使用Image.network+代理（Flutter 3.29移除了HTML renderer，CanvasKit需要CORS支持）
      return Container(
        width: width,
        height: height,
        alignment: Alignment.center,
        child: Image.network(
          proxyImageUrl(url),
          fit: fit,
          width: width,
          height: height,
          loadingBuilder: (context, child, loadingProgress) {
            if (loadingProgress == null) return child;
            return Container(
              width: width,
              color: Colors.grey[350],
              height: height,
              alignment: Alignment.center,
              child: const Text(
                "加载中",
                style: TextStyle(fontSize: 16, color: Colors.white),
              ),
            );
          },
          errorBuilder: (context, error, stackTrace) {
            return Container(
              width: width,
              color: Colors.grey[350],
              height: height,
              alignment: Alignment.center,
              child: const Icon(Icons.broken_image, color: Colors.white),
            );
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
        imageUrl: url,
        //填充方式
        fit: fit,
        width: width,
        height: height,
        //等待提示
        placeholder: (BuildContext context, String url) {
          return Container(
            width: width,
            color: Colors.grey[350],
            height: height,
            alignment: Alignment.center,
            //加载中
            child: const Text(
              "加载中",
              style: TextStyle(fontSize: 16, color: Colors.white),
            ),
          );
        },
      ),
    );
  }
}
