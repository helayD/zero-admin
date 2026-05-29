import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_mall/theme/app_theme.dart';
import 'package:flutter_mall/view/digital_card/digital_card_claim_page.dart';
import 'package:image_picker/image_picker.dart';
import 'package:mobile_scanner/mobile_scanner.dart';
import 'package:url_launcher/url_launcher.dart';

/// 首页「扫一扫」扫码页（Story 10.7 Task 10，AC7）
///
/// 职责：
/// - 用 `mobile_scanner` 实时识别二维码（仅 [BarcodeFormat.qrCode]）
/// - 支持从相册选取静态二维码图片识别（复用 `image_picker`）
/// - 扫码结果三路分流：
///   1. 提货卡分享链接（`/h5/digital-card/claim?token=xxx`）→ 提取 token 跳转
///      [DigitalCardClaimPage]，复用「预览 → 未登录引导 → canClaim → 确认领取 →
///      成功跳详情」全流程；
///   2. 普通 http(s) 外链 → 弹窗确认后用系统浏览器打开；
///   3. 非 URL / 无法识别文本 → 明确提示「不可识别」，可继续扫码。
///
/// 监管约束：本页属于 C 端，不展示任何区块链底层字段，仅做扫码识别与路由。
class DigitalCardScanPage extends StatefulWidget {
  const DigitalCardScanPage({super.key});

  @override
  State<DigitalCardScanPage> createState() => _DigitalCardScanPageState();
}

class _DigitalCardScanPageState extends State<DigitalCardScanPage>
    with WidgetsBindingObserver {
  final MobileScannerController _controller = MobileScannerController(
    // 只识别二维码，过滤条码等其它格式，提升识别速度与准确率。
    formats: const <BarcodeFormat>[BarcodeFormat.qrCode],
    // 去重，避免同一张二维码在连续帧里被反复触发。
    detectionSpeed: DetectionSpeed.noDuplicates,
  );

  /// 扫码结果流订阅；在页面销毁与切到后台时取消，避免泄漏与误触发。
  StreamSubscription<BarcodeCapture>? _scanSubscription;

  /// 已处理标记：一次只路由一个结果，避免连续帧重复跳转。
  bool _handled = false;

  /// 相册识别进行中标记，避免重复触发。
  bool _picking = false;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    _scanSubscription = _controller.barcodes.listen(_handleCapture);
    unawaited(_controller.start());
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    switch (state) {
      case AppLifecycleState.resumed:
        _scanSubscription = _controller.barcodes.listen(_handleCapture);
        unawaited(_controller.start());
      case AppLifecycleState.inactive:
      case AppLifecycleState.paused:
      case AppLifecycleState.hidden:
      case AppLifecycleState.detached:
        unawaited(_scanSubscription?.cancel());
        _scanSubscription = null;
        unawaited(_controller.stop());
    }
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    unawaited(_scanSubscription?.cancel());
    _scanSubscription = null;
    unawaited(_controller.dispose());
    super.dispose();
  }

  /// 处理实时扫码捕获结果。
  void _handleCapture(BarcodeCapture capture) {
    if (_handled) return;
    if (capture.barcodes.isEmpty) return;
    final String? raw = capture.barcodes.first.rawValue;
    if (raw == null || raw.trim().isEmpty) return;
    _handled = true;
    unawaited(_routeResult(raw.trim()));
  }

  /// 扫码结果路由分流。
  Future<void> _routeResult(String raw) async {
    final Uri? uri = Uri.tryParse(raw);

    // 1. 提货卡分享链接：提取 token 走领取闭环。
    final String? token = _extractClaimToken(uri);
    if (token != null && token.isNotEmpty) {
      unawaited(_controller.stop());
      if (!mounted) return;
      await Navigator.of(context).pushReplacement(
        MaterialPageRoute<void>(
          builder: (_) => DigitalCardClaimPage(token: token),
        ),
      );
      return;
    }

    // 2. 普通 http(s) 外链：弹窗确认后用浏览器打开。
    if (uri != null && (uri.scheme == 'http' || uri.scheme == 'https')) {
      await _confirmOpenExternal(uri);
      return;
    }

    // 3. 非 URL / 无法识别文本。
    await _showUnrecognized(raw);
  }

  /// 从二维码 URL 中提取提货卡领取 token。
  ///
  /// 兼容 `/h5/digital-card/claim?token=xxx` 与 `/digital-card/claim?token=xxx`
  /// 两种路径形态（后端 H5 落地页 / App Deep Link）。
  String? _extractClaimToken(Uri? uri) {
    if (uri == null) return null;
    if (!uri.path.contains('digital-card/claim')) return null;
    return uri.queryParameters['token'];
  }

  Future<void> _confirmOpenExternal(Uri uri) async {
    if (!mounted) {
      _resumeScan();
      return;
    }
    final bool? open = await showDialog<bool>(
      context: context,
      builder: (BuildContext ctx) => AlertDialog(
        title: const Text('识别到一个网页链接'),
        content: Text('是否在浏览器中打开？\n\n${uri.toString()}'),
        actions: <Widget>[
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('取消'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('打开'),
          ),
        ],
      ),
    );
    if (open == true) {
      final bool ok = await launchUrl(uri, mode: LaunchMode.externalApplication);
      if (!ok && mounted) {
        _showSnack('无法打开该链接');
      }
      if (mounted) {
        Navigator.of(context).pop();
      }
    } else {
      _resumeScan();
    }
  }

  Future<void> _showUnrecognized(String raw) async {
    if (!mounted) {
      _resumeScan();
      return;
    }
    await showDialog<void>(
      context: context,
      builder: (BuildContext ctx) => AlertDialog(
        title: const Text('无法识别'),
        content: Text('这不是有效的提货卡分享码。\n\n扫描内容：$raw'),
        actions: <Widget>[
          FilledButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text('继续扫码'),
          ),
        ],
      ),
    );
    _resumeScan();
  }

  /// 重置已处理标记，允许继续扫下一张码。
  void _resumeScan() {
    _handled = false;
  }

  /// 从相册选取一张二维码图片并识别。
  Future<void> _pickFromGallery() async {
    if (_picking) return;
    _picking = true;
    try {
      final ImagePicker picker = ImagePicker();
      final XFile? file = await picker.pickImage(source: ImageSource.gallery);
      if (file == null) {
        _picking = false;
        return;
      }
      final BarcodeCapture? capture = await _controller.analyzeImage(file.path);
      _picking = false;
      if (capture == null || capture.barcodes.isEmpty) {
        if (mounted) await _showUnrecognized('相册图片中未发现二维码');
        return;
      }
      final String? raw = capture.barcodes.first.rawValue;
      if (raw == null || raw.trim().isEmpty) {
        if (mounted) await _showUnrecognized('相册图片中的二维码无法解析');
        return;
      }
      _handled = true;
      await _routeResult(raw.trim());
    } catch (_) {
      _picking = false;
      if (mounted) _showSnack('读取相册图片失败');
    }
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
      backgroundColor: Colors.black,
      appBar: AppBar(
        backgroundColor: Colors.black,
        foregroundColor: Colors.white,
        elevation: 0,
        title: const Text('扫一扫'),
        actions: <Widget>[
          ValueListenableBuilder<MobileScannerState>(
            valueListenable: _controller,
            builder: (BuildContext context, MobileScannerState state, Widget? child) {
              final bool torchOn = state.torchState == TorchState.on;
              return IconButton(
                tooltip: torchOn ? '关闭手电筒' : '打开手电筒',
                onPressed: () => unawaited(_controller.toggleTorch()),
                icon: Icon(
                  torchOn ? Icons.flash_on : Icons.flash_off,
                  color: Colors.white,
                ),
              );
            },
          ),
        ],
      ),
      body: Stack(
        children: <Widget>[
          MobileScanner(
            controller: _controller,
            // 相机不可用 / 权限被拒：全屏提示并提供「去设置」入口。
            errorBuilder: (BuildContext context, MobileScannerException error, Widget? child) {
              return _buildCameraError(error);
            },
          ),
          _buildScanOverlay(),
          _buildBottomBar(),
        ],
      ),
    );
  }

  /// 扫描取景框装饰。
  Widget _buildScanOverlay() {
    return Center(
      child: Container(
        width: 240,
        height: 240,
        decoration: BoxDecoration(
          border: Border.all(color: Colors.white70, width: 2),
          borderRadius: BorderRadius.circular(AppRadii.lg),
        ),
      ),
    );
  }

  Widget _buildBottomBar() {
    return Positioned(
      left: 0,
      right: 0,
      bottom: 0,
      child: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(AppSpacing.xl),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              const Text(
                '将提货卡分享二维码放入框内即可自动识别',
                textAlign: TextAlign.center,
                style: TextStyle(color: Colors.white70, fontSize: 13),
              ),
              const SizedBox(height: AppSpacing.lg),
              OutlinedButton.icon(
                onPressed: _pickFromGallery,
                style: OutlinedButton.styleFrom(
                  foregroundColor: Colors.white,
                  side: const BorderSide(color: Colors.white54),
                ),
                icon: const Icon(Icons.photo_library_outlined),
                label: const Text('从相册选取'),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildCameraError(MobileScannerException error) {
    final String detail =
        error.errorDetails?.message ?? error.errorCode.name;
    return Container(
      color: Colors.black,
      child: Center(
        child: Padding(
          padding: const EdgeInsets.all(AppSpacing.xl),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              const Icon(
                Icons.no_photography_outlined,
                color: Colors.white70,
                size: 56,
              ),
              const SizedBox(height: AppSpacing.lg),
              const Text(
                '需要相机权限才能扫码',
                textAlign: TextAlign.center,
                style: TextStyle(color: Colors.white, fontSize: 16),
              ),
              const SizedBox(height: AppSpacing.sm),
              Text(
                detail,
                textAlign: TextAlign.center,
                style: const TextStyle(color: Colors.white54, fontSize: 12),
              ),
              const SizedBox(height: AppSpacing.lg),
              FilledButton(
                onPressed: () =>
                    unawaited(launchUrl(Uri.parse('app-settings:'))),
                child: const Text('去设置开启权限'),
              ),
              const SizedBox(height: AppSpacing.sm),
              TextButton(
                onPressed: _pickFromGallery,
                style: TextButton.styleFrom(foregroundColor: Colors.white70),
                child: const Text('或从相册选取二维码'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
